package main

import (
	"flag"
	"io"
	"net"
	"os"

	pl "github.com/fluffysnowman/prettylogger"
)

const (
	HOST = "0.0.0.0"
	PORT = "8084"
	TYPE = "tcp"
)

var targetAddress string

func handleTCPConnection(clientConn net.Conn) {
	defer clientConn.Close()

	pl.Log("recv conn from client: %v", clientConn.RemoteAddr())

	targetConn, err := net.Dial(TYPE, targetAddress)
	if err != nil {
		pl.LogError("failed to conn to target service: %v", err)
		return
	}
	defer targetConn.Close()

	pl.Log("connected to target service?: %v", targetAddress)

	done := make(chan bool, 2)

	go func() {
		_, err := io.Copy(targetConn, clientConn)
		if err != nil && err != io.EOF {
			pl.LogError("FAILED TO COPY CLIENT -> target: %v", err)
		}
		done <- true
	}()

	go func() {
		_, err := io.Copy(clientConn, targetConn)
		if err != nil && err != io.EOF {
			pl.LogError("ERROR COPYING TARGET -> client: %v", err)
		}
		done <- true
	}()

	<-done
	<-done
}

func init() {
	pl.InitPrettyLogger("SIMPLE2")

	flag.StringVar(&targetAddress, "to", "", "Address to proxy to")
	flag.Parse()

	if targetAddress == "" {
		pl.LogError("Addy is required. Specify with -to <ADDRESS >")
		os.Exit(1)
	}
}

func main() {
	pl.Log("Starting proxy on %s:%s -> %s", HOST, PORT, targetAddress)

	tcpListener, err := net.Listen(TYPE, HOST+":"+PORT)
	if err != nil {
		pl.LogError("[ main.main ] failed to initialize tcp listener: %v", err)
		os.Exit(1)
	}
	defer tcpListener.Close()

	pl.Log("Listening for connections...")

	for {
		clientConn, err := tcpListener.Accept()
		if err != nil {
			pl.LogError("failed to accept tcp conection: %v", err)
			continue
		}
		go handleTCPConnection(clientConn)
	}
}
