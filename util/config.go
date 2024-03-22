/*
 * Copyright 2024 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package util

import (
	sb_util "github.com/SENERGY-Platform/go-service-base/util"
	"github.com/y-du/go-log-level/level"
)

type MqttClientConfig struct {
	Server            string `json:"server" env_var:"MQTT_SERVER"`
	KeepAlive         int64  `json:"keep_alive" env_var:"MQTT_KEEP_ALIVE"`
	PingTimeout       int64  `json:"ping_timeout" env_var:"MQTT_PING_TIMEOUT"`
	ConnectTimeout    int64  `json:"connect_timeout" env_var:"MQTT_CONNECT_TIMEOUT"`
	ConnectRetryDelay int64  `json:"connect_retry_delay" env_var:"MQTT_CONNECT_RETRY_DELAY"`
	MaxReconnectDelay int64  `json:"max_reconnect_delay" env_var:"MQTT_MAX_RECONNECT_DELAY"`
	WaitTimeout       int64  `json:"wait_timeout" env_var:"MQTT_WAIT_TIMEOUT"`
	QOSLevel          byte   `json:"qos_level" env_var:"MQTT_QOS_LEVEL"`
}

type HttpClientConfig struct {
	CloudBaseUrl string `json:"cloud_base_url" env_var:"CLOUD_BASE_URL"`
	DmBaseUrl    string `json:"dm_base_url" env_var:"DM_BASE_URL"`
	AuthBaseUrl  string `json:"auth_base_url" env_var:"AUTH_BASE_URL"`
	LocalTimeout int64  `json:"local_timeout" env_var:"HTTP_LOCAL_TIMEOUT"`
	CloudTimeout int64  `json:"cloud_timeout" env_var:"HTTP_CLOUD_TIMEOUT"`
}

type AuthConfig struct {
	User     string               `json:"user" env_var:"USER"`
	Password sb_util.SecretString `json:"password" env_var:"PASSWORD"`
	ClientID string               `json:"client_id" env_var:"CLIENT_ID"`
}

type CloudDeviceHandlerConfig struct {
	HubID           string `json:"hub_id" env_var:"CDH_HUB_ID"`
	DefaultHubName  string `json:"default_hub_name" env_var:"CDH_DEFAULT_HUB_NAME"`
	WrkSpcPath      string `json:"wrk_spc_path" env_var:"CDH_WRK_SPC_PATH"`
	AttributeOrigin string `json:"attribute_origin" env_var:"CDH_ATTRIBUTE_ORIGIN"`
}

type LocalDeviceHandlerConfig struct {
	IDPrefix      string `json:"id_prefix" env_var:"LDH_ID_PREFIX"`
	QueryInterval int64  `json:"query_interval" env_var:"LDH_QUERY_INTERVAL"`
}

type Config struct {
	Logger             sb_util.LoggerConfig     `json:"logger" env_var:"LOGGER_CONFIG"`
	CloudMqttClient    MqttClientConfig         `json:"upstream_mqtt_client" env_var:"UPSTREAM_MQTT_CLIENT_CONFIG"`
	LocalMqttClient    MqttClientConfig         `json:"downstream_mqtt_client" env_var:"DOWNSTREAM_MQTT_CLIENT_CONFIG"`
	HttpClient         HttpClientConfig         `json:"http_client" env_var:"HTTP_CLIENT_CONFIG"`
	Auth               AuthConfig               `json:"auth" env_var:"AUTH_CONFIG"`
	CloudDeviceHandler CloudDeviceHandlerConfig `json:"cloud_device_handler" env_var:"CLOUD_DEVICE_HANDLER_CONFIG"`
	LocalDeviceHandler LocalDeviceHandlerConfig `json:"local_device_handler" env_var:"LOCAL_DEVICE_HANDLER_CONFIG"`
	MessageRelayBuffer int                      `json:"message_relay_buffer" env_var:"MESSAGE_RELAY_BUFFER"`
	MGWDeploymentID    string                   `json:"mgw_deployment_id" env_var:"MGW_DID"`
}

var defaultMqttClientConfig = MqttClientConfig{
	KeepAlive:         30,
	PingTimeout:       15000000000,  // 15s
	ConnectTimeout:    30000000000,  // 30s
	ConnectRetryDelay: 30000000000,  // 30s
	MaxReconnectDelay: 300000000000, // 5m
	WaitTimeout:       5000000000,   // 5s
	QOSLevel:          2,
}

func NewConfig(path string) (*Config, error) {
	cfg := Config{
		Logger: sb_util.LoggerConfig{
			Level:        level.Warning,
			Utc:          true,
			Microseconds: true,
			Terminal:     true,
		},
		CloudMqttClient: defaultMqttClientConfig,
		LocalMqttClient: defaultMqttClientConfig,
		HttpClient: HttpClientConfig{
			LocalTimeout: 10000000000, // 10s
			CloudTimeout: 30000000000, // 30s
		},
		CloudDeviceHandler: CloudDeviceHandlerConfig{
			WrkSpcPath:      "/opt/device-cloud-connector",
			AttributeOrigin: "device-cloud-connector",
		},
		MessageRelayBuffer: 5000,
	}
	err := sb_util.LoadConfig(path, &cfg, nil, nil, nil)
	return &cfg, err
}
