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

func SetLogger() {
	mqtt.ERROR = &mqttLogger{
		println: util.Logger.Errorln,
		printf:  util.Logger.Errorf,
	}
	mqtt.CRITICAL = &mqttLogger{
		println: util.Logger.Errorln,
		printf:  util.Logger.Errorf,
	}
	mqtt.WARN = &mqttLogger{
		println: util.Logger.Warningln,
		printf:  util.Logger.Warningf,
	}
	mqtt.DEBUG = &mqttLogger{
		println: util.Logger.Debugln,
		printf:  util.Logger.Debugf,
	}
}
