package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"

	yaml "gopkg.in/yaml.v2"
)

// PolarisRepository defines a particular Git repository plus branch
//
type PolarisRepository struct {
	Ref string
	URI string
}

// PolarisConfig is the structure defining the config
//
type PolarisConfig struct {
	Repositories map[string]PolarisRepository
}

// DefaultConfig is what you get when you have no config at the start
//
var DefaultConfig = PolarisConfig{
	Repositories: map[string]PolarisRepository{
		"core/stable": PolarisRepository{
			URI: "https://github.com/synthesis-labs/polaris-scaffolds",
			Ref: "refs/heads/stable",
		},
		"core/unstable": PolarisRepository{
			URI: "https://github.com/synthesis-labs/polaris-scaffolds",
			Ref: "refs/heads/unstable",
		},
	},
}

// GetConfig calculates the home folder of polaris (default: ~/.polaris)
// and also initializes it
//
func GetConfig() (string, *PolarisConfig) {
	polarisHome, found := os.LookupEnv("POLARIS_HOME")
	if !found {
		user, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		polarisHome = fmt.Sprintf("%s/.polaris", user.HomeDir)
	}

	//fmt.Println("Using POLARIS_HOME =", polarisHome)

	if _, err := os.Stat(polarisHome); os.IsNotExist(err) {
		fmt.Println("Setting up", polarisHome)

		// Make the directory
		//
		os.Mkdir(polarisHome, os.ModePerm)

		// Write the default config file
		//
		SaveConfig(polarisHome, &DefaultConfig)
	}

	// Parse the existing one and return it
	//
	configData, err := ioutil.ReadFile(fmt.Sprintf("%s/config.yaml", polarisHome))
	if err != nil {
		log.Fatal(err)
	}
	polarisConfig := PolarisConfig{}
	err = yaml.Unmarshal(configData, &polarisConfig)
	if err != nil {
		log.Fatal(err)
	}

	// If any of the repositories have not been synced for a while - then resync them
	//

	return polarisHome, &polarisConfig
}

// SaveConfig saves the contents to disk
//
func SaveConfig(polarisHome string, polarisConfig *PolarisConfig) {
	config, err := yaml.Marshal(*polarisConfig)
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile(fmt.Sprintf("%s/config.yaml", polarisHome), config, 0644)
}
