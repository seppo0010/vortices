package main

import (
	"encoding/hex"
	"fmt"
	"github.com/chifflier/nfqueue-go/nfqueue"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

var log string

func real_callback(payload *nfqueue.Payload) int {
	log += "Real callback\n"
	log += fmt.Sprintf("  id: %d\n", payload.Id)
	log += hex.Dump(payload.Data) + "\n"
	// Decode a packet
	packet := gopacket.NewPacket(payload.Data, layers.LayerTypeIPv4, gopacket.Default)
	// Get the TCP layer from this packet
	if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		// Get actual TCP data from this layer
		tcp, _ := tcpLayer.(*layers.TCP)
		log += fmt.Sprintf("From src port %d to dst port %d\n", tcp.SrcPort, tcp.DstPort)
	}
	if icmpLayer := packet.Layer(layers.LayerTypeICMPv4); icmpLayer != nil {
		log += "This is an ICMP packet!" + "\n"
		icmp, _ := icmpLayer.(*layers.ICMPv4)
		log += fmt.Sprintf("icmp id: %d\n", icmp.Id)
	}
	// Iterate over all layers, printing out each layer type
	for _, layer := range packet.Layers() {
		log += "PACKET LAYER:\n" + fmt.Sprintf("%v", layer.LayerType()) + "\n"
		log += gopacket.LayerDump(layer) + "\n"
	}
	log += "-- \n"
	if icmpLayer := packet.Layer(layers.LayerTypeICMPv4); icmpLayer != nil {
	payload.SetVerdict(nfqueue.NF_DROP)
} else {
	payload.SetVerdict(nfqueue.NF_ACCEPT)
}
	return 0
}

func listenToQueue(queue int) {
	q := new(nfqueue.Queue)

	q.SetCallback(real_callback)
	q.Init()

	q.Unbind(syscall.AF_INET)
	q.Bind(syscall.AF_INET)

	q.CreateQueue(queue)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			// sig is a ^C, handle it
			_ = sig
			q.StopLoop()
		}
	}()

	q.Loop()
	q.DestroyQueue()
	q.Close()
}

func main() {
	wg := sync.WaitGroup{}
	for _, q := range []int{1, 2} {
		wg.Add(1)
		go func(q int) {
			listenToQueue(q)
			wg.Done()
		}(q)
	}
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(http.StatusOK)
            w.Write([]byte(log))
            })
    go http.ListenAndServe(":8080", nil)
	wg.Wait()
	os.Exit(0)
}
