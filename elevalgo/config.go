package elevalgo

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/go-yaml/yaml"
)

var ConfigPath = path.Join("elevalgo", "elevator_config.yaml")

type clearRequestVariant int

const (
	// Assume everyone waiting for the elevator gets on the elevator, even if
	// they will be traveling in the "wrong" direction for a while
	clearAll clearRequestVariant = iota

	// Assume that only those that want to travel in the current direction
	// enter the elevator, and keep waiting outside otherwise
	clearSameDir
)

type config struct {
	ClearRequestVariant clearRequestVariant `yaml:"ClearRequestVariant"`
	DoorOpenDuration    time.Duration       `yaml:"DoorOpenDuration"`
}

func loadConfig() (config, error) {
	c := config{}
	file, err := os.Open(ConfigPath)
	if err != nil {
		fmt.Println("Error reading file")
		return c, err
	}
	defer file.Close()

	err = yaml.NewDecoder(file).Decode(&c)
	if err != nil {
		fmt.Println("Error decoding file")
		return c, err
	}
	return c, nil
}
