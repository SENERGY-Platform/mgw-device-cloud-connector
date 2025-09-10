package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	sb_logger "github.com/SENERGY-Platform/go-service-base/logger"
	"github.com/SENERGY-Platform/mgw-cloud-proxy/cert-manager/lib/client"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler/cloud_hdl"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler/cloud_mqtt_hdl"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler/local_device_hdl"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler/local_mqtt_hdl"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler/message_hdl"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler/msg_relay_hdl"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler/persistent_msg_relay_hdl"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler/persistent_msg_relay_hdl/sqlite"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/cloud_client"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/dm_client"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/paho_mqtt"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/topic"
	"github.com/SENERGY-Platform/mgw-go-service-base/srv-info-hdl"
	sb_util "github.com/SENERGY-Platform/mgw-go-service-base/util"
	"github.com/SENERGY-Platform/mgw-go-service-base/watchdog"
	"github.com/eclipse/paho.mqtt.golang"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
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
		var logFileError *sb_logger.LogFileError
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

	cloudClient := cloud_client.New(cloudHttpClient, config.HttpClient.CloudApiBaseUrl)
	certManagerClient := client.New(localHttpClient, config.HttpClient.LocalCmBaseUrl)
	cloudHdl := cloud_hdl.New(cloudClient, certManagerClient, time.Duration(config.CloudHandler.SyncInterval), config.CloudHandler.AttributeOrigin)

	wtchdg.Start()

	chCtx, cf := context.WithCancel(context.Background())
	defer cf()
	wtchdg.RegisterStopFunc(func() error {
		cf()
		return nil
	})

	networkID, userID, err := cloudHdl.Init(chCtx, time.Duration(config.CloudHandler.NetworkInitDelay))
	if err != nil {
		util.Logger.Error(err)
		ec = 1
		return
	}

	topic.InitTopicHandler(userID, networkID)

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

	cloudMqttHdl := cloud_mqtt_hdl.New(config.CloudMqttClient.SubscribeQOSLevel)

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
	paho_mqtt.SetCloudClientOptions(cloudMqttClientOpt, networkID, config.CloudMqttClient)
	cloudMqttClient := paho_mqtt.NewWrapper(mqtt.NewClient(cloudMqttClientOpt), time.Duration(config.CloudMqttClient.WaitTimeout))
	cloudMqttClientPubF := func(topic string, data []byte) error {
		return cloudMqttClient.Publish(topic, config.CloudMqttClient.PublishQOSLevel, false, data)
	}

	msgRelayHdlCtx, msgRelayHdlCf := context.WithCancel(context.Background())
	var deviceEventMessageRelayHdl handler.MessageRelayHandler
	switch config.RelayHandler.EventMessageHandlerType {
	case util.EventMsgRelayHdlInMemory:
		util.Logger.Infof("event message relay handler: %s", util.EventMsgRelayHdlInMemory)
		message_hdl.DeviceEventMaxAge = time.Duration(config.RelayHandler.MaxDeviceEventAge)
		deviceEventMsgRelayHdl := msg_relay_hdl.New(config.RelayHandler.EventMessageBuffer, message_hdl.HandleUpstreamDeviceEventAgeLimit, cloudMqttClientPubF)
		deviceEventMsgRelayHdl.Start()
		wtchdg.RegisterStopFunc(func() error {
			deviceEventMsgRelayHdl.Stop()
			return nil
		})
		deviceEventMessageRelayHdl = deviceEventMsgRelayHdl
	case util.EventMsgRelayHdlPersistent:
		util.Logger.Infof("event message relay handler: %s", util.EventMsgRelayHdlPersistent)
		size, err := util.ParseSize(config.RelayHandler.EventMessagePersistentStorageSize)
		if err != nil {
			util.Logger.Error(err)
			ec = 1
			return
		}
		storageHdl, err := sqlite.New(path.Join(config.RelayHandler.EventMessagePersistentWorkspacePath, "messages.db"))
		if err != nil {
			util.Logger.Error(err)
			ec = 1
			return
		}
		defer storageHdl.Close()
		err = storageHdl.Init(msgRelayHdlCtx, size)
		if err != nil {
			util.Logger.Error(err)
			ec = 1
			return
		}
		storageHdl.PeriodicOptimization(msgRelayHdlCtx, time.Hour)
		wtchdg.RegisterStopFunc(func() error {
			storageHdl.Stop()
			return nil
		})
		deviceEventPersistentMsgRelayHdl := persistent_msg_relay_hdl.New(
			config.RelayHandler.EventMessagePersistentWorkspacePath,
			config.RelayHandler.EventMessageBuffer,
			storageHdl,
			message_hdl.HandleUpstreamDeviceEvent,
			cloudMqttClientPubF,
			config.RelayHandler.EventMessagePersistentReadLimit,
		)
		err = deviceEventPersistentMsgRelayHdl.Init(msgRelayHdlCtx)
		if err != nil {
			util.Logger.Error(err)
			ec = 1
			return
		}
		deviceEventPersistentMsgRelayHdl.Start(msgRelayHdlCtx)
		wtchdg.RegisterHealthFunc(deviceEventPersistentMsgRelayHdl.Running)
		wtchdg.RegisterStopFunc(func() error {
			deviceEventPersistentMsgRelayHdl.Stop(time.Minute)
			return nil
		})
		deviceEventMessageRelayHdl = deviceEventPersistentMsgRelayHdl
	default:
		util.Logger.Errorf("unknown event message relay handler: %s", config.RelayHandler.EventMessageHandlerType)
		ec = 1
		return
	}

	message_hdl.DeviceCommandIDPrefix = fmt.Sprintf("%s_%s_", srvInfoHdl.GetName(), config.MGWDeploymentID)
	message_hdl.DeviceCommandMaxAge = time.Duration(config.RelayHandler.MaxDeviceCmdAge)
	message_hdl.LocalDeviceIDPrefix = config.LocalDeviceHandler.IDPrefix

	deviceCmdMsgRelayHdl := msg_relay_hdl.New(config.RelayHandler.MessageBuffer, message_hdl.HandleDownstreamDeviceCmd, localMqttClientPubF)
	processesCmdMsgRelayHdl := msg_relay_hdl.New(config.RelayHandler.MessageBuffer, message_hdl.HandleDownstreamProcessesCmd, localMqttClientPubF)
	deviceCmdRespMsgRelayHdl := msg_relay_hdl.New(config.RelayHandler.MessageBuffer, message_hdl.HandleUpstreamDeviceCmdResponse, cloudMqttClientPubF)
	processesStateMsgRelayHdl := msg_relay_hdl.New(config.RelayHandler.MessageBuffer, message_hdl.HandleUpstreamProcessesState, cloudMqttClientPubF)
	deviceConnectorErrMsgRelayHdl := msg_relay_hdl.New(config.RelayHandler.MessageBuffer, message_hdl.HandlerUpstreamDeviceConnectorErr, cloudMqttClientPubF)
	deviceErrMsgRelayHdl := msg_relay_hdl.New(config.RelayHandler.MessageBuffer, message_hdl.HandlerUpstreamDeviceErr, cloudMqttClientPubF)
	deviceCmdErrMsgRelayHdl := msg_relay_hdl.New(config.RelayHandler.MessageBuffer, message_hdl.HandlerUpstreamDeviceCmdErr, cloudMqttClientPubF)

	localMqttHdl.SetMqttClient(localMqttClient)
	localMqttHdl.SetMessageRelayHdl(
		deviceEventMessageRelayHdl,
		deviceCmdRespMsgRelayHdl,
		processesStateMsgRelayHdl,
		deviceConnectorErrMsgRelayHdl,
		deviceErrMsgRelayHdl,
		deviceCmdErrMsgRelayHdl)
	cloudMqttHdl.SetMqttClient(cloudMqttClient)
	cloudMqttHdl.SetMessageRelayHdl(deviceCmdMsgRelayHdl, processesCmdMsgRelayHdl)

	localDeviceHdl.SetDeviceSyncFunc(cloudHdl.Sync)
	cloudHdl.SetDeviceStateSyncFunc(cloudMqttHdl.HandleSubscriptions)

	localDeviceHdl.Start()

	deviceCmdMsgRelayHdl.Start()
	processesCmdMsgRelayHdl.Start()
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
		msgRelayHdlCf()
		localMqttClient.Disconnect(1000)
		cloudMqttClient.Disconnect(1000)
		return nil
	})
	wtchdg.RegisterStopFunc(func() error {
		deviceCmdMsgRelayHdl.Stop()
		processesCmdMsgRelayHdl.Stop()
		deviceCmdRespMsgRelayHdl.Stop()
		processesStateMsgRelayHdl.Stop()
		deviceConnectorErrMsgRelayHdl.Stop()
		deviceErrMsgRelayHdl.Stop()
		deviceCmdErrMsgRelayHdl.Stop()
		return nil
	})

	ec = wtchdg.Join()
}
