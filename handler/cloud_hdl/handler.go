package cloud_hdl

import (
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/cloud_client"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"time"
)

type Handler struct {
	cloudClient  cloud_client.ClientItf
	mqttClient   mqtt.Client
	timeout      time.Duration
	hubInfo      hubInfo
	wrkSpacePath string
}

func New(cloudClient cloud_client.ClientItf, mqttClient mqtt.Client, timeout time.Duration, wrkSpacePath string) *Handler {
	return &Handler{
		cloudClient:  cloudClient,
		mqttClient:   mqttClient,
		timeout:      timeout,
		wrkSpacePath: wrkSpacePath,
	}
}
