package builder

import (
	"context"
	"strings"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepGetVMDetails struct{}

func (s *stepGetVMDetails) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	config := state.Get("config").(*Config)

	ui.Message("Looking up VM IP address")
	out, err := HostmgrExec(ctx, ui, "vm", "details", config.DestinationImage, "--ip-address")
	ui.Message(string(out))

	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())

		return multistep.ActionHalt
	}

	state.Put("ip-address", strings.TrimSpace(string(out)))

	return multistep.ActionContinue
}

func (s *stepGetVMDetails) Cleanup(state multistep.StateBag) {
	// nothing to clean up
}
