package clients

import (
	models "cf-cloud-connector/clients/models"
	"cf-cloud-connector/log"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cloudfoundry/cli/plugin"
)

func GetServiceInstanceByUrl(cliConnection plugin.CliConnection, url string) (models.CFServiceInstance, error) {
	var accessToken string
	var request *http.Request
	var response *http.Response
	var err error
	var body []byte
	var serviceInstance models.CFServiceInstance

	// Setup request
	accessToken, err = cliConnection.AccessToken()
	if err != nil {
		return serviceInstance, err
	}

	log.Tracef("Making request to: %s\n", url)
	request, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return serviceInstance, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", accessToken)

	// Make request
	client, err := GetDefaultClient()
	if err != nil {
		return serviceInstance, err
	}
	response, err = client.Do(request)
	if err != nil {
		return serviceInstance, err
	}
	defer response.Body.Close()

	// Read response body
	body, err = io.ReadAll(response.Body)
	log.Trace(log.Response{Head: response, Body: body})
	if err != nil {
		return serviceInstance, err
	}

	// Failed to get service instance
	if response.StatusCode != 200 {
		return serviceInstance, fmt.Errorf("failed to get service instance by URL '%s': [%d] %+v", url, response.StatusCode, body)
	}

	// Unmarshal JSON
	err = json.Unmarshal(body, &serviceInstance)
	if err != nil {
		return serviceInstance, err
	}

	return serviceInstance, nil
}
