package commands

import (
	"cf-cloud-connector/clients/models"
	"cf-cloud-connector/log"
	"cf-cloud-connector/ui"

	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/plugin"
)

// ListCommand prints the list of HTML5 applications
// deployed using multiple instances of html5-apps-repo
// service app-host plan
type DestinationListCommand struct {
	DestinationCommand
}

// GetPluginCommand returns the plugin command details
func (c *DestinationCommand) GetPluginCommand() plugin.Command {
	return plugin.Command{
		Name:     "cloud-connector-list",
		HelpText: "Display list of HTML5 applications or file paths of specified application",
		UsageDetails: plugin.Usage{
			Usage: "cf html5-list [APP_NAME] [APP_VERSION] [APP_HOST_ID|-n APP_HOST_NAME] [-d|-di DESTINATION_SERVICE_INSTANCE_NAME|-a CF_APP_NAME [-rt RUNTIME] [-u]]",
			Options: map[string]string{
				"APP_NAME":                          "Application name, which file paths should be listed. If not provided, list of applications will be printed",
				"APP_VERSION":                       "Application version, which file paths should be listed. If not provided, current active version will be used",
				"APP_HOST_ID":                       "GUID of html5-apps-repo app-host service instance that contains application with specified name and version",
				"APP_HOST_NAME":                     "Name of html5-apps-repo app-host service instance that contains application with specified name and version",
				"DESTINATION_SERVICE_INSTANCE_NAME": "Name of destination service intance",
				"-destination, -d":                  "List HTML5 applications exposed via subaccount destinations with sap.cloud.service and html5-apps-repo.app_host_id properties",
				"-destination-instance, -di":        "List HTML5 applications exposed via service instance destinations with sap.cloud.service and html5-apps-repo.app_host_id properties",
				"-name, -n":                         "Use html5-apps-repo app-host service instance name instead of APP_HOST_ID",
				"-app, -a":                          "Cloud Foundry application name, which is bound to services that expose UI via html5-apps-repo",
				"-runtime, -rt":                     "Runtime service for which conventional URLs of applications will be shown. Default value is 'cpp'",
				"-url, -u":                          "Show conventional URLs of applications, when accessed via Cloud Foundry application specified with --app flag or when --destination or --destination-instance flag is used",
			},
		},
	}
}

// Execute executes plugin command
func (c *DestinationCommand) Execute(args []string) ExecutionStatus {
	log.Tracef("Executing command '%s': args: '%v'\n", c.Name, args)

	// List apps in the space
	if len(args) == 0 {
		return c.ListApps(nil)
	}

	ui.Failed("Too many arguments. See [cf html5-list --help] for more details")
	return Failure
}

// ListApps get list of applications for given app-host-id or current space
func (c *DestinationCommand) ListApps(appHostGUID *string) ExecutionStatus {
	// Get context
	log.Tracef("Getting context (org/space/username)\n")
	context, err := c.GetContext()
	if err != nil {
		ui.Failed("Could not get org and space: %s", err.Error())
		return Failure
	}

	appHostMessage := ""
	if appHostGUID != nil {
		appHostMessage = " with app-host-id " + terminal.EntityNameColor(*appHostGUID)
	}

	ui.Say("Getting list of Destinations applications%s in org %s / space %s as %s...",
		appHostMessage,
		terminal.EntityNameColor(context.Org),
		terminal.EntityNameColor(context.Space),
		terminal.EntityNameColor(context.Username))

	// Get HTML5 context
	destinationContext, err := c.GetDestinationContext(context)
	if err != nil {
		ui.Failed(err.Error())
		return Failure
	}
	_ = destinationContext
	// Find app-host service plan
	log.Tracef("Looking for app-host service plan\n")
	var appHostServicePlan *models.CFServicePlan
	_ = appHostServicePlan

	ui.Ok()
	ui.Say("")
	var onPremiseApps []models.DestinationApp
	for _, service := range destinationContext {
		if service.ProxyType == "OnPremise" {
			onPremiseApps = append(onPremiseApps, service)
		}
	}

	// Display information about HTML5 applications
	table := ui.Table([]string{"name", "description", "type", "URL"})
	for _, service := range onPremiseApps {
		table.Add(service.Name, service.Description, service.ProxyType, service.URL)
	}
	table.Print()

	return Success
}
