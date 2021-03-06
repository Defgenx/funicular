package main

import (
	"github.com/defgenx/funicular/internal/utils"
	"github.com/defgenx/funicular/pkg/clients"

	"fmt"
	"github.com/pkg/sftp"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"time"
)

const ENV_DIR = "../../.env"
const STREAM = "intra-new-outbound-vgm"
const CONSUMER_NAME = STREAM + "-consumer"
const OUTBOUND_VGM_DIR = "./outbound/vgm/"

func main() {
	utils.LoadEnvFile(ENV_DIR, os.Getenv("ENV"))

	fileChan := make(chan map[string]interface{})
	go func() {
		var port uint32
		if portInt, err := strconv.Atoi(os.Getenv("INTRA_PORT")); err == nil {
			port = uint32(portInt)
		}
		sftpManager := clients.NewSFTPManager(
			os.Getenv("INTRA_HOST"),
			port,
			clients.NewSSHConfig(
				os.Getenv("INTRA_USER"),
				os.Getenv("INTRA_PASSWORD"),
			),
		)
		sftpConn, err := sftpManager.AddClient()
		if err != nil {
			log.Fatalf("Error #%v", err)
		}
		defer func() {
			err := sftpConn.Close()
			if err != nil {
				log.Fatalf("Failed to close SFTP client: %v", err)
			}
		}()

		tmpReadFiles := make([]os.FileInfo, 0)
		for {
			dir, err := sftpConn.Client.ReadDir(OUTBOUND_VGM_DIR)
			if err != nil {
				log.Fatalf("Cannot read dir #%v", err)
			}
			if !reflect.DeepEqual(tmpReadFiles, dir) {
				log.Print("New files detected and send in stream")

				for _, file := range dir {
					if !stringInSlice(file.Name(), tmpReadFiles) {
						fHandler, err := sftpConn.Client.Open(OUTBOUND_VGM_DIR + file.Name())
						if err != nil {
							log.Printf("Cannot read file %s #%v", file.Name(), err)
						} else {
							fileChan <- map[string]interface{}{"fileInfo": file, "fileHandler": fHandler}
						}
					}
				}
				tmpReadFiles = dir
			}
			time.Sleep(3 * time.Second)
		}
	}()

	redisPort, _ := strconv.Atoi(os.Getenv("REDIS_PORT"))
	redisDb, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	redisCli, _ := clients.NewRedisWrapper(
		clients.RedisConfig{
			Host: os.Getenv("REDIS_HOST"),
			Port: uint16(redisPort),
			DB:   uint8(redisDb),
		},
		STREAM,
		CONSUMER_NAME,
	)
	defer func() {
		err := redisCli.Close()
		if err != nil {
			log.Fatalf("Failed to close redis client: %v", err)
		}
	}()

	for {
		select {
		case fileMap := <-fileChan:
			fmt.Printf("Got file message chan: %v\n", fileMap["fileInfo"].(os.FileInfo).Name())

			fByteData, err := ioutil.ReadAll(fileMap["fileHandler"].(*sftp.File))
			if err != nil {
				log.Printf("Cannot read file data %s #%v", fileMap["fileInfo"].(os.FileInfo).Name(), err)
			} else {
				msgData := map[string]interface{}{"filename": fileMap["fileInfo"].(os.FileInfo).Name(), "fileData": fByteData}
				_, err = redisCli.AddMessage(msgData)
				if err != nil {
					log.Printf("Cannot send message: %v", err)
				}
			}
		}
	}
}

func stringInSlice(a string, list []os.FileInfo) bool {
	for _, b := range list {
		if b.Name() == a {
			return true
		}
	}
	return false
}
