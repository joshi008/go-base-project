package incident

import (
	DBManager "go-base-project/db"
	"go-base-project/models"
	incidentComms "go-base-project/services/comms"
)

type IncidentService struct {
	db *DBManager.DB
}

func NewIncidentService(db *DBManager.DB) *IncidentService {
	return &IncidentService{
		db: db,
	}
}

func (i IncidentService) PostAlertHandling(alert *models.Alert) *models.Incident {
	newIncidentId := len(i.db.Incidents) + 1

	incident := models.Incident{
		ID:                 string(newIncidentId),
		ServiceId:          alert.ServiceId,
		Severity: alert.Severity,
		Status:             models.TRIGGERED,
		NotificationStatus: models.QUEUED,
	}

	i.db.Alerts[alert.ID] = alert
	i.db.Incidents[string(newIncidentId)] = &incident

	SendIncidentComms(incident)

	return &incident
}

func (i IncidentService) SendIncidentComms(incident models.Incident) {
	// SendComms
	switch(incident.Severity) {
		case models.CRITICAL:
			criticalStrategy := &incidentComms.CriticalStrat{}
			ok, err := criticalStrategy.SendComms(incident)
		case models.WARNING:
			generalStrat := &incidentComms.GeneralStrat{}
			ok, err := generalStrat.SendComms(incident)
		case models.INFO:
			generalStrat := &incidentComms.GeneralStrat{}
			ok, err := generalStrat.SendComms(incident)
	}

	if err!=nil {
		switch(err)
		newIncident := &incident
		newIncident.NotificationStatus
	}

	SendComms()
}
