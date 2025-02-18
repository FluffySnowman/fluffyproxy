package conf

import (
	"fmt"
	"strings"

	// "io"
	// "io/ioutil"
	"os"

	"github.com/fluffysnowman/fluffyproxy/data"
	pl "github.com/fluffysnowman/prettylogger"
)

// all the types of keys like port, hostname etc (the actual keys might change
// so don't rely on this comment to know what is what lol)
var ALL_KEY_TYPES_SERVER = []string{"listen", "control", "listen_ip", "listen_port", "control_ip", "control_port"}
var ALL_KEY_TYPES_CLIENT = []string{"local", "local_service_ip", "local_service_port", "server", "server_ip", "server_port"}

type Entry struct {
	Key   string
	Value string
}

// multiple entries hence a config
type Config []Entry

// the actual content of the config that's gonnab e read from a file
var CONFIG_CONTENT string

var CONFIG_FIELDS_ARRAY []string

func LoadConfigFile(configFilePath string) {

	if configFilePath == "" || len(configFilePath) < 1 {
		configFilePath = "proxy_c"
	}

	configData, err := os.ReadFile(configFilePath)
	if err != nil {
		pl.LogError("[ conf.LoadConfigFile ] failed to read config file: %v", err)
		os.Exit(1)
		return
	}

	stringifiedConfigData := string(configData)
	CONFIG_CONTENT = stringifiedConfigData

	lines := strings.Split(CONFIG_CONTENT, "\n")
	var tokens []string
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}
		lineTokens := strings.Fields(trimmedLine)
		tokens = append(tokens, lineTokens...)
	}
	CONFIG_FIELDS_ARRAY = tokens

	fmt.Println("config fields array below")
	fmt.Println(CONFIG_FIELDS_ARRAY)

	lexedConfigData := LexConfigFile()
	if lexedConfigData == nil || len(lexedConfigData) < 1 {
		pl.LogFatal("config file has nothing in it or has an unidentified syntax error.")
	}
	fmt.Println("lexed config data below")
	fmt.Println(lexedConfigData)

	ParseConfigFile(lexedConfigData)
}

func LexConfigFile() Config {

	configLength := len(CONFIG_FIELDS_ARRAY)

	var configData Config

	for i := 0; i < configLength; i += 2 {

		if i+1 >= configLength {
			pl.LogFatal("Missing value in config file. Make sure all keys have a corresponding value")
			os.Exit(1)
		}

		k := CONFIG_FIELDS_ARRAY[i]
		v := CONFIG_FIELDS_ARRAY[i+1]

		configData = append(configData, Entry{Key: k, Value: v})

		// %2 is a value. !%2 is a key
		// if i%2 == 1 {
		// 	fmt.Printf("Value: %s\n", CONFIG_FIELDS_ARRAY[i])
		// 	configData[i].Value = CONFIG_FIELDS_ARRAY[i]
		// }

	}
	return configData
}

func ParseConfigFile(confData Config) {
	for _, entry := range confData {
		fmt.Printf("Key: %s, Value: %s\n", entry.Key, entry.Value)
	}

	for i := 0; i < len(confData); i++ {
		switch confData[i].Key {

		// server shit
		case "listen":
			data.GLOBAL_SERVER_CONFIG.ServerListenAddress = confData[i].Value
		case "control":
			data.GLOBAL_SERVER_CONFIG.ServerControlAddress = confData[i].Value
		case "listen_ip":
			data.GLOBAL_SERVER_CONFIG.ServerListenIP = confData[i].Value
		case "listen_port":
			data.GLOBAL_SERVER_CONFIG.ServerListenPort = confData[i].Value
		case "control_ip":
			data.GLOBAL_SERVER_CONFIG.ServerControlIP = confData[i].Value
		case "control_port":
			data.GLOBAL_SERVER_CONFIG.ServerControlPort = confData[i].Value

		// client shit
		case "local_service_ip":
			data.GLOBAL_CLIENT_CONFIG.LocalServiceIP = confData[i].Value
		case "local_service_port":
			data.GLOBAL_CLIENT_CONFIG.LocalServicePort = confData[i].Value
		case "local":
			data.GLOBAL_CLIENT_CONFIG.LocalServiceAddress = confData[i].Value
		case "server":
			data.GLOBAL_CLIENT_CONFIG.ServerCtrlAddress = confData[i].Value
		}
	}

}

// type Token struct {
// }

func PrintAllKeyTypes() {
	fmt.Println("all key types", ALL_KEY_TYPES_SERVER)
}
