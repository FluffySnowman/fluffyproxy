package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fluffysnowman/fluffyproxy/conf"
	"github.com/fluffysnowman/fluffyproxy/data"
	pl "github.com/fluffysnowman/prettylogger"
	"github.com/hashicorp/yamux"
)

const FP_CURRENT_VERSION = "v1.1.0"

var (
	allowedClientIPs   []string
	allowedExternalIPs []string

	clientWhitelistFlag   string
	externalWhitelistFlag string

	isClientWhitelist   bool
	isExternalWhitelist bool

	allowAllExternalIPs bool
	allowAllClientIPs   bool
)

var (
	CLIENT_ENABLE bool
	SERVER_ENABLE bool
	configFile    string

	controlSession *yamux.Session
	sessionMutex   sync.RWMutex

	activeClient    string
	activeClientMux sync.RWMutex
)

func parseIPs(s string) []string {
	parts := strings.Split(s, ",")
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}
	return parts
}

func ipAllowed(ip string, allowedIPs []string) bool {
	for _, allowed := range allowedIPs {
		if allowed == "*" {
			return true
		}
		if ip == allowed {
			return true
		}
	}
	return false
}

func setActiveClient(clientAddr string) bool {
	activeClientMux.Lock()
	defer activeClientMux.Unlock()

	if activeClient != "" {
		return false
	}
	activeClient = clientAddr
	return true
}

func removeActiveClient(clientAddr string) {
	activeClientMux.Lock()
	defer activeClientMux.Unlock()

	if activeClient == clientAddr {
		activeClient = ""
		pl.Log("[ SERVER ] active client %v removed", clientAddr)
	}
}

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
	clientAddr := conn.RemoteAddr().String()
	ip, _, err := net.SplitHostPort(clientAddr)
	if err != nil {
		pl.LogError("[ SERVER ] failed to parse client IP: %v", err)
		conn.Close()
		return
	}

	if !ipAllowed(ip, allowedClientIPs) {
		pl.LogError("[ SERVER ] connection from %s rejected: IP not whitelisted", ip)
		conn.Write([]byte("REJECT\n"))
		conn.Close()
		return
	}

	if !setActiveClient(clientAddr) {
		pl.LogError("[ SERVER ] Connection rejected - imagine being a skid lol haha")
		conn.Write([]byte("REJECT\n"))
		conn.Close()
		return
	}

	pl.Log("[ SERVER ] Persistent ctrl conn established from : %v", clientAddr)

	config := yamux.DefaultConfig()
	config.KeepAliveInterval = 30 * time.Second
	config.ConnectionWriteTimeout = 10 * time.Second

	session, err := yamux.Server(conn, config)
	if err != nil {
		pl.LogError("[ SERVER ] Failed to multiplex session: %v", err)
		removeActiveClient(clientAddr)
		conn.Close()
		return
	}

	setControlSession(session)

	defer func() {
		setControlSession(nil)
		removeActiveClient(clientAddr)
		conn.Close()
	}()

	for {
		_, err := session.Accept()
		if err != nil {
			pl.LogError("[ SERVER ] Yamux session closed: %v", err)
			break
		}
	}
}

func controlListener() {
	// addr := data.GLOBAL_SERVER_CONFIG.ServerControlIP + ":" + data.GLOBAL_SERVER_CONFIG.ServerControlPort
	addr := data.GLOBAL_SERVER_CONFIG.ServerControlAddress
	pl.Log("[ SERVER ] starting ctrl listener on: %s", addr)
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
	// addr := data.GLOBAL_SERVER_CONFIG.ServerListenIP + ":" + data.GLOBAL_SERVER_CONFIG.ServerListenPort
	addr := data.GLOBAL_SERVER_CONFIG.ServerListenAddress
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

		extIP, _, err := net.SplitHostPort(extConn.RemoteAddr().String())
		if err != nil {
			pl.LogError("[ SERVER ] failed to parse external connection IP: %v", err)
			extConn.Close()
			continue
		}
		if !ipAllowed(extIP, allowedExternalIPs) {
			pl.LogError("[ SERVER ] external connection from %s rejected: ip not whitelisted", extIP)
			extConn.Close()
			continue
		}

		pl.Log("[ SERVER ] got external conn from %v", extConn.RemoteAddr())
		session := getControlSession()
		if session == nil {
			pl.LogError("[ SERVER ] no control/persisted sesh available. rejecting all reqs")
			extConn.Close()
			continue
		}

		stream, err := session.Open()
		if err != nil {
			pl.LogError("[ SERVER ] failed ot open substream/stream inside stream?idfk: %v", err)
			extConn.Close()
			continue
		}
		pl.Log("[ SERVER ] Opened new sub-stream for external connection %v", extConn.RemoteAddr())
		go bridgeConnections(extConn, stream)
	}
}

func runServer() {
	go controlListener()
	go externalListener()
	select {}
}

func checkForRejection(conn net.Conn) bool {
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	buf := make([]byte, 7)
	n, err := conn.Read(buf)
	if err == nil && n == 7 && string(buf) == "REJECT\n" {
		pl.LogError("[ CLIENT ] skid detected. thou shalt not pass")
		os.Exit(1)
		return true
	}
	conn.SetReadDeadline(time.Time{})
	return false
}

