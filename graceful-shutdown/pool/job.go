package pool

import "fmt"

// Job is a job that can be executed by a worker.
type Job struct {
	run func()
}

// NewJob creates a new job from a passed-in function.
func NewJob(f func()) Job {
	return Job{
		run: f,
	}
}

// String returns a string representation of the job.
func (j Job) String() string {
	return fmt.Sprintf("Job(%p)", j.run)
}
