package services

import (
	"go-base-project/models"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type System struct {
	Users    []*models.User
	Vehicles []*models.Vehicle
	Bookings []*models.Booking
	mu       sync.RWMutex
}

func NewSystem() *System {
	return &System{
		Users:    []*models.User{},
		Vehicles: []*models.Vehicle{},
		Bookings: []*models.Booking{},
	}
}

func (s *System) AddUser(name string) *models.User {
	id := uuid.New()
	user := &models.User{
		Id:   id,
		Name: name,
	}
	s.mu.Lock()
	s.Users = append(s.Users, user)
	s.mu.Unlock()

	return user
}

func (s *System) AddVehicle(modelName string, dailyRate float64) *models.Vehicle {
	id := uuid.New()
	vehicle := &models.Vehicle{
		Id:        id,
		ModelName: modelName,
		DailyRate: dailyRate,
	}
	s.mu.Lock()
	s.Vehicles = append(s.Vehicles, vehicle)
	s.mu.Unlock()

	return vehicle
}

func (s *System) AddBooking(userId uuid.UUID, vehicleId uuid.UUID, startDate time.Time, endDate time.Time) *models.Booking {
	if !s.CheckAvailability(vehicleId, startDate, endDate) {
		return nil
	}

	id := uuid.New()
	user := s.GetUserById(userId)
	vehicle := s.GetVehicleById(vehicleId)
	booking := &models.Booking{
		Id:        id,
		User:      user,
		Vehicle:   vehicle,
		StartDate: startDate,
		EndDate:   endDate,
		IsOngoing: true,
	}
	s.Bookings = append(s.Bookings, booking)

	return booking
}

func (s *System) CheckAvailability(vehicleId uuid.UUID, startDate time.Time, endDate time.Time) bool {
	for _, booking := range s.Bookings {
		if booking.Vehicle.Id == vehicleId {
			if startDate.Before(booking.StartDate) && endDate.After(booking.StartDate) {
				return false
			}
			if startDate.Before(booking.EndDate) && endDate.After(booking.EndDate) {
				return false
			}
			if startDate.After(booking.StartDate) && endDate.Before(booking.EndDate) {
				return false
			}
		}
	}
	return true
}

func (s *System) GetUserById(id uuid.UUID) *models.User {
	for _, user := range s.Users {
		if user.Id == id {
			return user
		}
	}
	return nil
}

func (s *System) GetVehicleById(id uuid.UUID) *models.Vehicle {
	for _, vehicle := range s.Vehicles {
		if vehicle.Id == id {
			return vehicle
		}
	}
	return nil
}

func (s *System) GetBookingById(id uuid.UUID) *models.Booking {
	for _, booking := range s.Bookings {
		if booking.Id == id {
			return booking
		}
	}
	return nil
}

func (s *System) SearchCar(name string) *[]models.Vehicle {
	var vehicles []models.Vehicle
	for _, vehicle := range s.Vehicles {
		if strings.Contains(vehicle.ModelName, name) {
			vehicles = append(vehicles, *vehicle)
		}
	}
	return &vehicles
}

func (s *System) ReturnBooking(bookingId uuid.UUID, returnTime time.Time) (cost float64) {
	booking := s.GetBookingById(bookingId)
	if booking == nil {
		return 0
	}
	pricingStrategy := GetPricingStrategy("default")
	cost = pricingStrategy.CalculatePrice(booking.Vehicle, s, booking)
	booking.IsOngoing = false
	booking.EndDate = returnTime
	return cost
}
