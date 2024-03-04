package dm_client

type Device struct {
	Name       string      `json:"name"`
	State      string      `json:"state"`
	Type       string      `json:"device_type"`
	ModuleID   string      `json:"module_id"`
	Attributes []Attribute `json:"attributes"`
}

type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
