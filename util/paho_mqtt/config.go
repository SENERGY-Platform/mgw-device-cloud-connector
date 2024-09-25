package paho_mqtt

import (
	"crypto/tls"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/eclipse/paho.mqtt.golang"
	"time"
)

func SetLocalClientOptions(co *mqtt.ClientOptions, clientID string, mqttConf util.LocalMqttClientConfig) {
	co.AddBroker(mqttConf.Server)
	co.SetClientID(clientID)
	co.SetKeepAlive(time.Duration(mqttConf.KeepAlive))
	co.SetPingTimeout(time.Duration(mqttConf.PingTimeout))
	co.SetConnectTimeout(time.Duration(mqttConf.ConnectTimeout))
	co.SetConnectRetryInterval(time.Duration(mqttConf.ConnectRetryDelay))
	co.SetMaxReconnectInterval(time.Duration(mqttConf.MaxReconnectDelay))
	co.SetWriteTimeout(time.Second * 5)
	co.ConnectRetry = true
	co.AutoReconnect = true
}

func SetCloudClientOptions(co *mqtt.ClientOptions, clientID string, mqttConf util.CloudMqttClientConfig, authConf *util.CloudAuthConfig, tlsConf *tls.Config) {
	co.AddBroker(mqttConf.Server)
	co.SetClientID(clientID)
	co.SetKeepAlive(time.Duration(mqttConf.KeepAlive))
	co.SetPingTimeout(time.Duration(mqttConf.PingTimeout))
	co.SetConnectTimeout(time.Duration(mqttConf.ConnectTimeout))
	co.SetConnectRetryInterval(time.Duration(mqttConf.ConnectRetryDelay))
	co.SetMaxReconnectInterval(time.Duration(mqttConf.MaxReconnectDelay))
	co.SetWriteTimeout(time.Second * 5)
	co.ConnectRetry = true
	co.AutoReconnect = true
	if authConf != nil {
		co.SetUsername(authConf.User)
		co.SetPassword(authConf.Password.Value())
	}
	if tlsConf == nil {
		tlsConf = &tls.Config{}
	}
	tlsConf.VerifyConnection = checkPeerCertValidityBounds
	co.SetTLSConfig(tlsConf)
}
