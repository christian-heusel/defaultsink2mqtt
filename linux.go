//go:build linux

package main

import (
	"log"
	"time"

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
			tmp, err := client.ServerInfo()
			if err != nil {
				log.Printf("ServerInfo: %v", err)
				continue
			}

			time.Sleep(200 * time.Millisecond)

			info, err := client.ServerInfo()
			if err != nil {
				log.Printf("ServerInfo: %v", err)
				continue
			}

			// Skip short changes, i.e. when just turning on
			if tmp.DefaultSink != info.DefaultSink {
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
