package main

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

type Driver struct {
	driverId                 int
	usdHourlyRatePerDelivery float64
}

type Delivery struct {
	deliveryId int
	driverId   int
	startTime  time.Time
	endTime    time.Time
	cost       float64
	paid       bool
}

// Constructor
func NewDriver(driverId int, usdHourlyRatePerDelivery float64) *Driver {
	return &Driver{
		driverId:                 driverId,
		usdHourlyRatePerDelivery: usdHourlyRatePerDelivery,
	}
}
func NewDelivery(deliveryId int, driverId int, startTime time.Time, endTime time.Time, cost float64) *Delivery {
	return &Delivery{
		deliveryId: deliveryId,
		driverId:   driverId,
		startTime:  startTime,
		endTime:    endTime,
		cost:       cost,
		paid:       false,
	}
}

// Storage
type DB struct {
	Drivers    map[int]*Driver
	Deliveries map[int]*Delivery
}

func NewDb() *DB {
	return &DB{
		Drivers:    make(map[int]*Driver),
		Deliveries: make(map[int]*Delivery),
	}
}

// DeliveryService
type DeliveryService struct {
	DB *DB
}

func NewDeliveryService(DB *DB) *DeliveryService {
	return &DeliveryService{
		DB: DB,
	}
}

func (d *DeliveryService) AddDriver(driverId int, usdHourlyRatePerDelivery float64) *Driver {
	if d.DB.Drivers[driverId] != nil {
		fmt.Println("Driver Already there")
		return nil
	}
	dr := NewDriver(driverId, usdHourlyRatePerDelivery)
	d.DB.Drivers[driverId] = dr

	return dr
}

func (d *DeliveryService) RecordDelivery(driverId int, startTime time.Time, endTime time.Time) *Delivery {
	timeDuration := math.Ceil(math.Round(endTime.Sub(startTime).Hours()*100) / 100)

	cost := d.DB.Drivers[driverId].usdHourlyRatePerDelivery * timeDuration

	dl := NewDelivery(driverId, driverId, startTime, endTime, cost)
	deliveryId := len(d.DB.Deliveries) + 1
	d.DB.Deliveries[deliveryId] = dl

	return dl
}

func (d *DeliveryService) GetTotalCost() string {
	cost := 0.0
	for _, dl := range d.DB.Deliveries {
		cost += dl.cost
	}
	return strconv.FormatFloat(cost, 'f', 2, 64)
}

func (d *DeliveryService) PayUpTo(payTime time.Time) string {
	cost := 0.0
	for _, dl := range d.DB.Deliveries {
		if payTime.Sub(dl.endTime) >= 0 {
			cost += dl.cost
			dl.paid = true
		}
	}
	return strconv.FormatFloat(cost, 'f', 2, 64)
}

func (d *DeliveryService) GetTotalCostUnpaid() string {
	cost := 0.0
	for _, dl := range d.DB.Deliveries {
		if !dl.paid {
			cost += dl.cost
		}
	}
	return strconv.FormatFloat(cost, 'f', 2, 64)
}

func main() {
	fmt.Println("Start of program")

	db := NewDb()

	ds := NewDeliveryService(db)

	dr1 := ds.AddDriver(1, 5)
	dr2 := ds.AddDriver(2, 6)

	ds.RecordDelivery(dr1.driverId, time.Now().Add(-2*time.Hour), time.Now().Add(-1*time.Hour))
	ds.RecordDelivery(dr2.driverId, time.Now().Add(-6*time.Hour), time.Now().Add(-3*time.Hour))

	cost := ds.GetTotalCost()

	fmt.Println("Final Cost: ", cost)

	cost1 := ds.PayUpTo(time.Now().Add(-2 * time.Hour))
	fmt.Println("Paytime Cost: ", cost1)

	cost2 := ds.GetTotalCostUnpaid()
	fmt.Println("Cost Unpaid : ", cost2)
	return
}
