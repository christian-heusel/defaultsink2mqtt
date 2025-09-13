package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	mqttBroker = flag.String("mqtt_broker",
		"tcp://scotty-the-fifth.lan:1883",
		"MQTT broker address for github.com/eclipse/paho.mqtt.golang")

	mqttPrefix = flag.String("mqtt_topic",
		"github.com/stapelberg/defaultsink2mqtt/",
		"MQTT topic prefix")

	gopassPath = flag.String("gopass_path",
		"/home/chris/.local/bin/gopass",
		"Path to the gopass executable")
)

func credentialProvider() (username string, password string) {
	password_path := "heuselfamily/mqtt/"
	username = "defaultsink2mqtt"
	out, err := exec.Command(*gopassPath, "show", "--password", password_path+username).Output()
	if err != nil {
		log.Fatal("Error in gopass cmd:", err)
	}
	password = string(out)
	return username, password
}

type NotificationCallback struct {
	mqttClient mqtt.Client
}

func (callback *NotificationCallback) Notify(sinkName string) {
	log.Printf("MQTT: %sdefault_sink -> %q", *mqttPrefix, sinkName)
	callback.mqttClient.Publish(
		*mqttPrefix+"default_sink",
		0,    /* qos */
		true, /* retained */
		sinkName)

}

func defaultsink2mqtt() error {
	opts := mqtt.NewClientOptions().AddBroker(*mqttBroker)
	clientID := "https://github.com/stapelberg/defaultsink2mqtt"
	if hostname, err := os.Hostname(); err == nil {
		clientID += "@" + hostname
	}
	opts.SetClientID(clientID)
	opts.SetCredentialsProvider(credentialProvider)
	opts.SetConnectRetry(true)
	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("MQTT connection failed: %v", token.Error())
	}
	log.Printf("Successfully registered on %q", *mqttBroker)

	if err := getUpdates(NotificationCallback{mqttClient}); err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Parse()
	if err := defaultsink2mqtt(); err != nil {
		log.Fatal(err)
	}
}
