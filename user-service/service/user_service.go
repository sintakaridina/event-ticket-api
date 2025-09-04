package service

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/ticket-system/user-service/config"
	"github.com/yourusername/ticket-system/user-service/middleware"
	"github.com/yourusername/ticket-system/user-service/model"
	"github.com/yourusername/ticket-system/user-service/repository"
)

// UserService defines the interface for user service operations
type UserService interface {
	Register(req model.RegisterRequest) (*model.UserResponse, error)
	Login(req model.LoginRequest) (*model.LoginResponse, error)
	GetUserByID(id uuid.UUID) (*model.UserResponse, error)
	UpdateProfile(id uuid.UUID, req model.UpdateProfileRequest) (*model.UserResponse, error)
	ChangePassword(id uuid.UUID, req model.ChangePasswordRequest) error
}

// userService implements UserService interface
type userService struct {
	userRepo repository.UserRepository
	rmq      *config.RabbitMQ
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository, rmq *config.RabbitMQ) UserService {
	return &userService{
		userRepo: userRepo,
		rmq:      rmq,
	}
}

// Register registers a new user
func (s *userService) Register(req model.RegisterRequest) (*model.UserResponse, error) {
	// Check if user with the same email already exists
	existingUser, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Create new user
	user := &model.User{
		Email:     req.Email,
		Password:  req.Password, // Will be hashed by GORM hook
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Role:      "user", // Default role
		Active:    true,
	}

	// Save user to database
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Publish user created event
	s.publishUserEvent("user.created", user)

	// Return user response
	userResponse := user.ToResponse()
	return &userResponse, nil
}

// Login authenticates a user and returns a JWT token
func (s *userService) Login(req model.LoginRequest) (*model.LoginResponse, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		return nil, errors.New("invalid email or password")
	}

	// Check if user is active
	if !user.Active {
		return nil, errors.New("user account is inactive")
	}

	// Check password
	if !user.CheckPassword(req.Password) {
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT token
	token, err := middleware.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Publish user login event
	s.publishUserEvent("user.login", user)

	// Return login response
	return &model.LoginResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

// GetUserByID gets a user by ID
func (s *userService) GetUserByID(id uuid.UUID) (*model.UserResponse, error) {
	// Find user by ID
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	// Return user response
	userResponse := user.ToResponse()
	return &userResponse, nil
}

// UpdateProfile updates a user's profile
func (s *userService) UpdateProfile(id uuid.UUID, req model.UpdateProfileRequest) (*model.UserResponse, error) {
	// Find user by ID
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	// Update user fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}

	// Save user to database
	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Publish user updated event
	s.publishUserEvent("user.updated", user)

	// Return user response
	userResponse := user.ToResponse()
	return &userResponse, nil
}

// ChangePassword changes a user's password
func (s *userService) ChangePassword(id uuid.UUID, req model.ChangePasswordRequest) error {
	// Find user by ID
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		return errors.New("user not found")
	}

	// Check current password
	if !user.CheckPassword(req.CurrentPassword) {
		return errors.New("current password is incorrect")
	}

	// Update password
	user.Password = req.NewPassword // Will be hashed by GORM hook

	// Save user to database
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Publish password changed event
	s.publishUserEvent("user.password_changed", user)

	return nil
}

// publishUserEvent publishes a user event to RabbitMQ
func (s *userService) publishUserEvent(eventType string, user *model.User) {
	// Create event payload
	event := map[string]interface{}{
		"event_type": eventType,
		"user_id":    user.ID.String(),
		"email":      user.Email,
		"timestamp":  user.UpdatedAt,
	}

	// Convert event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal user event")
		return
	}

	// Publish event to RabbitMQ
	err = s.rmq.PublishMessage("user_events", eventType, eventJSON)
	if err != nil {
		logrus.WithError(err).Error("Failed to publish user event")
	}
}