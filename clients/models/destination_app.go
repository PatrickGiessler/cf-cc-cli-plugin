package models

// HTML5ListApplicationsResponse response of list applications API
type DestinationAppResponse []DestinationApp

// HTML5App HTML5 application
type DestinationApp struct {
	Description string `json:"Description,omitempty"`
	Name        string `json:"Name,omitempty"`
	ProxyType   string `json:"ProxyType,omitempty"`
	URL         string `json:"URL,omitempty"`
}
