package clients_test

import (
	"github.com/defgenx/funicular/internal/utils"
	"github.com/go-redis/redis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"strconv"
	"time"

	. "github.com/defgenx/funicular/pkg/clients"
)

var _ = Describe("Redis", func() {
	// Declaring var for tests
	port, _ := strconv.Atoi(os.Getenv("REDIS_PORT"))
	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	var config = RedisConfig{
		Host: os.Getenv("REDIS_HOST"),
		Port: uint16(port),
		DB:   uint8(db),
	}
	var wrapper, nilErr = NewRedisWrapper(config, "test-channel")

	Describe("Using Manager", func() {
		var manualManager = &RedisManager{
			Clients: make(map[string][]*RedisWrapper),
		}
		var category= "test"
		var manager *RedisManager

		BeforeEach(func() {
			manager = NewRedisManager()
		})

		Context("From constructor function", func() {
			It("should create a valid instance", func() {
				Expect(manager).To(Equal(manualManager))
			})

			It("should contain zero clients", func() {
				Expect(len(manager.Clients)).To(BeZero())
			})
		})

		It("should fail to add client with empty category", func() {
			client, err := manager.AddClient(config, "", "test")
			Expect(err).To(HaveOccurred())
			Expect(client).To(BeNil())
		})

		Context("Without Redis client in the stack", func() {
			It("should use category as channel if channel is empty and add client to manager", func() {
				client, err := manager.AddClient(config, category, "")
				Expect(err).ToNot(HaveOccurred())
				Expect(client.GetChannel()).To(Equal(category))
				Expect(manager.Clients[category]).To(HaveLen(1))
				Expect(manager.Clients[category][0]).To(Equal(client))
			})

			It("should not close clients", func() {
				var err error
				stdout := utils.CaptureStdout(func() { err = manager.Close() })
				Expect(err).ToNot(HaveOccurred())
				Expect(stdout).To(ContainSubstring("Manager have no clients to close..."))
			})
		})

		Context("With Redis clients in the stack", func() {
			It("should use category as channel if channel is empty and add client to manager", func() {
				client, err := manager.AddClient(config, category, "")
				client2, err2 := manager.AddClient(config, category, "")
				Expect(err).ToNot(HaveOccurred())
				Expect(client.GetChannel()).To(Equal(category))
				Expect(err2).ToNot(HaveOccurred())
				Expect(client2.GetChannel()).To(Equal(category))
				Expect(manager.Clients[category]).To(HaveLen(2))
				Expect(manager.Clients[category][0]).To(Equal(client))
				Expect(manager.Clients[category][1]).To(Equal(client2))
			})

			It("should close all clients", func() {
				_, _ = manager.AddClient(config, category, "")
				_, _ = manager.AddClient(config, category, "")
				var err error
				stdout := utils.CaptureStdout(func() { err = manager.Close() })
				Expect(err).ToNot(HaveOccurred())
				Expect(stdout).ToNot(ContainSubstring("Manager have no clients to close..."))
			})
		})
	})


	Describe("Using Wrapper", func() {
		Context("From constructor function", func() {
			It("should create a valid instance", func() {
				Expect(nilErr).ToNot(HaveOccurred())
				Expect(wrapper.Client).To(BeAssignableToTypeOf(&redis.Client{}))
				Expect(wrapper.GetChannel()).To(Equal("test-channel"))
			})

			It("should fail with empty string for channel", func() {
				_, filledErr := NewRedisWrapper(config, "")
				Expect(filledErr).To(
					SatisfyAll(
						HaveOccurred(),
						MatchError("channel must be filled"),
					),
				)
			})
		})

		Context("When Redis stream channel is empty", func() {
			It("should fail to read", func() {
				_, readErr := wrapper.ReadMessage("$", 1, 100 * time.Millisecond)
				Expect(readErr).To(
					SatisfyAll(
						HaveOccurred(),
						MatchError("redis: nil"),
					),
				)

				_, readErr = wrapper.ReadRangeMessage("-", "+")
				Expect(readErr).ToNot(HaveOccurred())
			})
		})
	})
})