package migration

import (
	"context"
	"fmt"
	"time"

	"gin-mongo-aws/internal/logger"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// Migration defines a single migration
type Migration interface {
	Name() string
	Up(ctx context.Context, db *mongo.Database) error
	Down(ctx context.Context, db *mongo.Database) error
}

// MigrationRecord stores metadata about applied migrations
type MigrationRecord struct {
	Name      string    `bson:"name"`
	AppliedAt time.Time `bson:"applied_at"`
}

// Manager handles running migrations
type Manager struct {
	db         *mongo.Database
	collection *mongo.Collection
}

// NewManager creates a new migration manager
func NewManager(db *mongo.Database) *Manager {
	return &Manager{
		db:         db,
		collection: db.Collection("_migrations"),
	}
}

// Initialize sets up the migrations collection if it doesn't exist
func (m *Manager) Initialize(ctx context.Context) error {
	// The collection will be created automatically on first write
	// Create a unique index on the name field
	indexModel := mongo.IndexModel{
		Keys: bson.M{"name": 1},
		Options: options.Index().SetUnique(true),
	}

	_, err := m.collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		logger.Log.Error("Failed to create migration index", zap.Error(err))
		return err
	}

	logger.Log.Info("Migration manager initialized")
	return nil
}

// Run executes all pending migrations
func (m *Manager) Run(ctx context.Context, migrations []Migration) error {
	logger.Log.Info("Starting migration run", zap.Int("total_migrations", len(migrations)))

	for _, migration := range migrations {
		if applied, err := m.isApplied(ctx, migration.Name()); err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", migration.Name(), err)
		} else if applied {
			logger.Log.Info("Migration already applied", zap.String("name", migration.Name()))
			continue
		}

		logger.Log.Info("Running migration", zap.String("name", migration.Name()))
		if err := migration.Up(ctx, m.db); err != nil {
			logger.Log.Error("Migration failed", zap.String("name", migration.Name()), zap.Error(err))
			return fmt.Errorf("migration %s failed: %w", migration.Name(), err)
		}

		if err := m.recordMigration(ctx, migration.Name()); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migration.Name(), err)
		}

		logger.Log.Info("Migration completed", zap.String("name", migration.Name()))
	}

	logger.Log.Info("All migrations completed successfully")
	return nil
}

// Rollback reverts the last applied migration
func (m *Manager) Rollback(ctx context.Context, migrations []Migration) error {
	// Get the last applied migration
	var lastRecord MigrationRecord
	opts := options.FindOne().SetSort(bson.M{"applied_at": -1})
	err := m.collection.FindOne(ctx, bson.M{}, opts).Decode(&lastRecord)
	if err == mongo.ErrNoDocuments {
		logger.Log.Info("No migrations to rollback")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to find last migration: %w", err)
	}

	// Find the migration to rollback
	var migration Migration
	for _, m := range migrations {
		if m.Name() == lastRecord.Name {
			migration = m
			break
		}
	}

	if migration == nil {
		return fmt.Errorf("migration %s not found in migration list", lastRecord.Name)
	}

	logger.Log.Info("Rolling back migration", zap.String("name", migration.Name()))
	if err := migration.Down(ctx, m.db); err != nil {
		logger.Log.Error("Rollback failed", zap.String("name", migration.Name()), zap.Error(err))
		return fmt.Errorf("rollback of %s failed: %w", migration.Name(), err)
	}

	if err := m.removeMigration(ctx, migration.Name()); err != nil {
		return fmt.Errorf("failed to remove migration record %s: %w", migration.Name(), err)
	}

	logger.Log.Info("Rollback completed", zap.String("name", migration.Name()))
	return nil
}

// GetStatus returns the list of applied migrations
func (m *Manager) GetStatus(ctx context.Context) ([]MigrationRecord, error) {
	cursor, err := m.collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"applied_at": 1}))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch migration status: %w", err)
	}
	defer cursor.Close(ctx)

	var records []MigrationRecord
	if err := cursor.All(ctx, &records); err != nil {
		return nil, fmt.Errorf("failed to decode migration records: %w", err)
	}

	return records, nil
}

// isApplied checks if a migration has been applied
func (m *Manager) isApplied(ctx context.Context, name string) (bool, error) {
	count, err := m.collection.CountDocuments(ctx, bson.M{"name": name})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// recordMigration records that a migration has been applied
func (m *Manager) recordMigration(ctx context.Context, name string) error {
	record := MigrationRecord{
		Name:      name,
		AppliedAt: time.Now(),
	}
	_, err := m.collection.InsertOne(ctx, record)
	return err
}

// removeMigration removes a migration record
func (m *Manager) removeMigration(ctx context.Context, name string) error {
	_, err := m.collection.DeleteOne(ctx, bson.M{"name": name})
	return err
}
