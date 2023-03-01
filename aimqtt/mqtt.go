package aimqtt

import (
	"encoding/json"
	"fmt"
	logger "github.com/cjburchell/uatu-go"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"gitlab.com/cjburchell/aidirectortomqtt/AquaIllumination"
	"image/color"
	"strings"
)

func sanitize(text string) string {

	newText := strings.Replace(text, " ", "_", -1)
	newText = strings.Replace(newText, "/", "_", -1)
	newText = strings.Replace(newText, ".", "_", -1)
	newText = strings.Replace(newText, "&", "_", -1)
	return newText
}

type AiMqtt struct {
	data map[string]string
}

func (profiMqtt *AiMqtt) PublishMQTTOld(mqttClient mqtt.Client, log logger.ILog, topic string) {
	fullTopic := fmt.Sprintf("aimqtt/%s", topic)
	if profiMqtt.data == nil {
		return
	} else {
		_, ok := profiMqtt.data[fullTopic]
		if !ok {
			return
		}
	}

	t := mqttClient.Publish(fullTopic, 1, false, profiMqtt.data[fullTopic])
	// Handle the token in a go routine so this loop keeps sending messages regardless of delivery status
	go func() {
		_ = t.Wait() // Can also use '<-t.Done()' in releases > 1.2.0
		if t.Error() != nil {
			log.Warnf("ERROR PUBLISHING aimqtt/%s", fullTopic)
		}
	}()
}

func (profiMqtt *AiMqtt) PublishMQTT(mqttClient mqtt.Client, log logger.ILog, topic string, payload string, forceUpdate bool) {
	fullTopic := fmt.Sprintf("aimqtt/%s", topic)
	if profiMqtt.data == nil {
		profiMqtt.data = make(map[string]string)
	} else {
		if profiMqtt.data[fullTopic] == payload && !forceUpdate {
			return
		}
	}
	profiMqtt.data[fullTopic] = payload

	t := mqttClient.Publish(fullTopic, 1, false, payload)
	// Handle the token in a go routine so this loop keeps sending messages regardless of delivery status
	go func() {
		<-t.Done()
		if t.Error() != nil {
			log.Warnf("ERROR PUBLISHING aimqtt/%s", fullTopic)
		}
	}()
}

func (profiMqtt *AiMqtt) publishHA(mqttClient mqtt.Client, log logger.ILog, platform string, device string, topic string, payload []byte, forceUpdate bool) {
	fullTopic := fmt.Sprintf("homeassistant/%s/%s/%s/config", platform, device, topic)
	if profiMqtt.data == nil {
		profiMqtt.data = make(map[string]string)
	} else {
		if profiMqtt.data[fullTopic] == string(payload) && !forceUpdate {
			return
		}
	}
	profiMqtt.data[fullTopic] = string(payload)

	t := mqttClient.Publish(fullTopic, 1, false, payload)
	// Handle the token in a go routine so this loop keeps sending messages regardless of delivery status
	go func() {
		_ = t.Wait() // Can also use '<-t.Done()' in releases > 1.2.0
		if t.Error() != nil {
			log.Warnf("ERROR PUBLISHING %s", fullTopic)
		}
	}()
}

type Device struct {
	Identifiers  string `json:"identifiers"`
	Name         string `json:"name"`
	Model        string `json:"model"`
	Manufacturer string `json:"manufacturer"`
	Version      string `json:"hw_version"`
}

type HaBaseConfig struct {
	Device              Device `json:"device"`
	Name                string `json:"name"`
	UniqueId            string `json:"unique_id"`
	AvailabilityTopic   string `json:"availability_topic,omitempty"`
	DeviceClass         string `json:"device_class,omitempty"`
	PayloadAvailable    string `json:"payload_available"`
	PayloadNotAvailable string `json:"payload_not_available"`
	Icon                string `json:"icon,omitempty"`
}

type HaStateConfig struct {
	HaBaseConfig
	StateTopic        string `json:"state_topic"`
	UnitOfMeasurement string `json:"unit_of_measurement,omitempty"`
}

type HaButtonConfig struct {
	HaBaseConfig
	CommandTopic string `json:"command_topic"`
}

type HaSwitchConfig struct {
	HaBaseConfig
	StateTopic   string `json:"state_topic"`
	CommandTopic string `json:"command_topic"`
}

