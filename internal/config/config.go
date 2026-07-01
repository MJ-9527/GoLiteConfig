package config

import (
	"os"
	"strings"
)

type Config struct {
	Port          string
	EtcdEndpoints []string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	endpoints := splitEnv(os.Getenv("ETCD_ENDPOINTS"))
	if len(endpoints) == 0 {
		endpoints = []string{"localhost:2379"}
	}

	return Config{
		Port:          port,
		EtcdEndpoints: endpoints,
	}
}

func splitEnv(value string) []string {
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}
