package dealer

import (
	"bytes"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const maxWorkers = 5

func TestCanStartStop(t *testing.T) {

	logger, _ := zap.NewProduction()

	d := New(logger.Sugar(), maxWorkers)
	d.Start(true)
	d.Stop()

}

func TestCanExecuteJobsWithWorkerPool(t *testing.T) {
	t.Parallel()

	logger, _ := zap.NewProduction()
	d := New(logger.Sugar(), maxWorkers)
	d.WithStrategy(WorkerPool)

	d.Start(true)

	f := func() *JobResult {
		for i := 0; i < 50; i++ {
		}
		return NewJobResult(nil, nil)
	}

	j := newJob(f)
	for i := 0; i < 50; i++ {
		d.addJob(j)
		res := j.Wait()
		require.NoError(t, res.Err)
		require.Nil(t, res.Out)
	}

	d.Stop()

}

func TestWorkerPoolIsFasterThanSemaphore(t *testing.T) {
	t.Parallel()
	logger := zap.NewNop()

	//Semaphore pattern means for each job one goroutine will be created
	//thus, giving a lot of overhead if jobs are constantly being added.
	//
	//In this example, workerPool should win in time, because only once
	//100 workers will be created
	//Semaphore in a long run might cause big GC pauses and sheduling overall.
	localMaxWorkers := 200
	d := New(logger.Sugar(), localMaxWorkers)
	d.WithStrategy(Semaphore)

	d.Start(true)

	f := func() *JobResult {
		for i := 0; i < 50; i++ {
		}
		return NewJobResult(nil, nil)
	}

	startSemaphore := time.Now()
	j := newJob(f)
	for i := 0; i < 10000; i++ {

		d.addJob(j)
		res := j.Wait()
		require.NoError(t, res.Err)
		require.Nil(t, res.Out)
	}

	elapsedSemaphore := time.Now().Sub(startSemaphore).Microseconds()
	d.Stop()

	// ------------------------------

	d2 := New(logger.Sugar(), localMaxWorkers)
	d2.WithStrategy(WorkerPool)

	//sleep is required for all workers to spin-up
	d2.Start(true)
	time.Sleep(time.Millisecond * 50)

	f2 := func() *JobResult {
		for i := 0; i < 50; i++ {
		}
		return NewJobResult(nil, nil)
	}

	startPool := time.Now()
	j2 := newJob(f2)
	for i := 0; i < 10000; i++ {
		d2.addJob(j2)
		res := j2.Wait()
		require.NoError(t, res.Err)
		require.Nil(t, res.Out)

	}
	elapsedPool := time.Now().Sub(startPool).Microseconds()

	d2.Stop()

	//This test shows 3x-5x more perfomance with workers
	//See explanation at the top of the test
	//Semaphore might be usefull to do batch-jobs at once ASAP
	require.Less(t, elapsedPool, elapsedSemaphore)
}

func TestCanExecuteOneJobSync(t *testing.T) {
	t.Parallel()
	logger := zap.NewNop()

	d := New(logger.Sugar(), maxWorkers)
	d.Start(true)

	f := func() *JobResult {
		for i := 0; i < 50; i++ {
		}
		return NewJobResult(nil, nil)
	}

	j := newJob(f)
	for i := 0; i < 50; i++ {
		d.addJob(j)
		res := j.Wait()
		require.NoError(t, res.Err)
	}

	d.Stop()
}
func TestCanExecuteOneJobAsync(t *testing.T) {
	t.Parallel()
	logger := zap.NewNop()

	d := New(logger.Sugar(), maxWorkers)
	d.Start(true)

	f := func() *JobResult {
		for i := 0; i < 50; i++ {
		}
		return NewJobResult(nil, nil)
	}

	j := newJob(f)
	wg := new(sync.WaitGroup)

	for i := 0; i < 50; i++ {
		//async
		wg.Add(1)
		go func() {
			d.addJob(j)
			res := j.Wait()
			wg.Done() //important order
			require.NoError(t, res.Err)
		}()
	}

	wg.Wait()
	d.Stop()
}
func TestCanExecuteLongJobsAndReadErrorsAsync(t *testing.T) {
	t.Parallel()

	logger := zap.NewNop()

	d := New(logger.Sugar(), maxWorkers)
	jobCount := 50

	//Assuming that f has time.Sleep = 200ms,
	//having 5 workers and 50 jobs for 200ms results in 2 seconds of execute time.
	//Add 1% of time to each of jobs, so time becomes 200ms * 1.01 = 202ms
	//Resulting totally in (50 * 202ms) / (maxWorkers * 1000ms) = 2.02 seconds
	//The success will be if:
	// - limit(2.02seconds) - actualExecutionTime > 0
	limit := (float64(200) * float64(jobCount) * 1.01) / float64(maxWorkers*1000)

	d.Start(true)

	wg := new(sync.WaitGroup)
	start := time.Now()
	for i := 0; i < jobCount; i++ {
		wg.Add(1)
		go func() {
			f := func() *JobResult {
				time.Sleep(time.Millisecond * 200)
				err := errors.New("err")
				return NewJobResult(nil, err)
			}
			j := newJob(f)
			d.addJob(j)
			res := j.Wait()
			wg.Done()
			require.Error(t, res.Err)
		}()
	}
	wg.Wait()
	d.Stop()

	elapsed := time.Now().Sub(start).Seconds()
	t.Logf("%.4f; %.4f", limit, elapsed)
	require.Less(t, elapsed, limit)

}
func TestCanExecuteJobsAndReceiveOutput(t *testing.T) {
	t.Parallel()

	logger := zap.NewNop()

	d := New(logger.Sugar(), maxWorkers)
	d.Start(true)

	for i := 0; i < 50; i++ {
		j := newJob(func() *JobResult {
			buff := bytes.NewBuffer(nil)
			buff.WriteString("Hello world!")
			buff.WriteString("Hello world!")
			return NewJobResult(buff, nil)
		})
		d.addJob(j)

		result := j.Wait()

		str := result.Out.(*bytes.Buffer).String()
		require.Equal(t, "Hello world!Hello world!", str)
		require.NoError(t, result.Err)
	}

	d.Stop()
}
