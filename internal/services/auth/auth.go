package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/alexPavlikov/sso/internal/domain/models"
	"github.com/alexPavlikov/sso/internal/lib/jwt"
	"github.com/alexPavlikov/sso/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log         *slog.Logger
	userSaver   UserSaver
	userProvide UserProvider
	appProvider AppProvider
	tokenTTL    time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppID       = errors.New("invalid app id")
	ErrUserExists         = errors.New("user already exists")
)

func New(log *slog.Logger, userSaver UserSaver, userProvide UserProvider, appProvider AppProvider, tokenTTL time.Duration) *Auth {
	return &Auth{
		log:         log,
		userSaver:   userSaver,
		userProvide: userProvide,
		appProvider: appProvider,
		tokenTTL:    tokenTTL,
	}
}

func (a *Auth) Login(ctx context.Context, email string, password string, appID int) (string, error) {
	const op = "Auth.Login"

	log := a.log.With(slog.String("op", op), slog.String("username", email))

	log.Info("attempting to login user")

	user, err := a.userProvide.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", err)

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", err)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", err)

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		a.log.Info("invalid app_id", err)

		return "", fmt.Errorf("%s:%w", op, err)
	}

	a.log.Info("user logged successfully")

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", err)

		return "", fmt.Errorf("%s:%w", op, err)
	}

	return token, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context, email string, pass string) (int, error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With(slog.String("op", op), slog.String("email", email))

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", err)

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.userSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists", err)

			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}

		log.Error("failed to save user", err)

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered")

	return int(id), nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID int) (bool, error) {
	const op = "Auth.IsAdmin"

	log := a.log.With(slog.String("op", op), slog.String("user_id", fmt.Sprint(userID)))

	log.Info("checking if user is admin")

	isAdmin, err := a.userProvide.IsAdmin(ctx, int64(userID))
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("user not found", err)

			return false, fmt.Errorf("%s: %w", op, err)
		}

		log.Error("failed to checking if user is admin", err)

		return false, fmt.Errorf("%s: %w", op, ErrInvalidAppID)
	}

	log.Info("completed checking if user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}
