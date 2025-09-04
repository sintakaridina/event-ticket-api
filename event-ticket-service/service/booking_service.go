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
	"gorm.io/gorm"
)

// BookingService defines the interface for booking service operations
type BookingService interface {
	CreateBooking(userID uuid.UUID, req model.CreateBookingRequest) (*model.BookingResponse, error)
	GetBookingByID(id uuid.UUID) (*model.BookingResponse, error)
	GetUserBookings(userID uuid.UUID, page, pageSize int) ([]model.BookingResponse, int64, error)
	UpdateBookingStatus(id uuid.UUID, status string) (*model.BookingResponse, error)
	CancelBooking(id uuid.UUID) error
}

// bookingService implements BookingService interface
type bookingService struct {
	bookingRepo repository.BookingRepository
	eventRepo   repository.EventRepository
	ticketRepo  repository.TicketRepository
	db          *gorm.DB
	rmq         *config.RabbitMQ
}

// NewBookingService creates a new booking service
func NewBookingService(
	bookingRepo repository.BookingRepository,
	eventRepo repository.EventRepository,
	ticketRepo repository.TicketRepository,
	db *gorm.DB,
	rmq *config.RabbitMQ,
) BookingService {
	return &bookingService{
		bookingRepo: bookingRepo,
		eventRepo:   eventRepo,
		ticketRepo:  ticketRepo,
		db:          db,
		rmq:         rmq,
	}
}

// CreateBooking creates a new booking
func (s *bookingService) CreateBooking(userID uuid.UUID, req model.CreateBookingRequest) (*model.BookingResponse, error) {
	// Find event by ID
	event, err := s.eventRepo.FindByID(req.EventID)
	if err != nil {
		return nil, fmt.Errorf("failed to find event: %w", err)
	}

	if event == nil {
		return nil, fmt.Errorf("event not found")
	}

	// Check if event is active
	if event.Status != "active" {
		return nil, fmt.Errorf("event is not active")
	}

	// Check if event date is in the future
	if event.StartDate.Before(time.Now()) {
		return nil, fmt.Errorf("event has already started")
	}

	// Use transaction to ensure data consistency
	var booking *model.Booking
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Create booking
		booking = &model.Booking{
			UserID:     userID,
			EventID:    req.EventID,
			Status:     "pending",
			TotalPrice: 0,
		}

		// Find available tickets for each ticket type
		selectedTickets := make([]*model.Ticket, 0)
		totalPrice := 0.0

		for _, ticketReq := range req.Tickets {
			// Find available tickets of the requested type
			availableTickets, err := s.ticketRepo.FindAvailableByEventID(req.EventID, ticketReq.Type)
			if err != nil {
				return fmt.Errorf("failed to find available tickets: %w", err)
			}

			// Check if enough tickets are available
			if len(availableTickets) < ticketReq.Quantity {
				return fmt.Errorf("not enough tickets available for type %s", ticketReq.Type)
			}

			// Select tickets
			for i := 0; i < ticketReq.Quantity; i++ {
				ticket := &availableTickets[i]
				ticket.Status = "reserved"
				ticket.UserID = userID
				selectedTickets = append(selectedTickets, ticket)
				totalPrice += ticket.Price
			}
		}

		// Set total price
		booking.TotalPrice = totalPrice

		// Save booking to database
		if err := s.bookingRepo.Create(booking); err != nil {
			return fmt.Errorf("failed to create booking: %w", err)
		}

		// Update tickets with booking ID and status
		for _, ticket := range selectedTickets {
			ticket.BookingID = booking.ID
			ticket.Status = "reserved"
		}

		// Update tickets in database
		if err := s.ticketRepo.UpdateBatch(selectedTickets); err != nil {
			return fmt.Errorf("failed to update tickets: %w", err)
		}

		// Set tickets in booking
		booking.Tickets = make([]model.Ticket, len(selectedTickets))
		for i, ticket := range selectedTickets {
			booking.Tickets[i] = *ticket
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Publish booking created event
	s.publishBookingEvent("booking.created", booking)

	// Return booking response
	bookingResponse := booking.ToResponse(false)
	return &bookingResponse, nil
}

// GetBookingByID gets a booking by ID
func (s *bookingService) GetBookingByID(id uuid.UUID) (*model.BookingResponse, error) {
	// Find booking by ID
	booking, err := s.bookingRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find booking: %w", err)
	}

	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	// Find event
	event, err := s.eventRepo.FindByID(booking.EventID)
	if err != nil {
		return nil, fmt.Errorf("failed to find event: %w", err)
	}

	// Convert booking to response
	response := booking.ToResponse(true)

	// Set event in response
	if event != nil {
		eventResponse := event.ToResponse()
		response.Event = &eventResponse
	}

	return &response, nil
}

