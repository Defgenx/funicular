package utils

import (
	"github.com/joho/godotenv"
	"log"
)

func LoadEnvFile(envDir string, runningEnv string) {
	allowedRunningEnv := []string{"development", "testing"}
	exists, _ := InArray(runningEnv, allowedRunningEnv)
	if exists {
		err := godotenv.Load(envDir)
		if err != nil {
			log.Fatal("Error loading environment file")
		}
	} else {
		log.Print("Environment file not loaded for the current env")
	}
}