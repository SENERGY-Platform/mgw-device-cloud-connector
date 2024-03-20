package message_hdl

type CloudStandardEnvelope struct {
	Metadata string `json:"metadata,omitempty"`
	Data     string `json:"data"`
}

type CloudDeviceEventMsg = CloudStandardEnvelope

type CloudDeviceCmdMsg struct {
	CorrelationID      string                `json:"correlation_id"`
	CompletionStrategy string                `json:"completion_strategy"`
	Timestamp          float64               `json:"timestamp"`
	Payload            CloudStandardEnvelope `json:"payload"`
}

type CloudDeviceCmdResponseMsg struct {
	CorrelationID string                `json:"correlation_id"`
	Payload       CloudStandardEnvelope `json:"payload"`
}

type LocalDeviceCmdBase struct {
	CommandID string `json:"command_id"`
	Data      string `json:"data"`
}

type LocalDeviceCmdMsg struct {
	LocalDeviceCmdBase
	CompletionStrategy string `json:"completion_strategy"`
}

type LocalDeviceCmdResponseMsg = LocalDeviceCmdBase
