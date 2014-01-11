package beanstalkconn

import (
	"github.com/kr/beanstalk"
)

// TODO: from env.
const server = "localhost:11300"

func Get() (*beanstalk.Conn, error) {
	// TODO: connection pooling impl.
	return beanstalk.Dial("tcp", server)
}
