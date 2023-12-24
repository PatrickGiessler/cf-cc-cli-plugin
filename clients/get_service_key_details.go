package clients

import (
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	client := &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		return serviceKey.Credentials, err
	}
	defer response.Body.Close()

	// Read response body
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return serviceKey.Credentials, err
	}

	// Failed to get service key details
	if response.StatusCode != 200 {
		return serviceKey.Credentials, fmt.Errorf("Failed to get service key details: [%d] %+v", response.StatusCode, body)
	}

	// Unmarshal JSON
	err = json.Unmarshal(body, &serviceKey)
	if err != nil {
		return serviceKey.Credentials, err
	}

	return serviceKey.Credentials, nil
}