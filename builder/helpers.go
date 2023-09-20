package builder

import (
	"context"
	"fmt"
	"os/exec"
	"io"
	"bytes"
	"strings"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type uiMessageWriter struct {
	ui    packersdk.Ui
}

func (e uiMessageWriter) Write(p []byte) (int, error) {
	e.ui.Say(string(p))
	return len(p), nil
}

type uiErrorWriter struct {
	ui    packersdk.Ui
}

func (e uiErrorWriter) Write(p []byte) (int, error) {
	e.ui.Error(string(p))
	return len(p), nil
}

func HostmgrStreamingExec(ctx context.Context, ui packersdk.Ui, args ...string) error {
	ui.Message(fmt.Sprintf("Running %#v", strings.Join(append([]string{"hostmgr"}, args...), " ")))

	cmd := exec.CommandContext(ctx, "hostmgr", args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(uiMessageWriter { ui: ui }, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(uiErrorWriter { ui: ui }, &stderrBuf)

	err := cmd.Run()

	return err
}

func HostmgrExec(ctx context.Context, ui packersdk.Ui, args ...string) (string, error) {
	ui.Message(fmt.Sprintf("Running %#v", strings.Join(append([]string{"hostmgr"}, args...), " ")))

	cmd := exec.CommandContext(ctx, "hostmgr", args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(uiMessageWriter { ui: ui }, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(uiErrorWriter { ui: ui }, &stderrBuf)

	err := cmd.Run()

	return stdoutBuf.String(), err
}
