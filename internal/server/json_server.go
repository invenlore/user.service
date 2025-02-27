package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/invenlore/user.service/domain"
	"github.com/invenlore/user.service/internal/service"
	"github.com/sirupsen/logrus"
)

type UserResponse = domain.User

type APIFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

type JSONAPIServer struct {
	listenAddr string
	svc        service.UserService
}

func MakeAPIFunc(fn APIFunc) http.HandlerFunc {
	ctx := context.Background()

	return func(w http.ResponseWriter, r *http.Request) {
		ctx = context.WithValue(ctx, "requestID", uuid.New().String())

		if err := fn(ctx, w, r); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		}
	}
}

func NewJSONAPIServer(listenAddr string, svc service.UserService) *JSONAPIServer {
	return &JSONAPIServer{
		svc:        svc,
		listenAddr: listenAddr,
	}
}

func (s *JSONAPIServer) Run() {
	logrus.Info("Starting JSON server on ", s.listenAddr)

	http.HandleFunc("/api/v1/user", MakeAPIFunc(s.HandleGetUser))

	http.ListenAndServe(s.listenAddr, nil)
}

func (s *JSONAPIServer) HandleGetUser(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id := r.URL.Query().Get("id")
	if len(id) == 0 {
		return fmt.Errorf("invalid id")
	}

	email, err := s.svc.GetUser(ctx, id)
	if err != nil {
		return err
	}

	resp := UserResponse{
		Id:    id,
		Email: email,
	}

	return writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, s int, v any) error {
	w.WriteHeader(s)

	return json.NewEncoder(w).Encode(v)
}