type HaLightConfig struct {
	HaBaseConfig
	StateTopic             string `json:"state_topic"`
	CommandTopic           string `json:"command_topic"`
	OnCommandType          string `json:"on_command_type,omitempty"`
	BrightnessStateTopic   string `json:"brightness_state_topic,omitempty"`
	BrightnessCommandTopic string `json:"brightness_command_topic,omitempty"`
	PayloadOff             string `json:"payload_off,omitempty"`
	PayloadOn              string `json:"payload_on,omitempty"`
	RgbStateTopic          string `json:"rgb_state_topic,omitempty"`
	RgbCommandTopic        string `json:"rgb_command_topic,omitempty"`
	BrightnessScale        string `json:"brightness_scale,omitempty"`
}

func (profiMqtt *AiMqtt) UpdateHomeAssistant(aiDirector AquaIllumination.Director, mqttClient mqtt.Client, log logger.ILog, forceUpdate bool) {
	for _, tank := range aiDirector.Tank {
		for _, group := range tank.Groups {
			for _, device := range aiDirector.Devices {
				if device.GroupId != group.GroupId {
					continue
				}

				deviceModel := "Unknown"
				if device.DeviceId < 60000000 {
					switch device.Model {
					case 0:
						deviceModel = "Sol White"
					case 1:
						deviceModel = "Sol Blue"
					case 2:
						deviceModel = "Nano"
					}
				} else if device.DeviceId < 118000000 {
					deviceModel = "Hydra"
				} else if device.DeviceId < 119000000 {
					deviceModel = "Hydra 52"
				}

				deviceName := fmt.Sprintf("%s %s %X", tank.TankName, deviceModel, device.DeviceId)
				deviceId := fmt.Sprintf("%d", device.DeviceId)

				deviceHASS := Device{
					Identifiers:  deviceId,
					Version:      aiDirector.Version,
					Name:         deviceName,
					Model:        deviceModel,
					Manufacturer: "Aqua Illumination",
				}

				deviceTopic := fmt.Sprintf("aimqtt/%s/tank/%d/group/%d/device/%d", sanitize(aiDirector.Name), tank.TankId, group.GroupId, device.DeviceId)

				fanConfig := HaStateConfig{
					HaBaseConfig: HaBaseConfig{
						Device:              deviceHASS,
						Name:                fmt.Sprintf("%s Fan", deviceName),
						UniqueId:            strings.ToLower(fmt.Sprintf("%d_fan", device.DeviceId)),
						AvailabilityTopic:   "aimqtt/status",
						PayloadAvailable:    "online",
						PayloadNotAvailable: "offline",
						Icon:                "mdi:fan",
					},
					StateTopic: deviceTopic + "/fanspeed",
				}

				modeMsg, _ := json.Marshal(fanConfig)
				profiMqtt.publishHA(mqttClient, log, "binary_sensor", deviceId, "Fan", modeMsg, forceUpdate)

				signalConfig := HaStateConfig{
					HaBaseConfig: HaBaseConfig{
						Device:              deviceHASS,
						Name:                fmt.Sprintf("%s Signal Strength", deviceName),
						UniqueId:            strings.ToLower(fmt.Sprintf("%d_signal", device.DeviceId)),
						AvailabilityTopic:   "aimqtt/status",
						PayloadAvailable:    "online",
						PayloadNotAvailable: "offline",
						Icon:                "mdi:signal",
					},
					StateTopic:        deviceTopic + "/signal",
					UnitOfMeasurement: "%",
				}

				msg, _ := json.Marshal(signalConfig)
				profiMqtt.publishHA(mqttClient, log, "sensor", deviceId, "Signal", msg, forceUpdate)

				modeConfig := HaSwitchConfig{
					HaBaseConfig: HaBaseConfig{
						Device:              deviceHASS,
						Name:                fmt.Sprintf("%s Manual Mode", deviceName),
						UniqueId:            strings.ToLower(fmt.Sprintf("%d_mode", device.DeviceId)),
						AvailabilityTopic:   "aimqtt/status",
						PayloadAvailable:    "online",
						PayloadNotAvailable: "offline",
						Icon:                "mdi:wrench",
					},
					StateTopic:   deviceTopic + "/manualMode",
					CommandTopic: deviceTopic + "/manualModeCommand",
				}

				msg, _ = json.Marshal(modeConfig)
				profiMqtt.publishHA(mqttClient, log, "switch", deviceId, "ManualMode", msg, forceUpdate)

				for _, color := range aiDirector.GroupColors[group.GroupId].Colors {
					profiMqtt.PublishMQTT(mqttClient, log, fmt.Sprintf("%s/%d/device/%d/%s", sanitize(aiDirector.Name), tank.TankId, device.DeviceId, color.Color), fmt.Sprintf("%d", color.Intensity), forceUpdate)

					lightConfig := HaLightConfig{
						HaBaseConfig: HaBaseConfig{
							Device:              deviceHASS,
							Name:                fmt.Sprintf("%s %s", deviceName, strings.Title(strings.Replace(color.Color, "_", " ", -1))),
							UniqueId:            strings.ToLower(fmt.Sprintf("%d_%s_light", device.DeviceId, sanitize(color.Color))),
							AvailabilityTopic:   "aimqtt/status",
							PayloadAvailable:    "online",
							PayloadNotAvailable: "offline",
						},
						StateTopic:             deviceTopic + "/" + color.Color + "/state",
						CommandTopic:           deviceTopic + "/" + color.Color + "/toggle",
						PayloadOff:             "OFF",
						BrightnessStateTopic:   deviceTopic + "/" + color.Color + "/intensity",
						BrightnessCommandTopic: deviceTopic + "/" + color.Color + "/setintensity",
						OnCommandType:          "brightness",
						RgbStateTopic:          deviceTopic + "/" + color.Color + "/color",
						BrightnessScale:        "100",
					}

					msg, _ = json.Marshal(lightConfig)
					profiMqtt.publishHA(mqttClient, log, "light", deviceId, color.Color, msg, forceUpdate)
				}

			}
		}
	}
}

