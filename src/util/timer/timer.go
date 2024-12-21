package timer

import (
	"fmt"
	"log"
	"time"
)

type Timer struct {
	started   bool
	startTime time.Time
	label     string
}

func (t *Timer) Start(format string, a ...any) {
	if t.started {
		t.Stop()
	}
	t.started = true
	t.startTime = time.Now()
	t.label = fmt.Sprintf(format, a...)
}

func (t *Timer) Stop() {
	if !t.started {
		return
	}
	t.started = false
	duration := time.Since(t.startTime)
	log.Printf("[timer] %s %s", t.label, duration)
}
