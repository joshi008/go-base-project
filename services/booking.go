package services

import (
	"database/sql"
	"fmt"
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
	// Query to find an available car based on category and time slot
	query := `
		SELECT c.id
		FROM cars c
		LEFT JOIN slot_booking sb ON c.id = sb.car_id
		WHERE c.category_id = $1
		AND (sb.start_time IS NULL OR sb.end_time <= $2 OR sb.start_time >= $3)
		LIMIT 1
	`
	err = bs.DB.QueryRow(query, categoryId, startTime, endTime).Scan(&carId)
	fmt.Printf("Query executed: %s with categoryId=%d, startTime=%s, endTime=%s\n", query, categoryId, startTime, endTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // No available car found
		}
		return 0, err // Other database error
	}

	// Insert booking into the database (this is a simplified example)
	insertQuery := `
		INSERT INTO slot_booking (car_id, start_time, end_time)
		VALUES ($1, $2, $3) RETURNING id
	`
	var bookingId int
	err = bs.DB.QueryRow(insertQuery, carId, startTime, endTime).Scan(&bookingId)
	if err != nil {
		fmt.Printf("Error inserting booking into database: %v\n", err)
		return 0, err
	}

	return carId, nil
}
