package dealer

type JobResult struct {
	Out any
	Err error
}

func NewJobResult(out any, err error) *JobResult {
	return &JobResult{
		Out: out,
		Err: err,
	}
}

type JobFunc func() *JobResult

type Job struct {
	f        JobFunc
	resultch chan *JobResult
}

func (j *Job) Wait() *JobResult {
	return <-j.resultch
}

func newJob(f JobFunc) *Job {
	return &Job{
		f:        f,
		resultch: make(chan *JobResult),
	}
}
