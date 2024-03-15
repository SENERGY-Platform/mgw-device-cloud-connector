package paho_mqtt

import (
	"crypto/tls"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/eclipse/paho.mqtt.golang"
	"time"
)

func SetClientOptions(co *mqtt.ClientOptions, clientID string, mqttConf util.MqttClientConfig, authConf *util.AuthConfig, tlsConf *tls.Config) {
	co.AddBroker(mqttConf.Server)
	co.SetClientID(clientID)
	co.SetKeepAlive(time.Duration(mqttConf.KeepAlive))
	co.SetPingTimeout(time.Duration(mqttConf.PingTimeout))
	co.SetConnectTimeout(time.Duration(mqttConf.ConnectTimeout))
	co.SetConnectRetryInterval(time.Duration(mqttConf.ConnectRetryDelay))
	co.SetMaxReconnectInterval(time.Duration(mqttConf.MaxReconnectDelay))
	co.SetWriteTimeout(time.Duration(mqttConf.PublishTimeout))
	co.ConnectRetry = true
	co.AutoReconnect = true
	if authConf != nil {
		co.SetUsername(authConf.User)
		co.SetPassword(authConf.Password.String())
	}
	if tlsConf != nil {
		co.SetTLSConfig(tlsConf)
	}
}
