package clients

import (
	models "cf-cloud-connector/clients/models"
	"cf-cloud-connector/log"
	"encoding/json"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

// GetEnvironment get Cloud Foundry application environment
func GetEnvironment(cliConnection plugin.CliConnection, appGUID string) (*models.CFEnvironmentResponse, error) {
	var responseObject models.CFEnvironmentResponse
	var responseStrings []string
	var err error
	var url string

	url = "/v3/apps/" + appGUID + "/env"

	log.Tracef("Making request to: %s\n", url)
	responseStrings, err = cliConnection.CliCommandWithoutTerminalOutput("curl", url)
	if err != nil {
		return nil, err
	}

	body := []byte(strings.Join(responseStrings, ""))
	log.Trace(log.Response{Body: body})

	err = json.Unmarshal(body, &responseObject)
	if err != nil {
		return nil, err
	}

	return &responseObject, nil
}
