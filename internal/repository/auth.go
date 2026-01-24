package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/invenlore/core/pkg/config"
	"github.com/invenlore/identity.service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IdentityAuthRepository interface {
	InsertAuthKey(context.Context, *domain.AuthKey) error
	FindActiveAuthKey(context.Context) (*domain.AuthKey, error)
	ListActivePublicKeys(context.Context) ([]*domain.AuthKey, error)
	InsertUserCredentials(context.Context, *domain.User) (primitive.ObjectID, error)
	FindUserByEmail(context.Context, string) (*domain.User, error)
	FindUserByID(context.Context, primitive.ObjectID) (*domain.User, error)
	InsertRefreshSession(context.Context, *domain.RefreshSession) error
	FindRefreshSession(context.Context, string) (*domain.RefreshSession, error)
	RevokeRefreshSession(context.Context, string, time.Time) error
	RotateRefreshSession(context.Context, string, string, time.Time, time.Time) error
}

type identityAuthRepository struct {
	usersCol    *mongo.Collection
	keysCol     *mongo.Collection
	sessionsCol *mongo.Collection
	cfg         *config.MongoConfig
}

func NewIdentityAuthRepository(db *mongo.Client, cfg *config.MongoConfig) IdentityAuthRepository {
	database := db.Database(cfg.DatabaseName)

	return &identityAuthRepository{
		usersCol:    database.Collection("users"),
		keysCol:     database.Collection("auth_keys"),
		sessionsCol: database.Collection("refresh_sessions"),
		cfg:         cfg,
	}
}

func (r *identityAuthRepository) InsertAuthKey(ctx context.Context, key *domain.AuthKey) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.OperationTimeout)
	defer cancel()

	_, err := r.keysCol.InsertOne(ctx, key)
	return err
}

func (r *identityAuthRepository) FindActiveAuthKey(ctx context.Context) (*domain.AuthKey, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.OperationTimeout)
	defer cancel()

	filter := bson.M{"status": domain.AuthKeyStatusActive}
	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})

	var key domain.AuthKey
	if err := r.keysCol.FindOne(ctx, filter, opts).Decode(&key); err != nil {
		return nil, err
	}

	return &key, nil
}

func (r *identityAuthRepository) ListActivePublicKeys(ctx context.Context) ([]*domain.AuthKey, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.OperationTimeout)
	defer cancel()

	filter := bson.M{"status": bson.M{"$in": []domain.AuthKeyStatus{domain.AuthKeyStatusActive, domain.AuthKeyStatusRetiring}}}

	cur, err := r.keysCol.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	defer func() { _ = cur.Close(context.Background()) }()

	keys := make([]*domain.AuthKey, 0)

	for cur.Next(ctx) {
		var key domain.AuthKey

		if err := cur.Decode(&key); err != nil {
			return nil, err
		}

		keys = append(keys, &key)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return keys, nil
}

func (r *identityAuthRepository) InsertUserCredentials(ctx context.Context, user *domain.User) (primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.OperationTimeout)
	defer cancel()

	result, err := r.usersCol.InsertOne(ctx, user)
	if err != nil {
		return primitive.ObjectID{}, err
	}

	objID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return primitive.ObjectID{}, fmt.Errorf("unexpected inserted id type: %T", result.InsertedID)
	}

	return objID, nil
}

func (r *identityAuthRepository) FindUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.OperationTimeout)
	defer cancel()

	filter := bson.M{"email": email}
	var user domain.User

	if err := r.usersCol.FindOne(ctx, filter).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *identityAuthRepository) FindUserByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.OperationTimeout)
	defer cancel()

	filter := bson.M{"_id": id}
	var user domain.User

	if err := r.usersCol.FindOne(ctx, filter).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *identityAuthRepository) InsertRefreshSession(ctx context.Context, session *domain.RefreshSession) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.OperationTimeout)
	defer cancel()

	_, err := r.sessionsCol.InsertOne(ctx, session)
	return err
}

func (r *identityAuthRepository) FindRefreshSession(ctx context.Context, sessionID string) (*domain.RefreshSession, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.OperationTimeout)
	defer cancel()

	filter := bson.M{"session_id": sessionID}
	var session domain.RefreshSession

	if err := r.sessionsCol.FindOne(ctx, filter).Decode(&session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *identityAuthRepository) RevokeRefreshSession(ctx context.Context, sessionID string, revokedAt time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.OperationTimeout)
	defer cancel()

	filter := bson.M{"session_id": sessionID}
	update := bson.M{"$set": bson.M{"revoked_at": revokedAt}}

	result, err := r.sessionsCol.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *identityAuthRepository) RotateRefreshSession(ctx context.Context, sessionID string, tokenHash string, expiresAt time.Time, updatedAt time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.OperationTimeout)
	defer cancel()

	filter := bson.M{"session_id": sessionID}
	update := bson.M{"$set": bson.M{"refresh_token_hash": tokenHash, "expires_at": expiresAt, "updated_at": updatedAt}}

	result, err := r.sessionsCol.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}
