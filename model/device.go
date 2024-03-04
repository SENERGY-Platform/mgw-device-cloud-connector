package model

const (
	Online  = "online"
	Offline = "offline"
)

type Device struct {
	ID         string
	Name       string
	State      string
	LastState  string
	Type       string
	Hash       string
	Attributes []Attribute
}

type Attribute struct {
	Key   string
	Value string
}
