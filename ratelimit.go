package hermes

import (
	// "fmt"
	"time"
)

type RateLimiter struct {
	Interval       int64 // milisecond
	ResetOnRequest bool
	records        map[string]time.Time
}

func (rl *RateLimiter) Check(token string) bool {
	if token == "" {
		return false
	}
	lastRec, ok := rl.records[token]
	if !ok {
		//record does not exists
		rl.records[token] = time.Now()
		return true
	} else {
		if int64(time.Since(lastRec)) > rl.Interval*int64(time.Millisecond) {
			// passed rate
			rl.records[token] = time.Now()
			return true
		} else {
			if rl.ResetOnRequest {
				rl.records[token] = time.Now()
			}
			return false
		}
	}
	// return false
}

func (rl *RateLimiter) Reset(token string) {
	if token == "" {
		return
	}
	delete(rl.records, token)
}

func NewRateLimiter(interval int64) *RateLimiter {
	rl := &RateLimiter{}
	rl.Interval = interval
	rl.records = make(map[string]time.Time, 1000)
	return rl
}
