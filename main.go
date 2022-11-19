package main

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

const (
	protocolID = "/zozo/1.0.0"
	//ServiceName = "libp2p.ping"
)

func sendPing(s network.Stream) {
	var (
		message []byte = []byte("ping")
		err     error
	)

	fmt.Println("sending message: ", message)
	err = binary.Write(s, binary.BigEndian, &message)
	if err != nil {
		panic(err)
	}
}

func listenPing(s network.Stream) {
	var (
		messageByte   []byte = make([]byte, 4)
		messageString string
		err           error
	)
	err = binary.Read(s, binary.BigEndian, &messageByte)
	if err != nil {
		panic(err)
	}

	fmt.Println(messageByte)

	messageString = string(messageByte)

	fmt.Printf("Received %s from %s\n", messageString, s.ID())

	<-time.After(time.Second * 5)
}

func connectToPeer(node host.Host) {
	// Add -peer-address flag
	var peerAddr = flag.String("peer-address", "", "peer address")
	flag.Parse()
	fmt.Println(*peerAddr)
	// If we received a peer address, we should connect to it.
	if *peerAddr != "" {
		// Parse the multiaddr string.
		peerMA, err := multiaddr.NewMultiaddr(*peerAddr)
		if err != nil {
			panic(err)
		}
		peerAddrInfo, err := peer.AddrInfoFromP2pAddr(peerMA)
		if err != nil {
			panic(err)
		}

		// Connect to the node at the given address.
		if err := node.Connect(context.Background(), *peerAddrInfo); err != nil {
			panic(err)
		}
		fmt.Println("Connected to", peerAddrInfo.String())

		// Open a stream with the given peer.
		s, err := node.NewStream(context.Background(), peerAddrInfo.ID, protocolID)
		if err != nil {
			panic(err)
		}

		go sendPing(s)
	}
}

func main() {
	var (
		node host.Host
		err  error
	)
	if node, err = libp2p.New(libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0")); err != nil {
		panic(err)
	}
	defer node.Close()

	fmt.Println("Addresse/id:", node.Addrs()[0].String()+"/p2p/"+node.ID().String())

	connectToPeer(node)

	// This gets called every time a peer connects
	// and opens a stream to this node.
	node.SetStreamHandler(protocolID, func(s network.Stream) {
		go listenPing(s)
	})

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGKILL, syscall.SIGINT)
	fmt.Print(<-sigCh)
}
