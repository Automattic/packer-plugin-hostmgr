package builder

import (
	"context"
	"time"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepStartupVM struct{}

func (s *stepStartupVM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Message("Booting " + config.DestinationImage + " - once it's booted, you'll need to configure it")
	err := HostmgrStreamingExec(ctx, ui, "vm", "start", config.DestinationImage, "--persistent")
	time.Sleep(3 * time.Second)

	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())

		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepStartupVM) Cleanup(state multistep.StateBag) {
	// No cleanup required
}
