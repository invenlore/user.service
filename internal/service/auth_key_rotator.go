package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"time"

	"github.com/google/uuid"
	"github.com/invenlore/core/pkg/migrator"
	"github.com/invenlore/identity.service/internal/domain"
	"github.com/invenlore/identity.service/internal/repository"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthKeyRotator struct {
	repo             repository.IdentityAuthRepository
	locker           *migrator.Locker
	rotationInterval time.Duration
	retireAfter      time.Duration
	logger           *logrus.Entry
}

type AuthKeyRotatorConfig struct {
	LockKey          string
	LeaseFor         time.Duration
	RotationInterval time.Duration
	RetireAfter      time.Duration
}

func NewAuthKeyRotator(db *mongo.Database, repo repository.IdentityAuthRepository, owner string, cfg AuthKeyRotatorConfig, logger *logrus.Entry) *AuthKeyRotator {
	if cfg.LockKey == "" {
		cfg.LockKey = "identity:auth-key-rotation"
	}

	if cfg.LeaseFor <= 0 {
		cfg.LeaseFor = 30 * time.Second
	}

	if cfg.RotationInterval <= 0 {
		cfg.RotationInterval = 168 * time.Hour
	}

	if cfg.RetireAfter <= 0 {
		cfg.RetireAfter = time.Hour
	}

	if logger == nil {
		logger = logrus.WithField("scope", "auth-key-rotation")
	}

	return &AuthKeyRotator{
		repo:             repo,
		locker:           migrator.NewLocker(db, cfg.LockKey, owner, cfg.LeaseFor),
		rotationInterval: cfg.RotationInterval,
		retireAfter:      cfg.RetireAfter,
		logger:           logger,
	}
}

func (r *AuthKeyRotator) Tick(ctx context.Context) {
	acquired, err := r.locker.TryAcquire(ctx)
	if err != nil {
		r.logger.WithError(err).Warn("auth key rotation: lock acquire failed")
		return
	}

	if !acquired {
		return
	}

	if err := r.ensureActiveKey(ctx); err != nil {
		r.logger.WithError(err).Error("auth key rotation: ensure active key failed")
		return
	}

	if err := r.rotateIfNeeded(ctx); err != nil {
		r.logger.WithError(err).Error("auth key rotation: rotate failed")
		return
	}

	if err := r.revokeRetired(ctx); err != nil {
		r.logger.WithError(err).Error("auth key rotation: revoke retiring failed")
	}
}

func (r *AuthKeyRotator) ensureActiveKey(ctx context.Context) error {
	_, err := r.repo.FindActiveAuthKey(ctx)
	if err == nil {
		return nil
	}

	if err != mongo.ErrNoDocuments {
		return err
	}

	return EnsureActiveKey(ctx, r.repo)
}

func (r *AuthKeyRotator) rotateIfNeeded(ctx context.Context) error {
	key, err := r.repo.FindActiveAuthKey(ctx)
	if err != nil {
		return err
	}

	if time.Since(key.CreatedAt) < r.rotationInterval {
		return nil
	}

	newKey, err := buildAuthKey()
	if err != nil {
		return err
	}

	if err := r.repo.InsertAuthKey(ctx, newKey); err != nil {
		return err
	}

	now := time.Now().UTC()
	if err := r.repo.UpdateAuthKeyStatus(ctx, key.Id, domain.AuthKeyStatusRetiring, &now); err != nil {
		return err
	}

	r.logger.WithFields(logrus.Fields{
		"new_kid": newKey.Kid,
		"old_kid": key.Kid,
	}).Info("auth key rotation: rotated")

	return nil
}

func (r *AuthKeyRotator) revokeRetired(ctx context.Context) error {
	cutoff := time.Now().UTC().Add(-r.retireAfter)

	count, err := r.repo.RevokeRetiringBefore(ctx, cutoff)
	if err != nil {
		return err
	}

	if count > 0 {
		r.logger.WithField("count", count).Info("auth key rotation: revoked retiring keys")
	}

	return nil
}

func buildAuthKey() (*domain.AuthKey, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	privPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, err
	}

	pubPem := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	return &domain.AuthKey{
		Kid:           uuid.NewString(),
		Alg:           "RS256",
		Use:           "sig",
		PrivateKeyPEM: string(privPem),
		PublicKeyPEM:  string(pubPem),
		Status:        domain.AuthKeyStatusActive,
		CreatedAt:     time.Now().UTC(),
	}, nil
}
