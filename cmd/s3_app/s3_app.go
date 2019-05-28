package main

import (
	funiAWS "funicular/pkg/clients/aws"
	funiRedis "funicular/pkg/clients/redis"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"strings"
)

const ENV_DIR = ".env"
const STREAM = "intra-new-outbound-vgm"
const BUCKET_NAME = "development-buyco-app-uploads"
const STORE_PATH = "/outbound/test/"

func main() {
	err := godotenv.Load(ENV_DIR)
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fileChan := make(chan redis.XMessage)
	s3Chan := make(chan string)
	go func() {
		redisPort, _ := strconv.Atoi(os.Getenv("REDIS_PORT"))
		redisDb, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
		redisCli := funiRedis.NewWrapper(
			funiRedis.Config{
				Host: os.Getenv("REDIS_HOST"),
				Port: uint16(redisPort),
				DB:   uint8(redisDb),
			},
		)
		defer func() {
			err := redisCli.Client.Close()
			if err != nil {
				log.Fatalf("failed to close redis client: %v", err)
			}
		}()

		go func() {
			for {
				select {
				case filename := <-s3Chan:
					redisCli.Client.XDel(STREAM, filename)
					log.Printf("File message stream deleted for ID: %s", filename)
				}
			}
		}()

		for {
			vals, err := redisCli.Client.XRange(STREAM, "-", "+").Result()
			if err != nil {
				log.Fatalf("failed to read redis stream: %v", err)
			}
			for _, msg := range vals {
				log.Printf("Got message with file: %s",	msg.Values["filename"].(string))
				fileChan <- msg
			}
		}
	}()

	awsManager := funiAWS.NewAWSManager(uint8(3))
	s3Bucket := awsManager.S3Manager.AddS3BucketManager()

	for {
		select {
		case fileData := <-fileChan:
			upParams := &s3manager.UploadInput{
				Bucket: aws.String(BUCKET_NAME),
				Key:    aws.String(STORE_PATH + fileData.Values["filename"].(string)),
				Body:   strings.NewReader(fileData.Values["fileData"].(string)),
			}
			result, err := s3Bucket.Uploader.Upload(upParams)
			if err != nil {
				log.Fatalf("failed to upload file, %v", err)
			}
			log.Printf("file uploaded to, %s\n", aws.StringValue(&result.Location))
			s3Chan <- fileData.ID
		}
	}
}
