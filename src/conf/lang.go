package conf

import (
	"fmt"
	"strings"

	// "io"
	// "io/ioutil"
	"os"

	pl "github.com/fluffysnowman/prettylogger"
)

// all the types of keys like port, hostname etc (the actual keys might change
// so don't rely on this comment to know what is what lol)
var ALL_KEY_TYPES = []string{"server_ip", "server_port", "app_port", "app_ip"}

type Entry struct {
	Key   string
	Value string
}

// multiple entries hence a config
type Config []Entry

// the actual content of the config that's gonnab e read from a file
var CONFIG_CONTENT string

var CONFIG_FIELDS_ARRAY []string

func LoadConfigFile() {
	configData, err := os.ReadFile("proxy_c") // proxy_c = proxy client
	if err != nil {
		pl.LogError("[ conf.LoadConfigFile ] failed to read config file: %v", err)
		os.Exit(1)
		return
	}

	stringifiedConfigData := string(configData)
	CONFIG_CONTENT = stringifiedConfigData

	fieldedString := strings.Fields(CONFIG_CONTENT)
	CONFIG_FIELDS_ARRAY = fieldedString
	fmt.Println("config fields array below")
	fmt.Println(CONFIG_FIELDS_ARRAY)

	lexedConfigData := LexConfigFile()
	if (lexedConfigData == nil) || (len(lexedConfigData) < 1) {
		pl.LogFatal("config file has nothing in it or has an unidentified syntax error.")
	}
	fmt.Println("lexed config data below")
	fmt.Println(lexedConfigData)
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

// type Token struct {
// }

func PrintAllKeyTypes() {
	fmt.Println("all key types", ALL_KEY_TYPES)
}
