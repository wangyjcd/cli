package v6

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . SetSpaceIsolationSegmentActor

type SetSpaceIsolationSegmentActor interface {
	AssignIsolationSegmentToSpaceByNameAndSpace(isolationSegmentName string, spaceGUID string) (v3action.Warnings, error)
}

//go:generate counterfeiter . SetSpaceIsolationSegmentActorV2

type SetSpaceIsolationSegmentActorV2 interface {
	GetSpaceByOrganizationAndName(orgGUID string, spaceName string) (v2action.Space, v2action.Warnings, error)
}

type SetSpaceIsolationSegmentCommand struct {
	RequiredArgs    flag.SpaceIsolationArgs `positional-args:"yes"`
	usage           interface{}             `usage:"CF_NAME set-space-isolation-segment SPACE_NAME SEGMENT_NAME"`
	relatedCommands interface{}             `related_commands:"org, reset-space-isolation-segment, restart, set-org-default-isolation-segment, space"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SetSpaceIsolationSegmentActor
	ActorV2     SetSpaceIsolationSegmentActorV2
}

func (cmd *SetSpaceIsolationSegmentCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	client, _, err := shared.NewV3BasedClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v3action.NewActor(client, config, nil, nil)

	ccClientV2, uaaClientV2, err := shared.GetNewClientsAndConnectToCF(config, ui)
	if err != nil {
		return err
	}
	cmd.ActorV2 = v2action.NewActor(ccClientV2, uaaClientV2, config)

	return nil
}

func (cmd SetSpaceIsolationSegmentCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Updating isolation segment of space {{.SpaceName}} in org {{.OrgName}} as {{.CurrentUser}}...", map[string]interface{}{
		"SegmentName": cmd.RequiredArgs.IsolationSegmentName,
		"SpaceName":   cmd.RequiredArgs.SpaceName,
		"OrgName":     cmd.Config.TargetedOrganization().Name,
		"CurrentUser": user.Name,
	})

	space, v2Warnings, err := cmd.ActorV2.GetSpaceByOrganizationAndName(cmd.Config.TargetedOrganization().GUID, cmd.RequiredArgs.SpaceName)
	cmd.UI.DisplayWarnings(v2Warnings)
	if err != nil {
		return err
	}

	warnings, err := cmd.Actor.AssignIsolationSegmentToSpaceByNameAndSpace(cmd.RequiredArgs.IsolationSegmentName, space.GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayText("In order to move running applications to this isolation segment, they must be restarted.")

	return nil
}
