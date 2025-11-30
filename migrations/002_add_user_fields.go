package migrations

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// M002_AddUserFields adds additional fields to the users collection
type M002_AddUserFields struct{}

func (m *M002_AddUserFields) Name() string {
	return "002_add_user_fields"
}

func (m *M002_AddUserFields) Up(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("users")

	// Update all documents to add missing fields with default values
	_, err := collection.UpdateMany(ctx, bson.M{}, bson.M{
		"$set": bson.M{
			"is_active":   true,
			"last_login":  nil,
			"phone_number": "",
		},
	})

	return err
}

func (m *M002_AddUserFields) Down(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("users")

	// Remove the added fields
	_, err := collection.UpdateMany(ctx, bson.M{}, bson.M{
		"$unset": bson.M{
			"is_active":   "",
			"last_login":  "",
			"phone_number": "",
		},
	})

	return err
}
