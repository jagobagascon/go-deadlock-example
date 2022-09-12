// nolint
package main

import (
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/saltosystems/winrt-go"
	"github.com/saltosystems/winrt-go/windows/devices/bluetooth/advertisement"
	"github.com/saltosystems/winrt-go/windows/foundation"
)

func main() {
	if err := runExample(); err != nil {
		panic(err)
	}
}

func runExample() error {
	ole.RoInitialize(1)

	instDelegate := foundation.NewTypedEventHandler(ole.NewGUID(winrt.ParameterizedInstanceGUID(
		foundation.GUIDTypedEventHandler,
		advertisement.SignatureBluetoothLEAdvertisementWatcher,
		advertisement.SignatureBluetoothLEAdvertisementReceivedEventArgs,
	)), func(_ *foundation.TypedEventHandler, sender, argsPtr unsafe.Pointer) {
		args := (*advertisement.BluetoothLEAdvertisementReceivedEventArgs)(argsPtr)

		addr, _ := args.GetBluetoothAddress()
		println("Ble device: ", addr)
	})
	defer instDelegate.Release()

	watcher, err := advertisement.NewBluetoothLEAdvertisementWatcher()
	if err != nil {
		return err
	}
	defer func() {
		_ = watcher.Release()
	}()

	token, err := watcher.AddReceived(instDelegate)
	if err != nil {
		return err
	}
	defer watcher.RemoveReceived(token)

	// Wait for when advertisement has stopped by a call to StopScan().
	// Advertisement doesn't seem to stop right away, there is an
	// intermediate Stopping state.
	stoppingChan := make(chan struct{})
	// TypedEventHandler<BluetoothLEAdvertisementWatcher, BluetoothLEAdvertisementWatcherStoppedEventArgs>
	eventStoppedGUID := winrt.ParameterizedInstanceGUID(
		foundation.GUIDTypedEventHandler,
		advertisement.SignatureBluetoothLEAdvertisementWatcher,
		advertisement.SignatureBluetoothLEAdvertisementWatcherStoppedEventArgs,
	)
	stoppedHandler := foundation.NewTypedEventHandler(ole.NewGUID(eventStoppedGUID), func(_ *foundation.TypedEventHandler, _, _ unsafe.Pointer) {
		// Note: the args parameter has an Error property that should
		// probably be checked, but I'm not sure when stopping the
		// advertisement watcher could ever result in an error (except
		// for bugs).
		close(stoppingChan)
	})
	defer stoppedHandler.Release()

	token, err = watcher.AddStopped(stoppedHandler)
	if err != nil {
		return err
	}
	defer watcher.RemoveStopped(token)

	err = watcher.Start()
	if err != nil {
		return err
	}

	// Wait until advertisement has stopped, and finish.
	<-stoppingChan

	return nil
}
