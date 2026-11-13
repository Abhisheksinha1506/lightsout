package wait

import "time"

// Until calls done every interval until it returns true or timeout passes.
//
// It returns true as soon as done does, and false if done is still false once
// timeout has elapsed. done is always called at least once, and once more at
// the deadline, so a change that lands just before the timeout is not missed.
func Until(timeout, interval time.Duration, done func() bool) bool {
	deadline := time.Now().Add(timeout)
	for {
		if done() {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(interval)
	}
}
