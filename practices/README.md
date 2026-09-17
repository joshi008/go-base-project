# Practice questions and solutions

Compiled snapshots of every `hrishabh/practice-*` branch. Original practice branches were not modified.

Each folder is the **tip of that practice branch** at the time this `dev` branch was created. Later practices were often stacked on earlier ones and then cleaned, so keeping them in separate folders is how all questions and solutions coexist.

| # | Question | Source branch | Commit | Folder |
|---|----------|---------------|--------|--------|
| 1 | Expense / Splitwise (equal, exact, percentage splits) | `hrishabh/practice-1` | `af88c48` | [01-expense-splitwise](./01-expense-splitwise) |
| 2 | Car Booking System (Single City) | `hrishabh/practice-2` | `da64cf5` | [02-car-booking-single-city](./02-car-booking-single-city) |
| 3 | In-memory Car Rental System (with dynamic pricing) | `hrishabh/practice-3` | `04f3aca` | [03-car-rental-system](./03-car-rental-system) |
| 4 | Parking system (design notes) | `hrishabh/practice-4` | `7d5500e` | [04-parking-system](./04-parking-system) |
| 5 | Hierarchical Restaurant Menu LLD | `hrishabh/practice-5` | `0cacf35` | [05-restaurant-menu](./05-restaurant-menu) |
| 6 | Refresh / empty slate (no problem files) | `hrishabh/practice-6` | `167437d` | [06-refresh](./06-refresh) |
| 7 | Splitwise design (equal, percentage, exact) | `hrishabh/practice-7` | `c3bcda8` | [07-splitwise](./07-splitwise) |
| 8 | Splitwise | `hrishabh/practice-8` | `7899267` | [08-splitwise](./08-splitwise) |
| 9 | Parking Lot System | `hrishabh/practice-9` | `530e067` | [09-parking-lot](./09-parking-lot) |
| 10 | Splitwise | `hrishabh/practice-10` | `0445d40` | [10-splitwise](./10-splitwise) |
| 11 | Delivery cost dashboard | `hrishabh/practice-11` | `e6e7591` | [11-delivery-cost-dashboard](./11-delivery-cost-dashboard) |
| 12 | Splitwise | `hrishabh/practice-12` | `c840def` | [12-splitwise](./12-splitwise) |
| 13 | Library Book Management System | `hrishabh/practice-13` | `804ff4d` | [13-library-management](./13-library-management) |
| 14 | Fizz Buzz | `hrishabh/practice-14` | `10a4c3d` | [14-fizzbuzz](./14-fizzbuzz) |
| 15 | Odd / even concurrent printing | `hrishabh/practice-15` | `17c058c` | [15-odd-even](./15-odd-even) |
| 16 | Splitwise | `hrishabh/practice-16` | `baf912a` | [16-splitwise](./16-splitwise) |
| 17 | Eligibility Rule Engine | `hrishabh/practice-17` | `8181a65` | [17-eligibility-rule-engine](./17-eligibility-rule-engine) |

To run a snapshot that has its own `go.mod`, `cd` into that folder and use `go run`.

---

## 1. Expense / Splitwise

**Question:** Build expense splitting on top of the base HTTP project. Support equal, exact, and percentage splits when creating an expense with a lender and borrowers.

**Solution:** `01-expense-splitwise/`

- Models: `models/main.go`
- Strategies: `services/strategy/`
- HTTP: `controllers/expense_controller.go`, `controllers/user_controller.go`, `routes/routes.go`

## 2. Car Booking System (Single City)

**Question:** Design a backend for booking cars in a single city. Multiple branches, each with its own fleet (SEDAN / SUV / HATCHBACK) and per-category hourly pricing. Allocate the cheapest available car for a time window; overlapping bookings are not allowed; concurrent requests must stay consistent.

**Problem statement:** `02-car-booking-single-city/Car Booking System (Single City).md`

**Solution:** `02-car-booking-single-city/controllers/car_management.go`, `services/booking.go`, `models/models.go`

## 3. In-memory Car Rental System

**Question:** In-memory rental system: add vehicles and users, search availability, book for N days, return and compute cost. Bonus: dynamic pricing (+10% if less than 20% of the fleet is available) and thread-safe booking.

**Problem statement:** `03-car-rental-system/problem.txt`

**Solution:** `03-car-rental-system/cmd/main.go`, `services/` (`booking.go`, `pricing.go`, `defaultPricing.go`, `dynamicPricing.go`, `pricingFactory.go`)

