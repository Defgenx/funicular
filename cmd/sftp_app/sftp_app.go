package main

import (
	"fmt"
	"funicular/pkg/clients"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"time"
)

const ENV_DIR = ".env"

func main() {
	err := godotenv.Load(ENV_DIR)
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var port uint32
	if portInt, err := strconv.Atoi(os.Getenv("INTRA_PORT")); err == nil {
		port = uint32(portInt)
	}
	sftpManager := clients.NewSFTPManager(
		os.Getenv("INTRA_HOST"),
		port,
		os.Getenv("INTRA_USER"),
		os.Getenv("INTRA_PASSWORD"),
	)
	sftpConn, err := sftpManager.AddClient()
	if err != nil {
		log.Fatalf("Error #%v", err)
	}

	files := make(chan []os.FileInfo)
	go func() {
		for {
			dir, err := sftpConn.Client.ReadDir("./")
			if err != nil {
				log.Fatalf("Cannot read dir #%v", err)
			}
			files <- dir
			time.Sleep(3 * time.Second)
		}
	}()
	var counter int
	for {
		select {
		case res := <-files:
			for _, file := range res {
				fmt.Printf("%v\n", file.Name())
			}
		}
		counter++
		if counter == 10 {
			os.Exit(0)
		}
	}

}
