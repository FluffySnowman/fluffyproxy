/*
This package has everything related to the data that is being used for
configuration and other similar things like settings and getting the server ip
address, port, host etc along with client side configuration like the internal
service/application that is going to be proxied to and so on.
*/
package data

// Struct for all the configuration values that the server takes
type ServerConfig struct {
	ServerListenAddress string // direct ip:port address
	ServerListenIP      string // addr listening for external conns
	ServerListenPort    string // port listening for ext conns
	ServerControlIP     string // control ip the client connects to
	ServerControlPort   string // control port the client connects to
}

// Struct for all the configuration values that the client takes
//
// client doesn't need much config other than the server address
type ClientConfig struct {
	LocalServiceIP   string
	LocalServicePort string
	ServerAddress    string // direct ip:port
	// ServerIP      string
	// ServerPort    string
}

var GLOBAL_SERVER_CONFIG ServerConfig
var GLOBAL_CLIENT_CONFIG ClientConfig

// setting default shit

func SetDefaultServerConfig() {

	GLOBAL_SERVER_CONFIG.ServerListenIP = "0.0.0.0"
	GLOBAL_SERVER_CONFIG.ServerListenPort = "7000"

	GLOBAL_SERVER_CONFIG.ServerControlIP = "0.0.0.0"
	GLOBAL_SERVER_CONFIG.ServerControlPort = "42069"

}

func SetDefaultClientConfig() {

	GLOBAL_CLIENT_CONFIG.LocalServiceIP = "0.0.0.0"
	GLOBAL_CLIENT_CONFIG.LocalServicePort = "8080"

	GLOBAL_CLIENT_CONFIG.ServerAddress = "0.0.0.0:42069"

}
