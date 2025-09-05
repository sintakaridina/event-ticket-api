package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourusername/ticket-system/user-service/model"
)

// MockUserService is a mock implementation of UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(req model.RegisterRequest) (*model.UserResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserResponse), args.Error(1)
}

func (m *MockUserService) Login(req model.LoginRequest) (*model.LoginResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.LoginResponse), args.Error(1)
}

func (m *MockUserService) GetUserByID(id uuid.UUID) (*model.UserResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserResponse), args.Error(1)
}

func (m *MockUserService) UpdateProfile(id uuid.UUID, req model.UpdateProfileRequest) (*model.UserResponse, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserResponse), args.Error(1)
}

func (m *MockUserService) ChangePassword(id uuid.UUID, req model.ChangePasswordRequest) error {
	args := m.Called(id, req)
	return args.Error(0)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestUserHandler_Register(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)
	router := setupTestRouter()

	t.Run("successful registration", func(t *testing.T) {
		req := model.RegisterRequest{
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "John",
			LastName:  "Doe",
			Phone:     "+1234567890",
		}

		expectedResponse := &model.UserResponse{
			ID:        uuid.New(),
			Email:     req.Email,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Phone:     req.Phone,
			Role:      "user",
		}

		mockService.On("Register", req).Return(expectedResponse, nil)

		router.POST("/register", handler.Register)

		reqBody, _ := json.Marshal(req)
		request := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(reqBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusCreated, response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		router := setupTestRouter()
		router.POST("/register", handler.Register)

		request := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer([]byte("invalid json")))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusBadRequest, response.Code)
	})

	t.Run("service error", func(t *testing.T) {
		req := model.RegisterRequest{
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "John",
			LastName:  "Doe",
			Phone:     "+1234567890",
		}

		mockService := new(MockUserService)
		mockService.On("Register", req).Return(nil, errors.New("service error"))
		handler := NewUserHandler(mockService)

		router := setupTestRouter()
		router.POST("/register", handler.Register)

		reqBody, _ := json.Marshal(req)
		request := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(reqBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusInternalServerError, response.Code)
		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_Login(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)
	router := setupTestRouter()

	t.Run("successful login", func(t *testing.T) {
		req := model.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		expectedResponse := &model.LoginResponse{
			Token: "jwt-token",
			User: model.UserResponse{
				ID:        uuid.New(),
				Email:     req.Email,
				FirstName: "John",
				LastName:  "Doe",
				Role:      "user",
			},
		}

		mockService.On("Login", req).Return(expectedResponse, nil)

		router.POST("/login", handler.Login)

		reqBody, _ := json.Marshal(req)
		request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusOK, response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		req := model.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		mockService := new(MockUserService)
		mockService.On("Login", req).Return(nil, errors.New("invalid credentials"))
		handler := NewUserHandler(mockService)

		router := setupTestRouter()
		router.POST("/login", handler.Login)

		reqBody, _ := json.Marshal(req)
		request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusUnauthorized, response.Code)
		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_GetProfile(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)
	router := setupTestRouter()

	t.Run("successful get profile", func(t *testing.T) {
		userID := uuid.New()
		expectedResponse := &model.UserResponse{
			ID:        userID,
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
			Phone:     "+1234567890",
			Role:      "user",
		}

		mockService.On("GetUserByID", userID).Return(expectedResponse, nil)

		// Mock middleware to set user_id in context
		router.Use(func(c *gin.Context) {
			c.Set("user_id", userID.String())
			c.Next()
		})
		router.GET("/profile", handler.GetProfile)

		request := httptest.NewRequest(http.MethodGet, "/profile", nil)
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusOK, response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		userID := uuid.New()

		mockService := new(MockUserService)
		mockService.On("GetUserByID", userID).Return(nil, errors.New("user not found"))
		handler := NewUserHandler(mockService)

		router := setupTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", userID.String())
			c.Next()
		})
		router.GET("/profile", handler.GetProfile)

		request := httptest.NewRequest(http.MethodGet, "/profile", nil)
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusNotFound, response.Code)
		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_UpdateProfile(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)
	router := setupTestRouter()

	t.Run("successful update", func(t *testing.T) {
		userID := uuid.New()
		req := model.UpdateProfileRequest{
			FirstName: "Jane",
			LastName:  "Smith",
			Phone:     "+9876543210",
		}

		expectedResponse := &model.UserResponse{
			ID:        userID,
			Email:     "test@example.com",
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Phone:     req.Phone,
			Role:      "user",
		}

		mockService.On("UpdateProfile", userID, req).Return(expectedResponse, nil)

		// Mock middleware to set user_id in context
		router.Use(func(c *gin.Context) {
			c.Set("user_id", userID.String())
			c.Next()
		})
		router.PUT("/profile", handler.UpdateProfile)

		reqBody, _ := json.Marshal(req)
		request := httptest.NewRequest(http.MethodPut, "/profile", bytes.NewBuffer(reqBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusOK, response.Code)
		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_ChangePassword(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)
	router := setupTestRouter()

	t.Run("successful password change", func(t *testing.T) {
		userID := uuid.New()
		req := model.ChangePasswordRequest{
			CurrentPassword: "oldpassword",
			NewPassword:     "newpassword123",
		}

		mockService.On("ChangePassword", userID, req).Return(nil)

		// Mock middleware to set user_id in context
		router.Use(func(c *gin.Context) {
			c.Set("user_id", userID.String())
			c.Next()
		})
		router.PUT("/change-password", handler.ChangePassword)

		reqBody, _ := json.Marshal(req)
		request := httptest.NewRequest(http.MethodPut, "/change-password", bytes.NewBuffer(reqBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusOK, response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid current password", func(t *testing.T) {
		userID := uuid.New()
		req := model.ChangePasswordRequest{
			CurrentPassword: "wrongpassword",
			NewPassword:     "newpassword123",
		}

		mockService := new(MockUserService)
		mockService.On("ChangePassword", userID, req).Return(errors.New("invalid current password"))
		handler := NewUserHandler(mockService)

		router := setupTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", userID.String())
			c.Next()
		})
		router.PUT("/change-password", handler.ChangePassword)

		reqBody, _ := json.Marshal(req)
		request := httptest.NewRequest(http.MethodPut, "/change-password", bytes.NewBuffer(reqBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusBadRequest, response.Code)
		mockService.AssertExpectations(t)
	})
}