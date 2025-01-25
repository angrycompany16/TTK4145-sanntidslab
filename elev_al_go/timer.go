package elevalgo

import "time"

type Timer struct {
	startTime time.Time
	timeOut   time.Duration
	active    bool
}

func (t *Timer) Start() {
	t.startTime = time.Now()
	t.active = true
}

func (t *Timer) Stop() {
	t.active = false
}

func (t *Timer) TimedOut() bool {
	return t.active && time.Since(t.startTime) > t.timeOut
}

func MakeTimer(timeOut time.Duration) *Timer {
	return &Timer{
		time.Now(),
		timeOut,
		false,
	}
}
