//go:build linux

package main

func getUpdates(callback notificationCallback) error {

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
				callback.Notify(info.DefaultSink)
				lastDefaultSink = info.DefaultSink
			}
		}
	}()

}
