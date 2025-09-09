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
	"github.com/SENERGY-Platform/go-service-base/config-hdl"
	cfg_types "github.com/SENERGY-Platform/go-service-base/config-hdl/types"
	sb_logger "github.com/SENERGY-Platform/go-service-base/logger"
	envldr "github.com/y-du/go-env-loader"
	"github.com/y-du/go-log-level/level"
	"reflect"
)

type CloudMqttClientConfig struct {
	Server            string `json:"server" env_var:"CLOUD_MQTT_SERVER"`
	KeepAlive         int64  `json:"keep_alive" env_var:"CLOUD_MQTT_KEEP_ALIVE"`
	PingTimeout       int64  `json:"ping_timeout" env_var:"CLOUD_MQTT_PING_TIMEOUT"`
	ConnectTimeout    int64  `json:"connect_timeout" env_var:"CLOUD_MQTT_CONNECT_TIMEOUT"`
	ConnectRetryDelay int64  `json:"connect_retry_delay" env_var:"CLOUD_MQTT_CONNECT_RETRY_DELAY"`
	MaxReconnectDelay int64  `json:"max_reconnect_delay" env_var:"CLOUD_MQTT_MAX_RECONNECT_DELAY"`
	WaitTimeout       int64  `json:"wait_timeout" env_var:"CLOUD_MQTT_WAIT_TIMEOUT"`
	PublishQOSLevel   byte   `json:"publish_qos_level" env_var:"CLOUD_PUBLISH_QOS_LEVEL"`
	SubscribeQOSLevel byte   `json:"subscribe_qos_level" env_var:"CLOUD_SUBSCRIBE_QOS_LEVEL"`
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
	LocalMmBaseUrl   string `json:"local_mm_base_url" env_var:"LOCAL_MM_BASE_URL"`
	LocalDmBaseUrl   string `json:"local_dm_base_url" env_var:"LOCAL_DM_BASE_URL"`
	CloudApiBaseUrl  string `json:"cloud_api_base_url" env_var:"CLOUD_API_BASE_URL"`
	CloudAuthBaseUrl string `json:"cloud_auth_base_url" env_var:"CLOUD_AUTH_BASE_URL"`
	LocalTimeout     int64  `json:"local_timeout" env_var:"HTTP_LOCAL_TIMEOUT"`
	CloudTimeout     int64  `json:"cloud_timeout" env_var:"HTTP_CLOUD_TIMEOUT"`
}

type CloudAuthConfig struct {
	User     string           `json:"user" env_var:"CLOUD_USER"`
	Password cfg_types.Secret `json:"password" env_var:"CLOUD_PASSWORD"`
	ClientID string           `json:"client_id" env_var:"CLOUD_CLIENT_ID"`
}

type CloudHandlerConfig struct {
	NetworkID          string `json:"network_id" env_var:"CH_NETWORK_ID"`
	DefaultNetworkName string `json:"default_network_name" env_var:"CH_DEFAULT_NETWORK_NAME"`
	WrkSpcPath         string `json:"wrk_spc_path" env_var:"CH_WRK_SPC_PATH"`
	AttributeOrigin    string `json:"attribute_origin" env_var:"CH_ATTRIBUTE_ORIGIN"`
	SyncInterval       int64  `json:"sync_interval" env_var:"CH_SYNC_INTERVAL"`
	NetworkInitDelay   int64  `json:"network_init_delay" env_var:"CH_NETWORK_INIT_DELAY"`
}

type LocalDeviceHandlerConfig struct {
	IDPrefix        string `json:"id_prefix" env_var:"LDH_ID_PREFIX"`
	RefreshInterval int64  `json:"refresh_interval" env_var:"LDH_REFRESH_INTERVAL"`
}

type RelayHandlerConfig struct {
	MessageBuffer                       int    `json:"message_buffer" env_var:"RH_MESSAGE_BUFFER"`
	EventMessageBuffer                  int    `json:"event_message_buffer" env_var:"RH_EVENT_MESSAGE_BUFFER"`
	EventMessagePersistent              bool   `json:"event_message_persistent" env_var:"RH_EVENT_MESSAGE_PERSISTENT"`
	EventMessagePersistentWorkspacePath string `json:"event_message_persistent_workspace_path" env_var:"RH_EVENT_MESSAGE_PERSISTENT_WORKSPACE_PATH"`
	EventMessagePersistentStorageSize   string `json:"event_message_persistent_storage_size" env_var:"RH_EVENT_MESSAGE_PERSISTENT_STORAGE_SIZE"`
	EventMessagePersistentReadLimit     int    `json:"event_message_persistent_read_limit" env_var:"RH_EVENT_MESSAGE_PERSISTENT_READ_LIMIT"`
	MaxDeviceCmdAge                     int64  `json:"max_device_cmd_age" env_var:"RH_MAX_DEVICE_CMD_AGE"`
	MaxDeviceEventAge                   int64  `json:"max_device_event_age" env_var:"RH_MAX_DEVICE_EVENT_AGE"`
}

