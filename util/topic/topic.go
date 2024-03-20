package topic

const (
	CloudDeviceCmdSub          = "command/+/+"       // command/{local_device_id}/+
	CloudProcessesCmdSub       = "processes/+/cmd/#" // processes/{hub_id}/cmd/#
	CloudDeviceEventPub        = "event"             // event/{local_device_id}/{service_id}
	CloudDeviceCmdResponsePub  = "response"          // response/{local_device_id}/{service_id}
	CloudProcessesStatePub     = "processes"         // processes/{hub_id}/state/{sub_topic}
	CloudDeviceConnectorErrPub = "error"
	CloudDeviceErrPub          = "error/device"  //  error/device/{local_device_id}
	CloudDeviceCmdErrPub       = "error/command" // error/command/{correlation_id}
)

const (
	LocalDeviceEventSub        = "event/+/+"    // event/{local_device_id}/{service_id}
	LocalDeviceCmdResponseSub  = "response/+/+" // response/{local_device_id}/{service_id}
	LocalProcessesStateSub     = "processes/state/#"
	LocalDeviceConnectorErrSub = "error/client"
	LocalDeviceErrSub          = "error/device/+"
	LocalDeviceCmdErrSub       = "error/command/+"
	LocalDeviceCmdPub          = "command"
	LocalProcessesCmdPub       = "processes/cmd"
)
