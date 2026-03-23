package utils

import (
    "os"
)

type Config struct {
    NodeURL         string
    ContractAddress string
    PrivateKey      string
    Port            string
}

func LoadConfig() *Config {
    return &Config{
        NodeURL:         getEnv("NODE_URL", "http://localhost:8545"),
        ContractAddress: getEnv("CONTRACT_ADDRESS", ""),
        PrivateKey:      getEnv("PRIVATE_KEY", ""),
        Port:            getEnv("PORT", "8080"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

