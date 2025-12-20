package repository

import (
	"context"

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
	FindAll(context.Context) ([]*domain.User, error)
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

	return result.InsertedID.(primitive.ObjectID), nil
}

func (r *userRepository) FindOne(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.OperationTimeout)
	defer cancel()

	var user domain.User
	filter := bson.M{"_id": id}

	err := r.col.FindOne(ctx, filter).Decode(&user)
	if err != nil {
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

func (r *userRepository) FindAll(ctx context.Context) ([]*domain.User, error) {
	var users []*domain.User

	ctx, cancel := context.WithTimeout(ctx, r.cfg.OperationTimeout)
	defer cancel()

	cur, err := r.col.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var user domain.User

		if err := cur.Decode(&user); err != nil {
			return nil, err
		}

		users = append(users, &user)
	}

	return users, nil
}
