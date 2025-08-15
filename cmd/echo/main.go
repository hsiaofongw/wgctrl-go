package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	vnetns "github.com/vishvananda/netns"
)

func main() {
	conn, err := net.Dial("tcp", "192.168.31.103:44099")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	go func() {
		for {
			now := time.Now()
			conn.Write([]byte(fmt.Sprintf("%s\n", now.Format(time.RFC3339))))
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		io.Copy(os.Stdout, conn)
	}()

	time.Sleep(30 * time.Second)

	log.Println("Start switching netns")

	pid := os.Getpid()
	hostNetns, err := vnetns.GetFromPid(pid)
	if err != nil {
		panic(err)
	}
	defer hostNetns.Close()

	nsHandle, err := vnetns.GetFromPid(19035)
	if err != nil {
		panic(err)
	}
	defer nsHandle.Close()

	vnetns.Set(nsHandle)

	log.Println("netns switched")

	time.Sleep(30 * time.Second)

	log.Println("Start switching netns back")

	vnetns.Set(hostNetns)

	log.Println("netns switched back")

	time.Sleep(30 * time.Second)
}
