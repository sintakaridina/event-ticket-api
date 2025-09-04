package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/ticket-system/event-ticket-service/model"
	"github.com/yourusername/ticket-system/event-ticket-service/service"
)

// BookingHandler handles HTTP requests related to bookings
type BookingHandler struct {
	bookingService service.BookingService
}

// NewBookingHandler creates a new booking handler
func NewBookingHandler(bookingService service.BookingService) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
	}
}

// CreateBooking handles the creation of a new booking
func (h *BookingHandler) CreateBooking(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Parse user ID
	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Parse request body
	var req model.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	if req.EventID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "event ID is required"})
		return
	}

	if len(req.Tickets) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one ticket is required"})
		return
	}

	// Create booking
	booking, err := h.bookingService.CreateBooking(userUUID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, booking)
}

// GetBooking handles the retrieval of a booking by ID
func (h *BookingHandler) GetBooking(c *gin.Context) {
	// Get booking ID from URL
	bookingID := c.Param("id")
	if bookingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "booking ID is required"})
		return
	}

	// Parse booking ID
	bookingUUID, err := uuid.Parse(bookingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking ID"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get booking
	booking, err := h.bookingService.GetBookingByID(bookingUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Check if user owns the booking or is an admin
	userRole, _ := c.Get("userRole")
	if userID.(string) != booking.UserID && userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	c.JSON(http.StatusOK, booking)
}

// GetUserBookings handles the retrieval of bookings for a user
func (h *BookingHandler) GetUserBookings(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Parse user ID
	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Get bookings
	bookings, total, err := h.bookingService.GetUserBookings(userUUID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bookings": bookings,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// UpdateBookingStatus handles the update of a booking status
func (h *BookingHandler) UpdateBookingStatus(c *gin.Context) {
	// Get booking ID from URL
	bookingID := c.Param("id")
	if bookingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "booking ID is required"})
		return
	}

	// Parse booking ID
	bookingUUID, err := uuid.Parse(bookingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking ID"})
		return
	}

	// Parse request body
	var req model.UpdateBookingStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	if req.Status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
		return
	}

	// Check if status is valid
	validStatuses := map[string]bool{
		"pending":   true,
		"confirmed": true,
		"cancelled": true,
		"refunded":  true,
	}

	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}

	// Get user role from context
	userRole, exists := c.Get("userRole")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admins can update booking status"})
		return
	}

	// Update booking status
	booking, err := h.bookingService.UpdateBookingStatus(bookingUUID, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, booking)
}

// CancelBooking handles the cancellation of a booking
func (h *BookingHandler) CancelBooking(c *gin.Context) {
	// Get booking ID from URL
	bookingID := c.Param("id")
	if bookingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "booking ID is required"})
		return
	}

	// Parse booking ID
	bookingUUID, err := uuid.Parse(bookingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking ID"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get booking
	booking, err := h.bookingService.GetBookingByID(bookingUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Check if user owns the booking or is an admin
	userRole, _ := c.Get("userRole")
	if userID.(string) != booking.UserID && userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	// Cancel booking
	err = h.bookingService.CancelBooking(bookingUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "booking cancelled successfully"})
}

// SetupRoutes sets up the booking routes
func (h *BookingHandler) SetupRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	// Create booking routes group
	bookingRoutes := router.Group("/api/bookings")
	bookingRoutes.Use(authMiddleware)

	// Set up routes
	bookingRoutes.POST("", h.CreateBooking)
	bookingRoutes.GET("/user", h.GetUserBookings)
	bookingRoutes.GET("/:id", h.GetBooking)
	bookingRoutes.PUT("/:id/status", h.UpdateBookingStatus)
	bookingRoutes.POST("/:id/cancel", h.CancelBooking)
}