package dispatcher

import (
	"net/http"
	"sync"
	"time"
)

type Dispatcher struct {
	startOnce sync.Once
	queue     chan taskData
	opts      Opts
}

type Opts struct {
	QueueSize int
	MinWait   time.Duration
}

type taskData struct {
	req        *http.Request
	resultChan chan<- taskResult
}

type taskResult struct {
	resp *http.Response
	err  error
}

func NewDispatcher(opts Opts) Dispatcher {
	return Dispatcher{queue: make(chan taskData, opts.QueueSize), opts: opts}
}

func (x *Dispatcher) Do(req *http.Request) (*http.Response, error) {
	x.start()
	resultChan := make(chan taskResult, 1)
	x.queue <- taskData{req, resultChan}
	result := <-resultChan
	return result.resp, result.err
}

func (x *Dispatcher) start() {
	x.startOnce.Do(func() {
		ticker := time.NewTicker(x.opts.MinWait)
		defer ticker.Stop()
		go func() {
			for range ticker.C {
				task := <-x.queue
				resp, err := http.DefaultClient.Do(task.req)
				task.resultChan <- taskResult{resp, err}
				close(task.resultChan)
			}
		}()
	})
}
