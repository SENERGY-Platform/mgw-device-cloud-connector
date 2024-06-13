package topic

const (
	LocalDeviceEventSub        = "event/+/+"    // event/{local_device_id}/{service_id}
	LocalDeviceCmdResponseSub  = "response/+/+" // response/{local_device_id}/{service_id}
	LocalProcessesStateSub     = "processes/state/#"
	LocalDeviceConnectorErrSub = "error/client"
	LocalDeviceErrSub          = "error/device/+"  // error/device/{local_device_id}
	LocalDeviceCmdErrSub       = "error/command/+" // error/command/{correlation_id}
	CloudDeviceConnectorErrPub = "error"
)

type handler struct {
	UserID    string
	NetworkID string
}

var Handler *handler

func InitTopicHandler(userID, networkID string) {
	Handler = &handler{
		UserID:    userID,
		NetworkID: networkID,
	}
}

func (h *handler) CloudDeviceCmdSub() string {
	return "command/" + h.UserID + "/+/+"
}

func (h *handler) CloudDeviceServiceCmdSub(localDeviceID string) string {
	return "command/" + h.UserID + "/" + localDeviceID + "/+"
}

func (h *handler) CloudDeviceCmdResponsePub(localDeviceID, serviceID string) string {
	return "response/" + h.UserID + "/" + localDeviceID + "/" + serviceID
}

func (h *handler) CloudDeviceCmdErrPub(correlationID string) string {
	return "error/command/" + h.UserID + "/" + correlationID
}

func (h *handler) CloudDeviceEventPub(localDeviceID, serviceID string) string {
	return "event/" + h.UserID + "/" + localDeviceID + "/" + serviceID
}

func (h *handler) CloudDeviceErrPub(localDeviceID string) string {
	return "error/device/" + h.UserID + "/" + localDeviceID
}

func (h *handler) CloudProcessesCmdSub() string {
	return "processes/" + h.NetworkID + "/cmd/#"
}

func (h *handler) CloudProcessesStatePub(subTopic string) string {
	return "processes/" + h.NetworkID + "/state/" + subTopic
}

func (h *handler) LocalDeviceCmdPub(localDeviceID, serviceID string) string {
	return "command/" + localDeviceID + "/" + serviceID
}

func (h *handler) LocalProcessesCmdPub(subTopic string) string {
	return "processes/cmd/" + subTopic
}
