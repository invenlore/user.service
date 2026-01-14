package migrations

import (
	"context"

	"github.com/invenlore/core/pkg/migrator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func List() []migrator.Migration {
	return []migrator.Migration{
		{
			Version: 1,
			Name:    "users: unique email index",
			Up: func(ctx context.Context, db *mongo.Database) error {
				col := db.Collection("users")

				_, err := col.Indexes().CreateOne(ctx, mongo.IndexModel{
					Keys: bson.D{
						{Key: "email", Value: 1},
					},
					Options: options.Index().
						SetUnique(true).
						SetName("uniq_email"),
				})

				return err
			},
		},
	}
}
