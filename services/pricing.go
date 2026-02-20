package services

import "go-base-project/models"

type PricingStrategy interface {
	CalculatePrice(vehicle *models.Vehicle, system *System, booking *models.Booking) float64
}
