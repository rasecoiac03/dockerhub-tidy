package config

import (
	"os"
	"strconv"
)

var config = map[string]string{
	"LOG_LEVEL":          "debug",
	"DOCKERHUB_BASE_URL": "https://hub.docker.com",
	"DOCKER_USERNAME":    "",
	"DOCKER_PASSWORD":    "",
}

func init() {
	// Env vars
	for k := range config {
		v := os.Getenv(k)
		if v != "" {
			config[k] = v
		}
	}
}

func GetEnv(configKey string) string {
	return config[configKey]
}

func GetEnvInt(configKey string) int {
	i := 0
	i, _ = strconv.Atoi(GetEnv(configKey))
	return i
}
