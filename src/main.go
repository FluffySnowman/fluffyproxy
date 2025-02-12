package main

import (
	"fmt"
	// "fmt"
	"net"
	"os"

	// "net/http"

	pl "github.com/fluffysnowman/prettylogger"
)

// const TCP_LISTEN_ADDR = "0.0.0.0:8084"

const (
	HOST = "0.0.0.0"
	PORT = "8084"
	TYPE = "tcp"
)

func handleTCPConnection(conn net.Conn) {

	pl.Log("Received connection from remote: %v", conn.RemoteAddr())

	input_buf := make([]byte, 32)

	_, err := conn.Read(input_buf)
	if err != nil {
		pl.LogError("[ main.handleTCPConnection ] failed to read connection buffer: %v", err)
		os.Exit(1)
	}

    pl.LogDebug("Received data from remote: %s", input_buf)

	conn.Write([]byte(fmt.Sprintf("hello response to conn.\nHere's your data lol: %s", input_buf)))
	conn.Close()
}

func init() {
	pl.InitPrettyLogger("SIMPLE2")
}

func main() {

	pl.Log("Starting proxy...")

	// getting/starting the tcp listener and then doing shit so it listens to
	// me
	tcp_listener, err := net.Listen(TYPE, HOST+":"+PORT)
	if err != nil {
		pl.LogError("[ main.main ] failed to initialize tcp listener: %v", err)
		os.Exit(1)
	}

	defer tcp_listener.Close()

	pl.Log("Starting tcp listen...")

	for {
		TCP_CONN, err := tcp_listener.Accept()

		if err != nil {
			pl.Log("failed to accept tcp conection: %v", err)
			os.Exit(1)
		}

		go handleTCPConnection(TCP_CONN)
	}

}
