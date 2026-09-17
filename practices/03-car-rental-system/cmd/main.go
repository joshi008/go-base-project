package main

import (
	"fmt"
	"go-base-project/services"
	"time"
)

func main() {
	fmt.Println("Hello and welcome to our system!")

	system := services.NewSystem()

	u1 := system.AddUser("ABC")
	u1.PrintUser()
	u2 := system.AddUser("BCD")
	u2.PrintUser()
	u3 := system.AddUser("CDF")
	u3.PrintUser()

	c1 := system.AddVehicle("Car1", 20)
	c1.PrintVehicle()
	c2 := system.AddVehicle("Car2", 25)
	c2.PrintVehicle()
	c3 := system.AddVehicle("Car3", 30)
	c3.PrintVehicle()

	b1 := system.AddBooking(u1.Id, c1.Id, time.Now(), time.Now().AddDate(0, 0, 7))
	b1.PrintBooking()
	b2 := system.AddBooking(u2.Id, c2.Id, time.Now(), time.Now().AddDate(0, 0, 7))
	b2.PrintBooking()
	b3 := system.AddBooking(u3.Id, c3.Id, time.Now(), time.Now().AddDate(0, 0, 7))
	b3.PrintBooking()
	cost := system.ReturnBooking(b3.Id, time.Now())
	fmt.Println("Cost is: ", cost)

	b4 := system.AddBooking(u3.Id, c3.Id, time.Now(), time.Now().AddDate(0, 0, 7))
	b4.PrintBooking()

}
