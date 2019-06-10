package clients_test

import (
	"github.com/defgenx/funicular/internal/utils"
	. "github.com/defgenx/funicular/pkg/clients"

	"github.com/go-redis/redis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"strconv"
	"time"
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
		var category = "test"
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

		var group = "foo-group"
		var validDefaultMsgId = "1538561700640-0"
		var malformedMsgId = "foo:bar"

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

			It("should fail to read message", func() {
				_, readErr := wrapper.ReadMessage("$", 1, 100*time.Millisecond)
				Expect(readErr).To(
					SatisfyAll(
						HaveOccurred(),
						MatchError("redis: nil"),
					),
				)

				_, readErr = wrapper.ReadRangeMessage("-", "+")
				Expect(readErr).ToNot(HaveOccurred())
			})

			It("should not have messages to delete", func() {
				id, readErr := wrapper.DeleteMessage(validDefaultMsgId)
				Expect(id).To(BeZero())
				Expect(readErr).ToNot(HaveOccurred())
			})

			It("should fail to delete malformed message ID", func() {
				id, readErr := wrapper.DeleteMessage(malformedMsgId)
				Expect(id).To(BeZero())
				Expect(readErr).To(
					SatisfyAll(
						HaveOccurred(),
						MatchError("ERR Invalid stream ID specified as stream command argument"),
					),
				)
			})

			Context("When no group exists", func() {

				It("should not have messages to acknowledge", func() {
					id, readErr := wrapper.AckMessage(group, malformedMsgId)
					Expect(id).To(BeZero())
					Expect(readErr).ToNot(HaveOccurred())
				})

				It("should not have pending messages", func() {
					_, readErr := wrapper.PendingMessage(group)
					Expect(readErr).To(HaveOccurred())
				})
			})

			Context("When a group exists", func() {

				var cliAddGrpResponse string
				var errAddGrp error

				BeforeEach(func() {
					cliAddGrpResponse, errAddGrp = wrapper.CreateGroup(group, "$")
				})

				AfterEach(func() {
					_, _ = wrapper.DeleteGroup(group)
				})

				It("should have created a new group", func() {
					Expect(cliAddGrpResponse).To(Equal("OK"))
					Expect(errAddGrp).ToNot(HaveOccurred())
				})

				It("should fail to create same group", func() {
					failResp, errSameAddGrp := wrapper.CreateGroup(group, "$")
					Expect(failResp).To(BeEmpty())
					Expect(errSameAddGrp).To(
						SatisfyAll(
							HaveOccurred(),
							MatchError("BUSYGROUP Consumer Group name already exists"),
						),
					)
				})

				It("should delete the group", func() {
					delGrp, errDelGrp := wrapper.DeleteGroup(group)
					Expect(delGrp).To(Equal(int64(1)))
					Expect(errDelGrp).ToNot(HaveOccurred())
				})

				It("should fail to acknowledge malformed message ID", func() {
					ackMsgGrp, errAckMgsGrp := wrapper.AckMessage(group, malformedMsgId)
					Expect(ackMsgGrp).To(BeZero())
					Expect(errAckMgsGrp).To(
						SatisfyAll(
							HaveOccurred(),
							MatchError("ERR Invalid stream ID specified as stream command argument"),
						),
					)
				})

				It("should not have message to acknowledge", func() {
					ackMsgGrp, errAckMgsGrp := wrapper.AckMessage(group, validDefaultMsgId)
					Expect(ackMsgGrp).To(BeZero())
					Expect(errAckMgsGrp).ToNot(HaveOccurred())
				})
			})
		})

		Context("When Redis stream channel is filled", func() {

			message := map[string]interface{}{"foo": "bar"}
			var msgId string

			BeforeEach(func() {
				msgId, _ = wrapper.AddMessage(message)
			})

			It("should read message", func() {
				msg, readErr := wrapper.ReadMessage("$", 1, 100*time.Millisecond)
				Expect(readErr).ToNot(HaveOccurred())
				Expect(msg).To(
					SatisfyAll(
						HaveLen(1),
						ContainElement(message),
					),
				)
			})

			It("should delete message", func() {
				nbMsg, readErr := wrapper.DeleteMessage(msgId)
				Expect(nbMsg).To(Equal(int64(1)))
				Expect(readErr).ToNot(HaveOccurred())
			})
		})
	})
})
