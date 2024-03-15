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
	PublishTimeout    int64  `json:"publish_timeout" env_var:"MQTT_PUBLISH_TIMEOUT"`
}

type HttpClientConfig struct {
	DmBaseUrl    string `json:"dm_base_url" env_var:"DM_BASE_URL"`
	AuthBaseUrl  string `json:"auth_base_url" env_var:"AUTH_BASE_URL"`
	Timeout      int64  `json:"timeout" env_var:"HTTP_TIMEOUT"`
	CloudTimeout int64  `json:"cloud_timeout" env_var:"HTTP_CLOUD_TIMEOUT"`
}

type AuthConfig struct {
	User     string               `json:"user" env_var:"USER"`
	Password sb_util.SecretString `json:"password" env_var:"PASSWORD"`
	ClientID string               `json:"client_id" env_var:"CLIENT_ID"`
}

type CloudHandlerConfig struct {
	HubID           string `json:"hub_id" env_var:"CH_HUB_ID"`
	WrkSpcPath      string `json:"wrk_spc_path" env_var:"CH_WRK_SPC_PATH"`
	AttributeOrigin string `json:"attribute_origin" env_var:"CH_ATTRIBUTE_ORIGIN"`
}

type Config struct {
	Logger               sb_util.LoggerConfig `json:"logger" env_var:"LOGGER_CONFIG"`
	UpstreamMqttClient   MqttClientConfig     `json:"upstream_mqtt_client" env_var:"UPSTREAM_MQTT_CLIENT_CONFIG"`
	DownstreamMqttClient MqttClientConfig     `json:"downstream_mqtt_client" env_var:"DOWNSTREAM_MQTT_CLIENT_CONFIG"`
	HttpClient           HttpClientConfig     `json:"http_client" env_var:"HTTP_CLIENT_CONFIG"`
	Auth                 AuthConfig           `json:"auth" env_var:"AUTH_CONFIG"`
	CloudHandler         CloudHandlerConfig   `json:"cloud_handler" env_var:"CLOUD_HANDLER_CONFIG"`
	DeviceQueryInterval  int64                `json:"device_query_interval" env_var:"DEVICE_QUERY_INTERVAL"`
	MGWDeploymentID      string               `json:"mgw_deployment_id" env_var:"MGW_DID"`
}

var defaultMqttClientConfig = MqttClientConfig{
	KeepAlive:         30,
	PingTimeout:       15000000000,  // 15s
	ConnectTimeout:    30000000000,  // 30s
	ConnectRetryDelay: 30000000000,  // 30s
	MaxReconnectDelay: 300000000000, // 5m
	PublishTimeout:    0,
}

func NewConfig(path string) (*Config, error) {
	cfg := Config{
		Logger: sb_util.LoggerConfig{
			Level:        level.Warning,
			Utc:          true,
			Microseconds: true,
			Terminal:     true,
		},
		UpstreamMqttClient:   defaultMqttClientConfig,
		DownstreamMqttClient: defaultMqttClientConfig,
		HttpClient: HttpClientConfig{
			Timeout:      10000000000,
			CloudTimeout: 30000000000,
		},
		CloudHandler: CloudHandlerConfig{
			WrkSpcPath:      "/opt/device-cloud-connector",
			AttributeOrigin: "device-cloud-connector",
		},
	}
	err := sb_util.LoadConfig(path, &cfg, nil, nil, nil)
	return &cfg, err
}
