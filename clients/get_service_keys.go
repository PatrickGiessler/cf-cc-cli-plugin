package clients

import (
	models "cf-cloud-connector/clients/models"
	"cf-cloud-connector/log"
	"encoding/json"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

// GetServiceKeys get Cloud Foundry service keys
func GetServiceKeys(cliConnection plugin.CliConnection, serviceInstanceGUID string) ([]models.CFServiceKey, error) {
	var serviceKeys []models.CFServiceKey
	var responseObject models.CFResponse
	var serviceKeyCredentials models.CFCredentials
	var responseStrings []string
	var err error
	var nextURL *string
	var pathStart int
	var pathSlice string

	serviceKeys = make([]models.CFServiceKey, 0)
	firstURL := "/v3/service_credential_bindings?service_instance_guids=" + serviceInstanceGUID
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

		for _, serviceKey := range responseObject.Resources {
			serviceKeys = append(serviceKeys, models.CFServiceKey{
				Name: serviceKey.Name,
				GUID: serviceKey.GUID,
			})
		}
		if responseObject.Pagination.Next.Href != nil && *nextURL == *responseObject.Pagination.Next.Href {
			log.Tracef("Unexpected value of the next page URL (equal to previous): %s\n", *nextURL)
			break
		}
		nextURL = responseObject.Pagination.Next.Href
		if nextURL != nil {
			pathStart = strings.Index(*nextURL, "/v3/service_credential_bindings")
			if pathStart > 0 {
				pathSlice = (*nextURL)[pathStart:]
				nextURL = &pathSlice
			}
		}
	}

	for idx, serviceKey := range serviceKeys {
		serviceKeyCredentials, err = GetServiceKeyDetails(cliConnection, serviceKey.GUID)
		if err != nil {
			return serviceKeys, err
		}
		(&serviceKeys[idx]).Credentials = serviceKeyCredentials
	}

	return serviceKeys, nil
}
