package redisconn

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

// TODO: get this from env.
const server = "localhost:6379"

var pool = &redis.Pool{
	MaxIdle:     3,
	IdleTimeout: 240 * time.Second,
	Dial: func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", server)
		if err != nil {
			return nil, err
		}
		return c, err
	},
	TestOnBorrow: func(c redis.Conn, t time.Time) error {
		_, err := c.Do("PING")
		return err
	},
}

func Get() redis.Conn {
	return pool.Get()
}
