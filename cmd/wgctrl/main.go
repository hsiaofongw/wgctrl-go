// Command wgctrl is a testing utility for interacting with WireGuard via package
// wgctrl.
package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	vnetns "github.com/vishvananda/netns"
)

func main() {
	pid := os.Getpid()
	hostNetns, err := vnetns.GetFromPid(pid)
	if err != nil {
		panic(err)
	}
	defer hostNetns.Close()

	nsHandle, err := vnetns.GetFromDocker("b84b8f195b3683962f229c0503dad94fe161c48cf5bd7545b3a8f4ccc981acd0")
	if err != nil {
		panic(err)
	}
	defer nsHandle.Close()

	vnetns.Set(nsHandle)
	defer func() {
		vnetns.Set(hostNetns)

		log.Printf("---- set to host netns ----")

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

		for _, d := range devices {
			printDevice(d)

			for _, p := range d.Peers {
				printPeer(p)
			}
		}
	}()

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

	for _, d := range devices {
		printDevice(d)

		for _, p := range d.Peers {
			printPeer(p)
		}
	}
}

func printDevice(d *wgtypes.Device) {
	const f = `interface: %s (%s)
  public key: %s
  private key: (hidden)
  listening port: %d

`

	fmt.Printf(
		f,
		d.Name,
		d.Type.String(),
		d.PublicKey.String(),
		d.ListenPort)
}

func printPeer(p wgtypes.Peer) {
	const f = `peer: %s
  endpoint: %s
  allowed ips: %s
  latest handshake: %s
  transfer: %d B received, %d B sent

`

	fmt.Printf(
		f,
		p.PublicKey.String(),
		// TODO(mdlayher): get right endpoint with getnameinfo.
		p.Endpoint.String(),
		ipsString(p.AllowedIPs),
		p.LastHandshakeTime.String(),
		p.ReceiveBytes,
		p.TransmitBytes,
	)
}

func ipsString(ipns []net.IPNet) string {
	ss := make([]string, 0, len(ipns))
	for _, ipn := range ipns {
		ss = append(ss, ipn.String())
	}

	return strings.Join(ss, ", ")
}
