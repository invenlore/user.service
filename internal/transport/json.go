package transport

import (
	"net/http"
	"time"

	"github.com/alexliesenfeld/health"
	"github.com/sirupsen/logrus"
)

type JSONAPIServer struct {
	listenAddr string
}

func NewJSONAPIServer(listenAddr string) *JSONAPIServer {
	return &JSONAPIServer{
		listenAddr: listenAddr,
	}
}

func (s *JSONAPIServer) Run() {
	logrus.Info("Starting JSONAPI server on ", s.listenAddr)

	checker := health.NewChecker(
		health.WithCacheDuration(1*time.Second),
		health.WithTimeout(10*time.Second),
	)

	http.Handle("/health", health.NewHandler(checker))
	logrus.Fatalln(http.ListenAndServe(s.listenAddr, nil))
}
