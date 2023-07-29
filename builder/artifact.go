package builder

import (
	"path"
	"fmt"
)

// packersdk.Artifact implementation
type Artifact struct {
	Name string

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return []string{
		a.vmDirPath(),
	}
}

func (a *Artifact) Id() string {
	return a.Name
}

func (*Artifact) String() string {
	return ""
}

func (a *Artifact) State(name string) interface{} {
	return a.StateData[name]
}

func (a *Artifact) Destroy() error {
	return nil
}

func (a *Artifact) vmDirPath() string {
	return path.Join("opt", "ci", "vm-images", fmt.Sprintf("%s.bundle", a.Name))
}
