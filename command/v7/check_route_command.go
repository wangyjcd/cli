package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type CheckRouteCommand struct {
	BaseCommand

	RequiredArgs    flag.Domain      `positional-args:"yes"`
	Hostname        string           `long:"hostname" short:"n" description:"Hostname used to identify the HTTP route"`
	Path            flag.V7RoutePath `long:"path" description:"Path for the route"`
	Port            int              `long:"port" description:"Port used to identify the TCP route"`
	usage           interface{}      `usage:"Check an HTTP route:\n   CF_NAME check-route DOMAIN [--hostname HOSTNAME] [--path PATH]\n\nCheck a TCP route:\n   cf check-route DOMAIN --port PORT\n\nEXAMPLES:\n   CF_NAME check-route example.com                      # example.com\n   CF_NAME check-route example.com -n myhost --path foo # myhost.example.com/foo\n   CF_NAME check-route example.com --path foo           # example.com/foo\n   CF_NAME check-route example.com --port 5000          # example.com:5000"`
	relatedCommands interface{}      `related_commands:"create-route, delete-route, routes"`
}

func (cmd CheckRouteCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	_, err = cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayText("Checking for route...")

	path := cmd.Path.Path
	matches, warnings, err := cmd.Actor.CheckRoute(cmd.RequiredArgs.Domain, cmd.Hostname, path, cmd.Port)
	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		return err
	}

	formatParams := map[string]interface{}{
		"URL": desiredURL(cmd.RequiredArgs.Domain, cmd.Hostname, path, cmd.Port),
	}

	if matches {
		cmd.UI.DisplayText("Route '{{.URL}}' does exist.", formatParams)
	} else {
		cmd.UI.DisplayText("Route '{{.URL}}' does not exist.", formatParams)
	}

	cmd.UI.DisplayOK()

	return nil
}
