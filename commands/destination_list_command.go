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

	ui.Say("Getting list of HTML5 applications%s in org %s / space %s as %s...",
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

	/*	for _, plan := range html5Context.HTML5AppsRepoServicePlans {
			if plan.Name == "app-host" {
				appHostServicePlan = &plan
				break
			}
		}
		if appHostServicePlan == nil {
			ui.Failed("Could not find app-host service plan")
			return Failure
		}

		var appHostServiceInstances []models.CFServiceInstance
		if appHostGUID == nil {
			// Get list of service instances of app-host plan
			log.Tracef("Getting service instances of %s service app-host plan (%+v)\n", html5Context.ServiceName, appHostServicePlan)
			appHostServiceInstances, err = clients.GetServiceInstances(c.CliConnection, context.SpaceID, []models.CFServicePlan{*appHostServicePlan})
			if err != nil {
				ui.Failed("Could not get service instances for app-host plan: %+v", err)
				return Failure
			}
		} else {
			// Use service instance with provided app-host-id
			appHostServiceInstances = []models.CFServiceInstance{
				{
					GUID: *appHostGUID,
					Name: "-",
					LastOperation: models.CFLastOperation{
						State:       "-",
						Type:        "-",
						Description: "-",
						UpdatedAt:   "-",
						CreatedAt:   "-",
					},
					UpdatedAt: "-",
				},
			}
		}

		// Get list of applications for each app-host service instance
		var data Model
		data.Services = make([]Service, 0)
		for _, serviceInstance := range appHostServiceInstances {
			log.Tracef("Getting list of applications for app-host plan (%+v)\n", serviceInstance)
			applications, err := clients.ListApplicationsForAppHost(*html5Context.HTML5AppRuntimeServiceInstanceKeys[len(html5Context.HTML5AppRuntimeServiceInstanceKeys)-1].Credentials.URI,
				html5Context.HTML5AppRuntimeServiceInstanceKeyToken, serviceInstance.GUID)
			if err != nil {
				ui.Failed("Could not get list of applications for app-host instance %s: %+v", serviceInstance.Name, err)
				return Failure
			}
			apps := make([]App, 0)
			for _, app := range applications {
				apps = append(apps, App{Name: app.ApplicationName, Version: app.ApplicationVersion, Changed: app.ChangedOn, Public: app.IsPublic})
			}
			data.Services = append(data.Services, Service{Name: serviceInstance.Name, GUID: serviceInstance.GUID, UpdatedAt: serviceInstance.UpdatedAt, Apps: apps})
		}

		// Clean-up HTML5 context

		ui.Ok()
		ui.Say("")

		// Display information about HTML5 applications
		table := ui.Table([]string{"name", "version", "app-host-id", "service instance", "visibility", "last changed"})
		for _, service := range data.Services {
			if len(service.Apps) == 0 {
				table.Add("-", "-", service.GUID, service.Name, "-", service.UpdatedAt)
			} else {
				for _, app := range service.Apps {
					table.Add(app.Name, app.Version, service.GUID, service.Name, (map[bool]string{true: "public", false: "private"})[app.Public], app.Changed)
				}
			}
		}
		table.Print()*/

	return Success
}
