package dm_client

const (
	Online  State = "online"
	Offline State = "offline"
)

type State = string

type Device struct {
	Name       string      `json:"name"`
	State      State       `json:"state"`
	Type       string      `json:"device_type"`
	ModuleID   string      `json:"module_id"`
	Attributes []Attribute `json:"attributes"`
}

type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
