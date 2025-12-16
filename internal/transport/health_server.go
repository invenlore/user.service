package transport

import (
	"fmt"
	"net"
	"net/http"

	"github.com/invenlore/user.service/pkg/health"
	"github.com/sirupsen/logrus"
)

func StartHealthServer(listenAddr string, errChan chan error) (*http.Server, net.Listener, error) {
	logrus.Info("starting health server on ", listenAddr)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen on %s: %w", listenAddr, err)
	}

	mux := http.NewServeMux()
	mux.Handle("/health", health.GetHealthHandler())

	server := &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}

	go func() {
		logrus.Printf("health server serving on %s", listenAddr)

		if serveErr := server.Serve(ln); serveErr != nil && serveErr != http.ErrServerClosed {
			errChan <- fmt.Errorf("health server failed to serve: %w", serveErr)
		}
	}()

	return server, ln, nil
}
