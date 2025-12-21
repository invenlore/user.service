package transport

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/invenlore/core/pkg/config"
	"github.com/invenlore/core/pkg/health"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func NewHealthServer(
	cfg *config.AppConfig,
	mongoClient *mongo.Client,
) (*http.Server, net.Listener, error) {
	var (
		loggerEntry = logrus.WithField("scope", "health")
		listenAddr  = net.JoinHostPort(cfg.Health.Host, cfg.Health.Port)
	)

	loggerEntry.Info("starting health server...")

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen on %s: %w", listenAddr, err)
	}

	base := health.GetHealthHandler()

	mux := http.NewServeMux()
	mux.Handle("GET /health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pingCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := mongoClient.Ping(pingCtx, readpref.Primary()); err != nil {
			loggerEntry.Errorf("MongoDB ping failed: %v", err)
			http.Error(w, "MongoDB unavailable", http.StatusServiceUnavailable)

			return
		}

		loggerEntry.Debug("healthcheck: MongoDB available")
		base.ServeHTTP(w, r)
	}))

	server := &http.Server{
		Addr:              listenAddr,
		Handler:           mux,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
	}

	return server, ln, nil
}
