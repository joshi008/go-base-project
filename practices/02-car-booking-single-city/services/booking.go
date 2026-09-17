package services

import (
	"database/sql"
	"fmt"
	"time"
)

// BookingService handles booking-related operations
type BookingService struct {
	DB *sql.DB
}

// NewBookingService creates a new BookingService
func NewBookingService(db *sql.DB) *BookingService {
	return &BookingService{DB: db}
}

func (bs *BookingService) CreateBooking(categoryId int, startTime, endTime string) (carId int, err error) {
	// Parse time strings to validate format
	startTimeParsed, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return 0, fmt.Errorf("invalid start time format: %w", err)
	}
	endTimeParsed, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		return 0, fmt.Errorf("invalid end time format: %w", err)
	}

	// Calculate duration in hours for pricing
	durationHours := endTimeParsed.Sub(startTimeParsed).Hours()
	if durationHours <= 0 {
		return 0, fmt.Errorf("end time must be after start time")
	}

	// Query to find available cars with pricing (lowest price first, then by car ID for tie-breaking)
	query := `
		SELECT c.id, p.price
		FROM cars c
		INNER JOIN pricing p ON c.branch_id = p.branch_id AND c.category_id = p.category_id
		LEFT JOIN slot_booking sb ON c.id = sb.car_id 
			AND NOT (sb.end_time <= $2 OR sb.start_time >= $3)
		WHERE c.category_id = $1 AND sb.car_id IS NULL
		ORDER BY p.price ASC, c.id ASC
		LIMIT 1
	`
	
	var pricePerHour float64
	err = bs.DB.QueryRow(query, categoryId, startTime, endTime).Scan(&carId, &pricePerHour)
	fmt.Printf("Query executed: %s with categoryId=%d, startTime=%s, endTime=%s\n", query, categoryId, startTime, endTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("no available car found for category %d", categoryId)
		}
		return 0, fmt.Errorf("database error: %w", err)
	}

	// Calculate total price
	totalPrice := pricePerHour * durationHours

	// Insert booking into the database with price
	insertQuery := `
		INSERT INTO slot_booking (car_id, start_time, end_time, price)
		VALUES ($1, $2, $3, $4) RETURNING id
	`
	var bookingId int
	err = bs.DB.QueryRow(insertQuery, carId, startTime, endTime, totalPrice).Scan(&bookingId)
	if err != nil {
		fmt.Printf("Error inserting booking into database: %v\n", err)
		return 0, fmt.Errorf("failed to create booking: %w", err)
	}

	fmt.Printf("Booking created: ID=%d, CarID=%d, Price=%.2f\n", bookingId, carId, totalPrice)
	return carId, nil
}
