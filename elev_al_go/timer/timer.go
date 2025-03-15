package timer

import (
	"time"
)

// Communicate with rest of system via channels

var (
	startTime     = time.Now()
	active        = true
	timedOutCache = false
	timeout       = time.Second       // Default value
	TimeoutChan   = make(chan int, 1) // Default value
)

func StartTimer() {
	startTime = time.Now()
	active = true
}

func StopTimer() {
	active = false
}

// TODO: Turn this into a (small) process
func CheckTimeout() {
	timedOut := active && time.Since(startTime) > timeout
	if timedOut && timedOut != timedOutCache {
		TimeoutChan <- 1
	}
	timedOutCache = timedOut
}

func SetTimeout(_timeout time.Duration) {
	timeout = _timeout
}
