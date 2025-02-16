package main

import (
	"flag"
	"io"
	"net"
	"os"

	// "os"

	pl "github.com/fluffysnowman/prettylogger"
)

const (
	HOST = "0.0.0.0"
	PORT = "8084"
	TYPE = "tcp"
)

const SERVER_IP = "192.168.1.96"
const SERVER_PORT = "42069"

const CLIENT_IP = "172.17.0.1"
const CLIENT_PORT = "6969"

const INTERNAL_SERVICE_HOST = "10.69.42.16"
const INTERNAL_SERVICE_PORT = "8000"

var targetAddress string
var CLIENT_ENABLE bool
var SERVER_ENABLE bool

// Takes a tcp connection and forwards it to the target addy
func handleTCPConnection(clientConn net.Conn) {
	defer clientConn.Close()

	pl.Log("[ HANDLE ] received connection: %v", clientConn.RemoteAddr())

	targetConn, err := net.Dial(TYPE, CLIENT_IP+":"+CLIENT_PORT)
	if err != nil {
		pl.LogError("[ HANDLE ] failed to connect to target service: %v", err)
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

// func forwardConnTcpToIP() {
// }

// function that listens on the server port and then forwards all connections to
// the client port etc
func SERVER_listenForConnections() {
	pl.Log("[ SERVER ] Listening for connections on %s:%s", SERVER_IP, SERVER_PORT)
	serverListener, err := net.Listen(TYPE, SERVER_IP+":"+SERVER_PORT)
	if err != nil {
		pl.LogError("failed to listen on server port: %v", err)
		return
	}

	for {
		serverConn, err := serverListener.Accept()
		if err != nil {
			pl.LogError("failed to accept server connection: %v", err)
			continue
		}
		pl.Log("[ SERVER ] Received connection from %s", serverConn.RemoteAddr())
		go handleTCPConnection(serverConn)
	}
}

// function that listens for connections coming from the server and then
// forwards them to the internal service and then sends them back
func CLIENT_listenAndForwardToInternalService() {
	pl.Log("[ CLIENT ] Listening for connections on %s:%s", CLIENT_IP, CLIENT_PORT)
	clientListener, err := net.Listen(TYPE, CLIENT_IP+":"+CLIENT_PORT)
	if err != nil {
		pl.LogError("failed to listen on client port: %v", err)
		return
	}

	for {
		clientConn, err := clientListener.Accept()
		if err != nil {
			pl.LogError("failed to accept client connection: %v", err)
			continue
		}

		go func(clientConn net.Conn) {
			defer clientConn.Close()
			pl.Log("[ CLIENT ] Received connection from %s", clientConn.RemoteAddr())

			internalServiceConn, err := net.Dial(TYPE, INTERNAL_SERVICE_HOST+":"+INTERNAL_SERVICE_PORT)
			if err != nil {
				pl.LogError("failed to connect to internal service: %v", err)
				return
			}
			defer internalServiceConn.Close()

			done := make(chan bool, 2)

			go func() {
				_, err := io.Copy(internalServiceConn, clientConn)
				if err != nil && err != io.EOF {
					pl.LogError("failed to copy client -> internal service: %v", err)
				}
				done <- true
			}()

			go func() {
				_, err := io.Copy(clientConn, internalServiceConn)
				if err != nil && err != io.EOF {
					pl.LogError("failed to copy internal service -> client: %v", err)
				}
				done <- true
			}()

			<-done
			<-done
		}(clientConn)
	}
}

func init() {
	pl.InitPrettyLogger("SIMPLE2")

	// flag.StringVar(&targetAddress, "to", "", "Address to proxy to")
	flag.BoolVar(&CLIENT_ENABLE, "client", false, "Run client")
	flag.BoolVar(&SERVER_ENABLE, "server", false, "Run server")
	flag.Parse()

	pl.Log("[ init ] CLIENT_ENABLE: %v", CLIENT_ENABLE)
	pl.Log("[ init ] SERVER_ENABLE: %v", SERVER_ENABLE)

	// if targetAddress == "" {
	// 	pl.LogError("Addy is required. Specify with -to <ADDRESS >")
	// 	os.Exit(1)
	// }
}

func main() {

	pl.Log("Starting fluffyproxy...")

	// log internal service addr
	pl.Log("Internal service address: %s:%s", INTERNAL_SERVICE_HOST, INTERNAL_SERVICE_PORT)

	if CLIENT_ENABLE {
		go CLIENT_listenAndForwardToInternalService()
	}

	if SERVER_ENABLE {
		go SERVER_listenForConnections()
	}

	if !CLIENT_ENABLE && !SERVER_ENABLE {
		pl.LogError("No option specified- please run the client or the server")
		pl.Log("Exiting...")
		os.Exit(1)
		return
	}

	if CLIENT_ENABLE && SERVER_ENABLE {
		pl.LogError("Cannot run both the client and the server at the same time")
		pl.Log("Exiting...")
		os.Exit(1)
		return
	}

	select {}

}

/*
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
*/
