package commands

import (
	"cf-cloud-connector/cache"
	clients "cf-cloud-connector/clients"
	"cf-cloud-connector/clients/models"
	"cf-cloud-connector/log"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

// HTML5Command base struct for HTML5 repository operations
type DestinationCommand struct {
	BaseCommand
}
type DestinationContext struct {
	// Pointer to destination service
	DestinationServices []models.CFService
	// Pointer to 'lite' plan of destination service
	DestinationServicePlan *models.CFServicePlan
	// List of destination service instances
	DestinationServiceInstances []models.CFServiceInstance
	// Pointer to destination service instance created during context initialization
	DestinationServiceInstance *models.CFServiceInstance
	// Pointer to destination service key created during context initialization
	DestinationServiceInstanceKey *models.CFServiceKey
	// Access token of destination service key
	DestinationServiceInstanceKeyToken string
}

// Initialize initializes the command with the specified name and CLI connection
func (c *DestinationCommand) Initialize(name string, cliConnection plugin.CliConnection) (err error) {
	log.Tracef("Initializing command '%s'\n", name)
	err = c.InitializeBase(name, cliConnection)
	if err != nil {
		return
	}
	isInsecure, _ := cliConnection.IsSSLDisabled()

	// TLS configuration
	clients.SetInsecure(isInsecure)
	customCAPath := os.Getenv("SSL_CERT_FILE")
	if customCAPath == "" {
		customCAPath = os.Getenv("SSL_CERT_DIR")
		if customCAPath != "" {
			customCAPath = filepath.Join(customCAPath, "server.crt")
		}
	}
	if customCAPath != "" {
		if _, err := os.Stat(customCAPath); err != nil {
			log.Tracef("Failed to read file with additional root CAs: %s\n", err.Error())
			return fmt.Errorf("certificate file %q is not accessible. Please check 'SSL_CERT_FILE' or 'SSL_CERT_DIR' environment variable is pointing to existing file or directory", customCAPath)
		}
	}
	clients.SetCustomCAPath(customCAPath)

	// Cache
	return nil
}

// Dispose disposes command and saves cache if needed
func (c *DestinationCommand) Dispose(name string) {
	log.Tracef("Disposing command '%s'\n", name)
	c.DisposeBase(name)

}

// GetDestinationContext get destination context
func (c *DestinationCommand) GetDestinationContext(context Context) (models.DestinationAppResponse, error) {

	// Context to return
	var destinationApps = models.DestinationAppResponse{}

	// Get all services
	log.Tracef("Getting list of services\n")
	services, err := clients.GetServices(c.CliConnection)
	if err != nil {
		return destinationApps, errors.New("Could not get services: " + err.Error())
	}

	// Find destination service
	log.Tracef("Looking for 'destination' service\n")
	var destinationServices []models.CFService
	for _, service := range services {
		if service.Name == "destination" {
			destinationServices = append(destinationServices, service)

		}
	}
	if destinationServices == nil {
		return destinationApps, fmt.Errorf("destination service is not in the list of available services." +
			" Make sure your subaccount has entitlement to use it")
	}

	for _, destinationService := range destinationServices {
		log.Tracef("Getting service plans for 'destination' service (GUID: %s)\n", destinationService.GUID)
		var liteServicePlan *models.CFServicePlan
		destinationServicePlans, err := clients.GetServicePlans(c.CliConnection, destinationService.GUID)
		if err != nil {
			return destinationApps, fmt.Errorf("could not get service plans: %s", err.Error())
		}
		for _, servicePlan := range destinationServicePlans {
			if servicePlan.Name == "lite" {
				liteServicePlan = &servicePlan
				break
			}
		}
		if liteServicePlan == nil {
			return destinationApps, fmt.Errorf("destination service does not have a 'lite' plan")
		}
		var destinationContext = DestinationContext{}
		destinationContext.DestinationServices = destinationServices
		destinationContext.DestinationServicePlan = liteServicePlan

		log.Tracef("Getting service instances of 'destination' service 'lite' plan (%+v)\n", liteServicePlan)
		var destinationServiceInstances []models.CFServiceInstance
		destinationServiceInstances, err = clients.GetServiceInstances(c.CliConnection, context.SpaceID, []models.CFServicePlan{*liteServicePlan})
		if err != nil {
			return destinationApps, fmt.Errorf("could not get service instances for 'lite' plan: %s", err.Error())
		}
		destinationContext.DestinationServiceInstances = destinationServiceInstances
		for _, instance := range destinationServiceInstances {
			// get service keys
			destinationServiceInstanceKeys, err := clients.GetServiceKeys(c.CliConnection, instance.GUID)
			if err != nil {
				return destinationApps, fmt.Errorf("could not get service keys of %s service instance: %s",
					instance.Name,
					err.Error())
			}
			if len(destinationServiceInstanceKeys) == 0 {
				log.Tracef("Creating service key for %s service instance\n", instance.Name)
				destinationServiceInstanceKey, err := clients.CreateServiceKey(c.CliConnection, instance.GUID, nil)
				if err != nil {
					return destinationApps, fmt.Errorf("could not create service key of %s service instance: %s",
						instance.Name,
						err.Error())
				}
				destinationServiceInstanceKeys = append(destinationServiceInstanceKeys, *destinationServiceInstanceKey)
			}
			if len(destinationServiceInstanceKeys) > 0 {
				log.Tracef("Found %d service keys for service %s, using service key with GUID=%s\n",
					len(destinationServiceInstanceKeys),
					instance.Name,
					destinationServiceInstanceKeys[len(destinationServiceInstanceKeys)-1].GUID)
				destinationServiceInstanceKeyToken, err := clients.GetToken(destinationServiceInstanceKeys[len(destinationServiceInstanceKeys)-1].Credentials)
				if err != nil {
					return destinationApps, fmt.Errorf("could not obtain access token: %s", err.Error())
				}
				//destinationContext.DestinationServiceInstanceKey = destinationServiceInstanceKeys[len(destinationServiceInstanceKeys)-1]
				destinationApps, err = clients.ListDestinationDetails(*destinationServiceInstanceKeys[len(destinationServiceInstanceKeys)-1].Credentials.URI,
					destinationServiceInstanceKeyToken, instance.GUID)
				break
			}

		}

	}
	return destinationApps, nil
	//log.Tracef("Destination services found: %+v\n", destinationContext.DestinationServices.)
	//destinationContext.DestinationService = destinationService

	// Get list of service instances of 'lite' plan

	// Sort destination service instance so that the requested instance to be the first one in the list.
	// If specific destinaton service instance name is required, but not found - return error
	/*if destinationInstanceName != "" {
		found := false
		for idx, instance := range destinationServiceInstances {
			if instance.Name == destinationInstanceName {
				tmp := destinationServiceInstances[0]
				destinationServiceInstances[0] = instance
				destinationServiceInstances[idx] = tmp
				found = true
				break
			}
		}
		if !found {
			return destinationContext, fmt.Errorf("Could not find service instance of 'destination' service 'lite' plan with name '%s'", destinationInstanceName)
		}
	}*/

	// Create instance of 'lite' plan if needed
	/*if len(destinationServiceInstances) == 0 {
		log.Tracef("Creating service instance of 'destination' service 'lite' plan\n")
		destinationServiceInstance, err := clients.CreateServiceInstance(c.CliConnection, context.SpaceID, *liteServicePlan, nil, "")
		if err != nil {
			return destinationContext, fmt.Errorf("Could not create service instance of 'destination' service 'lite' plan: %s", err.Error())
		}
		destinationServiceInstances = append(destinationServiceInstances, *destinationServiceInstance)
		destinationContext.DestinationServiceInstance = destinationServiceInstance
	} else {
		log.Tracef("Using service instance of 'destination' service 'lite' plan: %+v\n", destinationServiceInstances[0])
	}*/

	// TODO: chech if there is an existing service key and use it, if found

	// Create service key
	/*log.Tracef("Creating service key for 'destination' service 'lite' plan\n")
	destinationServiceInstanceKey, err := clients.CreateServiceKey(c.CliConnection, destinationServiceInstances[0].GUID, nil)
	if err != nil {
		return destinationContext, fmt.Errorf("Could not create service key of %s service instance: %s",
			destinationServiceInstances[0].Name,
			err.Error())
	}
	destinationContext.DestinationServiceInstanceKey = destinationServiceInstanceKey

	// Get destination service lite plan key access token
	log.Tracef("Getting token for service key %s\n", destinationServiceInstanceKey.Name)
	destinationServiceInstanceKeyToken, err := clients.GetToken(destinationServiceInstanceKey.Credentials)
	if err != nil {
		return destinationContext, fmt.Errorf("Could not obtain access token: %s", err.Error())
	}
	log.Tracef("Access token for service key %s: %s\n",
		destinationServiceInstanceKey.Name,
		log.Sensitive{Data: destinationServiceInstanceKeyToken})
	destinationContext.DestinationServiceInstanceKeyToken = destinationServiceInstanceKeyToken*/

}

// CleanDestinationContext clean destination context
func (c *DestinationCommand) CleanDestinationContext(destinationContext DestinationContext) error {
	var err error

	// Delete service key
	if destinationContext.DestinationServiceInstanceKey != nil {
		log.Tracef("Deleting service key %s\n", destinationContext.DestinationServiceInstanceKey.Name)
		err = clients.DeleteServiceKey(c.CliConnection, destinationContext.DestinationServiceInstanceKey.GUID, maxRetryCount)
		if err != nil {
			return errors.New("Could not delete service key" + destinationContext.DestinationServiceInstanceKey.Name + ": " + err.Error())
		}
		destinationContext.DestinationServiceInstanceKey = nil
	}

	// Delete service instance
	if destinationContext.DestinationServiceInstance != nil {
		log.Tracef("Deleting service instance %s\n", destinationContext.DestinationServiceInstance.Name)
		err = clients.DeleteServiceInstance(c.CliConnection, destinationContext.DestinationServiceInstance.GUID, maxRetryCount)
		if err != nil {
			return errors.New("Could not delete service instance of lite plan: " + err.Error())
		}
		log.Tracef("Service instance %s successfully deleted\n", destinationContext.DestinationServiceInstance.Name)
		destinationContext.DestinationServiceInstance = nil
	}

	return nil
}

// GetHTML5Context get HTML5 context
func (c *DestinationCommand) GetHTML5Context(context Context) (HTML5Context, error) {
	log.Tracef("Getting HTML5 context\n")

	// Try to load context from cache
	if html5ContextFromCache, ok := cache.Get("GetHTML5Context:" + context.OrgID + ":" + context.SpaceID); ok {
		log.Tracef("Returning cached HTML5 context\n")
		return html5ContextFromCache.(HTML5Context), nil
	}

	// Context to return
	html5Context := HTML5Context{}

	// Get name of html5-apps-repo service
	serviceName := os.Getenv("HTML5_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "destination"
	}
	html5Context.ServiceName = serviceName

	// Get list of services
	log.Tracef("Getting list of services\n")
	services, err := clients.GetServices(c.CliConnection)
	if err != nil {
		return html5Context, errors.New("Could not get services: " + err.Error())
	}
	html5Context.Services = services

	// Find html5-apps-repo service
	log.Tracef("Looking for '%s' service\n", serviceName)
	var html5AppsRepoService *models.CFService
	for _, service := range services {
		if service.Name == serviceName {
			html5AppsRepoService = &service
			break
		}
	}

	//hhhahhahaha
	var conService *models.CFService
	for _, service := range services {
		if service.Name == "destination" {
			conService = &service
		}
	}
	servicePlans_con, err := clients.GetServicePlans(c.CliConnection, conService.GUID)
	_ = servicePlans_con
	_ = err
	var appa, err2 = clients.GetServiceInstances(c.CliConnection, context.SpaceID, servicePlans_con)
	//log.Tracef(err.Error())
	_ = appa
	_ = err2

	var destService *models.CFService
	for _, service := range services {
		if service.Name == "connectivity" {
			destService = &service
			break
		}
	}
	_ = destService

	destService_con, err := clients.GetServicePlans(c.CliConnection, destService.GUID)
	_ = destService_con

	var appa2, err3 = clients.GetServiceInstances(c.CliConnection, context.SpaceID, destService_con)
	//log.Tracef(err.Error())
	_ = appa2
	_ = err3

	if err != nil {
		return html5Context, errors.New("Could not get service instances for app-runtime plan: " + err.Error())
	}

	if html5AppsRepoService == nil {
		return html5Context, errors.New(serviceName + " service is not in the list of available services")
	}
	html5Context.HTML5AppsRepoService = html5AppsRepoService

	// Get list of service plans
	log.Tracef("Getting service plans for '%s' service (GUID: %s)\n", serviceName, html5AppsRepoService.GUID)
	servicePlans, err := clients.GetServicePlans(c.CliConnection, html5AppsRepoService.GUID)
	if err != nil {
		return html5Context, errors.New("Could not get service plans: " + err.Error())
	}
	html5Context.HTML5AppsRepoServicePlans = servicePlans

	// Find app-runtime service plan
	log.Tracef("Looking for app-runtime service plan\n")
	var appRuntimeServicePlan *models.CFServicePlan
	for _, plan := range servicePlans {
		if plan.Name == "app-runtime" {
			appRuntimeServicePlan = &plan
			break
		}
	}
	if appRuntimeServicePlan == nil {
		return html5Context, errors.New("could not find app-runtime service plan")
	}
	html5Context.HTML5AppRuntimeServicePlan = appRuntimeServicePlan

	// Get list of service instances of app-runtime plan
	log.Tracef("Getting service instances of '%s' service app-runtime plan (%+v)\n", serviceName, appRuntimeServicePlan)
	var appRuntimeServiceInstances []models.CFServiceInstance
	appRuntimeServiceInstances, err = clients.GetServiceInstances(c.CliConnection, context.SpaceID, []models.CFServicePlan{*appRuntimeServicePlan})
	if err != nil {
		return html5Context, errors.New("Could not get service instances for app-runtime plan: " + err.Error())
	}

	// Filter out service instances that were recently failed to delete
	validAppRuntimeServiceInstances := make([]models.CFServiceInstance, 0)
	for _, serviceInstance := range appRuntimeServiceInstances {
		if serviceInstance.LastOperation.Type == "delete" && serviceInstance.LastOperation.State == "failed" {
			log.Tracef("Service instance %s is potentially broken and will not be reused\n", serviceInstance.Name)
			continue
		}
		validAppRuntimeServiceInstances = append(validAppRuntimeServiceInstances, serviceInstance)
	}
	html5Context.HTML5AppRuntimeServiceInstances = validAppRuntimeServiceInstances

	// Create instance of app-runtime plan if needed
	var appRuntimeServiceInstance *models.CFServiceInstance
	if len(validAppRuntimeServiceInstances) == 0 {
		log.Tracef("Creating service instance of %s service app-runtime plan\n", serviceName)
		appRuntimeServiceInstance, err = clients.CreateServiceInstance(c.CliConnection, context.SpaceID, *appRuntimeServicePlan, nil, "")
		if err != nil {
			return html5Context, errors.New("Could not create service instance of app-runtime plan: " + err.Error())
		}
		validAppRuntimeServiceInstances = append(validAppRuntimeServiceInstances, *appRuntimeServiceInstance)
	}
	html5Context.HTML5AppRuntimeServiceInstance = appRuntimeServiceInstance

	// Get service key
	log.Tracef("Getting list of service keys for service %s\n", validAppRuntimeServiceInstances[len(validAppRuntimeServiceInstances)-1].Name)
	appRuntimeServiceInstanceKeys, err := clients.GetServiceKeys(c.CliConnection, validAppRuntimeServiceInstances[len(validAppRuntimeServiceInstances)-1].GUID)
	if err != nil {
		return html5Context, errors.New("Could not get service keys of " +
			validAppRuntimeServiceInstances[len(validAppRuntimeServiceInstances)-1].Name + " service instance: " + err.Error())
	}
	if len(appRuntimeServiceInstanceKeys) > 0 {
		log.Tracef("Found %d service keys for service %s, using service key with GUID=%s\n",
			len(appRuntimeServiceInstanceKeys),
			validAppRuntimeServiceInstances[len(validAppRuntimeServiceInstances)-1].Name,
			appRuntimeServiceInstanceKeys[len(appRuntimeServiceInstanceKeys)-1].GUID)
		html5Context.HTML5AppRuntimeServiceInstanceKeys = appRuntimeServiceInstanceKeys
	}

	// Create service key if needed
	if len(appRuntimeServiceInstanceKeys) == 0 {
		var keyParams interface{}
		keyParamsJson := os.Getenv("HTML5_APP_RUNTIME_KEY_PARAMETERS")
		if keyParamsJson != "" {
			log.Tracef("Using service key configuration %s\n", keyParamsJson)
			err = json.Unmarshal([]byte(keyParamsJson), &keyParams)
			if err != nil {
				return html5Context, errors.New("Service key configuration is not a valid JSON: " + err.Error())
			}
		}
		log.Tracef("Creating service key for %s service\n", validAppRuntimeServiceInstances[len(validAppRuntimeServiceInstances)-1].Name)
		appRuntimeServiceInstanceKey, err := clients.CreateServiceKey(c.CliConnection, validAppRuntimeServiceInstances[len(validAppRuntimeServiceInstances)-1].GUID, keyParams)
		if err != nil {
			return html5Context, errors.New("Could not create service key of " +
				validAppRuntimeServiceInstances[len(validAppRuntimeServiceInstances)-1].Name + " service instance: " + err.Error())
		}
		html5Context.HTML5AppRuntimeServiceInstanceKeys = append(html5Context.HTML5AppRuntimeServiceInstanceKeys, *appRuntimeServiceInstanceKey)
		html5Context.HTML5AppRuntimeServiceInstanceKey = appRuntimeServiceInstanceKey
	}

	// Get app-runtime access token
	log.Tracef("Getting token for service key %s\n", html5Context.HTML5AppRuntimeServiceInstanceKeys[len(html5Context.HTML5AppRuntimeServiceInstanceKeys)-1].Name)
	appRuntimeServiceInstanceKeyToken, err := clients.GetToken(html5Context.HTML5AppRuntimeServiceInstanceKeys[len(html5Context.HTML5AppRuntimeServiceInstanceKeys)-1].Credentials)
	if err != nil {
		return html5Context, errors.New("Could not obtain access token: " + err.Error())
	}
	html5Context.HTML5AppRuntimeServiceInstanceKeyToken = appRuntimeServiceInstanceKeyToken
	log.Tracef("Access token for service key %s: %s\n",
		html5Context.HTML5AppRuntimeServiceInstanceKeys[len(html5Context.HTML5AppRuntimeServiceInstanceKeys)-1].Name,
		log.Sensitive{Data: appRuntimeServiceInstanceKeyToken})

	// Fill cache
	cache.Set("GetHTML5Context:"+context.OrgID+":"+context.SpaceID, html5Context)

	return html5Context, nil
}

// CleanHTML5Context clean-up temporary service keys and service instances
// created to form HTML5 context

// HTML5Context HTML5 context struct

// GetRuntimeURL base runtime URL for HTML5 applications
func (ctx *HTML5Context) GetRuntimeURLDest(runtime string) string {
	runtimeURL := os.Getenv("HTML5_RUNTIME_URL")
	if runtimeURL == "" {
		uri := *ctx.HTML5AppRuntimeServiceInstanceKey.Credentials.URI
		if runtime == "" {
			runtime = "cpp"
		}
		runtimeURL = "https://" + ctx.HTML5AppRuntimeServiceInstanceKey.Credentials.UAA.IdentityZone + "." + runtime + uri[strings.Index(uri, "."):]
	}
	return runtimeURL
}
