package hw5

import (
	"errors"
	"fmt"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

type errCounter struct {
	count int
	limit int
	mx    sync.Mutex
}

func (ec *errCounter) inc() {
	ec.mx.Lock()
	defer ec.mx.Unlock()
	ec.count++
}
func (ec *errCounter) isLimit() bool {
	ec.mx.Lock()
	defer ec.mx.Unlock()
	return ec.count >= ec.limit
}

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if n <= 0 {
		return fmt.Errorf("n must be positive: %d", n)
	}

	if len(tasks) == 0 {
		return nil
	}

	ec := errCounter{
		count: 0,
		limit: m,
		mx:    sync.Mutex{},
	}

	wg := sync.WaitGroup{}
	tasksChan := make(chan Task, len(tasks))

	worker := func() {
		defer wg.Done()
		for task := range tasksChan {
			if ec.isLimit() {
				return
			}
			if err := task(); err != nil {
				ec.inc()
			}
		}
	}

	for i := 0; i < n; i++ {
		wg.Add(1)
		go worker()
	}
	for _, task := range tasks {
		if ec.isLimit() {
			break
		}
		tasksChan <- task
	}

	close(tasksChan)
	wg.Wait()
	if ec.isLimit() {
		return ErrErrorsLimitExceeded
	}
	return nil
}
