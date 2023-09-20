//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package provisioner

import (
	"context"
	"strings"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	ctx interpolate.Context

	// Should we enable passwordless sudo? It's super helpful for CI nodes, because we can install software at runtime
	EnablePasswordlessSudo bool `mapstructure:"enable_passwordless_sudo"`

	// Should we install Homebrew on this machine?
	InstallHomebrew bool `mapstructure:"install_homebrew"`

	// Which homebrew dependencies should we install? If this value isn't empty, it'll cause Homebrew to be installed
	HomebrewDependencies []string `mapstructure:"homebrew_dependencies"`
}

type HostmgrProvisioner struct {
	config Config
}

func (p *HostmgrProvisioner) ConfigSpec() hcldec.ObjectSpec {
	return p.config.FlatMapstructure().HCL2Spec()
}

func (p *HostmgrProvisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         "packer.provisioner.hostmgr",
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}
	return nil
}

func (p *HostmgrProvisioner) Provision(ctx context.Context, ui packer.Ui, communicator packer.Communicator, generatedData map[string]interface{}) error {
	
	copyErr := copySSHProvisioningKeyToVM(ctx, ui, communicator, generatedData)

	if copyErr != nil {
		ui.Error("Failed to copy SSH provisioning key to VM – aborting")
		return copyErr
	}

	if p.config.EnablePasswordlessSudo {
		sudoErr := enablePasswordlessSudo(ctx, ui, communicator, generatedData)

		if sudoErr != nil {
			ui.Error("Failed to enable passwordless sudo – aborting")
			return sudoErr
		}
	}

	if p.config.InstallHomebrew || len(p.config.HomebrewDependencies) > 0 {
		homebrewErr := installHomebrew(ctx, ui, communicator, generatedData)

		if homebrewErr != nil {
			ui.Error("Failed to install Homebrew – aborting")
			return homebrewErr
		}
	}

	for _, dependency := range p.config.HomebrewDependencies {
		dependencyErr := installHomebrewDependency(ctx, ui, communicator, dependency)

		if dependencyErr != nil {
			ui.Error(fmt.Sprintf("Failed to install %s – aborting", dependency))
			return dependencyErr
		}
	}


	return nil
}

func copySSHProvisioningKeyToVM(_ context.Context, ui packer.Ui, communicator packer.Communicator, generatedData map[string]interface{}) error {
	ui.Message("Copying SSH Key to VM")

	username := generatedData["User"].(string)
	sshPublicKey := generatedData["SSHPublicKey"].(string)

	return communicator.Upload(fmt.Sprintf("/Users/%s/.ssh/authorized_keys", username), strings.NewReader(sshPublicKey), nil)
}

func enablePasswordlessSudo(ctx context.Context, ui packer.Ui, communicator packer.Communicator, generatedData map[string]interface{}) error {
	ui.Message("Enabling Password-less sudo")

	username := generatedData["User"].(string)
	password := generatedData["Password"].(string)

	scriptPath := fmt.Sprintf("/Users/%s/enable-passwordless-sudo.sh", username)

	uploadErr := communicator.Upload(scriptPath, strings.NewReader(enablePasswordlessSudoCommand()), nil)

	if uploadErr != nil {
		return uploadErr
	}

	commandString := fmt.Sprintf("echo %s | sudo -S sh -eux %s", password, scriptPath)

	command := packer.RemoteCmd {
		Command: commandString,
	}

	return command.RunWithUi(ctx, communicator, ui)
}

func enablePasswordlessSudoCommand() string {
	return `fgrep '%admin		ALL = (ALL) NOPASSWD: ALL' /etc/sudoers && exit 0
	sed -i.bak 's/%admin		ALL = (ALL) ALL/%admin		ALL = (ALL) NOPASSWD: ALL/' /etc/sudoers`
}

func installHomebrew(ctx context.Context, ui packer.Ui, communicator packer.Communicator, generatedData map[string]interface{}) error {
	ui.Message("Installing Homebrew")

	command := packer.RemoteCmd {
		Command: "NONINTERACTIVE=1 curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install.sh | /bin/bash",
	}

	return command.RunWithUi(ctx, communicator, ui)
}

func installHomebrewDependency(ctx context.Context, ui packer.Ui, communicator packer.Communicator, dependency string) error {
	ui.Message(fmt.Sprintf("Installing %s", dependency))

	command := packer.RemoteCmd {
		Command: fmt.Sprintf("/opt/homebrew/bin/brew install %s", dependency),
	}

	return command.RunWithUi(ctx, communicator, ui)
}
