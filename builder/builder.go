//go:generate packer-sdc mapstructure-to-hcl2 -type Config
//go:generate packer-sdc struct-markdown

package builder

import (
	"context"
	"errors"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

const BuilderId = "hostmgr"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm communicator.Config `mapstructure:",squash"`

	// The name of the image we're cloning from
	SourceImage string `mapstructure:"source_image" required:"true"`

	// The name of the new image we're creating
	DestinationImage string `mapstructure:"destination_image" required:"true"`

	ctx interpolate.Context
}

type HostmgrBuilder struct {
	config Config
	runner multistep.Runner
}

func (b *HostmgrBuilder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *HostmgrBuilder) Prepare(raws ...interface{}) (generatedVars []string, warnings []string, err error) {
	err = config.Decode(&b.config, &config.DecodeOpts{
		PluginType:  "packer.builder.hostmgr",
		Interpolate: true,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{"boot_command"},
		},
	}, raws...)

	if b.config.SourceImage == "" {
		return nil, nil, errors.New("You must specify a `source_image` parameter")
	}

	if b.config.DestinationImage == "" {
		return nil, nil, errors.New("You must specify a `destination_image` parameter")
	}

	if err != nil {
		return nil, nil, err
	}

	if errs := b.config.Comm.Prepare(&b.config.ctx); len(errs) != 0 {
		return nil, nil, packer.MultiErrorAppend(nil, errs...)
	}

	return nil, nil, nil
}

func (b *HostmgrBuilder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	steps := []multistep.Step{}

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)

	state.Put("config", &b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)

	steps = append(steps,
		&communicator.StepSSHKeyGen{
			CommConf:            &b.config.Comm,
			SSHTemporaryKeyPair: b.config.Comm.SSH.SSHTemporaryKeyPair,
		},
		new(stepCloneVM),
		new(stepStartupVM),
		new(stepGetVMDetails), // Looks up the VM IP and stores it in `ip-address`
		&communicator.StepConnect{
			Config: &b.config.Comm,
			Host: communicator.CommHost(b.config.Comm.Host(), "ip-address"),
			SSHConfig: b.config.Comm.SSHConfigFunc(),
		},
		new(commonsteps.StepProvision),
	)

	// Run!
	b.runner = commonsteps.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if err, ok := state.GetOk("error"); ok {
		return nil, err.(error)
	}

	artifact := &Artifact{
		Name: b.config.DestinationImage,
		// Add the builder generated data to the artifact StateData so that post-processors
		// can access them.
		StateData: map[string]interface{}{
			"generated_data": map[string]interface{} {},
		},
	}

	return artifact, nil
}