func runClient() {
	for {
		pl.Log("[ CLIENT ] dialing server at: %s", data.GLOBAL_CLIENT_CONFIG.ServerCtrlAddress)
		conn, err := net.Dial("tcp", data.GLOBAL_CLIENT_CONFIG.ServerCtrlAddress)
		if err != nil {
			pl.LogError("[ CLIENT ] failed to connect to server: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if checkForRejection(conn) {
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

	// internalAddr := data.GLOBAL_CLIENT_CONFIG.LocalServiceIP + ":" + data.GLOBAL_CLIENT_CONFIG.LocalServicePort
	internalAddr := data.GLOBAL_CLIENT_CONFIG.LocalServiceAddress
	// fmt.Println("internal address: ", internalAddr)
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

	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version" || os.Args[1] == "-version") {
    fmt.Println(FP_CURRENT_VERSION)
		os.Exit(0)
	}

	pl.InitPrettyLogger("SIMPLE2")

	// data.SetDefaultServerConfig()
	// data.SetDefaultClientConfig()
	originalUsage := flag.Usage

	flag.Usage = func() {
		originalUsage()
		fmt.Println("\nExample:")
		fmt.Println("  [SERVER] fp -server -listen '192.168.1.96:8989' -control '0.0.0.0:42069' -client-whitelist '192.168.1.100,192.168.1.101' -external-whitelist '1.1.1.1,8.8.8.8'")
		fmt.Println("  ^ Whitelist IP's should be comma seperated")
		fmt.Println("  [CLIENT] fp -client -server-control-addr '0.0.0.0:42069' -local '10.69.42.16:8000'")
	}

	flag.BoolVar(&CLIENT_ENABLE, "client", false, "Run in client mode")
	flag.BoolVar(&SERVER_ENABLE, "server", false, "Run in server mode")
	flag.StringVar(&configFile, "f", "", "Path to config file")
	flag.StringVar(&configFile, "file", "", "Path to config file")
	flag.StringVar(&configFile, "c", "", "Same as -f")
	flag.StringVar(&configFile, "config", "", "Same as -f OR --file")
	flag.StringVar(&data.GLOBAL_SERVER_CONFIG.ServerListenAddress, "listen", "0.0.0.0:7000", "[Server] listen Address IP:PORT")
	flag.StringVar(&data.GLOBAL_SERVER_CONFIG.ServerControlAddress, "control", "0.0.0.0:42069", "[Server] control Address IP:PORT")
	flag.StringVar(&data.GLOBAL_SERVER_CONFIG.ClientWhitelistIPs, "client-whitelist", "", "[Server] Comma-separated list of allowed client IPs")
	flag.StringVar(&data.GLOBAL_SERVER_CONFIG.ExternalWhitelistIPs, "external-whitelist", "", "[Server] Comma-separated list of allowed external IPs")
	flag.StringVar(&data.GLOBAL_CLIENT_CONFIG.ServerCtrlAddress, "server-control-addr", "0.0.0.0:42069", "[Client] Server control address IP:PORT")
	flag.StringVar(&data.GLOBAL_CLIENT_CONFIG.LocalServiceAddress, "local", "0.0.0.0:8080", "[Client] Local service Address IP:PORT")
	flag.Parse()

	if data.GLOBAL_SERVER_CONFIG.ClientWhitelistIPs != "" {
		allowedClientIPs = parseIPs(data.GLOBAL_SERVER_CONFIG.ClientWhitelistIPs)
	} else {
		allowedClientIPs = []string{"192.168.1.1"}
	}
	if data.GLOBAL_SERVER_CONFIG.ExternalWhitelistIPs != "" {
		allowedExternalIPs = parseIPs(data.GLOBAL_SERVER_CONFIG.ExternalWhitelistIPs)
	} else {
		allowedExternalIPs = []string{"1.1.1.1"}
	}

	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "-help") {
		flag.Usage()
		os.Exit(0)
	}

	pl.Log("[ init ] CLIENT_ENABLE: %v", CLIENT_ENABLE)
	pl.Log("[ init ] SERVER_ENABLE: %v", SERVER_ENABLE)
}

func main() {

	fmt.Printf("all config data:\nServer: %v\nClient: %v\n", data.GLOBAL_SERVER_CONFIG, data.GLOBAL_CLIENT_CONFIG)

	conf.PrintAllKeyTypes()
	if configFile != "" {
		conf.LoadConfigFile(configFile)
	} else {
		pl.LogInfo("No config file specified")
	}

	if data.GLOBAL_SERVER_CONFIG.ClientWhitelistIPs != "" {
		allowedClientIPs = parseIPs(data.GLOBAL_SERVER_CONFIG.ClientWhitelistIPs)
	} else {
		allowedClientIPs = []string{"192.168.1.1"}
	}
	if data.GLOBAL_SERVER_CONFIG.ExternalWhitelistIPs != "" {
		allowedExternalIPs = parseIPs(data.GLOBAL_SERVER_CONFIG.ExternalWhitelistIPs)
	} else {
		allowedExternalIPs = []string{"1.1.1.1"}
	}

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