type LoggerConfig struct {
	Level        level.Level `json:"level" env_var:"LOGGER_LEVEL"`
	Utc          bool        `json:"utc" env_var:"LOGGER_UTC"`
	Path         string      `json:"path" env_var:"LOGGER_PATH"`
	FileName     string      `json:"file_name" env_var:"LOGGER_FILE_NAME"`
	Terminal     bool        `json:"terminal" env_var:"LOGGER_TERMINAL"`
	Microseconds bool        `json:"microseconds" env_var:"LOGGER_MICROSECONDS"`
	Prefix       string      `json:"prefix" env_var:"LOGGER_PREFIX"`
}

type Config struct {
	Logger             LoggerConfig             `json:"logger" env_var:"LOGGER_CONFIG"`
	CloudMqttClient    CloudMqttClientConfig    `json:"cloud_mqtt_client" env_var:"CLOUD_MQTT_CLIENT_CONFIG"`
	LocalMqttClient    LocalMqttClientConfig    `json:"local_mqtt_client" env_var:"LOCAL_MQTT_CLIENT_CONFIG"`
	HttpClient         HttpClientConfig         `json:"http_client" env_var:"HTTP_CLIENT_CONFIG"`
	CloudAuth          CloudAuthConfig          `json:"cloud_auth" env_var:"CLOUD_AUTH_CONFIG"`
	CloudHandler       CloudHandlerConfig       `json:"cloud_handler" env_var:"CLOUD_HANDLER_CONFIG"`
	LocalDeviceHandler LocalDeviceHandlerConfig `json:"local_device_handler" env_var:"LOCAL_DEVICE_HANDLER_CONFIG"`
	RelayHandler       RelayHandlerConfig       `json:"relay_handler" env_var:"RELAY_HANDLER_CONFIG"`
	MGWDeploymentID    string                   `json:"mgw_deployment_id" env_var:"MGW_DID"`
	MQTTLog            bool                     `json:"mqtt_log" env_var:"MQTT_LOG"`
	MQTTDebugLog       bool                     `json:"mqtt_debug_log" env_var:"MQTT_DEBUG_LOG"`
}

var defaultCloudMqttClientConfig = CloudMqttClientConfig{
	KeepAlive:         30000000000, // 30s
	PingTimeout:       10000000000, // 10s
	ConnectTimeout:    30000000000, // 30s
	ConnectRetryDelay: 30000000000, // 30s
	MaxReconnectDelay: 60000000000, // 1m
	WaitTimeout:       30000000000, // 30s
	PublishQOSLevel:   0,
	SubscribeQOSLevel: 2,
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
		Logger: LoggerConfig{
			Level:        level.Warning,
			Utc:          true,
			Microseconds: true,
			Terminal:     true,
		},
		CloudMqttClient: defaultCloudMqttClientConfig,
		LocalMqttClient: defaultLocalMqttClientConfig,
		HttpClient: HttpClientConfig{
			LocalMmBaseUrl: "http://module-manager",
			LocalTimeout:   10000000000, // 10s
			CloudTimeout:   30000000000, // 30s
		},
		CloudHandler: CloudHandlerConfig{
			WrkSpcPath:       "/opt/connector/ch-data",
			AttributeOrigin:  "dcc",
			SyncInterval:     1800000000000, // 30m
			NetworkInitDelay: 10000000000,   // 10s
		},
		LocalDeviceHandler: LocalDeviceHandlerConfig{
			RefreshInterval: 5000000000, // 5s
		},
		RelayHandler: RelayHandlerConfig{
			MessageBuffer:                       50000,
			EventMessageBuffer:                  100000,
			EventMessagePersistentWorkspacePath: "/opt/connector/event-data",
			EventMessagePersistentStorageSize:   "4GB",
			EventMessagePersistentReadLimit:     100,
			MaxDeviceCmdAge:                     30000000000,  // 30s
			MaxDeviceEventAge:                   300000000000, // 5m
		},
	}
	err := config_hdl.Load(&cfg, nil, map[reflect.Type]envldr.Parser{reflect.TypeOf(level.Off): sb_logger.LevelParser}, nil, path)
	return &cfg, err
}
