package clients_test

import (
	"github.com/defgenx/funicular/internal/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"strconv"

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
	//var wrapper = NewRedisWrapper(config, "test-channel")

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

			It("should contains zero clients", func() {
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

			It("should close all clients", func() {
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
})