package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/ticket-system/user-service/model"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user repository operations
type UserRepository interface {
	Create(user *model.User) error
	FindByID(id uuid.UUID) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	Update(user *model.User) error
	Delete(id uuid.UUID) error
}

// userRepository implements UserRepository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) UserRepository {
	// Auto migrate the user model
	if err := db.AutoMigrate(&model.User{}); err != nil {
		logrus.WithError(err).Fatal("Failed to migrate user model")
	}

	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(user *model.User) error {
	result := r.db.Create(user)
	if result.Error != nil {
		logrus.WithError(result.Error).Error("Failed to create user")
		return result.Error
	}
	return nil
}

// FindByID finds a user by ID
func (r *userRepository) FindByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	result := r.db.First(&user, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logrus.WithError(result.Error).Error("Failed to find user by ID")
		return nil, result.Error
	}
	return &user, nil
}

// FindByEmail finds a user by email
func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	result := r.db.First(&user, "email = ?", email)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logrus.WithError(result.Error).Error("Failed to find user by email")
		return nil, result.Error
	}
	return &user, nil
}

// Update updates a user
func (r *userRepository) Update(user *model.User) error {
	result := r.db.Save(user)
	if result.Error != nil {
		logrus.WithError(result.Error).Error("Failed to update user")
		return result.Error
	}
	return nil
}

// Delete deletes a user
func (r *userRepository) Delete(id uuid.UUID) error {
	result := r.db.Delete(&model.User{}, "id = ?", id)
	if result.Error != nil {
		logrus.WithError(result.Error).Error("Failed to delete user")
		return result.Error
	}
	return nil
}