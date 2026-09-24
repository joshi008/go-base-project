package DBManager

import "go-base-project/models"

type DB struct {
	Incidents map[string]*models.Incident
	Alerts    map[string]*models.Alert
}

func NewDB() *DB {
	return &DB{
		Incidents: make(map[string]*models.Incident),
		Alerts:    make(map[string]*models.Alert),
	}
}