func boolToOnOff(value bool) string {
	if value {
		return "ON"
	}

	return "OFF"
}

var colorMap = map[string]string{
	"uv":         "#eb00ff",
	"violet":     "#9a00e2",
	"deep_blue":  "#313b9d",
	"royal":      "#0084ff",
	"green":      "#9cd11e",
	"deep_red":   "#a60f24",
	"cool_white": "#c6cccc",
	"blue":       "#bfe6f0",
}

func rgbToHexColor(color string) string {
	rgb := strings.Split(color, ",")
	return fmt.Sprintf("#%02x%02x%02x", rgb[0], rgb[1], rgb[3])
}

func ParseHexColor(s string) (c color.RGBA, err error) {
	c.A = 0xff
	switch len(s) {
	case 7:
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits:
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default:
		err = fmt.Errorf("invalid length, must be 7 or 4")

	}
	return
}

func hexColorToRgb(hexColor string) string {
	rgb, _ := ParseHexColor(hexColor)
	return fmt.Sprintf("%d,%d,%d", rgb.R, rgb.G, rgb.B)
}

func (profiMqtt *AiMqtt) UpdateMQTT(aiDirector AquaIllumination.Director, mqttClient mqtt.Client, log logger.ILog, forceUpdate bool) {
	profiMqtt.PublishMQTT(mqttClient, log, "status", "online", forceUpdate)
	for _, tank := range aiDirector.Tank {
		for _, group := range tank.Groups {
			for _, device := range aiDirector.Devices {
				if device.GroupId != group.GroupId {
					continue
				}

				deviceTopic := fmt.Sprintf("%s/tank/%d/group/%d/device/%d", sanitize(aiDirector.Name), tank.TankId, group.GroupId, device.DeviceId)

				if device.FanSpeed == "unknown" {
					if forceUpdate {
						profiMqtt.PublishMQTTOld(mqttClient, log, fmt.Sprintf("%s/fanspeed", deviceTopic))
					}
				} else {
					profiMqtt.PublishMQTT(mqttClient, log, fmt.Sprintf("%s/fanspeed", deviceTopic), strings.ToUpper(device.FanSpeed), forceUpdate)
				}

				profiMqtt.PublishMQTT(mqttClient, log, fmt.Sprintf("%s/signal", deviceTopic), fmt.Sprintf("%d", aiDirector.DeviceStats[device.DeviceId].CommPercent), forceUpdate)
				profiMqtt.PublishMQTT(mqttClient, log, fmt.Sprintf("%s/manualMode", deviceTopic), boolToOnOff(!aiDirector.GroupMode[device.GroupId]), forceUpdate)

				for _, color := range aiDirector.GroupColors[group.GroupId].Colors {
					profiMqtt.PublishMQTT(mqttClient, log, fmt.Sprintf("%s/%s/intensity", deviceTopic, color.Color), fmt.Sprintf("%d", color.Intensity), forceUpdate)
					profiMqtt.PublishMQTT(mqttClient, log, fmt.Sprintf("%s/%s/state", deviceTopic, color.Color), boolToOnOff(color.Intensity != 0), forceUpdate)
					profiMqtt.PublishMQTT(mqttClient, log, fmt.Sprintf("%s/%s/color", deviceTopic, color.Color), hexColorToRgb(colorMap[color.Color]), forceUpdate)
				}

			}
		}
	}
}
