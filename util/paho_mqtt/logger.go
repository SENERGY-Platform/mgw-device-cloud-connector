package paho_mqtt

import (
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/eclipse/paho.mqtt.golang"
)

type mqttLogger struct {
	println func(v ...any)
	printf  func(format string, v ...any)
}

func (l *mqttLogger) Println(v ...any) {
	l.println(v...)
}

func (l *mqttLogger) Printf(format string, v ...any) {
	l.printf(format, v...)
}

func SetLogger(debug bool) {
	mqtt.ERROR = &mqttLogger{
		println: util.Logger.Error,
		printf:  util.Logger.Errorf,
	}
	mqtt.CRITICAL = &mqttLogger{
		println: util.Logger.Error,
		printf:  util.Logger.Errorf,
	}
	mqtt.WARN = &mqttLogger{
		println: util.Logger.Warning,
		printf:  util.Logger.Warningf,
	}
	if debug {
		mqtt.DEBUG = &mqttLogger{
			println: util.Logger.Debug,
			printf:  util.Logger.Debugf,
		}
	}
}
