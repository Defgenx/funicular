package clients

import (
	"github.com/go-redis/redis"
	"log"
	"net"
	"strconv"
)

type RedisManager struct {
	Clients []*RedisWrapper
}

type RedisWrapper struct {
	Client *redis.Client
	config *RedisConfig
}

type RedisConfig struct {
	Host string
	Port uint16
	DB uint8
}


func NewRedisClient(config RedisConfig) *RedisWrapper {
	client := redis.NewClient(config.ToOption())
	return &RedisWrapper{
		Client: client,
		config: &config,
	}
}

func NewRedisManager() *RedisManager {
	return &RedisManager{
		Clients: make([]*RedisWrapper, 0),
	}
}

func (rw *RedisManager) AddClient(config RedisConfig) *RedisWrapper {
	client := NewRedisClient(config)
	rw.Clients = append(rw.Clients, client)
	return client
}

func (rw *RedisManager) Close() {
	var manageClientsCopy []*RedisWrapper
	copy(manageClientsCopy, rw.Clients)
	for _, client := range manageClientsCopy {
		err := client.Client.Close()
		if err != nil {
			log.Fatalf("An error occurred while closing client connexion pool: %v", err)
		}
		rw.Clients = rw.Clients[1:]
	}
}

func (rc *RedisConfig) ToOption() *redis.Options {
	return &redis.Options{
		Addr: net.JoinHostPort(rc.Host, strconv.Itoa(int(rc.Port))),
		DB: int(rc.DB),
	}
}