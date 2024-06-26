package clients

import (
	models "cf-cloud-connector/clients/models"
	"cf-cloud-connector/log"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

// GetServiceInstanceByName get Cloud Foundry service instance by name
func GetServiceInstanceByName(cliConnection plugin.CliConnection, spaceGUID string, serviceInstanceName string) (models.CFServiceInstance, error) {
	var serviceInstances []models.CFServiceInstance
	var responseObject models.CFResponse
	var responseStrings []string
	var err error
	var nextURL *string

	serviceInstances = make([]models.CFServiceInstance, 0)
	firstURL := "/v3/service_instances?names=" + serviceInstanceName + "&space_guids=" + spaceGUID
	nextURL = &firstURL

	for nextURL != nil {
		log.Tracef("Making request to: %s\n", *nextURL)
		responseStrings, err = cliConnection.CliCommandWithoutTerminalOutput("curl", *nextURL)
		if err != nil {
			return models.CFServiceInstance{}, err
		}

		responseObject = models.CFResponse{}
		body := []byte(strings.Join(responseStrings, ""))
		log.Trace(log.Response{Body: body})
		err = json.Unmarshal(body, &responseObject)
		if err != nil {
			return models.CFServiceInstance{}, err
		}

		for _, serviceInstance := range responseObject.Resources {
			serviceInstances = append(serviceInstances, models.CFServiceInstance{
				Name:          serviceInstance.Name,
				GUID:          serviceInstance.GUID,
				UpdatedAt:     serviceInstance.UpdatedAt,
				LastOperation: serviceInstance.LastOperation,
			})
		}
		if responseObject.Pagination.Next.Href != nil && *nextURL == *responseObject.Pagination.Next.Href {
			log.Tracef("Unexpected value of the next page URL (equal to previous): %s\n", *nextURL)
			break
		}
		nextURL = responseObject.Pagination.Next.Href
	}

	if len(serviceInstances) == 0 {
		return models.CFServiceInstance{}, fmt.Errorf("service instance with name '%s' not found", serviceInstanceName)
	}

	return serviceInstances[0], nil
}
