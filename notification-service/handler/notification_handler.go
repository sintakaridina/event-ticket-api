package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"notification-service/middleware"
	"notification-service/model"
	"notification-service/service"
)

// NotificationHandler handles HTTP requests for notifications
type NotificationHandler struct {
	Service service.NotificationService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(service service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		Service: service,
	}
}

// SetupRoutes sets up the notification routes
func (h *NotificationHandler) SetupRoutes(router *gin.Engine) {
	// API group with JWT authentication
	apiGroup := router.Group("/api/v1")
	apiGroup.Use(middleware.JWTAuth())

	// Notification routes
	notifications := apiGroup.Group("/notifications")
	notifications.POST("", h.CreateNotification)
	notifications.POST("/:id/send", h.SendNotification)
	notifications.GET("/:id", h.GetNotificationByID)
	notifications.GET("/user", h.GetUserNotifications)
	notifications.PUT("/:id/status", h.UpdateNotificationStatus)

	// Template routes
	templates := apiGroup.Group("/templates")
	templates.POST("", h.CreateTemplate)
	templates.GET("/:id", h.GetTemplateByID)
	templates.GET("/code/:code", h.GetTemplateByCode)
	templates.GET("", h.GetAllTemplates)
	templates.PUT("/:id", h.UpdateTemplate)
	templates.DELETE("/:id", h.DeleteTemplate)
}

// CreateNotification creates a new notification
func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	// Parse request
	var req model.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Set user ID if not provided
	if req.UserID == "" {
		req.UserID = userID.(string)
	}

	// Check if user has permission to create notification for another user
	if req.UserID != userID.(string) {
		userRole, exists := c.Get("userRole")
		if !exists || userRole.(string) != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to create notifications for other users"})
			return
		}
	}

	// Create notification
	notification, err := h.Service.CreateNotification(req)
	if err != nil {
		logrus.Errorf("Failed to create notification: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, notification)
}

// SendNotification sends a notification
func (h *NotificationHandler) SendNotification(c *gin.Context) {
	// Get notification ID
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Notification ID is required"})
		return
	}

	// Get notification
	notification, err := h.Service.GetNotificationByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	// Check if user has permission to send this notification
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userRole, _ := c.Get("userRole")
	if notification.UserID != userID.(string) && userRole.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to send this notification"})
		return
	}

	// Send notification
	if err := h.Service.SendNotification(id); err != nil {
		logrus.Errorf("Failed to send notification: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send notification: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification sent successfully"})
}

// GetNotificationByID gets a notification by ID
func (h *NotificationHandler) GetNotificationByID(c *gin.Context) {
	// Get notification ID
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Notification ID is required"})
		return
	}

	// Get notification
	notification, err := h.Service.GetNotificationByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	// Check if user has permission to view this notification
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userRole, _ := c.Get("userRole")
	if notification.UserID != userID.(string) && userRole.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view this notification"})
		return
	}

	c.JSON(http.StatusOK, notification)
}

// GetUserNotifications gets notifications for the current user
func (h *NotificationHandler) GetUserNotifications(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Get notifications
	notifications, total, err := h.Service.GetNotificationsByUserID(userID.(string), page, limit)
	if err != nil {
		logrus.Errorf("Failed to get notifications: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"total":         total,
		"page":          page,
		"limit":         limit,
	})
}

// UpdateNotificationStatus updates a notification's status
func (h *NotificationHandler) UpdateNotificationStatus(c *gin.Context) {
	// Get notification ID
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Notification ID is required"})
		return
	}

	// Parse request
	var req model.UpdateNotificationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Get notification
	notification, err := h.Service.GetNotificationByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	// Check if user has permission to update this notification
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userRole, _ := c.Get("userRole")
	if notification.UserID != userID.(string) && userRole.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this notification"})
		return
	}

	// Update status
	if err := h.Service.UpdateNotificationStatus(id, req.Status); err != nil {
		logrus.Errorf("Failed to update notification status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification status updated successfully"})
}

// CreateTemplate creates a new notification template
func (h *NotificationHandler) CreateTemplate(c *gin.Context) {
	// Check if user has admin role
	userRole, exists := c.Get("userRole")
	if !exists || userRole.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only administrators can manage templates"})
		return
	}

	// Parse request
	var req model.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Create template
	template, err := h.Service.CreateTemplate(req)
	if err != nil {
		logrus.Errorf("Failed to create template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// GetTemplateByID gets a template by ID
func (h *NotificationHandler) GetTemplateByID(c *gin.Context) {
	// Get template ID
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID is required"})
		return
	}

	// Get template
	template, err := h.Service.GetTemplateByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	c.JSON(http.StatusOK, template)
}

// GetTemplateByCode gets a template by code
func (h *NotificationHandler) GetTemplateByCode(c *gin.Context) {
	// Get template code
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template code is required"})
		return
	}

	// Get template
	template, err := h.Service.GetTemplateByCode(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	c.JSON(http.StatusOK, template)
}

// GetAllTemplates gets all templates with pagination
func (h *NotificationHandler) GetAllTemplates(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Get templates
	templates, total, err := h.Service.GetAllTemplates(page, limit)
	if err != nil {
		logrus.Errorf("Failed to get templates: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

// UpdateTemplate updates a template
func (h *NotificationHandler) UpdateTemplate(c *gin.Context) {
	// Check if user has admin role
	userRole, exists := c.Get("userRole")
	if !exists || userRole.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only administrators can manage templates"})
		return
	}

	// Get template ID
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID is required"})
		return
	}

	// Parse request
	var req model.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Update template
	template, err := h.Service.UpdateTemplate(id, req)
	if err != nil {
		logrus.Errorf("Failed to update template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update template: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

// DeleteTemplate deletes a template
func (h *NotificationHandler) DeleteTemplate(c *gin.Context) {
	// Check if user has admin role
	userRole, exists := c.Get("userRole")
	if !exists || userRole.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only administrators can manage templates"})
		return
	}

	// Get template ID
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID is required"})
		return
	}

	// Delete template
	if err := h.Service.DeleteTemplate(id); err != nil {
		logrus.Errorf("Failed to delete template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete template: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template deleted successfully"})
}