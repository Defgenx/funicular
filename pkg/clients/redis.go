package clients

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

type RedisManager struct {
	Clients map[string][]*RedisWrapper
}

type RedisWrapper struct {
	Client  *redis.Client
	config  *RedisConfig
	channel string
}

type RedisConfig struct {
	Host string
	Port uint16
	DB uint8
}


func NewRedisWrapper(config RedisConfig, channel string) *RedisWrapper {
	client := redis.NewClient(config.ToOption())
	return &RedisWrapper{
		Client: client,
		config: &config,
		channel: channel,
	}
}

func NewRedisManager() *RedisManager {
	return &RedisManager{
		Clients: make(map[string][]*RedisWrapper, 0),
	}
}

func (rw *RedisManager) add(redisWrapper *RedisWrapper, category string) {
	mm, ok := rw.Clients[category]
	if !ok {
		mm = make([]*RedisWrapper, 0)
		mm = append(mm, redisWrapper)
		rw.Clients[category] = mm
	}
}

func (rw *RedisManager) AddClient(config RedisConfig, category string, channel string) *RedisWrapper {
	if channel == "" {
		log.Fatal("Category must be filled")
	}
	if channel == "" {
		channel = category
	}
	client := NewRedisWrapper(config, channel)
	rw.add(client, category)
	return client
}

func (rw *RedisManager) Close() {
	var manageClientsCopy map[string][]*RedisWrapper
	manageClientsCopy = copyRedisClients(rw.Clients)
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

func (rc *RedisConfig) ToOption() *redis.Options {
	return &redis.Options{
		Addr: net.JoinHostPort(rc.Host, strconv.Itoa(int(rc.Port))),
		DB: int(rc.DB),
	}
}

func (w *RedisWrapper) SendMessage(data map[string]interface{}) (string, error) {
	xAddArgs := &redis.XAddArgs{
		Stream: w.channel,
		Values: data,
	}
	result := w.Client.XAdd(xAddArgs)
	return result.Result()
}

func (w *RedisWrapper) ReadMessage(last_id string, count int64, block time.Duration) ([]redis.XStream, error) {
	var channels = make([]string, 0)
	channels = append(channels, w.channel)
	channels = append(channels, last_id)
	xReadArgs := &redis.XReadArgs{
		Streams: channels,
		Count: count,
		Block: block,
	}
	result := w.Client.XRead(xReadArgs)
	return result.Result()
}

func (w *RedisWrapper) ReadRangeMessage(start string, stop string) ([]redis.XMessage, error) {
	result := w.Client.XRange(w.channel, start, stop)
	return result.Result()
}

func (w *RedisWrapper) DeleteMessage(ids ...string) (int64, error) {
	result := w.Client.XDel(w.channel, ids...)
	return result.Result()
}

func (w *RedisWrapper) CreateGroup(group string, start string) (string, error) {
	result := w.Client.XGroupCreate(w.channel, group, start)
	return result.Result()
}

func (w *RedisWrapper) DeleteGroup(group string) (int64, error) {
	result := w.Client.XGroupDestroy(w.channel, group)
	return result.Result()
}

func (w *RedisWrapper) PendingMessage(group string) (*redis.XPending, error) {
	result := w.Client.XPending(w.channel, group)
	return result.Result()
}

func (w *RedisWrapper) AckMessage(group string, ids ...string) (int64, error) {
	result := w.Client.XAck(w.channel, group, ids...)
	return result.Result()
}

func copyRedisClients(originalMap map[string][]*RedisWrapper) map[string][]*RedisWrapper {
	var newMap = make(map[string][]*RedisWrapper)
	for k,v := range originalMap {
		newMap[k] = v
	}
	return newMap
}