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

func GetServiceKeyDetails(cliConnection plugin.CliConnection, serviceKeyGUID string) (models.CFCredentials, error) {
	var apiEndpoint string
	var accessToken string
	var request *http.Request
	var response *http.Response
	var err error
	var url string
	var body []byte
	var serviceKey models.CFServiceKey

	// Setup request
	apiEndpoint, err = cliConnection.ApiEndpoint()
	if err != nil {
		return serviceKey.Credentials, err
	}
	accessToken, err = cliConnection.AccessToken()
	if err != nil {
		return serviceKey.Credentials, err
	}
	url = apiEndpoint + "/v3/service_credential_bindings/" + serviceKeyGUID + "/details"

	log.Tracef("Making request to: %s\n", url)
	request, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return serviceKey.Credentials, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", accessToken)

	// Make request
	client, err := GetDefaultClient()
	if err != nil {
		return serviceKey.Credentials, err
	}
	response, err = client.Do(request)
	if err != nil {
		return serviceKey.Credentials, err
	}
	defer response.Body.Close()

	// Read response body
	body, err = io.ReadAll(response.Body)
	log.Trace(log.Response{Head: response, Body: body})
	if err != nil {
		return serviceKey.Credentials, err
	}

	// Failed to get service key details
	if response.StatusCode != 200 {
		return serviceKey.Credentials, fmt.Errorf("failed to get service key details: [%d] %+v", response.StatusCode, body)
	}

	// Unmarshal JSON
	err = json.Unmarshal(body, &serviceKey)
	if err != nil {
		return serviceKey.Credentials, err
	}

	return serviceKey.Credentials, nil
}
