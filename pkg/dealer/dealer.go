// package dealer helps to work with worker pool or semaphores
// it provides a lot of functionality and manageable workflow

package dealer

import (
	"sync"
	"sync/atomic"

	"go.uber.org/zap"
)

// Strategy Default is Semaphore
type Strategy int

const (
	Semaphore Strategy = iota
	WorkerPool
)

type Dealer struct {
	sem      chan struct{}
	shutdown chan interface{}
	jobq     chan *Job
	logger   *zap.SugaredLogger
	wg       *sync.WaitGroup
	strategy Strategy
	// 0 - stopped
	// 1 - started
	started    int32
	maxWorkers int
}

func New(logger *zap.SugaredLogger, maxWorkers int) *Dealer {
	return &Dealer{
		started:    0,
		logger:     logger,
		maxWorkers: maxWorkers,
		sem:        make(chan struct{}, maxWorkers),
		jobq:       make(chan *Job, maxWorkers),
		shutdown:   make(chan interface{}),
		wg:         new(sync.WaitGroup),
	}
}

// WithStrategy sets strategy to a dealer instance
func (d *Dealer) WithStrategy(strategy Strategy) {
	d.strategy = strategy
}

func (d *Dealer) Start() {
	atomic.StoreInt32(&d.started, 1)
	switch d.strategy {
	case Semaphore:
		go d.startWithSemaphore()
		d.logger.Debugf("dealing has started with semaphore")
	case WorkerPool:
		go d.startWorkerPool()
		d.logger.Debugf("dealing has started with workerPool")
	}
}

func (d *Dealer) Stop() {
	atomic.StoreInt32(&d.started, 0)
	//If worker pool is selected, should close jobq first, to stop all workers
	switch d.strategy {
	case Semaphore:
		d.wg.Wait()
		close(d.jobq)
		close(d.shutdown)
	case WorkerPool:
		close(d.jobq)
		d.wg.Wait()
		close(d.shutdown)
	}
}

func (d *Dealer) Run(f JobFunc) *Job {
	j := newJob(f)
	d.addJob(j)
	return j
}

func (d *Dealer) addJob(j *Job) {
	if atomic.LoadInt32(&d.started) == 0 {
		panic("dealer has not started yet!")
	}
	d.jobq <- j
}

func (d *Dealer) startWorkerPool() {
	for n := 1; n <= d.maxWorkers; n++ {
		d.wg.Add(1)
		go d.startWorker()
	}
}

// TODO: add timeout
func (d *Dealer) startWorker() {
	for j := range d.jobq {
		j.resultch <- j.f()
	}
	defer d.wg.Done()
}

func (d *Dealer) startWithSemaphore() {
	for j := range d.jobq {
		d.acquire()
		d.wg.Add(1)
		go func(j *Job) {
			j.resultch <- j.f()
			defer func() {
				d.wg.Done()
				d.release()
			}()
		}(j)
	}
}

func (d *Dealer) acquire() {
	d.sem <- struct{}{}
}

func (d *Dealer) release() {
	<-d.sem
}
