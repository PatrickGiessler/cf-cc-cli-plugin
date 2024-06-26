package commands

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/plugin"
	plugin_models "github.com/cloudfoundry/cli/plugin/models"
)

const (
	maxConcurrentConnections = 50
	maxRetryCount            = 3
)

// BaseCommand base command for all commands
type BaseCommand struct {
	Name          string
	CliConnection plugin.CliConnection
}

// Initialize default initialization method which may be overriden in more specific commands
func (c *BaseCommand) Initialize(name string, cliConnection plugin.CliConnection) error {
	return c.InitializeBase(name, cliConnection)
}

// InitializeBase initializes the command with the specified name and CLI connection.
func (c *BaseCommand) InitializeBase(name string, cliConnection plugin.CliConnection) error {
	c.Name = name
	c.CliConnection = cliConnection
	return nil
}

// Dispose default dispose method which may be overriden in more specific commands
func (c *BaseCommand) Dispose(name string) {
	c.DisposeBase(name)
}

// DisposeBase dispose command
func (c *BaseCommand) DisposeBase(_ string) {
	// Do nothing
}

// Context holding the username, Org and Space of the current used
type Context struct {
	Username string
	Org      string
	OrgID    string
	Space    string
	SpaceID  string
}

// GetContext initializes and retrieves the Context
func (c *BaseCommand) GetContext() (Context, error) {
	username, err := c.GetUsername()
	if err != nil {
		return Context{}, err
	}
	org, err := c.GetOrg()
	if err != nil {
		return Context{}, err
	}
	space, err := c.GetSpace()
	if err != nil {
		return Context{}, err
	}
	return Context{Org: org.Name, OrgID: org.Guid, Space: space.Name, SpaceID: space.Guid, Username: username}, nil
}

// GetOrg gets the current org name from the CLI connection
func (c *BaseCommand) GetOrg() (plugin_models.Organization, error) {
	org, err := c.CliConnection.GetCurrentOrg()
	if err != nil {
		return plugin_models.Organization{}, fmt.Errorf("could not get current org: %s", err)
	}
	if org.Name == "" {
		return plugin_models.Organization{}, fmt.Errorf("no org and space targeted, use '%s' to target an org and a space", terminal.CommandColor("cf target -o ORG -s SPACE"))
	}
	return org, nil
}

// GetSpace gets the current space name from the CLI connection
func (c *BaseCommand) GetSpace() (plugin_models.Space, error) {
	space, err := c.CliConnection.GetCurrentSpace()
	if err != nil {
		return plugin_models.Space{}, fmt.Errorf("could not get current space: %s", err)
	}

	if space.Name == "" || space.Guid == "" {
		return plugin_models.Space{}, fmt.Errorf("no space targeted, use '%s' to target a space", terminal.CommandColor("cf target -s"))
	}
	return space, nil
}

// GetUsername gets the username from the CLI connection
func (c *BaseCommand) GetUsername() (string, error) {
	username, err := c.CliConnection.Username()
	if err != nil {
		return "", fmt.Errorf("could not get username: %s", err)
	}
	if username == "" {
		return "", fmt.Errorf("not logged in. Use '%s' to log in", terminal.CommandColor("cf login"))
	}
	return username, nil
}
