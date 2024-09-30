// cmd/logger/main.go: scans for BLE announcements, optionally for know devices, and logs them
//
// 2024/09/29 first version, based on cmd/scanner/main.go [DurvalMenezes]
//

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Seitanas/wosensortho-exporter/pkg/btle"
	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
	"github.com/pkg/errors"
	"log"
	"os"
	"strings"
	"time"
)

var (
	bledev_name = flag.String("bledev", "default", "BLE device name")
	timeout     = flag.Duration("timeout", 10*time.Second, "Scan duration")
)

func usage() {
	fmt.Printf("Usage: %s [-bledev bluetooth_device] [-timeout=timeout] [mac=name]\n", os.Args[0])
}

func main() {
	macname := make(map[string]string)
	macdone := make(map[string]bool)

	flag.Usage = usage
	flag.Parse()
	if len(flag.Args()) > 0 {
		for _, arg := range flag.Args() {
			s := strings.Split(arg, "=")
			if len(s) != 2 {
				fmt.Printf("%s: invalid argument '%s'\n", os.Args[0], arg)
				os.Exit(-1)
			}
			macname[s[0]] = s[1]
			macdone[s[0]] = false
		}
	}

	bledev, err := dev.NewDevice(*bledev_name)
	if err != nil {
		log.Fatalf("Can't create BLE device: %s", err)
	}
	ble.SetDefaultDevice(bledev)

	fmt.Printf("Scanning for %s...\n", timeout)
	left := len(macname)
	for timeleft := *timeout; timeleft >= 0; timeleft -= 1 * time.Second {
		ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), 1*time.Second))
		switch errors.Cause(ble.Scan(ctx, true, btle.Handler, nil)) {
		case nil:
		case context.DeadlineExceeded:
		case context.Canceled:
			fmt.Printf("Cancelled\n")
		default:
			log.Fatal(err.Error())
		}

		fmt.Printf("Found SwitchBot devices:\n")
		for mac, data := range btle.BTDevice {
			name, exists := macname[mac]
			if exists {
				if !macdone[mac] {
					macdone[mac] = true
					left--
					fmt.Printf("%s, Temperature: %f Humidity: %f Battery: %f\n", name, data.Temperature, data.Humidity, data.Battery)
				}
			} else {
				fmt.Printf("%s, Temperature: %f Humidity: %f Battery: %f\n", mac, data.Temperature, data.Humidity, data.Battery)
			}
			if left == 0 {
				break
			}
		}

		if left == 0 && len(macname) != 0 {
			break
		}
	}

	if left != 0 {
		fmt.Printf("Timeout, %d devices missing\n", left)
		os.Exit(1)
	} else {
		fmt.Printf("Done\n")
		os.Exit(0)
	}
}

//Eof cmd/logger/main.go
