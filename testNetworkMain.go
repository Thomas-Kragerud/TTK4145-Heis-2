package main

import (
	"Project/network/bcast"
	"Project/network/localip"
	"Project/network/peers"
	"flag"
	"fmt"
	"os"
	"time"
)

// We define some custom struct to send over the network
// Note that all members we want to transmit must be public
// Any private members will be received as zero-values (not sure)

type HelloMsg struct {
	Message string
	Iter    int
}

func main() {
	// Our id can be anything. Here we pass it on the command line
	// 'go run testNetworkMain.go -id=our_id'
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()
	port1 := 15647
	// ... or alternatively, we can use the local IP address.
	// (but since we can run multiple programs on the same PC, we also append the process ID)
	// Automatically assign a process ID through os.Getpid()
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peers-%s-%d", localIP, os.Getpid())
	}

	// We make a channel for receiving updates on the id´s of the peers that are alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has started
	// This can be used to signal that we are somehow "unavailable"
	peerTxEnable := make(chan bool)
	go peers.Transmitter(port1, id, peerTxEnable)
	go peers.Receiver(port1, peerUpdateCh)

	// Add multiple pairs to the network
	id2 := "70"
	id3 := "71"
	port2 := 6060
	port3 := 7070

	go peers.Transmitter(port1, id2, peerTxEnable)
	go peers.Receiver(port2, peerUpdateCh)
	go peers.Transmitter(port1, id3, peerTxEnable)
	go peers.Receiver(port3, peerUpdateCh)

	// *** This is somewhat totally unrelated ***
	// We make channels for sending and receiving out custom data types
	helloTx := make(chan HelloMsg)
	helloRx := make(chan HelloMsg)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	// start multiple transmitter/receivers on the same port
	go bcast.Transmitter(16569, helloTx)
	go bcast.Receiver(16569, helloRx)

	// The example message. We just send one of these every second
	go func() {
		helloMsg := HelloMsg{"Hello from " + id, 0}
		for {
			helloMsg.Iter++
			helloTx <- helloMsg // Send on the transmitter
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		time.Sleep(5 * time.Second)
		//peerTxEnable <- false
	}()

	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf(" Peers: %q\n", p.Peers)
			fmt.Printf(" New: %q\n", p.New)
			fmt.Printf(" Lost: %q\n", p.Lost)

		case a := <-helloRx: // Receive on the receiver
			fmt.Printf("Recived: %#v\n", a)
		}
	}

}
