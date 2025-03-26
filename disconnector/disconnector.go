package disconnector

import (
	"fmt"

	"github.com/eiannone/keyboard"
)

func RunDisconnector(
	disconnectChan chan<- int,
) {
	keyboard.Open()
	defer keyboard.Close()
	for {
		char, key, _ := keyboard.GetKey()
		fmt.Println(" ;; Received key stroke", string(char))
		if string(char) == "d" {
			disconnectChan <- 1
			fmt.Println(" ;; Sent to disconnect")
		} else if key == keyboard.KeyCtrlC {
			return
		}
	}
}
