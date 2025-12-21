package repository

import (
	"context"
	"fmt"

	"github.com/invenlore/core/pkg/config"
	"github.com/invenlore/user.service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type userRepository struct {
	col *mongo.Collection
	cfg *config.MongoConfig
}

type UserRepository interface {
	Insert(context.Context, *domain.User) (primitive.ObjectID, error)
	FindOne(context.Context, primitive.ObjectID) (*domain.User, error)
	DeleteOne(context.Context, primitive.ObjectID) (int64, error)
	StreamAll(ctx context.Context, fn func(*domain.User) error) error
}

func NewUserRepository(db *mongo.Client, cfg *config.MongoConfig) UserRepository {
	return &userRepository{
		col: db.Database(cfg.DatabaseName).Collection("users"),
		cfg: cfg,
	}
}

func (r *userRepository) Insert(ctx context.Context, user *domain.User) (primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.OperationTimeout)
	defer cancel()

	result, err := r.col.InsertOne(ctx, user)
	if err != nil {
		return primitive.ObjectID{}, err
	}

	objID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return primitive.ObjectID{}, fmt.Errorf("unexpected inserted id type: %T", result.InsertedID)
	}

	return objID, nil
}

func (r *userRepository) FindOne(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.OperationTimeout)
	defer cancel()

	var user domain.User
	filter := bson.M{"_id": id}

	if err := r.col.FindOne(ctx, filter).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) DeleteOne(ctx context.Context, id primitive.ObjectID) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.OperationTimeout)
	defer cancel()

	filter := bson.M{"_id": id}

	result, err := r.col.DeleteOne(ctx, filter)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

func (r *userRepository) StreamAll(ctx context.Context, fn func(*domain.User) error) error {
	cur, err := r.col.Find(ctx, bson.D{})
	if err != nil {
		return err
	}

	defer func() { _ = cur.Close(context.Background()) }()

	for cur.Next(ctx) {
		var u domain.User
		if err := cur.Decode(&u); err != nil {
			return err
		}
		if err := fn(&u); err != nil {
			return err
		}
	}

	if err := cur.Err(); err != nil {
		return err
	}

	return nil
}
