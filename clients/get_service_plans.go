package clients

import (
	"cf-cloud-connector/cache"
	models "cf-cloud-connector/clients/models"
	"cf-cloud-connector/log"
	"encoding/json"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

// GetServicePlans get Cloud Foundry services
func GetServicePlans(cliConnection plugin.CliConnection, serviceGUID string) ([]models.CFServicePlan, error) {
	var servicePlans []models.CFServicePlan
	var responseObject models.CFResponse
	var responseStrings []string
	var err error
	var nextURL *string
	var pathStart int
	var pathSlice string

	if cachedServicePlans, ok := cache.Get("GetServicePlans:" + serviceGUID); ok {
		log.Tracef("Returning cached list of service plans\n")
		servicePlans = cachedServicePlans.([]models.CFServicePlan)
		return servicePlans, nil
	}

	servicePlans = make([]models.CFServicePlan, 0)
	firstURL := "/v3/service_plans?service_offering_guids=" + serviceGUID
	nextURL = &firstURL

	for nextURL != nil {
		log.Tracef("Making request to: %s\n", *nextURL)
		responseStrings, err = cliConnection.CliCommandWithoutTerminalOutput("curl", *nextURL)
		if err != nil {
			return nil, err
		}

		responseObject = models.CFResponse{}
		body := []byte(strings.Join(responseStrings, ""))
		log.Trace(log.Response{Body: body})
		err = json.Unmarshal(body, &responseObject)
		if err != nil {
			return nil, err
		}

		for _, servicePlan := range responseObject.Resources {
			servicePlans = append(servicePlans, models.CFServicePlan{
				Name: servicePlan.Name,
				GUID: servicePlan.GUID,
			})
		}
		if responseObject.Pagination.Next.Href != nil && *nextURL == *responseObject.Pagination.Next.Href {
			log.Tracef("Unexpected value of the next page URL (equal to previous): %s\n", *nextURL)
			break
		}
		nextURL = responseObject.Pagination.Next.Href
		if nextURL != nil {
			pathStart = strings.Index(*nextURL, "/v3/service_plans")
			if pathStart > 0 {
				pathSlice = (*nextURL)[pathStart:]
				nextURL = &pathSlice
			}
		}
	}

	cache.Set("GetServicePlans:"+serviceGUID, servicePlans)

	return servicePlans, nil
}
