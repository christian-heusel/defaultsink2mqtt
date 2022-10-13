//go:build windows

package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	ole "github.com/go-ole/go-ole"

	"github.com/moutend/go-wca/pkg/wca"
)

func getUpdates(callback NotificationCallback) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		return err
	}

	defer ole.CoUninitialize()

	var mmde *wca.IMMDeviceEnumerator

	if err := wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &mmde); err != nil {
		return err
	}

	defer mmde.Release()

	cw := callbackWrapper{callback, ""}
	wcaCallbacks := wca.IMMNotificationClientCallback{
		OnDefaultDeviceChanged: cw.onDefaultDeviceChanged,
	}

	mmnc := wca.NewIMMNotificationClient(wcaCallbacks)

	if err := mmde.RegisterEndpointNotificationCallback(mmnc); err != nil {
		return err
	}

	select {
	case <-quit:
		fmt.Println("Received keyboard interrupt.")
	case <-time.After(5 * time.Minute):
		fmt.Println("Received timeout signal")
	}
	fmt.Println("Done")
	return nil
}

type callbackWrapper struct {
	notificationCallback NotificationCallback
	lastDefaultSink      string
}

func (c *callbackWrapper) onDefaultDeviceChanged(flow wca.EDataFlow, role wca.ERole, pwstrDeviceId string) error {
	err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
	if err != nil {
		return err
	}
	defer ole.CoUninitialize()

	var mmde *wca.IMMDeviceEnumerator
	if err = wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &mmde); err != nil {
		return err
	}
	defer mmde.Release()

	var mmd *wca.IMMDevice
	if err = mmde.GetDefaultAudioEndpoint(wca.ERender, wca.EConsole, &mmd); err != nil {
		return err
	}
	defer mmd.Release()

	var ps *wca.IPropertyStore
	if err = mmd.OpenPropertyStore(wca.STGM_READ, &ps); err != nil {
		return err
	}
	defer ps.Release()

	var pv wca.PROPVARIANT
	if err = ps.GetValue(&wca.PKEY_Device_FriendlyName, &pv); err != nil {
		return err
	}
	sink := pv.String()

	if sink != c.lastDefaultSink {
		c.notificationCallback.Notify(sink)
		c.lastDefaultSink = sink
	}
	return nil
}
