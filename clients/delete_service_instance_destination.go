package clients

import (
	"bytes"
	"cf-cloud-connector/log"
	"fmt"
	"io"
	"net/http"
)

// DeleteServiceInstanceDestination delete destination service instance level destination
func DeleteServiceInstanceDestination(serviceURL string, accessToken string, destinationName string) error {
	var err error
	var request *http.Request
	var response *http.Response
	var destinationsURL string
	var payload = []byte{}
	var body []byte

	destinationsURL = serviceURL + "/destination-configuration/v1/instanceDestinations/" + destinationName
	log.Tracef("Making request to: %s\n", destinationsURL)
	request, err = http.NewRequest("DELETE", destinationsURL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+accessToken)

	client, err := GetDefaultClient()
	if err != nil {
		return err
	}
	response, err = client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	log.Trace(log.Response{Head: response})
	if response.StatusCode > 201 {
		body, err = io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("could not delete destination: [%s] %+v", response.Status, body)
		}
	}

	return nil
}
