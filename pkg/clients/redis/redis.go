package redis

import (
	"github.com/go-redis/redis"
	"log"
	"net"
	"strconv"
)

type Manager struct {
	Clients []*Wrapper
}

type Wrapper struct {
	Client *redis.Client
	config *Config
}

type Config struct {
	Host string
	Port uint16
	DB uint8
}


func NewWrapper(config Config) *Wrapper {
	client := redis.NewClient(config.ToOption())
	return &Wrapper{
		Client: client,
		config: &config,
	}
}

func NewManager() *Manager {
	return &Manager{
		Clients: make([]*Wrapper, 0),
	}
}

func (rw *Manager) AddClient(config Config) *Wrapper {
	client := NewWrapper(config)
	rw.Clients = append(rw.Clients, client)
	return client
}

func (rw *Manager) Close() {
	var manageClientsCopy []*Wrapper
	copy(manageClientsCopy, rw.Clients)
	for _, client := range manageClientsCopy {
		err := client.Client.Close()
		if err != nil {
			log.Fatalf("An error occurred while closing client connexion pool: %v", err)
		}
		rw.Clients = rw.Clients[1:]
	}
}

func (rc *Config) ToOption() *redis.Options {
	return &redis.Options{
		Addr: net.JoinHostPort(rc.Host, strconv.Itoa(int(rc.Port))),
		DB: int(rc.DB),
	}
}