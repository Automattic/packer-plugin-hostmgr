//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package uploader

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	MockOption          string `mapstructure:"mock"`
	ctx                 interpolate.Context
}

type PostProcessor struct {
	config Config
}

func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         "packer.post-processor.hostmgr",
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

func (p *PostProcessor) PostProcess(ctx context.Context, ui packersdk.Ui, source packersdk.Artifact) (packersdk.Artifact, bool, bool, error) {
	
	ui.Message(fmt.Sprintf("Shutting down VM %+v", source.Id()))
	stopErr := HostmgrStreamingExec(ctx, ui, "vm", "stop", source.Id())

	if stopErr != nil {
		return source, false, false, stopErr
	}

	ui.Message(fmt.Sprintf("Packaging VM %+v", source.Id()))
	packageErr := HostmgrStreamingExec(ctx, ui, "vm", "package", source.Id())

	if packageErr != nil {
		return source, false, false, packageErr
	}

	ui.Message(fmt.Sprintf("Uploading VM %+v", source.Id()))
	uploadErr := HostmgrStreamingExec(ctx, ui, "vm", "publish", source.Id())

	if uploadErr != nil {
		return source, false, false, packageErr
	}

	return source, true, true, nil
}
