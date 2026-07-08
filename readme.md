The parking lot has multiple parking spots, including compact, regular, and oversized spots.
The parking lot supports parking for motorcycles, cars, and trucks.
Customers can park their vehicles in spots assigned based on vehicle size.
Customers receive a parking ticket with vehicle details and entry time at the entry point and pay a fee based on duration, vehicle size, and time of day at the exit point.

Entities

Vehicle
- ID
- Type
- Appointment

ParkingSlots
- ID
- VehicleType
- TotalSlots
- AvailableSlots

Pricing
- ID
- VehicleType
- AmountPerHour

Appointment
- ID
- VehicleId
- EntryTime
- ExitTime
- FinalAmount

APIs
- AddVehicle
- GetAvailableSlots
- ExitVehicle (GetFInalPrice)
