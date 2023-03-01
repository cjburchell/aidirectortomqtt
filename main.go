package main

import (
	"fmt"
	"github.com/cjburchell/tools-go/env"
	logger "github.com/cjburchell/uatu-go"
	"github.com/eclipse/paho.mqtt.golang"
	"gitlab.com/cjburchell/aidirectortomqtt/AquaIllumination"
	"gitlab.com/cjburchell/aidirectortomqtt/aimqtt"
	appSettings "gitlab.com/cjburchell/aidirectortomqtt/settings"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)
import "github.com/cjburchell/settings-go"

func main() {
	log := logger.Create(logger.Settings{
		MinLogLevel:  logger.INFO,
		ServiceName:  "AI Director to MQTT",
		LogToConsole: true,
		UseHTTP:      false,
		UsePubSub:    false,
	})

	fmt.Printf("This is a test")
	set := settings.Get(env.Get("ConfigFile", ""))
	appConfig := appSettings.Get(set)
	mqttOptions := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("mqtt://%s:%d", appConfig.MqttHost, appConfig.MqttPort)).SetClientID("ai-mqtt")
	mqttOptions.SetOrderMatters(false)       // Allow out of order messages (use this option unless in order delivery is essential)
	mqttOptions.ConnectTimeout = time.Second // Minimal delays on connect
	mqttOptions.WriteTimeout = time.Second   // Minimal delays on writes
	mqttOptions.KeepAlive = 10               // Keepalive every 10 seconds so we quickly detect network outages
	mqttOptions.PingTimeout = time.Second    // local broker so response should be quick
	mqttOptions.ConnectRetry = true
	mqttOptions.AutoReconnect = true

	mqttOptions.OnConnectionLost = func(cl mqtt.Client, err error) {
		log.Warn("connection lost")
	}
	mqttOptions.OnConnect = func(mqtt.Client) {
		log.Print("connection established")
	}
	mqttOptions.OnReconnecting = func(mqtt.Client, *mqtt.ClientOptions) {
		log.Print("attempting to reconnect")
	}

	mqttClient := mqtt.NewClient(mqttOptions)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	log.Print("Connection is up")
	defer mqttClient.Disconnect(100)

	var client http.Client
	var queue aimqtt.AiMqtt

	mqttClient.SubscribeMultiple(map[string]byte{
		"aimqtt/+/tank/+/group/+/device/+/manualModeCommand": 1}, func(_ mqtt.Client, message mqtt.Message) {
		tokens := strings.Split(message.Topic(), "/")
		groupId, _ := strconv.Atoi(tokens[5])
		AquaIllumination.SetManualMode(groupId, string(message.Payload()) == "ON", client, log, appConfig.DirectorHost)
		var result, err = AquaIllumination.GetAll(client, appConfig.DirectorHost)
		if err != nil {
			log.Errorf(err, "Unable to update")
		} else {
			queue.UpdateHomeAssistant(*result, mqttClient, log, false)
		}
	})

	mqttClient.SubscribeMultiple(map[string]byte{
		"aimqtt/+/tank/+/group/+/device/+/+/toggle": 1}, func(_ mqtt.Client, message mqtt.Message) {
		tokens := strings.Split(message.Topic(), "/")
		groupId, _ := strconv.Atoi(tokens[5])
		colorId := tokens[8]
		AquaIllumination.ToggleLight(groupId, colorId, string(message.Payload()) == "ON", client, log, appConfig.DirectorHost)
		var result, err = AquaIllumination.GetAll(client, appConfig.DirectorHost)
		if err != nil {
			log.Errorf(err, "Unable to update")
		} else {
			queue.UpdateHomeAssistant(*result, mqttClient, log, false)
		}
	})

	mqttClient.SubscribeMultiple(map[string]byte{
		"aimqtt/+/tank/+/group/+/device/+/+/setintensity": 1}, func(_ mqtt.Client, message mqtt.Message) {
		tokens := strings.Split(message.Topic(), "/")
		groupId, _ := strconv.Atoi(tokens[5])
		colorId := tokens[8]
		AquaIllumination.SetIntensity(groupId, colorId, string(message.Payload()), client, log, appConfig.DirectorHost)
		var result, err = AquaIllumination.GetAll(client, appConfig.DirectorHost)
		if err != nil {
			log.Errorf(err, "Unable to update")
		} else {
			queue.UpdateHomeAssistant(*result, mqttClient, log, false)
		}
	})

	var result, err = AquaIllumination.GetAll(client, appConfig.DirectorHost)
	if err != nil {
		log.Errorf(err, "Unable to update")
	} else {
		queue.UpdateMQTT(*result, mqttClient, log, true)
		queue.UpdateHomeAssistant(*result, mqttClient, log, true)
	}

	go RunUpdateConfig(client, mqttClient, log, appConfig, &queue)
	RunUpdate(client, mqttClient, log, appConfig, &queue)
}

const logRate = time.Second * 1
const logAllRate = time.Minute * 1

func RunUpdateConfig(client http.Client, mqttClient mqtt.Client, log logger.ILog, config appSettings.Config, queue *aimqtt.AiMqtt) {
	c := make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier are not blocked
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case <-c:
			queue.PublishMQTT(mqttClient, log, "status", "offline", true)
			log.Debug("Exit Application")
			return
		case <-time.After(logAllRate):
			log.Print("Updating Config")
			var result, err = AquaIllumination.GetAll(client, config.DirectorHost)
			if err != nil {
				log.Errorf(err, "Unable to update")
			} else {
				queue.UpdateMQTT(*result, mqttClient, log, true)
				queue.UpdateHomeAssistant(*result, mqttClient, log, true)
			}
		}
	}
}

func RunUpdate(client http.Client, mqttClient mqtt.Client, log logger.ILog, config appSettings.Config, queue *aimqtt.AiMqtt) {
	c := make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier are not blocked
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case <-c:
			log.Debug("Exit Application")
			return
		case <-time.After(logRate):
			log.Print("Updating State")
			var result, err = AquaIllumination.GetAll(client, config.DirectorHost)
			if err != nil {
				log.Errorf(err, "Unable to update state")
			} else {
				queue.UpdateMQTT(*result, mqttClient, log, false)
			}
		}
	}
}
