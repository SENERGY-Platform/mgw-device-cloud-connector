package message_hdl

type CloudStandardEnvelope struct {
	Metadata string `json:"metadata,omitempty"`
	Data     string `json:"data"`
}

type CloudDeviceEventMsg = CloudStandardEnvelope

type CloudDeviceCmdResponseMsg struct {
	CorrelationID string                `json:"correlation_id"`
	Payload       CloudStandardEnvelope `json:"payload"`
}

type LocalDeviceCmdResponseMsg struct {
	CommandID string `json:"command_id"`
	Data      string `json:"data"`
}
