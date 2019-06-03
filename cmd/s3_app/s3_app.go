package main

import (
	"github.com/defgenx/funicular/pkg/clients"
	"github.com/defgenx/funicular/internal/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/go-redis/redis"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const ENV_DIR = "../../.env"
const STREAM = "intra-new-outbound-vgm"
const BUCKET_NAME = "development-buyco-app-uploads"
const STORE_PATH = "/outbound/test/"

func main() {
	utils.LoadEnvFile(ENV_DIR, os.Getenv("ENV"))

	fileChan := make(chan redis.XMessage)
	s3Chan := make(chan string)
	go func() {
		redisPort, _ := strconv.Atoi(os.Getenv("REDIS_PORT"))
		redisDb, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
		redisCli := clients.NewRedisWrapper(
			clients.RedisConfig{
				Host: os.Getenv("REDIS_HOST"),
				Port: uint16(redisPort),
				DB:   uint8(redisDb),
			},
			STREAM,
		)
		defer func() {
			err := redisCli.Client.Close()
			if err != nil {
				log.Fatalf("Failed to close redis client: %v", err)
			}
		}()

		go func() {
			for {
				select {
				case filename := <-s3Chan:
					_, err := redisCli.DeleteMessage(filename)
					if err != nil {
						log.Fatalf("Failed to delete stream message: %v", err)
					}
					log.Printf("File message stream deleted for ID: %s", filename)
				}
			}
		}()
		lastId := "$"
		for {
			vals, err := redisCli.ReadMessage(lastId, 5, 3000 * time.Millisecond)
			if err != nil {
				log.Printf("Redis read error: %v", err)
			} else {
				NbStream := len(vals)
				NbMsgLastStreamEntry := len(vals[NbStream - 1].Messages)
				lastId = vals[NbStream - 1].Messages[NbMsgLastStreamEntry - 1].ID
				for _, msgs := range vals {
					for _, msg := range msgs.Messages {
						log.Printf("Got message with file: %s", msg.Values["filename"].(string))
						fileChan <- msg
					}
				}
			}
		}
	}()

	awsManager := clients.NewAWSManager(uint8(3))
	s3Bucket := awsManager.S3Manager.AddS3BucketManager(BUCKET_NAME)

	for {
		select {
		case fileData := <-fileChan:
			result, err := s3Bucket.UploadFile(
				STORE_PATH,
				fileData.Values["filename"].(string),
				strings.NewReader(fileData.Values["fileData"].(string)),
				)
			if err != nil {
				log.Printf("Failed to upload file, %v", err)
			} else {
				log.Printf("File uploaded to, %s\n", aws.StringValue(&result.Location))
				s3Chan <- fileData.ID
			}
		}
	}
}
