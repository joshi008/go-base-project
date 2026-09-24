package models

// Alert
// - serviceId
// - severity
// - title
// - dedupeKey

type SEVERITY int

const (
	CRITICAL SEVERITY = iota
	INFO
	WARNING
)

type SERVICES string

const (
	PAYMENTS SERVICES = "svc-payments"
	CHECKOUT SERVICES = "svc-checkout"
	SEARCH   SERVICES = "svc-search"
)

type Alert struct {
	ID        string
	ServiceId string
	Severity  SEVERITY
	Title     string
	DedupeKey string
}

type INCIDENT_STATUS int

const (
	TRIGGERED INCIDENT_STATUS = iota
	NOTIFIED
)

type NOTIFICATION_STATUS int

const (
	SENT NOTIFICATION_STATUS = iota
	FAILED
	SKIPPED
	QUEUED
)

type Incident struct {
	ID                 string
	ServiceId          string
	Severity           SEVERITY
	Status             INCIDENT_STATUS
	NotificationStatus NOTIFICATION_STATUS
}
