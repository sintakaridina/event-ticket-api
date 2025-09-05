package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/yourusername/ticket-system/user-service/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to test database")
	}

	// Auto migrate the user model
	err = db.AutoMigrate(&model.User{})
	if err != nil {
		panic("failed to migrate test database")
	}

	return db
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB()
	repo := NewUserRepository(db)

	t.Run("successful creation", func(t *testing.T) {
		user := &model.User{
			Email:     "test@example.com",
			Password:  "hashedpassword",
			FirstName: "John",
			LastName:  "Doe",
			Phone:     "+1234567890",
			Role:      "user",
			Active:    true,
		}

		err := repo.Create(user)

		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, user.ID)
	})

	t.Run("duplicate email", func(t *testing.T) {
		// Create first user
		user1 := &model.User{
			Email:     "duplicate@example.com",
			Password:  "hashedpassword",
			FirstName: "John",
			LastName:  "Doe",
			Phone:     "+1234567890",
			Role:      "user",
			Active:    true,
		}

		err := repo.Create(user1)
		assert.NoError(t, err)

		// Try to create second user with same email
		user2 := &model.User{
			Email:     "duplicate@example.com",
			Password:  "hashedpassword2",
			FirstName: "Jane",
			LastName:  "Smith",
			Phone:     "+9876543210",
			Role:      "user",
			Active:    true,
		}

		err = repo.Create(user2)
		assert.Error(t, err)
	})
}

func TestUserRepository_FindByID(t *testing.T) {
	db := setupTestDB()
	repo := NewUserRepository(db)

	t.Run("user found", func(t *testing.T) {
		// Create a user first
		user := &model.User{
			Email:     "findbyid@example.com",
			Password:  "hashedpassword",
			FirstName: "John",
			LastName:  "Doe",
			Phone:     "+1234567890",
			Role:      "user",
			Active:    true,
		}

		err := repo.Create(user)
		assert.NoError(t, err)

		// Find the user by ID
		foundUser, err := repo.FindByID(user.ID)

		assert.NoError(t, err)
		assert.NotNil(t, foundUser)
		assert.Equal(t, user.ID, foundUser.ID)
		assert.Equal(t, user.Email, foundUser.Email)
	})

	t.Run("user not found", func(t *testing.T) {
		nonExistentID := uuid.New()

		foundUser, err := repo.FindByID(nonExistentID)

		assert.Error(t, err)
		assert.Nil(t, foundUser)
	})
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db := setupTestDB()
	repo := NewUserRepository(db)

	t.Run("user found", func(t *testing.T) {
		// Create a user first
		user := &model.User{
			Email:     "findbyemail@example.com",
			Password:  "hashedpassword",
			FirstName: "John",
			LastName:  "Doe",
			Phone:     "+1234567890",
			Role:      "user",
			Active:    true,
		}

		err := repo.Create(user)
		assert.NoError(t, err)

		// Find the user by email
		foundUser, err := repo.FindByEmail(user.Email)

		assert.NoError(t, err)
		assert.NotNil(t, foundUser)
		assert.Equal(t, user.Email, foundUser.Email)
		assert.Equal(t, user.FirstName, foundUser.FirstName)
	})

	t.Run("user not found", func(t *testing.T) {
		foundUser, err := repo.FindByEmail("nonexistent@example.com")

		assert.Error(t, err)
		assert.Nil(t, foundUser)
	})
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB()
	repo := NewUserRepository(db)

	t.Run("successful update", func(t *testing.T) {
		// Create a user first
		user := &model.User{
			Email:     "update@example.com",
			Password:  "hashedpassword",
			FirstName: "John",
			LastName:  "Doe",
			Phone:     "+1234567890",
			Role:      "user",
			Active:    true,
		}

		err := repo.Create(user)
		assert.NoError(t, err)

		// Update the user
		user.FirstName = "Jane"
		user.LastName = "Smith"
		user.Phone = "+9876543210"

		err = repo.Update(user)
		assert.NoError(t, err)

		// Verify the update
		updatedUser, err := repo.FindByID(user.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Jane", updatedUser.FirstName)
		assert.Equal(t, "Smith", updatedUser.LastName)
		assert.Equal(t, "+9876543210", updatedUser.Phone)
	})
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupTestDB()
	repo := NewUserRepository(db)

	t.Run("successful deletion", func(t *testing.T) {
		// Create a user first
		user := &model.User{
			Email:     "delete@example.com",
			Password:  "hashedpassword",
			FirstName: "John",
			LastName:  "Doe",
			Phone:     "+1234567890",
			Role:      "user",
			Active:    true,
		}

		err := repo.Create(user)
		assert.NoError(t, err)

		// Delete the user
		err = repo.Delete(user.ID)
		assert.NoError(t, err)

		// Verify the user is deleted
		deletedUser, err := repo.FindByID(user.ID)
		assert.Error(t, err)
		assert.Nil(t, deletedUser)
	})

	t.Run("delete non-existent user", func(t *testing.T) {
		nonExistentID := uuid.New()

		err := repo.Delete(nonExistentID)
		// This should not return an error in GORM (it just affects 0 rows)
		assert.NoError(t, err)
	})
}