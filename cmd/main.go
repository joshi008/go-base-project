package main

import (
	"fmt"
	DBManager "go-base-project/db"
	"go-base-project/services/incident"
)

func main() {
	fmt.Println("Starting of the program")

	db := DBManager.NewDB()

	incidentService := incident.NewIncidentService(db)

}
