package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/ticket-system/event-ticket-service/config"
	"github.com/yourusername/ticket-system/event-ticket-service/model"
	"github.com/yourusername/ticket-system/event-ticket-service/repository"
)

// EventService defines the interface for event service operations
type EventService interface {
	CreateEvent(req model.CreateEventRequest) (*model.EventResponse, error)
	GetEventByID(id uuid.UUID) (*model.EventResponse, error)
	UpdateEvent(id uuid.UUID, req model.UpdateEventRequest) (*model.EventResponse, error)
	DeleteEvent(id uuid.UUID) error
	SearchEvents(req model.SearchEventRequest) ([]model.EventResponse, int64, error)
	GetAllEvents(page, pageSize int) ([]model.EventResponse, int64, error)
}

// eventService implements EventService interface
type eventService struct {
	eventRepo  repository.EventRepository
	ticketRepo repository.TicketRepository
	rmq        *config.RabbitMQ
}

// NewEventService creates a new event service
func NewEventService(eventRepo repository.EventRepository, ticketRepo repository.TicketRepository, rmq *config.RabbitMQ) EventService {
	return &eventService{
		eventRepo:  eventRepo,
		ticketRepo: ticketRepo,
		rmq:        rmq,
	}
}

// CreateEvent creates a new event
func (s *eventService) CreateEvent(req model.CreateEventRequest) (*model.EventResponse, error) {
	// Create event
	event := &model.Event{
		Name:        req.Name,
		Description: req.Description,
		Location:    req.Location,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Category:    req.Category,
		Organizer:   req.Organizer,
		ImageURL:    req.ImageURL,
		Status:      "active",
	}

	// Save event to database
	if err := s.eventRepo.Create(event); err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	// Create tickets for the event
	tickets := make([]*model.Ticket, 0)
	for _, ticketReq := range req.Tickets {
		for i := 0; i < ticketReq.Quantity; i++ {
			tickets = append(tickets, &model.Ticket{
				EventID: event.ID,
				Type:    ticketReq.Type,
				Price:   ticketReq.Price,
				Status:  "available",
			})
		}
	}

	// Save tickets to database
	if err := s.ticketRepo.CreateBatch(tickets); err != nil {
		return nil, fmt.Errorf("failed to create tickets: %w", err)
	}

	// Set tickets in event
	event.Tickets = make([]model.Ticket, len(tickets))
	for i, ticket := range tickets {
		event.Tickets[i] = *ticket
	}

	// Publish event created event
	s.publishEventEvent("event.created", event)

	// Return event response
	eventResponse := event.ToResponse()
	return &eventResponse, nil
}

// GetEventByID gets an event by ID
func (s *eventService) GetEventByID(id uuid.UUID) (*model.EventResponse, error) {
	// Find event by ID
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find event: %w", err)
	}

	if event == nil {
		return nil, fmt.Errorf("event not found")
	}

	// Return event response
	eventResponse := event.ToResponse()
	return &eventResponse, nil
}

// UpdateEvent updates an event
func (s *eventService) UpdateEvent(id uuid.UUID, req model.UpdateEventRequest) (*model.EventResponse, error) {
	// Find event by ID
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find event: %w", err)
	}

	if event == nil {
		return nil, fmt.Errorf("event not found")
	}

	// Update event fields
	if req.Name != "" {
		event.Name = req.Name
	}
	if req.Description != "" {
		event.Description = req.Description
	}
	if req.Location != "" {
		event.Location = req.Location
	}
	if !req.StartDate.IsZero() {
		event.StartDate = req.StartDate
	}
	if !req.EndDate.IsZero() {
		event.EndDate = req.EndDate
	}
	if req.Category != "" {
		event.Category = req.Category
	}
	if req.Organizer != "" {
		event.Organizer = req.Organizer
	}
	if req.ImageURL != "" {
		event.ImageURL = req.ImageURL
	}
	if req.Status != "" {
		event.Status = req.Status
	}

	// Save event to database
	if err := s.eventRepo.Update(event); err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	// Publish event updated event
	s.publishEventEvent("event.updated", event)

	// Return event response
	eventResponse := event.ToResponse()
	return &eventResponse, nil
}

// DeleteEvent deletes an event
func (s *eventService) DeleteEvent(id uuid.UUID) error {
	// Find event by ID
	event, err := s.eventRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("failed to find event: %w", err)
	}

	if event == nil {
		return fmt.Errorf("event not found")
	}

	// Delete event from database
	if err := s.eventRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	// Publish event deleted event
	s.publishEventEvent("event.deleted", event)

	return nil
}

// SearchEvents searches for events based on criteria
func (s *eventService) SearchEvents(req model.SearchEventRequest) ([]model.EventResponse, int64, error) {
	// Search events
	events, total, err := s.eventRepo.Search(req.Keyword, req.Category, req.Location, req.StartDate, req.EndDate, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search events: %w", err)
	}

	// Convert events to responses
	responses := make([]model.EventResponse, len(events))
	for i, event := range events {
		responses[i] = event.ToResponse()
	}

	return responses, total, nil
}

// GetAllEvents gets all events with pagination
func (s *eventService) GetAllEvents(page, pageSize int) ([]model.EventResponse, int64, error) {
	// Get all events
	events, total, err := s.eventRepo.FindAll(page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get events: %w", err)
	}

	// Convert events to responses
	responses := make([]model.EventResponse, len(events))
	for i, event := range events {
		responses[i] = event.ToResponse()
	}

	return responses, total, nil
}

// publishEventEvent publishes an event event to RabbitMQ
func (s *eventService) publishEventEvent(eventType string, event *model.Event) {
	// Create event payload
	payload := map[string]interface{}{
		"event_type": eventType,
		"event_id":   event.ID.String(),
		"name":       event.Name,
		"timestamp":  time.Now(),
	}

	// Convert payload to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal event event")
		return
	}

	// Publish event to RabbitMQ
	err = s.rmq.PublishMessage("ticket_events", eventType, payloadJSON)
	if err != nil {
		logrus.WithError(err).Error("Failed to publish event event")
	}
}