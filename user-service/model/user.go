package model

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Email     string    `gorm:"type:varchar(100);unique;not null" json:"email"`
	Password  string    `gorm:"type:varchar(100);not null" json:"-"`
	FirstName string    `gorm:"type:varchar(50)" json:"first_name"`
	LastName  string    `gorm:"type:varchar(50)" json:"last_name"`
	Phone     string    `gorm:"type:varchar(20)" json:"phone"`
	Role      string    `gorm:"type:varchar(20);default:'user'" json:"role"`
	Active    bool      `gorm:"default:true" json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate is a GORM hook that runs before creating a new user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// Generate a new UUID if not set
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	// Hash the password if not already hashed
	if len(u.Password) > 0 && len(u.Password) < 60 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
	}

	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a user
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// Hash the password if it has changed and is not already hashed
	if len(u.Password) > 0 && len(u.Password) < 60 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
	}

	return nil
}

// CheckPassword checks if the provided password matches the stored hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// UserResponse is the response structure for user data
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse converts a User to a UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Phone:     u.Phone,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}

// RegisterRequest represents the request structure for user registration
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone" binding:"required"`
}

// LoginRequest represents the request structure for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UpdateProfileRequest represents the request structure for updating a user profile
type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

// ChangePasswordRequest represents the request structure for changing a user's password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// LoginResponse represents the response structure for user login
type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}