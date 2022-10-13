package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/lawl/pulseaudio"
)

var (
	mqttBroker = flag.String("mqtt_broker",
		"tcp://scotty-the-fourth.fritz.box:1883",
		"MQTT broker address for github.com/eclipse/paho.mqtt.golang")

	mqttPrefix = flag.String("mqtt_topic",
		"github.com/stapelberg/defaultsink2mqtt/",
		"MQTT topic prefix")
)

func credentialProvider() (username string, password string) {
	password_path := "heuselfamily/mqtt/"
	username = "defaultsink2mqtt"
	out, err := exec.Command("/home/chris/.local/bin/gopass", "show", "--password", password_path+username).Output()
	if err != nil {
		log.Fatal(err)
	}
	password = string(out)
	return username, password
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

	client, err := pulseaudio.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	updates, err := client.Updates()
	if err != nil {
		return err
	}
	func() {
		var lastDefaultSink string
		for ; ; <-updates {
			info, err := client.ServerInfo()
			if err != nil {
				log.Printf("ServerInfo: %v", err)
				continue
			}
			if info.DefaultSink != lastDefaultSink {
				log.Printf("default sink changed from %s to %s", lastDefaultSink, info.DefaultSink)

				mqttClient.Publish(
					*mqttPrefix+"default_sink",
					0,    /* qos */
					true, /* retained */
					string(info.DefaultSink))

				lastDefaultSink = info.DefaultSink
			}
		}
	}()

	return nil
}

func main() {
	flag.Parse()
	if err := defaultsink2mqtt(); err != nil {
		log.Fatal(err)
	}
}
