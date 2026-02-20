package services

import (
	"go-base-project/models"
)

type DynamicPricingStrategy struct{}

func (d *DynamicPricingStrategy) CalculatePrice(vehicle *models.Vehicle, system *System, booking *models.Booking) float64 {
	return vehicle.DailyRate * float64(booking.EndDate.Sub(booking.StartDate).Hours()/24)
}
