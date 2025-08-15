// Command wgctrl is a testing utility for interacting with WireGuard via package
// wgctrl.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	vnetns "github.com/vishvananda/netns"

	"github.com/mdlayher/netlink"
)

func main() {
	flag.Parse()

	var err error
	var c *wgctrl.Client
	var netnsfd *int

	var netnsPid *int
	pid, err := strconv.ParseInt(flag.Arg(0), 10, 32)
	if err == nil {
		pidInt := int(pid)
		netnsPid = &pidInt
	}

	if netnsPid != nil {
		log.Printf("netns pid: %d", *netnsPid)
		nsHandle, err := vnetns.GetFromPid(*netnsPid)
		if err != nil {
			panic(err)
		}
		defer nsHandle.Close()

		fd := int(nsHandle)
		netnsfd = &fd
		log.Printf("netns fd: %d", *netnsfd)
	}

	if netnsfd != nil {
		log.Printf("opening wgctrl with netlink config (netns fd: %d)", *netnsfd)
		c, err = wgctrl.NewWithNetlinkConfig(&netlink.Config{
			NetNS: *netnsfd,
		})
		if err != nil {
			log.Fatalf("failed to open wgctrl with netlink config (netns fd: %d): %v", *netnsfd, err)
		}
		defer c.Close()
	} else {
		c, err = wgctrl.New()
		if err != nil {
			log.Fatalf("failed to open wgctrl: %v", err)
		}
		defer c.Close()
	}

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
