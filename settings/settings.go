package settings

import (
	"github.com/cjburchell/settings-go"
)

func Get(settings settings.ISettings) Config {
	return Config{
		MqttHost:     settings.Get("MQTT_HOST", "localhost"),
		MqttPort:     settings.GetInt("MQTT_PORT", 1883),
		MqttUser:     settings.Get("MQTT_USER", ""),
		MqttPassword: settings.Get("MQTT_PASSWORD", ""),
		DirectorHost: settings.Get("DIRECTOR_HOST", ""),
	}
}

type Config struct {
	MqttHost     string
	MqttPort     int
	MqttUser     string
	MqttPassword string
	DirectorHost string
}
