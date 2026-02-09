package models

import "time"

type Branches struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

//Category ID
//SEDAN : 1
//SUV : 2
//HATCHBACK : 3

type Cars struct {
	ID         int `json:"id"`
	CategoryId int `json:"category_id"`
	BranchID   int `json:"branch_id"`
}

type Pricing struct {
	ID         int     `json:"id"`
	Price      float64 `json:"price"`
	BranchID   int     `json:"branch_id"`
	CategoryId int     `json:"category_id"`
}

type SlotBooking struct {
	ID        int       `json:"id"`
	CarID     int       `json:"car_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Price     float64   `json:"price"`
}
