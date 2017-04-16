package redis

import (
	"time"

	"github.com/admpub/ddns/store"
	"github.com/garyburd/redigo/redis"
)

var _ store.Storer = &RedisConnection{}

const HostExpirationSeconds int = 10 * 24 * 60 * 60 // 10 Days

type RedisConnection struct {
	*redis.Pool
}

func OpenConnection(server string) *RedisConnection {
	return &RedisConnection{newPool(server)}
}

func newPool(server string) *redis.Pool {
	return &redis.Pool{
		MaxIdle: 3,

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
}

func (self *RedisConnection) GetHost(name string) *store.Host {
	conn := self.Get()
	defer conn.Close()

	host := store.Host{Hostname: name}

	if self.HostExist(name) {
		data, err := redis.Values(conn.Do("HGETALL", host.Hostname))
		store.HandleErr(err)

		store.HandleErr(redis.ScanStruct(data, &host))
	}

	return &host
}

func (self *RedisConnection) SaveHost(host *store.Host) {
	conn := self.Get()
	defer conn.Close()

	_, err := conn.Do("HMSET", redis.Args{}.Add(host.Hostname).AddFlat(host)...)
	store.HandleErr(err)

	_, err = conn.Do("EXPIRE", host.Hostname, HostExpirationSeconds)
	store.HandleErr(err)
}

func (self *RedisConnection) HostExist(name string) bool {
	conn := self.Get()
	defer conn.Close()

	exists, err := redis.Bool(conn.Do("EXISTS", name))
	store.HandleErr(err)

	return exists
}
