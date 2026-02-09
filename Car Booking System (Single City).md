Car Booking System (Single City)
Design and implement a backend system for booking cars within a single city.
 The city operates multiple branches, each managing its own fleet of cars.
Domain Overview
Branches

Each branch is uniquely identified by its name
A branch owns a fleet of cars and maintains its own pricing configuration
Cars
• Every car belongs to exactly one branch
• Supported car categories:
SEDAN
SUV
HATCHBACK
• Each car has a unique identifier within its branch
Pricing Rules
Pricing is configured per branch and per car category (-)

Price is defined as price per hour
All cars of the same category within a branch share the same price (-)
Pricing may change over time 
Booking Rules
• A user requests a car by specifying:
Car category
Time window (startDateTime, endDateTime)
• The system must allocate one available car that satisfies the request
• Among all valid options:
Select the car with the lowest price
If multiple cars have the same price, apply a deterministic tie-breaking rule
• A booking, once confirmed, blocks that car for the given time window
• Overlapping bookings for the same car are not allowed
Expectations

The system should support onboarding new branches, cars, and pricing configurations
Users should be able to request bookings without needing to know which branch or car they are getting
The system must ensure correctness even when multiple booking requests happen close together in time (Consistent - Lock) 

Entities:
Branches
    - ID
    - NAME

Cars
    - ID
    - CategoryId
    - BranchID

Pricing (Regular Updates)
    - ID
    - Price (Per hour)
    - BranchID
    - CategoryId

SlotBooking
    - StartTime
    - CarId
    - Duration (5h) ----
    - Price
    - ID


API Structures

Creation for branches, cars, pricing
Updation for pricing
/get-car - CategoryId, StartTime, EndTime -> CarId
