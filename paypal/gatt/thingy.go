// +build

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
)

var done = make(chan struct{})

func onStateChanged(d gatt.Device, s gatt.State) {
	fmt.Println("State:", s)
	switch s {
	case gatt.StatePoweredOn:
		fmt.Println("Scanning...")
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	id := strings.ToUpper(flag.Args()[0])
	if strings.ToUpper(p.ID()) != id {
		return
	}

	// Stop scanning once we've got the peripheral we're looking for.
	p.Device().StopScanning()

	fmt.Printf("\nPeripheral ID:%s, NAME:(%s)\n", p.ID(), p.Name())
	fmt.Println("  Local Name        =", a.LocalName)
	fmt.Println("  TX Power Level    =", a.TxPowerLevel)
	fmt.Println("  Manufacturer Data =", a.ManufacturerData)
	fmt.Println("  Service Data      =", a.ServiceData)
	fmt.Println("")

	p.Device().Connect(p)
}

func onPeriphConnected(p gatt.Peripheral, err error) {
	fmt.Println("Connected")
	defer p.Device().CancelConnection(p)

	if err := p.SetMTU(500); err != nil {
		fmt.Printf("Failed to set MTU, err: %s\n", err)
	}

	// Discovery services
	ss, err := p.DiscoverServices(nil)
	if err != nil {
		fmt.Printf("Failed to discover services, err: %s\n", err)
		return
	}
	var ledChar *gatt.Characteristic
	for _, s := range ss {
		msg := "Service: " + s.UUID().String()
		if len(s.Name()) > 0 {
			msg += " (" + s.Name() + ")"
		}
		fmt.Println(msg)

		// Discovery characteristics
		cs, err := p.DiscoverCharacteristics(nil, s)
		if err != nil {
			fmt.Printf("Failed to discover characteristics, err: %s\n", err)
			continue
		}

		for _, c := range cs {
			msg := "  Characteristic  " + c.UUID().String()
			if "ef6803019b3549339b1052ffa9740042" == c.UUID().String() {
				ledChar = c
				msg += "\n!!!!    get LED Characteristic    "
				err = p.WriteCharacteristic(ledChar, []byte("\x01\xff\x00\x00"), true)
				if err != nil {
					fmt.Printf("Failed to discover descriptors, err: %s\n", err)
				}
				b, err := p.ReadLongCharacteristic(ledChar)
				if err != nil {
					fmt.Printf("Failed to read characteristic, err: %s\n", err)
				}
				msg += fmt.Sprintf("!!!!LED  value         %x | %q\n", b, b)

			}
			if len(c.Name()) > 0 {
				msg += " (" + c.Name() + ")"
			}
			msg += "\n    properties    " + c.Properties().String()
			fmt.Println(msg)

			// Read the characteristic, if possible.
			if (c.Properties() & gatt.CharRead) != 0 {
				b, err := p.ReadCharacteristic(c)
				if err != nil {
					fmt.Printf("Failed to read characteristic, err: %s\n", err)
					continue
				}
				fmt.Printf("    value         %x | %q\n", b, b)
			}

			// Discovery descriptors
			ds, err := p.DiscoverDescriptors(nil, c)
			if err != nil {
				fmt.Printf("Failed to discover descriptors, err: %s\n", err)
				continue
			}

			for _, d := range ds {
				msg := "  Descriptor      " + d.UUID().String()
				if len(d.Name()) > 0 {
					msg += " (" + d.Name() + ")"
				}
				fmt.Println(msg)

				// Read descriptor (could fail, if it's not readable)
				b, err := p.ReadDescriptor(d)
				if err != nil {
					fmt.Printf("Failed to read descriptor, err: %s\n", err)
					continue
				}
				fmt.Printf("    value         %x | %q\n", b, b)
			}

		}
		fmt.Println()
	}

	fmt.Printf("Waiting for 5 seconds to get some notifiations, if any.\n")
	err = p.WriteCharacteristic(ledChar, []byte("\x01\xff\x00\x00"), true)
	if err != nil {
		fmt.Printf("Failed to discover descriptors, err: %s\n", err)
	}
	b, err := p.ReadLongCharacteristic(ledChar)
	if err != nil {
		fmt.Printf("Failed to read characteristic, err: %s\n", err)
	}
	fmt.Printf("LED  value         %x | %q\n", b, b)
	time.Sleep(5 * time.Second)

	err = p.WriteCharacteristic(ledChar, []byte("\x01\x00\xff\x00"), true)
	if err != nil {
		fmt.Printf("Failed to discover descriptors, err: %s\n", err)
	}
	b, err = p.ReadLongCharacteristic(ledChar)
	if err != nil {
		fmt.Printf("Failed to read characteristic, err: %s\n", err)
	}
	fmt.Printf("LED  value         %x | %q\n", b, b)
	time.Sleep(2 * time.Second)

	err = p.WriteCharacteristic(ledChar, []byte("\x01\x00\x00\xff"), true)
	if err != nil {
		fmt.Printf("Failed to discover descriptors, err: %s\n", err)
	}
	b, err = p.ReadLongCharacteristic(ledChar)
	if err != nil {
		fmt.Printf("Failed to read characteristic, err: %s\n", err)
	}
	fmt.Printf("LED  value         %x | %q\n", b, b)
	time.Sleep(2 * time.Second)
}

func onPeriphDisconnected(p gatt.Peripheral, err error) {
	fmt.Println("Disconnected")
	close(done)
}

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		log.Fatalf("usage: %s [options] peripheral-id\n", os.Args[0])
	}

	d, err := gatt.NewDevice(option.DefaultClientOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return
	}

	// Register handlers.
	d.Handle(
		gatt.PeripheralDiscovered(onPeriphDiscovered),
		gatt.PeripheralConnected(onPeriphConnected),
		gatt.PeripheralDisconnected(onPeriphDisconnected),
	)

	d.Init(onStateChanged)
	<-done
	fmt.Println("Done")
}
