package jobs

import (
	"encoding/json"
	"io"
	"time"

	"github.com/samcday/hosted-javadocsets/beanstalkconn"
)

type Job struct {
	Payload  func() map[string]string
	Complete func()
}

// QueueJob will queue up a job in beanstalk with given payload.
func QueueJob(payload map[string]string) error {
	beanstalkConn, err := beanstalkconn.Get()
	if err != nil {
		return err
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = beanstalkConn.Put(b, 1, 0, 0)
	return err
}

// TakeJob will grb the next job in queue and return a Job interface to read
// the Job payload and mark the Job as complete. If timeout is 0 a default of
// 5 seconds will be used when reserving job.
func TakeJob(timeout time.Duration) (*Job, error) {
	beanstalkConn, err := beanstalkconn.Get()
	if err != nil {
		return nil, err
	}

	if timeout == 0 {
		timeout = 5 * time.Second
	}

	id, body, err := beanstalkConn.Reserve(timeout)
	if err != nil {
		return nil, err
	}

	var payload map[string]string
	if err = json.Unmarshal(body, &payload); err != nil && err != io.EOF {
		return nil, err
	}

	return &Job{
		Payload: func() map[string]string {
			return payload
		},
		Complete: func() {
			_ = beanstalkConn.Delete(id)
		},
	}, nil
}
