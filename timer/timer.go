package timer

import (
	"fmt"
	"time"
)

type timer struct {
	startTime     time.Time
	active        bool
	timedOutCache bool
	timeout       time.Duration
}

// Runs a timer which can be reset (automatically starts and resets countdown) and
// stopped. Can also be set to panic when timing out
func RunTimer(
	resetChan <-chan int,
	stopChan <-chan int,
	timeoutChan chan<- int,

	timeout time.Duration,
	panicOnTimeout bool,
	name string,
) {
	timerInstance := newTimer(timeout)

	for {
		select {
		case <-stopChan:
			timerInstance.active = false
		case <-resetChan:
			timerInstance.startTime = time.Now()
			timerInstance.active = true
		default:
			timedOut := checkTimeout(timerInstance)
			if timedOut {
				timerInstance.active = false
				if panicOnTimeout {
					panic(fmt.Sprintf("Panicking timer %s timed out", name))
				}
				timeoutChan <- 1
			}
			timerInstance.timedOutCache = timedOut
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func checkTimeout(_timer timer) bool {
	timedOut := _timer.active && time.Since(_timer.startTime) > _timer.timeout
	return timedOut && timedOut != _timer.timedOutCache
}

func newTimer(timeout time.Duration) timer {
	return timer{
		startTime:     time.Now(),
		active:        false,
		timedOutCache: false,
		timeout:       timeout,
	}
}
