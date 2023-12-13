package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Job represents a task with a request function that returns a response of type T and an error.
type Job[T any] struct {
	Request func(ctx context.Context) (T, error)
}

type SchedulerResponse[T any] struct {
	Response *T
	Err      error
}

// Scheduler handles scheduling and executing Jobs with exponential backoff.
type Scheduler[T any] struct {
	requests        []*Job[T]
	initialInterval time.Duration
	maxInterval     time.Duration
	maxElapsedTime  time.Duration
	multiplier      float64
}

// RunExponentialBackOff executes all Jobs in the scheduler with exponential backoff.
func (s *Scheduler[T]) RunExponentialBackOff() []SchedulerResponse[T] {
	limiter := rate.NewLimiter(rate.Every(s.initialInterval), 1)
	var wg sync.WaitGroup
	schedulerResponses := make([]SchedulerResponse[T], len(s.requests))

	for i, job := range s.requests {
		wg.Add(1)
		go func(j *Job[T], index int) {
			defer wg.Done()

			interval := s.initialInterval
			maxRetries := 5 // Define the maximum number of retries
			var err error
			for attempt := 0; attempt < maxRetries; attempt++ {
				if err := limiter.Wait(context.Background()); err != nil {
					fmt.Println(err)
					schedulerResponses[index] = SchedulerResponse[T]{Response: nil, Err: err}
					return
				}

				response, err := j.Request(context.Background())
				if err == nil {
					schedulerResponses[index] = SchedulerResponse[T]{Response: &response, Err: nil}
					return
				}
				if err != nil {
					fmt.Println(err)
				}

				// Log the error or handle it as needed
				time.Sleep(interval)
				interval = time.Duration(float64(interval) * s.multiplier)
				if interval > s.maxInterval {
					interval = s.maxInterval
				}
			}

			if err != nil {
				fmt.Println(err)
				schedulerResponses[index] = SchedulerResponse[T]{Response: nil, Err: err}
			}
		}(job, i)
	}

	wg.Wait()
	return schedulerResponses
}

// CreateScheduler creates a new Scheduler with the given Jobs and default configuration.
func CreateScheduler[T any](jobs []*Job[T]) Scheduler[T] {
	return Scheduler[T]{
		requests:        jobs,
		initialInterval: 500 * time.Nanosecond,
		maxInterval:     60 * time.Second,
		maxElapsedTime:  15 * time.Minute,
		multiplier:      2,
	}
}
