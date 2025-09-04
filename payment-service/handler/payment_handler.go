package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/ticket-system/payment-service/model"
	"github.com/yourusername/ticket-system/payment-service/service"
)

// PaymentHandler handles HTTP requests related to payments
type PaymentHandler struct {
	paymentService service.PaymentService
}

// NewPaymentHandler creates a new payment handler
func NewPaymentHandler(paymentService service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// CreatePayment handles the creation of a new payment
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
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
	var req model.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	if req.BookingID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "booking ID is required"})
		return
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be greater than 0"})
		return
	}

	if req.Currency == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "currency is required"})
		return
	}

	if req.PaymentMethod == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment method is required"})
		return
	}

	// Create payment
	payment, err := h.paymentService.CreatePayment(userUUID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, payment)
}

// ProcessPayment handles the processing of a payment
func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
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
	var req model.ProcessPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	if req.BookingID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "booking ID is required"})
		return
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be greater than 0"})
		return
	}

	if req.Currency == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "currency is required"})
		return
	}

	if req.PaymentMethod == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment method is required"})
		return
	}

	// Validate card details for credit card payments
	if req.PaymentMethod == "credit_card" {
		if req.CardNumber == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "card number is required"})
			return
		}

		if req.CardExpiry == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "card expiry is required"})
			return
		}

		if req.CardCVC == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "card CVC is required"})
			return
		}

		if req.CardHolder == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "card holder is required"})
			return
		}
	}

	// Process payment
	payment, err := h.paymentService.ProcessPayment(userUUID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// GetPayment handles the retrieval of a payment by ID
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	// Get payment ID from URL
	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment ID is required"})
		return
	}

	// Parse payment ID
	paymentUUID, err := uuid.Parse(paymentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment ID"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get payment
	payment, err := h.paymentService.GetPaymentByID(paymentUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Check if user owns the payment or is an admin
	userRole, _ := c.Get("userRole")
	if userID.(string) != payment.UserID.String() && userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// GetPaymentByBooking handles the retrieval of a payment by booking ID
func (h *PaymentHandler) GetPaymentByBooking(c *gin.Context) {
	// Get booking ID from URL
	bookingID := c.Param("bookingId")
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

	// Get payment
	payment, err := h.paymentService.GetPaymentByBookingID(bookingUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Check if user owns the payment or is an admin
	userRole, _ := c.Get("userRole")
	if userID.(string) != payment.UserID.String() && userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// GetUserPayments handles the retrieval of payments for a user
func (h *PaymentHandler) GetUserPayments(c *gin.Context) {
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

	// Get payments
	payments, total, err := h.paymentService.GetUserPayments(userUUID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payments": payments,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// UpdatePaymentStatus handles the update of a payment status
func (h *PaymentHandler) UpdatePaymentStatus(c *gin.Context) {
	// Get payment ID from URL
	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment ID is required"})
		return
	}

	// Parse payment ID
	paymentUUID, err := uuid.Parse(paymentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment ID"})
		return
	}

	// Parse request body
	var req model.UpdatePaymentStatusRequest
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
		"completed": true,
		"failed":    true,
		"refunded":  true,
	}

	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}

	// Get user role from context
	userRole, exists := c.Get("userRole")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admins can update payment status"})
		return
	}

	// Update payment status
	payment, err := h.paymentService.UpdatePaymentStatus(paymentUUID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// RefundPayment handles the refund of a payment
func (h *PaymentHandler) RefundPayment(c *gin.Context) {
	// Get payment ID from URL
	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment ID is required"})
		return
	}

	// Parse payment ID
	paymentUUID, err := uuid.Parse(paymentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment ID"})
		return
	}

	// Parse request body
	var req model.RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user role from context
	userRole, exists := c.Get("userRole")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admins can refund payments"})
		return
	}

	// Refund payment
	payment, err := h.paymentService.RefundPayment(paymentUUID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// SetupRoutes sets up the payment routes
func (h *PaymentHandler) SetupRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	// Create payment routes group
	paymentRoutes := router.Group("/api/payments")
	paymentRoutes.Use(authMiddleware)

	// Set up routes
	paymentRoutes.POST("", h.CreatePayment)
	paymentRoutes.POST("/process", h.ProcessPayment)
	paymentRoutes.GET("/user", h.GetUserPayments)
	paymentRoutes.GET("/booking/:bookingId", h.GetPaymentByBooking)
	paymentRoutes.GET("/:id", h.GetPayment)
	paymentRoutes.PUT("/:id/status", h.UpdatePaymentStatus)
	paymentRoutes.POST("/:id/refund", h.RefundPayment)
}