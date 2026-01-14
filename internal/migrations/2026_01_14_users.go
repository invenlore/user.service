package migrations

import (
	"context"
	"fmt"

	"github.com/invenlore/core/pkg/migrator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Migration_20260114_UsersCollection_1 = migrator.Migration{
		Version: 1,
		Name:    "users: create collection",
		Up: func(ctx context.Context, db *mongo.Database) error {
			names, err := db.ListCollectionNames(ctx, bson.D{{Key: "name", Value: "users"}})
			if err != nil {
				return err
			}

			if len(names) > 0 {
				return nil
			}

			return db.CreateCollection(ctx, "users")
		},
	}

	Migration_20260114_UsersUniqueEmailIndex_1 = migrator.Migration{
		Version: 2,
		Name:    "users: unique email index",
		Up: func(ctx context.Context, db *mongo.Database) error {
			col := db.Collection("users")

			indexes, err := migrator.ListMongoIndexes(ctx, col)
			if err != nil {
				return err
			}

			state, existingName, err := migrator.MongoSingleFieldIndexState(indexes, "email", 1, "uniq_email")
			if err != nil {
				return fmt.Errorf("MongoDB migrations: %w", err)
			}

			switch state {
			case migrator.MongoIndexUnique:
				return nil
			case migrator.MongoIndexNonUnique:
				return fmt.Errorf(
					"MongoDB migrations: email index already exists but is not unique (name=%q), make it unique or drop it before running migrations",
					existingName,
				)
			case migrator.MongoIndexAbsent:
				// continue
			}

			_, err = col.Indexes().CreateOne(ctx, mongo.IndexModel{
				Keys: bson.D{{Key: "email", Value: 1}},
				Options: options.Index().
					SetUnique(true).
					SetName("uniq_email"),
			})
			if err == nil {
				return nil
			}

			// duplicate key (E11000)
			if migrator.IsMongoDuplicateKeyError(err) {
				diag, derr := migrator.DiagnoseMongoUniqueStringField(ctx, col, "email", 5)

				if derr == nil {
					if len(diag.DuplicateStrings) > 0 {
						return fmt.Errorf(
							"MongoDB migrations: cannot create unique email index; found duplicate emails (sample up to 5): %v: %w",
							diag.DuplicateStrings, err,
						)
					}

					if diag.NullCount > 1 {
						return fmt.Errorf(
							"MongoDB migrations: cannot create unique email index; found %d documents with email=null (unique index would fail): %w",
							diag.NullCount, err,
						)
					}
				}
			}

			return err
		},
	}
)
