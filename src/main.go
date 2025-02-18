package main

import (
	"flag"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/fluffysnowman/fluffyproxy/conf"
	_ "github.com/fluffysnowman/fluffyproxy/data"

	pl "github.com/fluffysnowman/prettylogger"
	"github.com/hashicorp/yamux"
)

const (
	// address that the internal service is accessible by
	SERVER_LISTEN_IP    = "192.168.1.96"
	SERVER_LISTEN_PORT  = "42000"

	// address the client connects to
	SERVER_CONTROL_IP = "192.168.1.96"
  SERVER_CONTROL_PORT = "6969"

  // shit that the proxy actually goes to
	INTERNAL_SERVICE_IP = "10.69.42.16"
	INTERNAL_SERVICE_PORT = "8000"
)

var CLIENT_ENABLE bool
var SERVER_ENABLE bool

var (
	controlSession *yamux.Session
	sessionMutex   sync.RWMutex
)

func setControlSession(session *yamux.Session) {
	sessionMutex.Lock()
	controlSession = session
	sessionMutex.Unlock()
}

func getControlSession() *yamux.Session {
	sessionMutex.RLock()
	defer sessionMutex.RUnlock()
	return controlSession
}

func bridgeConnections(conn1, conn2 net.Conn) {
	defer conn1.Close()
	defer conn2.Close()

	done := make(chan struct{}, 2)
	go func() {
		_, _ = io.Copy(conn1, conn2)
		done <- struct{}{}
	}()
	go func() {
		_, _ = io.Copy(conn2, conn1)
		done <- struct{}{}
	}()
	<-done
}

func handleControlConnection(conn net.Conn) {
	pl.Log(
		"[ SERVER ] Persistent ctrl conn established from %v",
		conn.RemoteAddr(),
	)

	config := yamux.DefaultConfig()
	config.KeepAliveInterval = 30 * time.Second
	config.ConnectionWriteTimeout = 10 * time.Second

	session, err := yamux.Server(conn, config)
	if err != nil {
		pl.LogError("[ SERVER ] Failed to multiplex session: %v", err)
		conn.Close()
		return
	}

	setControlSession(session)

	for {
		_, err := session.Accept()
		if err != nil {
			pl.LogError("[ SERVER ] Yamux session closed: %v", err)
			break
		}
	}
	setControlSession(nil)
}

func controlListener() {
	addr := SERVER_CONTROL_IP + ":" + SERVER_CONTROL_PORT
	pl.Log("[ SERVER ] starting ctrl listeneron: %s", addr)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		pl.LogError("[ SERVER ] failed to start ctrl listener: %v", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			pl.LogError("[ SERVER ] failed to accept ctrl conn: %v", err)
			continue
		}
		go handleControlConnection(conn)
	}
}

func externalListener() {
	addr := SERVER_LISTEN_IP + ":" + SERVER_LISTEN_PORT
	pl.Log("[ SERVER ] starting external listener on : %s", addr)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		pl.LogError("[ SERVER ] failed to start external listener: %v", err)
		return
	}
	defer listener.Close()

	for {
		extConn, err := listener.Accept()
		if err != nil {
			pl.LogError("[ SERVER ] failed to accept external conn: %v", err)
			continue
		}
		pl.Log("[ SERVER ] got external conn from %v", extConn.RemoteAddr())

		session := getControlSession()
		if session == nil {
			pl.LogError(
				"[ SERVER ] no control/persisted sesh available. rejecting all reqs",
			)
			extConn.Close()
			continue
		}

		stream, err := session.Open()
		if err != nil {
			pl.LogError(
				"[ SERVER ] failed ot open substream/stream inside stream?idfk: %v",
				err,
			)
			extConn.Close()
			continue
		}
		pl.Log(
			"[ SERVER ] Opened new sub-stream for external connection %v",
			extConn.RemoteAddr(),
		)

		go bridgeConnections(extConn, stream)
	}
}

func runServer() {
	go controlListener()
	go externalListener()
	select {}
}

func runClient() {
	for {
		addr := SERVER_CONTROL_IP + ":" + SERVER_CONTROL_PORT
		pl.Log("[ CLIENT ] dialing server at: %s", addr)
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			pl.LogError("[ CLIENT ] failed to connect to server: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		pl.Log("[ CLIENT ] connected to server. starting yamux sesh")

		config := yamux.DefaultConfig()
		config.KeepAliveInterval = 30 * time.Second
		config.ConnectionWriteTimeout = 10 * time.Second

		session, err := yamux.Client(conn, config)
		if err != nil {
			pl.LogError("[ CLIENT ] failed to create yamux sesh: %v", err)
			conn.Close()
			time.Sleep(5 * time.Second)
			continue
		}

		for {
			stream, err := session.Accept()
			if err != nil {
				pl.LogError("[ CLIENT ] yamux session accept err: %v", err)
				break
			}
			pl.Log("[ CLIENT ] recv new stream for server")
			go handleStream(stream)
		}
		session.Close()
		pl.Log("[ CLIENT ] yamux sesh closed. reconnectiong in 5s")
		time.Sleep(5 * time.Second)
	}
}

func handleStream(stream net.Conn) {
	defer stream.Close()

	if tcpStream, ok := stream.(*net.TCPConn); ok {
		tcpStream.SetNoDelay(true)
	}

	internalAddr := INTERNAL_SERVICE_IP + ":" + INTERNAL_SERVICE_PORT
	pl.Log("[ CLIENT ] connecting to internal service at : %s", internalAddr)
	internalConn, err := net.Dial("tcp", internalAddr)
	if err != nil {
		pl.LogError("[ CLIENT ] failed to conn to internal service: %v", err)
		return
	}
	defer internalConn.Close()

	if tcpInternal, ok := internalConn.(*net.TCPConn); ok {
		tcpInternal.SetNoDelay(true)
	}

	pl.Log("[ CLIENT ] bridging substream w/ internal service")
	bridgeConnections(stream, internalConn)
}

func init() {
	pl.InitPrettyLogger("SIMPLE2")
	flag.BoolVar(&CLIENT_ENABLE, "client", false, "Run in client mode")
	flag.BoolVar(&SERVER_ENABLE, "server", false, "Run in server mode")
	flag.Parse()
	pl.Log("[ init ] CLIENT_ENABLE: %v", CLIENT_ENABLE)
	pl.Log("[ init ] SERVER_ENABLE: %v", SERVER_ENABLE)
}

func main() {
	conf.PrintAllKeyTypes()
	conf.LoadConfigFile()
	pl.Log("Starting rev tunnel proxy...")
	if CLIENT_ENABLE && SERVER_ENABLE {
		pl.LogError("cant run client & server at the same time lol")
		os.Exit(1)
	}
	if !CLIENT_ENABLE && !SERVER_ENABLE {
		pl.LogError("Specify -server or -client to actually do something")
		os.Exit(1)
	}

	if SERVER_ENABLE {
		runServer()
	} else if CLIENT_ENABLE {
		runClient()
	}
}
