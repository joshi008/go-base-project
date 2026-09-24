package incidentComms

import (
	"go-base-project/models"

	"github.com/chromedp/cdproto/fetch"
	"google.golang.org/genproto/googleapis/api"
)

type IncidentStrategy interface {
	SendComms(models.Incident) (bool, error)
}

type CriticalStrat struct {}


// Base URL is handed to you at the start of the call.
// POST {EXOTEL_BASE_URL}/v1/calls
// {
//  "to":      "+919900112233",
//  "message": "CRITICAL: Payment gateway timeout on svc-payments"
// }


func (c CriticalStrat) SendComms(incident models.Incident) (bool, error) {
	IncidentToPhoneMap := map[string]string{
		 "svc-payments": "+919900112233",
		 "svc-checkout": "+919900445566",
		 "svc-search":   "+919900778899",
	}

	proto := api.fetch("{EXOTEL_BASE_URL}/v1/calls", {
		"to": IncIncidentToPhoneMap[incident.ServiceId],
		"message": string(incident.Status + ": " incident.ServiceId)
		})

	res, err := json.Unmarshal(proto)

	if err!= nil {
		return false, err
	}

	httpErr := HTTPHandler(res);

	if httpErr != nil {
		return false, err
	}

	return true, nil
}

func HTTPExotelHandler(res HTML.response) err {
	// HTTP CALLS /call
	// Translation for http to golang struct errors
	// res.Header.Code
	// res.BODY
	// err {}
	// HTTP CALLS WITH 200 Code
	// HTTP CALLS with 4XX Count increase
	// Prometheus.PublishMetrics
}


type GeneralStrat struct {}
func (c GeneralStrat) SendComms(incident models.Incident) (bool, error) {
	return true, nil
}
