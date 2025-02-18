/*
This package has everything related to the data that is being used for
configuration and other similar things like settings and getting the server ip
address, port, host etc along with client side configuration like the internal
service/application that is going to be proxied to and so on.
*/
package data

// Struct for all the configuration values that the server takes
type ServerConfig struct {
	ServerIP          string
	ServerPort        string
	ServerControlPort string
}

// Struct for all the configuration values that the client takes
type ClientConfig struct {
	ClientIP          string
	ClientPort        string
	ClientControlPort string
}
