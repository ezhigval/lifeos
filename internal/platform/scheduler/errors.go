package scheduler

import "time"

type DeferJobError struct {
	Delay time.Duration
}

func (e DeferJobError) Error() string {
	return "defer job"
}

func Defer(delay time.Duration) error {
	return DeferJobError{Delay: delay}
}
