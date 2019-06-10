package utils

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func LoadEnvFile(envFile string, runningEnv string) {
	allowedRunningEnv := []string{"development"}
	exists, _ := InArray(runningEnv, allowedRunningEnv)
	if exists {
		path, _ := os.Getwd()
		err := godotenv.Load(fmt.Sprintf("%s/%s", path, envFile))
		if err != nil {
			log.Fatal("Error loading environment file")
		}
	} else {
		log.Print("Environment file not loaded for the current env")
	}
}
