package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/invenlore/core/pkg/config"
	"github.com/invenlore/core/pkg/db"
	"github.com/invenlore/user.service/internal/repository"
	"github.com/invenlore/user.service/internal/service"
	"github.com/invenlore/user.service/internal/transport"
	"github.com/sirupsen/logrus"
)

func Start() {
	loggerEntry := logrus.WithField("scope", "service")
	loggerEntry.Info("service starting...")

	cfg, err := config.Config()
	if err != nil {
		loggerEntry.Fatalf("failed to load configuration: %v", err)
	}

	appCfg := cfg.GetConfig()
	mongoCfg := appCfg.GetMongoConfig()

	baseCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	g, ctx := errgroup.WithContext(baseCtx)

	mongoClient, err := db.MongoDBConnect(ctx, mongoCfg)
	if err != nil {
		loggerEntry.Fatalf("MongoDB connect failed: %v", err)
	}

	loggerEntry.Info("MongoDB connected successfully")

	mongoReadiness := db.NewMongoReadiness(mongoClient, mongoCfg.HealthCheckTimeout)

	g.Go(func() error {
		mongoReadiness.Run(ctx, mongoCfg.HealthCheckInterval)

		return nil
	})

	g.Go(func() error {
		<-ctx.Done()

		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return mongoClient.Disconnect(stopCtx)
	})

	repo := repository.NewUserRepository(mongoClient, mongoCfg)
	svc := service.NewUserService(repo)

	grpcSrv, grpcLn, err := transport.NewGRPCServer(appCfg.GetGRPCConfig(), svc, mongoReadiness)
	if err != nil {
		loggerEntry.Fatalf("gRPC server init failed: %v", err)
	}

	healthSrv, healthLn, err := transport.NewHealthServer(appCfg.GetHealthConfig())
	if err != nil {
		_ = grpcLn.Close()

		loggerEntry.Fatalf("health server init failed: %v", err)
	}

	g.Go(func() error {
		loggerEntry.Infof("gRPC server serving on %s...", grpcLn.Addr().String())

		if err := grpcSrv.Serve(grpcLn); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			return fmt.Errorf("gRPC serve failed: %w", err)
		}

		return nil
	})

	g.Go(func() error {
		loggerEntry.Infof("health server serving on %s...", healthSrv.Addr)

		if err := healthSrv.Serve(healthLn); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("health serve failed: %w", err)
		}

		return nil
	})

	g.Go(func() error {
		<-ctx.Done()

		loggerEntry.Trace("attempting graceful shutdown...")

		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		grpcSrv.GracefulStop()
		_ = healthSrv.Shutdown(stopCtx)

		loggerEntry.Info("clean service shutdown complete")
		return nil
	})

	if err := g.Wait(); err != nil {
		loggerEntry.Errorf("service stopped with error: %v", err)

		os.Exit(1)
	}

	loggerEntry.Info("service stopped gracefully")
}
