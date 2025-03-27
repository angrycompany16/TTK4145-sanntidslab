package elevalgo

import (
	"fmt"
	"os"
	"time"

	"github.com/go-yaml/yaml"
)

type clearRequestVariant int

const (
	NumFloors      = 4 // Default value
	NumCabButtons  = 1
	NumHallButtons = 2
	NumButtons     = NumCabButtons + NumHallButtons
)

const (
	// Assume everyone waiting for the elevator gets on the elevator, even if
	// they will be traveling in the "wrong" direction for a while
	clearAll clearRequestVariant = iota

	// Assume that only those that want to travel in the current direction
	// enter the elevator, and keep waiting outside otherwise
	clearSameDir
)

type Config struct {
	ClearRequestVariant clearRequestVariant `yaml:"ClearRequestVariant"`
	DoorOpenDuration    time.Duration       `yaml:"DoorOpenDuration"`
}

func LoadConfig(configPath string) (Config, error) {
	_config := Config{}
	file, err := os.Open(configPath)
	if err != nil {
		fmt.Println("Error reading file")
		return _config, err
	}
	defer file.Close()

	err = yaml.NewDecoder(file).Decode(&_config)
	if err != nil {
		fmt.Println("Error decoding file")
		return _config, err
	}
	return _config, nil
}