## 4. Parking system

**Question:** Parking system to book/reserve a spot and calculate fare by time. Covers fare members, where entry time lives, API design, schema, and entities (`ParkingSystem`, `Vehicle`, `Fare`, `Booking`, `Slot`).

**Problem / notes:** `04-parking-system/cmd/something.txt`

**Solution:** mostly design notes; `cmd/main.go` is a placeholder Hello World.

## 5. Hierarchical Restaurant Menu LLD

**Question:** Model a nested restaurant menu, compute dynamic cart pricing without hardcoded rules, and serve the nested menu quickly at scale.

**Problem / plan:** `05-restaurant-menu/plan.md`

**Solution:** `05-restaurant-menu/menu/` (`entity.go`, `handler.go`, `repository.go`, `service.go`)

## 6. Refresh

Empty slate after practice 5 (`go.mod` / `go.sum` only). No question or solution on this branch tip.

## 7. Splitwise (design + first implementation)

**Question:** Users add splits (equal, percentage, exact) and can check balances / who owes whom. Target ~1M DAU; consistency over availability.

**Problem statement:** `07-splitwise/readme.txt`

**Solution:** `07-splitwise/cmd/main.go`, `models/`, `strategy/`

## 8. Splitwise

**Question / entities:** User, Transaction, Split (lender, borrower, raw value, amount).

**Problem statement:** `08-splitwise/readme.md`

**Solution:** `08-splitwise/cmd/main.go`, `guide.go`

## 9. Parking Lot System

**Question:** Multi-size lot (compact, regular, oversized) for motorcycles, cars, and trucks. Assign spots by vehicle size. Ticket at entry; fee at exit based on duration, vehicle size, and time of day.

**Problem statement:** `09-parking-lot/readme.md`

**Solution:** `09-parking-lot/cmd/main.go`

## 10. Splitwise

**Question:** Another Splitwise attempt (`<!--Splitwise Here we go again-->`).

**Problem statement:** `10-splitwise/readme.md`

**Solution:** `10-splitwise/cmd/main.go`

## 11. Delivery cost dashboard

**Question:** Live accounting dashboard of total delivery cost.

- `AddDriver(driverId, usdHourlyRatePerDelivery)`
- `RecordDelivery(driverId, startTime, endTime)` — 1s precision, max 3 hours
- `GetTotalCost() -> string` — aggregated cost to 2 decimal places

Follow-up: `PayUpTo(payTime)` and `GetTotalCostUnpaid()`.

**Problem statement:** `11-delivery-cost-dashboard/readme.md`

**Solution:** `11-delivery-cost-dashboard/cmd/main.go`

## 12. Splitwise

**Question:** Splitwise (later rewrite).

**Problem statement:** `12-splitwise/readme.md`

**Solution:** `12-splitwise/cmd/main.go`

## 13. Library Book Management System

**Question:** Backend to manage books, users, borrow/return, and availability. Max 3 concurrent borrows per user; no two copies of the same book; REST API.

**Problem statement:** `13-library-management/readme.md`

**Design docs:** `13-library-management/docs/superpowers/`

**Solution:** `13-library-management/internal/` (HTTP server, library service, memory + postgres repos) and `cmd/main.go`

## 14. Fizz Buzz

**Question:** Fizz Buzz.

**Problem statement:** `14-fizzbuzz/readme.md`

**Solution:** `14-fizzbuzz/cmd/main.go`

## 15. Odd / even concurrent printing

**Question:** Print odd and even numbers concurrently with correct ordering (goroutines + channels).

**Solution:** `15-odd-even/cmd/main.go` and `15-odd-even/main.go`  
(readme on this branch still says Fizz Buzz from practice 14.)

## 16. Splitwise

**Question:** Splitwise (equal / exact).

**Problem statement:** `16-splitwise/readme.md`

**Solution:** `16-splitwise/cmd/main.go`

## 17. Eligibility Rule Engine

**Question:** In-memory engine that decides feature eligibility from nested AND / OR / NOT rules over user attributes. Comparators: `=`, `!=`, `>`, `<`, `>=`, `<=`. Types: string, bool, number, version strings. Programmatic rule construction (no string parser). Handle missing attributes explicitly.

**Problem statement:** `17-eligibility-rule-engine/readme.md`

**Solution:** `17-eligibility-rule-engine/cmd/main.go`
