package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/go-service-base/srv-info-hdl"
	sb_util "github.com/SENERGY-Platform/go-service-base/util"
	"github.com/SENERGY-Platform/go-service-base/watchdog"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler/cloud_hdl"
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
	dep_adv_client "github.com/SENERGY-Platform/mgw-module-manager/clients/dep-adv-client"
	mm_model "github.com/SENERGY-Platform/mgw-module-manager/lib/model"
	"github.com/eclipse/paho.mqtt.golang"
	"net"
	"net/http"
	"net/url"
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

	localHttpClient := &http.Client{
		Timeout: time.Duration(config.HttpClient.LocalTimeout),
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	dmClient := dm_client.New(localHttpClient, config.HttpClient.LocalDmBaseUrl)
	ldhCtx, cf := context.WithCancel(context.Background())
	defer cf()
	localDeviceHdl := local_device_hdl.New(ldhCtx, dmClient, time.Duration(config.LocalDeviceHandler.RefreshInterval), config.LocalDeviceHandler.IDPrefix)

	cloudHttpClient := &http.Client{
		Timeout: time.Duration(config.HttpClient.CloudTimeout),
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	cloutAuthClient := auth_client.New(cloudHttpClient, config.HttpClient.CloudAuthBaseUrl, config.CloudAuth.User, config.CloudAuth.Password.String(), config.CloudAuth.ClientID)
	cloudClient := cloud_client.New(cloudHttpClient, config.HttpClient.CloudApiBaseUrl, cloutAuthClient)
	cloudHdl := cloud_hdl.New(cloudClient, cloutAuthClient, time.Duration(config.CloudHandler.SyncInterval), config.CloudHandler.WrkSpcPath, config.CloudHandler.AttributeOrigin)

	wtchdg.Start()

	chCtx, cf := context.WithCancel(context.Background())
	defer cf()
	wtchdg.RegisterStopFunc(func() error {
		cf()
		return nil
	})

	networkID, userID, err := cloudHdl.Init(chCtx, config.CloudHandler.NetworkID, config.CloudHandler.DefaultNetworkName, time.Duration(config.CloudHandler.NetworkInitDelay))
	if err != nil {
		util.Logger.Error(err)
		ec = 1
		return
	}

	depAdvClient := dep_adv_client.New(localHttpClient, config.HttpClient.LocalMmBaseUrl)
	daCtx, cf2 := context.WithCancel(context.Background())
	defer cf2()
	err = depAdvClient.PutDepAdvertisement(daCtx, config.MGWDeploymentID, mm_model.DepAdvertisementBase{
		Ref: "network",
		Items: map[string]string{
			"id": networkID,
		},
	})
	if err != nil {
		util.Logger.Error(err)
	}

	localMqttHdl := local_mqtt_hdl.New(config.LocalMqttClient.QOSLevel)

	localMqttClientOpt := mqtt.NewClientOptions()
	localMqttClientOpt.SetConnectionAttemptHandler(func(_ *url.URL, tlsCfg *tls.Config) *tls.Config {
		util.Logger.Infof("%s connect to broker (%s)", local_mqtt_hdl.LogPrefix, config.LocalMqttClient.Server)
		return tlsCfg
	})
	localMqttClientOpt.SetOnConnectHandler(func(_ mqtt.Client) {
		util.Logger.Infof("%s connected to broker (%s)", local_mqtt_hdl.LogPrefix, config.LocalMqttClient.Server)
		localMqttHdl.HandleSubscriptions()
	})
	localMqttClientOpt.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		util.Logger.Warningf("%s connection lost: %s", local_mqtt_hdl.LogPrefix, err)
	})
	paho_mqtt.SetLocalClientOptions(localMqttClientOpt, fmt.Sprintf("%s_%s", srvInfoHdl.GetName(), config.MGWDeploymentID), config.LocalMqttClient)
	localMqttClient := paho_mqtt.NewWrapper(mqtt.NewClient(localMqttClientOpt), time.Duration(config.LocalMqttClient.WaitTimeout))
	localMqttClientPubF := func(topic string, data []byte) error {
		return localMqttClient.Publish(topic, config.LocalMqttClient.QOSLevel, false, data)
	}

	cloudMqttHdl := cloud_mqtt_hdl.New(config.CloudMqttClient.SubscribeQOSLevel, networkID, userID)

	cloudMqttClientOpt := mqtt.NewClientOptions()
	cloudMqttClientOpt.SetConnectionAttemptHandler(func(_ *url.URL, tlsCfg *tls.Config) *tls.Config {
		util.Logger.Infof("%s connect to broker (%s)", cloud_mqtt_hdl.LogPrefix, config.CloudMqttClient.Server)
		return tlsCfg
	})
	cloudMqttClientOpt.SetOnConnectHandler(func(_ mqtt.Client) {
		util.Logger.Infof("%s connected to broker (%s)", cloud_mqtt_hdl.LogPrefix, config.CloudMqttClient.Server)
	})
	cloudMqttClientOpt.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		util.Logger.Warningf("%s connection lost: %s", cloud_mqtt_hdl.LogPrefix, err)
		cloudMqttHdl.HandleOnDisconnect()
	})
	paho_mqtt.SetCloudClientOptions(cloudMqttClientOpt, networkID, config.CloudMqttClient, &config.CloudAuth, &tls.Config{InsecureSkipVerify: true})
	cloudMqttClient := paho_mqtt.NewWrapper(mqtt.NewClient(cloudMqttClientOpt), time.Duration(config.CloudMqttClient.WaitTimeout))
	cloudMqttClientPubF := func(topic string, data []byte) error {
		return cloudMqttClient.Publish(topic, config.CloudMqttClient.PublishQOSLevel, false, data)
	}

	message_hdl.DeviceEventMaxAge = time.Duration(config.RelayHandler.MaxDeviceEventAge)
	message_hdl.DeviceCommandIDPrefix = fmt.Sprintf("%s_%s_", srvInfoHdl.GetName(), config.MGWDeploymentID)
	message_hdl.DeviceCommandMaxAge = time.Duration(config.RelayHandler.MaxDeviceCmdAge)
	message_hdl.NetworkID = networkID
	message_hdl.UserID = userID
	message_hdl.LocalDeviceIDPrefix = config.LocalDeviceHandler.IDPrefix

	deviceCmdMsgRelayHdl := msg_relay_hdl.New(config.RelayHandler.MessageBuffer, message_hdl.HandleDownstreamDeviceCmd, localMqttClientPubF)
	processesCmdMsgRelayHdl := msg_relay_hdl.New(config.RelayHandler.MessageBuffer, message_hdl.HandleDownstreamProcessesCmd, localMqttClientPubF)
	deviceEventMsgRelayHdl := msg_relay_hdl.New(config.RelayHandler.EventMessageBuffer, message_hdl.HandleUpstreamDeviceEvent, cloudMqttClientPubF)
	deviceCmdRespMsgRelayHdl := msg_relay_hdl.New(config.RelayHandler.MessageBuffer, message_hdl.HandleUpstreamDeviceCmdResponse, cloudMqttClientPubF)
	processesStateMsgRelayHdl := msg_relay_hdl.New(config.RelayHandler.MessageBuffer, message_hdl.HandleUpstreamProcessesState, cloudMqttClientPubF)
	deviceConnectorErrMsgRelayHdl := msg_relay_hdl.New(config.RelayHandler.MessageBuffer, message_hdl.HandlerUpstreamDeviceConnectorErr, cloudMqttClientPubF)
	deviceErrMsgRelayHdl := msg_relay_hdl.New(config.RelayHandler.MessageBuffer, message_hdl.HandlerUpstreamDeviceErr, cloudMqttClientPubF)
	deviceCmdErrMsgRelayHdl := msg_relay_hdl.New(config.RelayHandler.MessageBuffer, message_hdl.HandlerUpstreamDeviceCmdErr, cloudMqttClientPubF)

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

	localDeviceHdl.SetDeviceSyncFunc(cloudHdl.Sync)
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
	wtchdg.RegisterHealthFunc(cloudHdl.HasNetwork)
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

	ec = wtchdg.Join()
}
