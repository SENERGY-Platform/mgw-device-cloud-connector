package topic

const (
	LocalDeviceEventSub        = "event/+/+"    // event/{local_device_id}/{service_id}
	LocalDeviceCmdResponseSub  = "response/+/+" // response/{local_device_id}/{service_id}
	LocalProcessesStateSub     = "processes/state/#"
	LocalDeviceConnectorErrSub = "error/client"
	LocalDeviceErrSub          = "error/device/+"  // error/device/{local_device_id}
	LocalDeviceCmdErrSub       = "error/command/+" // error/command/{correlation_id}
)
