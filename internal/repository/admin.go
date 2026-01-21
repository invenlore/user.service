package repository

import (
	"context"
	"fmt"

	"github.com/invenlore/core/pkg/config"
	"github.com/invenlore/identity.service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type identityAdminRepository struct {
	usersCol *mongo.Collection
	cfg      *config.MongoConfig
}

type IdentityAdminRepository interface {
	InsertUser(context.Context, *domain.User) (primitive.ObjectID, error)
	FindOneUser(context.Context, primitive.ObjectID) (*domain.User, error)
	DeleteOneUser(context.Context, primitive.ObjectID) (int64, error)
	ListUsers(ctx context.Context) ([]*domain.User, error)
}

func NewIdentityAdminRepository(db *mongo.Client, cfg *config.MongoConfig) IdentityAdminRepository {
	return &identityAdminRepository{
		usersCol: db.Database(cfg.DatabaseName).Collection("users"),
		cfg:      cfg,
	}
}

func (r *identityAdminRepository) InsertUser(ctx context.Context, user *domain.User) (primitive.ObjectID, error) {
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

func (r *identityAdminRepository) FindOneUser(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.OperationTimeout)
	defer cancel()

	var user domain.User
	filter := bson.M{"_id": id}

	if err := r.usersCol.FindOne(ctx, filter).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *identityAdminRepository) DeleteOneUser(ctx context.Context, id primitive.ObjectID) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.OperationTimeout)
	defer cancel()

	filter := bson.M{"_id": id}

	result, err := r.usersCol.DeleteOne(ctx, filter)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

func (r *identityAdminRepository) ListUsers(ctx context.Context) ([]*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.OperationTimeout)
	defer cancel()

	cur, err := r.usersCol.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	defer func() { _ = cur.Close(context.Background()) }()

	users := make([]*domain.User, 0)

	for cur.Next(ctx) {
		var u domain.User

		if err := cur.Decode(&u); err != nil {
			return nil, err
		}

		users = append(users, &u)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
