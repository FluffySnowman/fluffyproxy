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
var ALL_KEY_TYPES = []string{"hello", "there"}

// the actual content of the config that's gonnab e read from a file
var CONFIG_CONTENT string

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
	fmt.Println("fielded string below")
	fmt.Println(fieldedString)
}

func ParseConfigFile() {

}

// type Token struct {
// }

func PrintAllKeyTypes() {
	fmt.Println("all key types", ALL_KEY_TYPES)
}
