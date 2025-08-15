// Command wgctrl is a testing utility for interacting with WireGuard via package
// wgctrl.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	vnetns "github.com/vishvananda/netns"
)

func main() {
	flag.Parse()

	var nsHandle *vnetns.NsHandle = nil

	pid := os.Getpid()
	hostNetns, err := vnetns.GetFromPid(pid)
	if err != nil {
		panic(err)
	}
	defer hostNetns.Close()

	dockerId := flag.Arg(0)
	if dockerId != "" {
		hd, err := vnetns.GetFromDocker(dockerId)
		if err != nil {
			panic(err)
		}
		defer hd.Close()
		nsHandle = &hd
	}

	if nsHandle != nil {
		vnetns.Set(*nsHandle)
		defer nsHandle.Close()
		defer vnetns.Set(hostNetns)
	}

	c, err := wgctrl.New()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	var devices []*wgtypes.Device
	devices, err = c.Devices()
	if err != nil {
		log.Fatalf("failed to get devices: %v", err)
	}

	devStatuses := make([]*wgtypes.DeviceStatus, 0)
	for _, dev := range devices {
		devStatuses = append(devStatuses, wgtypes.NewDeviceStatus(dev))
	}

	if err := json.NewEncoder(os.Stdout).Encode(devStatuses); err != nil {
		panic(err)
	}
}
