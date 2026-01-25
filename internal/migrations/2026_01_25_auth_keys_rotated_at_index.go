package migrations

import (
	"context"

	"github.com/invenlore/core/pkg/migrator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Migration_20260125_AuthKeysRotatedAtIndex_1 = migrator.Migration{
		Version: 8,
		Name:    "auth_keys: rotated_at index",
		Up: func(ctx context.Context, db *mongo.Database) error {
			col := db.Collection("auth_keys")
			model := mongo.IndexModel{
				Keys:    bson.D{{Key: "rotated_at", Value: 1}},
				Options: options.Index().SetName("idx_rotated_at"),
			}

			_, err := col.Indexes().CreateOne(ctx, model)
			return err
		},
	}
)
