package redis

import (
	"github.com/go-redis/redis"
	"log"
	"net"
	"strconv"
	"time"
)

type SendResponse struct {
	Id string
	Val map[string]interface{}
}

type Manager struct {
	Clients map[string][]*Wrapper
}

type Wrapper struct {
	Client  *redis.Client
	config  *Config
	channel string
}

type Config struct {
	Host string
	Port uint16
	DB uint8
}


func NewWrapper(config Config, channel string) *Wrapper {
	client := redis.NewClient(config.ToOption())
	return &Wrapper{
		Client: client,
		config: &config,
		channel: channel,
	}
}

func NewManager() *Manager {
	return &Manager{
		Clients: make(map[string][]*Wrapper, 0),
	}
}

func (rw *Manager) add(redisWrapper *Wrapper, category string) {
	mm, ok := rw.Clients[category]
	if !ok {
		mm = make([]*Wrapper, 0)
		mm = append(mm, redisWrapper)
		rw.Clients[category] = mm
	}
}

func (rw *Manager) AddClient(config Config, category string, channel string) *Wrapper {
	if channel == "" {
		log.Fatal("Category must be filled")
	}
	if channel == "" {
		channel = category
	}
	client := NewWrapper(config, channel)
	rw.add(client, category)
	return client
}

func (rw *Manager) Close() {
	var manageClientsCopy map[string][]*Wrapper
	manageClientsCopy = copy(rw.Clients)
	for category, clients := range manageClientsCopy {
		for _, client := range clients {
			err := client.Client.Close()
			if err != nil {
				log.Fatalf("An error occurred while closing client connexion pool: %v", err)
			}
			rw.Clients[category] = rw.Clients[category][1:]
		}
		delete(rw.Clients, category)
	}
}

func (rc *Config) ToOption() *redis.Options {
	return &redis.Options{
		Addr: net.JoinHostPort(rc.Host, strconv.Itoa(int(rc.Port))),
		DB: int(rc.DB),
	}
}

func (w *Wrapper) SendStreamMessage(data map[string]interface{}) (string, error) {
	xAddArgs := &redis.XAddArgs{
		Stream: w.channel,
		Values: data,
	}
	result := w.Client.XAdd(xAddArgs)
	return result.Result()
}

func (w *Wrapper) ReadMessage(count int64, block time.Duration) ([]redis.XStream, error) {
	var channels = make([]string, 0)
	channels = append(channels, w.channel)
	xReadArgs := &redis.XReadArgs{
		Streams: channels,
		Count: count,
		Block: block,
	}
	result := w.Client.XRead(xReadArgs)
	return result.Result()
}

func (w *Wrapper) ReadRangeMessage(start string, stop string) ([]redis.XMessage, error) {
	result := w.Client.XRange(w.channel, start, stop)
	return result.Result()
}

func (w *Wrapper) DeleteMessages(ids ...string) (int64, error) {
	result := w.Client.XDel(w.channel, ids...)
	return result.Result()
}

func (w *Wrapper) CreateGroup(group string, start string) (string, error) {
	result := w.Client.XGroupCreate(w.channel, group, start)
	return result.Result()
}

func (w *Wrapper) DeleteGroup(group string) (int64, error) {
	result := w.Client.XGroupDestroy(w.channel, group)
	return result.Result()
}

func (w *Wrapper) PendingMessages(group string) (*redis.XPending, error) {
	result := w.Client.XPending(w.channel, group)
	return result.Result()
}

func (w *Wrapper) AckMessages(group string, ids ...string) (int64, error) {
	result := w.Client.XAck(w.channel, group, ids...)
	return result.Result()
}

func copy(originalMap map[string][]*Wrapper) map[string][]*Wrapper {
	var newMap = make(map[string][]*Wrapper)
	for k,v := range originalMap {
		newMap[k] = v
	}
	return newMap
}