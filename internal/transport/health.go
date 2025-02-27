package transport

import (
	"net/http"

	"github.com/invenlore/user.service/pkg/health"
	"github.com/sirupsen/logrus"
)

type HealthServer struct {
	listenAddr string
}

func NewHealthServer(listenAddr string) *HealthServer {
	return &HealthServer{
		listenAddr: listenAddr,
	}
}

func (s *HealthServer) Run() {
	logrus.Info("starting health server on ", s.listenAddr)

	http.Handle("/health", health.GetHealthHandler())
	logrus.Fatalln(http.ListenAndServe(s.listenAddr, nil))
}
