package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/alexPavlikov/sso/internal/app/grpc"
	"github.com/alexPavlikov/sso/internal/services/auth"
	"github.com/alexPavlikov/sso/internal/storage/sqlite"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, storagePath string, tokenTTL time.Duration) *App {
	// TODO: инициализировать хранилиище (storage)

	storage, _ := sqlite.New(storagePath)

	authSerive := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authSerive, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}
