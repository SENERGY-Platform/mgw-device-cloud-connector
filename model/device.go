package model

const (
	Online  = "online"
	Offline = "offline"
)

type Device struct {
	ID         string
	Name       string
	State      string
	Type       string
	Attributes []Attribute
}

type Attribute struct {
	Key   string
	Value string
}
