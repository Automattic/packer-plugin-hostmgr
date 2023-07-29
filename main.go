package main

import (
	"fmt"
	"os"

	"github.com/automattic/packer-plugin-hostmgr/builder"
	"github.com/automattic/packer-plugin-hostmgr/provisioner"
	"github.com/automattic/packer-plugin-hostmgr/uploader"

	pluginVersion "github.com/automattic/packer-plugin-hostmgr/version"
	"github.com/hashicorp/packer-plugin-sdk/plugin"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder("builder", new(builder.HostmgrBuilder))
	pps.RegisterProvisioner("provisioner", new(provisioner.HostmgrProvisioner))
	pps.RegisterPostProcessor("uploader", new(uploader.PostProcessor))

	pps.SetVersion(pluginVersion.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
