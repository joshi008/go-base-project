You are building the core of a small incident-alerting service, similar in spirit to PagerDuty. Monitored services emit alerts. Your service ingests each alert, records it as an incident, and — for the most severe alerts — places an automated phone call to the on-call person through an external telephony provider (we use a mocked Exotel-style API).
Build a small HTTP service around this.
1.1 Requirements
R1  —  POST /alert accepts an alert, creates an incident, applies the notification rule below, and returns the created incident.
R2  —  Severity decides the notification. The threshold is CRITICAL:
CRITICAL → place a phone call via the Exotel mock, to the on-call number for that service, with the alert message as the spoken text. Record the outcome (SENT or FAILED) on the incident.
Below CRITICAL (WARNING, INFO) → no phone call. Record the incident with notificationStatus = SKIPPED.
R3  —  Treat this as production code.  Log at meaningful boundaries and at appropriate levels, keeping sensitive data out of the logs. Handle errors where they can genuinely occur — and make sure an incident is never lost because something outside your service misbehaved.
1.2 Seed data — hardcode this
A static map of serviceId → on-call phone number. You do not need an API to manage it.
{
 "svc-payments": "+919900112233",
 "svc-checkout": "+919900445566",
 "svc-search":   "+919900778899"
}
1.3 Request shape
POST /alert
{
 "serviceId":  "svc-payments",
 "severity":   "CRITICAL",              // INFO | WARNING | CRITICAL
 "title":      "Payment gateway timeout",
 "dedupKey":   "payments-latency-001"
}
 
1.4 Incident shape — what you persist and return
{
 "incidentId":         "inc-1029",
 "serviceId":          "svc-payments",
"severity":          "CRITICAL",
 "status":             "NOTIFIED",     // TRIGGERED | NOTIFIED
 "notificationStatus": "SENT"          // SENT | FAILED | SKIPPED
}
 
1.5 The Exotel mock — external phone-call API
Base URL is handed to you at the start of the call.
POST {EXOTEL_BASE_URL}/v1/calls
{
 "to":      "+919900112233",
 "message": "CRITICAL: Payment gateway timeout on svc-payments"
}

Responses
200  { "callSid": "CA-7781", "status": "queued" }
 
4xx  { "error": "invalid_number", "message": "..." }
      The request itself is bad. Retrying will not help.
 
5xx  { "error": "provider_unavailable", "message": "..." }
      Transient.


<!--Non Functional Requirements-->
- Consistency > Availability
- Scalable Code

<!--Functional-->
- POST API call for /alerts ingest into in-memory DB 
- Severity based API calls to exotel
- Logs in the codebase
- API handling from exotel

<!--Entities-->
Alert
- serviceId
- severity
- title
- dedupeKey

Incident
- incidentId
- serviceId
-
