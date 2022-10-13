//go:build linux

package main

import (
	"log"

	"github.com/lawl/pulseaudio"
)

func getUpdates(callback NotificationCallback) error {

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
				log.Printf("default sink changed from %q to %q", lastDefaultSink, info.DefaultSink)
				callback.Notify(info.DefaultSink)
				lastDefaultSink = info.DefaultSink
			}
		}
	}()
	return nil
}
