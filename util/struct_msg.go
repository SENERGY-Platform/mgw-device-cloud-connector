/*
 * Copyright 2025 InfAI (CC SES)
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
	"time"

	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
)

type Message struct {
	TopicVal   string
	PayloadVal string
	TimeVal    time.Time
}

func (m Message) Topic() string {
	return m.TopicVal
}

func (m Message) Payload() []byte {
	return []byte(m.PayloadVal)
}

func (m Message) Timestamp() time.Time {
	return m.TimeVal
}

func MessageFromHandler(msg handler.Message) Message {
	return Message{
		TopicVal:   msg.Topic(),
		PayloadVal: string(msg.Payload()),
		TimeVal:    msg.Timestamp(),
	}
}