// GetUserBookings gets bookings for a user
func (s *bookingService) GetUserBookings(userID uuid.UUID, page, pageSize int) ([]model.BookingResponse, int64, error) {
	// Find bookings by user ID
	bookings, total, err := s.bookingRepo.FindByUserID(userID, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find bookings: %w", err)
	}

	// Convert bookings to responses
	responses := make([]model.BookingResponse, len(bookings))
	for i, booking := range bookings {
		responses[i] = booking.ToResponse(false)
	}

	return responses, total, nil
}

// UpdateBookingStatus updates a booking status
func (s *bookingService) UpdateBookingStatus(id uuid.UUID, status string) (*model.BookingResponse, error) {
	// Find booking by ID
	booking, err := s.bookingRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find booking: %w", err)
	}

	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	// Update booking status
	booking.Status = status

	// Save booking to database
	if err := s.bookingRepo.Update(booking); err != nil {
		return nil, fmt.Errorf("failed to update booking: %w", err)
	}

	// Update ticket status based on booking status
	ticketStatus := "reserved"
	switch status {
	case "confirmed":
		ticketStatus = "sold"
	case "cancelled":
		ticketStatus = "available"
	case "refunded":
		ticketStatus = "available"
	}

	// Find tickets for booking
	tickets, err := s.ticketRepo.FindByBookingID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find tickets: %w", err)
	}

	// Update ticket status
	for i := range tickets {
		tickets[i].Status = ticketStatus
		if ticketStatus == "available" {
			tickets[i].UserID = uuid.Nil
			tickets[i].BookingID = uuid.Nil
		}
	}

	// Convert tickets to pointers
	ticketPtrs := make([]*model.Ticket, len(tickets))
	for i := range tickets {
		ticketPtrs[i] = &tickets[i]
	}

	// Update tickets in database
	if err := s.ticketRepo.UpdateBatch(ticketPtrs); err != nil {
		return nil, fmt.Errorf("failed to update tickets: %w", err)
	}

	// Publish booking updated event
	s.publishBookingEvent("booking.updated", booking)

	// Return booking response
	bookingResponse := booking.ToResponse(false)
	return &bookingResponse, nil
}

// CancelBooking cancels a booking
func (s *bookingService) CancelBooking(id uuid.UUID) error {
	// Update booking status to cancelled
	_, err := s.UpdateBookingStatus(id, "cancelled")
	if err != nil {
		return err
	}

	// Publish booking cancelled event
	booking, err := s.bookingRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("failed to find booking: %w", err)
	}

	if booking != nil {
		s.publishBookingEvent("booking.cancelled", booking)
	}

	return nil
}

// publishBookingEvent publishes a booking event to RabbitMQ
func (s *bookingService) publishBookingEvent(eventType string, booking *model.Booking) {
	// Create event payload
	payload := map[string]interface{}{
		"event_type": eventType,
		"booking_id": booking.ID.String(),
		"user_id":    booking.UserID.String(),
		"event_id":   booking.EventID.String(),
		"status":     booking.Status,
		"timestamp":  time.Now(),
	}

	// Convert payload to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal booking event")
		return
	}

	// Publish event to RabbitMQ
	err = s.rmq.PublishMessage("ticket_events", eventType, payloadJSON)
	if err != nil {
		logrus.WithError(err).Error("Failed to publish booking event")
	}
}