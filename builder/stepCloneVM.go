package builder

import (
	"context"
	"fmt"
	"time"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepCloneVM struct{}

func (s *stepCloneVM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Message(fmt.Sprintf("Cloning %s as %s", config.SourceImage, config.DestinationImage))
	err := HostmgrStreamingExec(ctx, ui, "vm", "clone", config.SourceImage, config.DestinationImage)
	time.Sleep(1 * time.Second)

	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())

		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepCloneVM) Cleanup(state multistep.StateBag) {
	// No cleanup required
}
