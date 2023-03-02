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

func GetServiceKeyByUrl(cliConnection plugin.CliConnection, url string) (models.CFServiceKey, error) {
	var accessToken string
	var request *http.Request
	var response *http.Response
	var err error
	var body []byte
	var serviceKey models.CFServiceKey

	// Setup request
	accessToken, err = cliConnection.AccessToken()
	if err != nil {
		return serviceKey, err
	}

	log.Tracef("Making request to: %s\n", url)
	request, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return serviceKey, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", accessToken)

	// Make request
	client := &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		return serviceKey, err
	}
	defer response.Body.Close()

	// Read response body
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return serviceKey, err
	}

	// Failed to get service key
	if response.StatusCode != 200 {
		return serviceKey, fmt.Errorf("Failed to get service key by URL '%s': [%d] %+v", url, response.StatusCode, body)
	}

	// Unmarshal JSON
	err = json.Unmarshal(body, &serviceKey)
	if err != nil {
		return serviceKey, err
	}

	return serviceKey, nil
}
