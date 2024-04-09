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

type CloudMqttClientConfig struct {
	Server            string `json:"server" env_var:"CLOUD_MQTT_SERVER"`
	KeepAlive         int64  `json:"keep_alive" env_var:"CLOUD_MQTT_KEEP_ALIVE"`
	PingTimeout       int64  `json:"ping_timeout" env_var:"CLOUD_MQTT_PING_TIMEOUT"`
	ConnectTimeout    int64  `json:"connect_timeout" env_var:"CLOUD_MQTT_CONNECT_TIMEOUT"`
	ConnectRetryDelay int64  `json:"connect_retry_delay" env_var:"CLOUD_MQTT_CONNECT_RETRY_DELAY"`
	MaxReconnectDelay int64  `json:"max_reconnect_delay" env_var:"CLOUD_MQTT_MAX_RECONNECT_DELAY"`
	WaitTimeout       int64  `json:"wait_timeout" env_var:"CLOUD_MQTT_WAIT_TIMEOUT"`
	QOSLevel          byte   `json:"qos_level" env_var:"CLOUD_MQTT_QOS_LEVEL"`
}

type LocalMqttClientConfig struct {
	Server            string `json:"server" env_var:"LOCAL_MQTT_SERVER"`
	KeepAlive         int64  `json:"keep_alive" env_var:"LOCAL_MQTT_KEEP_ALIVE"`
	PingTimeout       int64  `json:"ping_timeout" env_var:"LOCAL_MQTT_PING_TIMEOUT"`
	ConnectTimeout    int64  `json:"connect_timeout" env_var:"LOCAL_MQTT_CONNECT_TIMEOUT"`
	ConnectRetryDelay int64  `json:"connect_retry_delay" env_var:"LOCAL_MQTT_CONNECT_RETRY_DELAY"`
	MaxReconnectDelay int64  `json:"max_reconnect_delay" env_var:"LOCAL_MQTT_MAX_RECONNECT_DELAY"`
	WaitTimeout       int64  `json:"wait_timeout" env_var:"LOCAL_MQTT_WAIT_TIMEOUT"`
	QOSLevel          byte   `json:"qos_level" env_var:"LOCAL_MQTT_QOS_LEVEL"`
}

type HttpClientConfig struct {
	LocalDmBaseUrl   string `json:"local_dm_base_url" env_var:"LOCAL_DM_BASE_URL"`
	CloudApiBaseUrl  string `json:"cloud_api_base_url" env_var:"CLOUD_API_BASE_URL"`
	CloudAuthBaseUrl string `json:"cloud_auth_base_url" env_var:"CLOUD_AUTH_BASE_URL"`
	LocalTimeout     int64  `json:"local_timeout" env_var:"HTTP_LOCAL_TIMEOUT"`
	CloudTimeout     int64  `json:"cloud_timeout" env_var:"HTTP_CLOUD_TIMEOUT"`
}

type CloudAuthConfig struct {
	User     string               `json:"user" env_var:"CLOUD_USER"`
	Password sb_util.SecretString `json:"password" env_var:"CLOUD_PASSWORD"`
	ClientID string               `json:"client_id" env_var:"CLOUD_CLIENT_ID"`
}

type CloudDeviceHandlerConfig struct {
	HubID           string `json:"hub_id" env_var:"CDH_HUB_ID"`
	DefaultHubName  string `json:"default_hub_name" env_var:"CDH_DEFAULT_HUB_NAME"`
	WrkSpcPath      string `json:"wrk_spc_path" env_var:"CDH_WRK_SPC_PATH"`
	AttributeOrigin string `json:"attribute_origin" env_var:"CDH_ATTRIBUTE_ORIGIN"`
	SyncInterval    int64  `json:"sync_interval" env_var:"CDH_SYNC_INTERVAL"`
}

type LocalDeviceHandlerConfig struct {
	IDPrefix      string `json:"id_prefix" env_var:"LDH_ID_PREFIX"`
	QueryInterval int64  `json:"query_interval" env_var:"LDH_QUERY_INTERVAL"`
}

type Config struct {
	Logger                  sb_util.LoggerConfig     `json:"logger" env_var:"LOGGER_CONFIG"`
	CloudMqttClient         CloudMqttClientConfig    `json:"cloud_mqtt_client" env_var:"CLOUD_MQTT_CLIENT_CONFIG"`
	LocalMqttClient         LocalMqttClientConfig    `json:"local_mqtt_client" env_var:"LOCAL_MQTT_CLIENT_CONFIG"`
	HttpClient              HttpClientConfig         `json:"http_client" env_var:"HTTP_CLIENT_CONFIG"`
	CloudAuth               CloudAuthConfig          `json:"cloud_auth" env_var:"CLOUD_AUTH_CONFIG"`
	CloudDeviceHandler      CloudDeviceHandlerConfig `json:"cloud_device_handler" env_var:"CLOUD_DEVICE_HANDLER_CONFIG"`
	LocalDeviceHandler      LocalDeviceHandlerConfig `json:"local_device_handler" env_var:"LOCAL_DEVICE_HANDLER_CONFIG"`
	MessageRelayBuffer      int                      `json:"message_relay_buffer" env_var:"MESSAGE_RELAY_BUFFER"`
	EventMessageRelayBuffer int                      `json:"event_message_relay_buffer" env_var:"EVENT_MESSAGE_RELAY_BUFFER"`
	MGWDeploymentID         string                   `json:"mgw_deployment_id" env_var:"MGW_DID"`
	MaxDeviceCmdAge         int64                    `json:"max_device_cmd_age" env_var:"MAX_DEVICE_CMD_AGE"`
	MQTTLog                 bool                     `json:"mqtt_log" env_var:"MQTT_LOG"`
	MQTTDebugLog            bool                     `json:"mqtt_debug_log" env_var:"MQTT_DEBUG_LOG"`
}

var defaultCloudMqttClientConfig = CloudMqttClientConfig{
	KeepAlive:         30000000000, // 30s
	PingTimeout:       10000000000, // 10s
	ConnectTimeout:    30000000000, // 30s
	ConnectRetryDelay: 30000000000, // 30s
	MaxReconnectDelay: 60000000000, // 1m
	WaitTimeout:       5000000000,  // 5s
	QOSLevel:          2,
}

var defaultLocalMqttClientConfig = LocalMqttClientConfig{
	KeepAlive:         30000000000, // 30s
	PingTimeout:       10000000000, // 10s
	ConnectTimeout:    30000000000, // 30s
	ConnectRetryDelay: 30000000000, // 30s
	MaxReconnectDelay: 30000000000, // 30s
	WaitTimeout:       5000000000,  // 5s
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
		CloudMqttClient: defaultCloudMqttClientConfig,
		LocalMqttClient: defaultLocalMqttClientConfig,
		HttpClient: HttpClientConfig{
			LocalTimeout: 10000000000, // 10s
			CloudTimeout: 30000000000, // 30s
		},
		CloudDeviceHandler: CloudDeviceHandlerConfig{
			WrkSpcPath:      "/opt/connector/cdh-data",
			AttributeOrigin: "dcc",
			SyncInterval:    1800000000000, // 30m
		},
		LocalDeviceHandler: LocalDeviceHandlerConfig{
			QueryInterval: 5000000000, // 5s
		},
		MessageRelayBuffer:      2500,
		EventMessageRelayBuffer: 5000,
		MaxDeviceCmdAge:         60000000000, // 60s
		MQTTLog:                 true,
	}
	err := sb_util.LoadConfig(path, &cfg, nil, nil, nil)
	return &cfg, err
}
