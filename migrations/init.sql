-- type Branches struct {
-- 	ID        int       `json:"id"`
-- 	Name 	string    `json:"name"`
-- }

CREATE TABLE IF NOT EXISTS branches (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);

-- type Cars struct {
-- 	ID        int       `json:"id"`
-- 	CategoryId int       `json:"category_id"`
-- 	BranchID int       `json:"branch_id"`
-- }
CREATE TABLE IF NOT EXISTS cars (
    id SERIAL PRIMARY KEY,
    category_id INT NOT NULL,
    branch_id INT NOT NULL,
    FOREIGN KEY (branch_id) REFERENCES branches(id)
);

-- type Pricing struct {
-- 	ID        int       `json:"id"`
-- 	Price 	float64    `json:"price"`
-- 	BranchID int       `json:"branch_id"`
-- 	CategoryId int       `json:"category_id"`
-- }
CREATE TABLE IF NOT EXISTS pricing (
    id SERIAL PRIMARY KEY,
    price FLOAT NOT NULL,
    branch_id INT NOT NULL,
    category_id INT NOT NULL,
    FOREIGN KEY (branch_id) REFERENCES branches(id)
);

-- type SlotBooking struct {
-- 	ID        int       `json:"id"`
-- 	CarID 	int       `json:"car_id"`
-- 	StartTime time.Time `json:"start_time"`
-- 	EndTime   time.Time `json:"end_time"`
-- 	Price    float64   `json:"price"`
-- }

CREATE TABLE IF NOT EXISTS slot_booking (
    id SERIAL PRIMARY KEY,
    car_id INT NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    price FLOAT NOT NULL,
    FOREIGN KEY (car_id) REFERENCES cars(id)
);