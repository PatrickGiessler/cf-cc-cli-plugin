package clients

import (
	models "cf-cloud-connector/clients/models"
	"cf-cloud-connector/log"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

// GetApplication get Cloud Foundry application
func GetApplication(cliConnection plugin.CliConnection, spaceGUID string, appName string) (*models.CFApplication, error) {
	var application *models.CFApplication
	var responseObject models.CFResponse
	var responseStrings []string
	var err error
	var url string

	url = "/v3/apps?names=" + appName

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
	if len(responseObject.Resources) > 0 {
		log.Tracef("Number of applications with name %s: %d\n", appName, len(responseObject.Resources))
	} else {
		return nil, fmt.Errorf("Application with name %s does not exist in current organization and space", appName)
	}
	application = &models.CFApplication{GUID: responseObject.Resources[0].GUID, Name: responseObject.Resources[0].Name}

	return application, nil
}
