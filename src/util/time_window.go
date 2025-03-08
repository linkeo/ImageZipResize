package util

import (
	"sync"
	"time"
)

type TimeWindow struct {
	data        []time.Time
	index       int
	length      int
	mutex       sync.Mutex
	initAverage time.Duration
}

func NewTimeWindow(capacity int, defaultAverage time.Duration) *TimeWindow {
	return &TimeWindow{data: make([]time.Time, capacity), index: 0, length: 0, initAverage: defaultAverage}
}

func (tw *TimeWindow) Append(t time.Time) {
	tw.mutex.Lock()
	defer tw.mutex.Unlock()
	tw.data[tw.index] = t
	tw.index = (tw.index + 1) % cap(tw.data)
	tw.length = min(tw.length+1, cap(tw.data))
}

func (tw *TimeWindow) Average() time.Duration {
	tw.mutex.Lock()
	defer tw.mutex.Unlock()
	if tw.length < 2 {
		return tw.initAverage
	}
	head := 0
	tail := (tw.index - 1 + cap(tw.data)) % cap(tw.data)
	if tw.length == cap(tw.data) {
		head = tw.index
	}
	duration := tw.data[tail].Sub(tw.data[head])
	return duration / time.Duration(tw.length-1)
}

func (tw *TimeWindow) Reset() {
	tw.mutex.Lock()
	defer tw.mutex.Unlock()
	tw.index = 0
	tw.length = 0
}

func (tw *TimeWindow) SetDefaultAverage(defaultAverage time.Duration) {
	tw.mutex.Lock()
	defer tw.mutex.Unlock()
	tw.initAverage = defaultAverage
}
