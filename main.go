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
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler/local_device_hdl"
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

	paho_mqtt.SetLogger()

	localMqttClientOpt := mqtt.NewClientOptions()
	paho_mqtt.SetClientOptions(localMqttClientOpt, fmt.Sprintf("%s_%s", srvInfoHdl.GetName(), config.MGWDeploymentID), config.UpstreamMqttClient, nil, nil)
	localMqttClient := mqtt.NewClient(localMqttClientOpt)

	cloudMqttClientOpt := mqtt.NewClientOptions()
	paho_mqtt.SetClientOptions(cloudMqttClientOpt, fmt.Sprintf("%s_%s", srvInfoHdl.GetName(), config.MGWDeploymentID), config.UpstreamMqttClient, &config.Auth, &tls.Config{InsecureSkipVerify: true})
	cloudMqttClient := mqtt.NewClient(cloudMqttClientOpt)

	dmClient := dm_client.New(http.DefaultClient, config.HttpClient.DmBaseUrl)

	localDeviceHdl := local_device_hdl.New(dmClient, time.Duration(config.HttpClient.Timeout), time.Duration(config.DeviceQueryInterval))

	cloudClient := cloud_client.New(http.DefaultClient, config.HttpClient.CloudBaseUrl, auth_client.New(http.DefaultClient, config.HttpClient.AuthBaseUrl, config.Auth.User, config.Auth.Password.String(), config.Auth.ClientID))

	cloudDeviceHdl := cloud_device_hdl.New(cloudClient, time.Duration(config.HttpClient.CloudTimeout), config.CloudHandler.WrkSpcPath, config.CloudHandler.AttributeOrigin)

	localDeviceHdl.SetSyncFunc(cloudDeviceHdl.Sync)
	localDeviceHdl.SetStateFunc(cloudDeviceHdl.UpdateStates)

	chCtx, cf := context.WithCancel(context.Background())
	defer cf()
	if err = cloudDeviceHdl.Init(chCtx, config.CloudHandler.HubID, config.CloudHandler.DefaultHubName); err != nil {
		util.Logger.Error(err)
		ec = 1
		return
	}

	ldhCtx, cf := context.WithCancel(context.Background())
	defer cf()
	if err = localDeviceHdl.RefreshDevices(ldhCtx); err != nil {
		util.Logger.Errorf("initial local device refresh failed: %s", err)
	}
	localDeviceHdl.Start()

	localMqttClient.Connect()
	cloudMqttClient.Connect()

	wtchdg.RegisterHealthFunc(localDeviceHdl.Running)
	wtchdg.RegisterStopFunc(func() error {
		localDeviceHdl.Stop()
		return nil
	})
	wtchdg.RegisterStopFunc(func() error {
		cloudMqttClient.Disconnect(500)
		return nil
	})
	wtchdg.RegisterStopFunc(func() error {
		localMqttClient.Disconnect(500)
		return nil
	})

	wtchdg.Start()

	ec = wtchdg.Join()
}
