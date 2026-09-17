package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Vehicle struct {
	Id        uuid.UUID
	ModelName string
	DailyRate float64
}

type User struct {
	Id   uuid.UUID
	Name string
}

type Booking struct {
	Id        uuid.UUID
	User      *User
	Vehicle   *Vehicle
	StartDate time.Time
	EndDate   time.Time
	IsOngoing bool
}

func (u *User) PrintUser() {
	fmt.Printf("User ID: %s, Name: %s\n", u.Id, u.Name)
}

func (v *Vehicle) PrintVehicle() {
	fmt.Printf("Vehicle ID: %s, Model Name: %s, Daily Rate: %.2f\n", v.Id, v.ModelName, v.DailyRate)
}

func (b *Booking) PrintBooking() {
	if b == nil {
		fmt.Println("Not booked")
		return
	}
	fmt.Printf("Booking ID: %s, User: %s, Vehicle: %s, Start Date: %s, End Date: %s, Is Ongoing: %t\n", b.Id, b.User.Name, b.Vehicle.ModelName, b.StartDate.Format("2006-01-02"), b.EndDate.Format("2006-01-02"), b.IsOngoing)
}
