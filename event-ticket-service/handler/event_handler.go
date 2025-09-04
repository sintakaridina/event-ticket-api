package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/ticket-system/event-ticket-service/model"
	"github.com/yourusername/ticket-system/event-ticket-service/service"
)

// EventHandler handles HTTP requests related to events
type EventHandler struct {
	eventService service.EventService
}

// NewEventHandler creates a new event handler
func NewEventHandler(eventService service.EventService) *EventHandler {
	return &EventHandler{
		eventService: eventService,
	}
}

// CreateEvent handles the creation of a new event
func (h *EventHandler) CreateEvent(c *gin.Context) {
	// Check if user is admin
	userRole, exists := c.Get("userRole")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admins can create events"})
		return
	}

	// Parse request body
	var req model.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	if req.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "description is required"})
		return
	}

	if req.Location == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "location is required"})
		return
	}

	if req.StartDate.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start date is required"})
		return
	}

	if req.EndDate.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end date is required"})
		return
	}

	if req.StartDate.After(req.EndDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start date must be before end date"})
		return
	}

	if req.StartDate.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start date must be in the future"})
		return
	}

	if len(req.TicketTypes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one ticket type is required"})
		return
	}

	// Create event
	event, err := h.eventService.CreateEvent(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, event)
}

// GetEvent handles the retrieval of an event by ID
func (h *EventHandler) GetEvent(c *gin.Context) {
	// Get event ID from URL
	eventID := c.Param("id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "event ID is required"})
		return
	}

	// Parse event ID
	eventUUID, err := uuid.Parse(eventID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
		return
	}

	// Get event
	event, err := h.eventService.GetEventByID(eventUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, event)
}

// UpdateEvent handles the update of an event
func (h *EventHandler) UpdateEvent(c *gin.Context) {
	// Check if user is admin
	userRole, exists := c.Get("userRole")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admins can update events"})
		return
	}

	// Get event ID from URL
	eventID := c.Param("id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "event ID is required"})
		return
	}

	// Parse event ID
	eventUUID, err := uuid.Parse(eventID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
		return
	}

	// Parse request body
	var req model.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update event
	event, err := h.eventService.UpdateEvent(eventUUID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, event)
}

// DeleteEvent handles the deletion of an event
func (h *EventHandler) DeleteEvent(c *gin.Context) {
	// Check if user is admin
	userRole, exists := c.Get("userRole")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admins can delete events"})
		return
	}

	// Get event ID from URL
	eventID := c.Param("id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "event ID is required"})
		return
	}

	// Parse event ID
	eventUUID, err := uuid.Parse(eventID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
		return
	}

	// Delete event
	err = h.eventService.DeleteEvent(eventUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "event deleted successfully"})
}

// SearchEvents handles the search of events
func (h *EventHandler) SearchEvents(c *gin.Context) {
	// Parse search parameters
	keyword := c.Query("keyword")
	category := c.Query("category")
	location := c.Query("location")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	// Parse pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Create search request
	req := model.SearchEventRequest{
		Keyword:  keyword,
		Category: category,
		Location: location,
		Page:     page,
		PageSize: pageSize,
	}

	// Parse start date if provided
	if startDate != "" {
		startDateParsed, err := time.Parse(time.RFC3339, startDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format"})
			return
		}
		req.StartDate = &startDateParsed
	}

	// Parse end date if provided
	if endDate != "" {
		endDateParsed, err := time.Parse(time.RFC3339, endDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format"})
			return
		}
		req.EndDate = &endDateParsed
	}

	// Search events
	events, total, err := h.eventService.SearchEvents(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events":   events,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetAllEvents handles the retrieval of all events with pagination
func (h *EventHandler) GetAllEvents(c *gin.Context) {
	// Parse pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Get all events
	events, total, err := h.eventService.GetAllEvents(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events":   events,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// SetupRoutes sets up the event routes
func (h *EventHandler) SetupRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	// Create public event routes group
	publicEventRoutes := router.Group("/api/events")

	// Set up public routes
	publicEventRoutes.GET("", h.GetAllEvents)
	publicEventRoutes.GET("/search", h.SearchEvents)
	publicEventRoutes.GET("/:id", h.GetEvent)

	// Create protected event routes group
	protectedEventRoutes := router.Group("/api/events")
	protectedEventRoutes.Use(authMiddleware)

	// Set up protected routes
	protectedEventRoutes.POST("", h.CreateEvent)
	protectedEventRoutes.PUT("/:id", h.UpdateEvent)
	protectedEventRoutes.DELETE("/:id", h.DeleteEvent)
}