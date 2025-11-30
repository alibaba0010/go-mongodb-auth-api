package migrations

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// M001_CreateUsersCollection creates the users collection with indexes
type M001_CreateUsersCollection struct{}

func (m *M001_CreateUsersCollection) Name() string {
	return "001_create_users_collection"
}

func (m *M001_CreateUsersCollection) Up(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("users")

	// Create unique index on email
	emailIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	// Create index on created_at
	createdAtIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "created_at", Value: -1}},
	}

	_, err := collection.Indexes().CreateMany(ctx, []mongo.IndexModel{emailIndex, createdAtIndex})
	return err
}

func (m *M001_CreateUsersCollection) Down(ctx context.Context, db *mongo.Database) error {
	return db.Collection("users").Drop(ctx)
}
