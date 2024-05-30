package main

import (
	"context"
	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	ssov1 "github.com/synthao/sso/gen/go/sso"
	"github.com/synthao/sso/internal/adapter/postgres/repository"
	"github.com/synthao/sso/internal/config"
	"github.com/synthao/sso/internal/service"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"net"
	"os"
)

func main() {
	fx.New(
		fx.Provide(
			config.NewLoggerConfig,
			config.NewJWTConfig,
			newLogger,
			NewDBConnection,
			repository.NewRepository,
			service.NewSSOService,
		),
		fx.Invoke(createGRPCServer),
	).Run()
}

func createGRPCServer(lc fx.Lifecycle, logger *zap.Logger, ssoService *service.SSOService) {
	server := grpc.NewServer()

	ssov1.RegisterServiceServer(server, ssoService)

	listener, err := net.Listen("tcp", net.JoinHostPort("", os.Getenv("GRPC_PORT")))
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting server ...")

			go server.Serve(listener)

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping server ...")

			server.Stop()

			return nil
		},
	})
}

func NewDBConnection() *sqlx.DB {
	return sqlx.MustConnect("postgres", config.GetDSN())
}

func newLogger(cnf *config.Logger) (*zap.Logger, error) {
	atomicLogLevel, err := zap.ParseAtomicLevel(cnf.Level)
	if err != nil {
		return nil, err
	}

	atom := zap.NewAtomicLevelAt(atomicLogLevel.Level())
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	return zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.Lock(os.Stdout),
			atom,
		),
		zap.WithCaller(true),
		zap.AddStacktrace(zap.ErrorLevel),
	), nil
}
