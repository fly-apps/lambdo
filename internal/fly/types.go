package fly

import (
	"golang.org/x/exp/slices"
)

const TypeShared = "shared"
const TypePerf = "performance"

type Organization struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type App struct {
	Name         string       `json:"name"`
	Organization Organization `json:"organization"`
}

type Machine struct {
	Id        string        `json:"id,omitempty"`
	Name      string        `json:"name,omitempty"`
	State     string        `json:"state,omitempty"`
	Region    string        `json:"region"`
	PrivateIp string        `json:"private_ip,omitempty"`
	Config    MachineConfig `json:"config"`
}

// IsInitialized checks the state of the machine to see if it
// is finished being created.
// See https://fly.io/docs/reference/machines/#machine-states
func (m *Machine) IsInitialized() bool {
	initValues := []string{"started", "stopped", "stopping"}

	return slices.Contains(initValues, m.State)
}

type MachineConfig struct {
	Image       string            `json:"image"`
	Guest       MachineSize       `json:"guest,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	Services    []MachineService  `json:"services,omitempty"`
	Processes   []MachineProcess  `json:"processes,omitempty"`
	MetaData    map[string]string `json:"metadata,omitempty"`
	Files       []MachineFile     `json:"files,omitempty"`
	AutoDestroy bool              `json:"auto_destroy,omitempty"`
}

type MachineSize struct {
	CpuCount int    `json:"cpus"`
	RAM      int    `json:"memory_mb"`
	Type     string `json:"cpu_kind"`
}

type MachineService struct {
	InternalPort int    `json:"internal_port"` // 8000
	Protocol     string `json:"protocol"`      // "tcp"
	Ports        []Port `json:"ports"`
}

type MachineProcess struct {
	Name       string            `json:"name,omitempty"`
	Entrypoint []string          `json:"entrypoint,omitempty"`
	Cmd        []string          `json:"cmd,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
	User       string            `json:"user,omitempty"`
}

type MachineFile struct {
	GuestPath  string `json:"guest_path"`
	RawValue   string `json:"raw_value,omitempty"`
	SecretName string `json:"secret_name,omitempty"`
}

type Port struct {
	Port     int      `json:"port"`     // 80, 443
	Handlers []string `json:"handlers"` // ["http"], ["tls", "http"]
}
