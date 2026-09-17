package main

import (
	"fmt"
	"time"
)

type VehicleType int

const (
	MOTORCYCLES VehicleType = iota
	CARS
	TRUCKS
)

type Vehicle struct {
	ID   string
	Type VehicleType
}

type ParkingSlots struct {
	ID             string
	VehicleType    VehicleType
	TotalSlots     int
	AvailableSlots int
}

type Pricing struct {
	ID            string
	VehicleType   VehicleType
	AmountPerHour int
}

type Appointment struct {
	ID          string
	VehicleId   string
	EntryTime   time.Time
	ExitTime    time.Time
	FinalAmount int
}

// Constructor
func NewParkingSlot(ID string, VehicleType VehicleType, TotalSlots int) *ParkingSlots {
	return &ParkingSlots{
		ID:             ID,
		VehicleType:    VehicleType,
		TotalSlots:     TotalSlots,
		AvailableSlots: TotalSlots,
	}
}

func NewVehicle(ID string, VehicleType VehicleType) *Vehicle {
	return &Vehicle{
		ID:   ID,
		Type: VehicleType,
	}
}

func NewPricing(ID string, VehicleType VehicleType, AmountPerHour int) *Pricing {
	return &Pricing{
		ID:            ID,
		VehicleType:   VehicleType,
		AmountPerHour: AmountPerHour,
	}
}

func NewAppointment(ID string, VehicleID string, EntryTime time.Time) *Appointment {
	return &Appointment{
		ID:          ID,
		VehicleId:   VehicleID,
		EntryTime:   EntryTime,
		ExitTime:    EntryTime,
		FinalAmount: -1,
	}
}

// In Memory Storage
type storage struct {
	vehicle      map[string]*Vehicle
	parkingSlots map[string]*ParkingSlots
	pricing      map[string]*Pricing
	appointments map[string]*Appointment
}

func NewStorage() *storage {
	return &storage{
		vehicle:      make(map[string]*Vehicle),
		parkingSlots: make(map[string]*ParkingSlots),
		pricing:      make(map[string]*Pricing),
		appointments: make(map[string]*Appointment),
	}
}
func (s *storage) AddVehicle(V *Vehicle) *Vehicle {
	fmt.Println("New Vehicle created: ", V.ID, V.Type)
	s.vehicle[V.ID] = V
	return V
}
func (s *storage) AddParkingSlot(P *ParkingSlots) {
	fmt.Println("New Parking Slot added: ", P.ID, P.VehicleType, P.TotalSlots)
	s.parkingSlots[P.ID] = P
}
func (s *storage) AddPricing(V *Pricing) {
	fmt.Println("New Pricing Added: ", V.ID, V.VehicleType, V.AmountPerHour)
	s.pricing[V.ID] = V
}
func (s *storage) GetPriceByType(vehicleType VehicleType) *Pricing {
	for _, p := range s.pricing {
		if p.VehicleType == vehicleType {
			return p
		}
	}
	return nil
}
func (s *storage) GetParkingLotByType(vehicleType VehicleType) *ParkingSlots {
	for _, p := range s.parkingSlots {
		if p.VehicleType == vehicleType {
			return p
		}
	}
	return nil
}
func (s *storage) GetVehicleByVehicleID(vehicleID string) *Vehicle {
	for _, p := range s.vehicle {
		if p.ID == vehicleID {
			return p
		}
	}
	return nil
}

// Sevice
type ParkingService struct {
	db *storage
}

func NewParkingService(db *storage) *ParkingService {
	return &ParkingService{
		db: db,
	}
}

func (P *ParkingService) EnterVehicle(Vehicle *Vehicle, entryTime time.Time) *Appointment {
	price := P.db.GetPriceByType(Vehicle.Type)
	if price == nil {
		return nil
	}

	lot := P.db.GetParkingLotByType(Vehicle.Type)
	if lot == nil || lot.AvailableSlots == 0 {
		fmt.Println("Slots Full")
		return nil
	}

	lot.AvailableSlots = lot.AvailableSlots - 1

	fmt.Println("Vehicle Entered : ", Vehicle)
	aid := fmt.Sprintf("aid-%d", len(P.db.appointments))
	a := NewAppointment(aid, Vehicle.ID, entryTime)

	return a
}
func (P *ParkingService) ExitVehicle(Appointment *Appointment, exitTime time.Time) {
	vehicle := P.db.GetVehicleByVehicleID(Appointment.VehicleId)
	price := P.db.GetPriceByType(vehicle.Type)

	lot := P.db.GetParkingLotByType(vehicle.Type)

	timeInHours := int(time.Duration(exitTime.Sub(Appointment.EntryTime)).Hours())

	Appointment.FinalAmount = price.AmountPerHour * timeInHours

	fmt.Println("Vehicle Exited : ", vehicle)
	fmt.Println("Pay: ", Appointment.FinalAmount)
	lot.AvailableSlots = lot.AvailableSlots + 1
}

func main() {
	fmt.Println("Starting Program!!!")

	db := NewStorage()

	v1 := db.AddVehicle(NewVehicle("DL1", CARS))
	v2 := db.AddVehicle(NewVehicle("UP1", TRUCKS))
	db.AddVehicle(NewVehicle("UP9", TRUCKS))
	db.AddVehicle(NewVehicle("DL2", CARS))
	db.AddVehicle(NewVehicle("DL3", CARS))
	v3 := db.AddVehicle(NewVehicle("KA3", MOTORCYCLES))
	db.AddVehicle(NewVehicle("KA5", MOTORCYCLES))

	db.AddParkingSlot(NewParkingSlot("P1", CARS, 25))
	db.AddParkingSlot(NewParkingSlot("P2", MOTORCYCLES, 50))
	db.AddParkingSlot(NewParkingSlot("P3", TRUCKS, 20))

	db.AddPricing(NewPricing("PC1", CARS, 20))
	db.AddPricing(NewPricing("PC2", MOTORCYCLES, 10))
	db.AddPricing(NewPricing("PC3", TRUCKS, 50))

	ps := NewParkingService(db)

	a1 := ps.EnterVehicle(v1, time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC))
	ps.EnterVehicle(v2, time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC))
	ps.EnterVehicle(v3, time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC))
	ps.ExitVehicle(a1, time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC))

}
