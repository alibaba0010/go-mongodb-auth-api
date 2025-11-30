package migrations

import (
	"gin-mongo-aws/internal/migration"
)

// GetAll returns all migrations in order
func GetAll() []migration.Migration {
	return []migration.Migration{
		&M001_CreateUsersCollection{},
		&M002_AddUserFields{},
	}
}
