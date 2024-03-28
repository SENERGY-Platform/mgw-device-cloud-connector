package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	srv_info_hdl "github.com/SENERGY-Platform/go-service-base/srv-info-hdl"
	sb_util "github.com/SENERGY-Platform/go-service-base/util"
	"github.com/SENERGY-Platform/go-service-base/watchdog"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler/cloud_device_hdl"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler/cloud_mqtt_hdl"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler/local_device_hdl"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler/local_mqtt_hdl"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler/message_hdl"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler/msg_relay_hdl"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/auth_client"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/cloud_client"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/dm_client"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/paho_mqtt"
	"github.com/eclipse/paho.mqtt.golang"
	"net/http"
	"os"
	"syscall"
	"time"
)

var version string

func main() {
	srvInfoHdl := srv_info_hdl.New("device-cloud-connector", version)

	ec := 0
	defer func() {
		os.Exit(ec)
	}()

	util.ParseFlags()

	config, err := util.NewConfig(util.Flags.ConfPath)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		ec = 1
		return
	}

	logFile, err := util.InitLogger(config.Logger)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		var logFileError *sb_util.LogFileError
		if errors.As(err, &logFileError) {
			ec = 1
			return
		}
	}
	if logFile != nil {
		defer logFile.Close()
	}

	util.Logger.Printf("%s %s", srvInfoHdl.GetName(), srvInfoHdl.GetVersion())

	util.Logger.Debugf("config: %s", sb_util.ToJsonStr(config))

	watchdog.Logger = util.Logger
	wtchdg := watchdog.New(syscall.SIGINT, syscall.SIGTERM)

	if config.MQTTLog {
		paho_mqtt.SetLogger(config.MQTTDebugLog)
	}

	dmClient := dm_client.New(http.DefaultClient, config.HttpClient.LocalDmBaseUrl)
	ldhCtx, cf := context.WithCancel(context.Background())
	defer cf()
	localDeviceHdl := local_device_hdl.New(ldhCtx, dmClient, time.Duration(config.HttpClient.LocalTimeout), time.Duration(config.LocalDeviceHandler.QueryInterval), config.LocalDeviceHandler.IDPrefix)

	cloudClient := cloud_client.New(http.DefaultClient, config.HttpClient.CloudApiBaseUrl, auth_client.New(http.DefaultClient, config.HttpClient.CloudAuthBaseUrl, config.CloudAuth.User, config.CloudAuth.Password.String(), config.CloudAuth.ClientID))
	cloudDeviceHdl := cloud_device_hdl.New(cloudClient, time.Duration(config.HttpClient.CloudTimeout), time.Duration(config.CloudDeviceHandler.SyncInterval), config.CloudDeviceHandler.WrkSpcPath, config.CloudDeviceHandler.AttributeOrigin)

	chCtx, cf := context.WithCancel(context.Background())
	defer cf()
	hubID, err := cloudDeviceHdl.Init(chCtx, config.CloudDeviceHandler.HubID, config.CloudDeviceHandler.DefaultHubName)
	if err != nil {
		util.Logger.Error(err)
		ec = 1
		return
	}

	localMqttHdl := local_mqtt_hdl.New(config.LocalMqttClient.QOSLevel)

	localMqttClientOpt := mqtt.NewClientOptions()
	localMqttClientOpt.SetOnConnectHandler(func(_ mqtt.Client) {
		localMqttHdl.HandleSubscriptions()
	})
	paho_mqtt.SetLocalClientOptions(localMqttClientOpt, fmt.Sprintf("%s_%s", srvInfoHdl.GetName(), config.MGWDeploymentID), config.LocalMqttClient)
	localMqttClient := paho_mqtt.NewWrapper(mqtt.NewClient(localMqttClientOpt), time.Duration(config.LocalMqttClient.WaitTimeout))
	localMqttClientPubF := func(topic string, data []byte) error {
		return localMqttClient.Publish(topic, config.LocalMqttClient.QOSLevel, false, data)
	}

	cloudMqttHdl := cloud_mqtt_hdl.New(config.CloudMqttClient.QOSLevel, hubID)

	cloudMqttClientOpt := mqtt.NewClientOptions()
	cloudMqttClientOpt.SetConnectionLostHandler(func(_ mqtt.Client, _ error) {
		cloudMqttHdl.HandleOnDisconnect()
	})
	paho_mqtt.SetCloudClientOptions(cloudMqttClientOpt, hubID, config.CloudMqttClient, &config.CloudAuth, &tls.Config{InsecureSkipVerify: true})
	cloudMqttClient := paho_mqtt.NewWrapper(mqtt.NewClient(cloudMqttClientOpt), time.Duration(config.CloudMqttClient.WaitTimeout))
	cloudMqttClientPubF := func(topic string, data []byte) error {
		return cloudMqttClient.Publish(topic, config.CloudMqttClient.QOSLevel, false, data)
	}

	message_hdl.DeviceCommandIDPrefix = fmt.Sprintf("%s_%s_", srvInfoHdl.GetName(), config.MGWDeploymentID)
	message_hdl.DeviceCommandMaxAge = time.Duration(config.MaxDeviceCmdAge)
	message_hdl.HubID = hubID
	message_hdl.LocalDeviceIDPrefix = config.LocalDeviceHandler.IDPrefix

	deviceCmdMsgRelayHdl := msg_relay_hdl.New(config.MessageRelayBuffer, message_hdl.HandleDownstreamDeviceCmd, localMqttClientPubF)
	processesCmdMsgRelayHdl := msg_relay_hdl.New(config.MessageRelayBuffer, message_hdl.HandleDownstreamProcessesCmd, localMqttClientPubF)
	deviceEventMsgRelayHdl := msg_relay_hdl.New(config.EventMessageRelayBuffer, message_hdl.HandleUpstreamDeviceEvent, cloudMqttClientPubF)
	deviceCmdRespMsgRelayHdl := msg_relay_hdl.New(config.MessageRelayBuffer, message_hdl.HandleUpstreamDeviceCmdResponse, cloudMqttClientPubF)
	processesStateMsgRelayHdl := msg_relay_hdl.New(config.MessageRelayBuffer, message_hdl.HandleUpstreamProcessesState, cloudMqttClientPubF)
	deviceConnectorErrMsgRelayHdl := msg_relay_hdl.New(config.MessageRelayBuffer, message_hdl.HandlerUpstreamDeviceConnectorErr, cloudMqttClientPubF)
	deviceErrMsgRelayHdl := msg_relay_hdl.New(config.MessageRelayBuffer, message_hdl.HandlerUpstreamDeviceErr, cloudMqttClientPubF)
	deviceCmdErrMsgRelayHdl := msg_relay_hdl.New(config.MessageRelayBuffer, message_hdl.HandlerUpstreamDeviceCmdErr, cloudMqttClientPubF)

	localMqttHdl.SetMqttClient(localMqttClient)
	localMqttHdl.SetMessageRelayHdl(
		deviceEventMsgRelayHdl,
		deviceCmdRespMsgRelayHdl,
		processesStateMsgRelayHdl,
		deviceConnectorErrMsgRelayHdl,
		deviceErrMsgRelayHdl,
		deviceCmdErrMsgRelayHdl)
	cloudMqttHdl.SetMqttClient(cloudMqttClient)
	cloudMqttHdl.SetMessageRelayHdl(deviceCmdMsgRelayHdl, processesCmdMsgRelayHdl)

	localDeviceHdl.SetDeviceSyncFunc(cloudDeviceHdl.Sync)
	localDeviceHdl.SetDeviceStateSyncFunc(cloudMqttHdl.HandleSubscriptions)

	localDeviceHdl.Start()

	deviceCmdMsgRelayHdl.Start()
	processesCmdMsgRelayHdl.Start()
	deviceEventMsgRelayHdl.Start()
	deviceCmdRespMsgRelayHdl.Start()
	processesStateMsgRelayHdl.Start()
	deviceConnectorErrMsgRelayHdl.Start()
	deviceErrMsgRelayHdl.Start()
	deviceCmdErrMsgRelayHdl.Start()

	localMqttClient.Connect()
	cloudMqttClient.Connect()

	wtchdg.RegisterHealthFunc(localDeviceHdl.Running)
	wtchdg.RegisterHealthFunc(cloudDeviceHdl.HasHub)
	wtchdg.RegisterStopFunc(func() error {
		localDeviceHdl.Stop()
		return nil
	})
	wtchdg.RegisterStopFunc(func() error {
		localMqttClient.Disconnect(1000)
		cloudMqttClient.Disconnect(1000)
		return nil
	})
	wtchdg.RegisterStopFunc(func() error {
		deviceCmdMsgRelayHdl.Stop()
		processesCmdMsgRelayHdl.Stop()
		deviceEventMsgRelayHdl.Stop()
		deviceCmdRespMsgRelayHdl.Stop()
		processesStateMsgRelayHdl.Stop()
		deviceConnectorErrMsgRelayHdl.Stop()
		deviceErrMsgRelayHdl.Stop()
		deviceCmdErrMsgRelayHdl.Stop()
		return nil
	})

	wtchdg.Start()

	ec = wtchdg.Join()
}
