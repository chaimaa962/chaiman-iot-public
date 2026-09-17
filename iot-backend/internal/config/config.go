package config

import (
    "os"
    "strconv"
)

type Config struct {
    NodeURL         string
    ContractAddress string
    PrivateKey      string
    Port            string
    TLSCertPath     string
    TLSKeyPath      string
    TLSInsecureDev  bool
    AIBackendSecret string
    ModelPubkeyPath string
    BackendURL      string
    Env             string
    DebugMode       bool
}

func LoadConfig() *Config {
    return &Config{
        NodeURL:         getEnv("NODE_URL", "http://localhost:8545"),
        ContractAddress: getEnv("CONTRACT_ADDRESS", ""),
        PrivateKey:      getEnv("PRIVATE_KEY", ""),
        Port:            getEnv("PORT", "8080"),
        TLSCertPath:     getEnv("TLS_CERT_PATH", ""),
        TLSKeyPath:      getEnv("TLS_KEY_PATH", ""),
        TLSInsecureDev:  getEnv("TLS_INSECURE_DEV", "false") == "true",
        AIBackendSecret: getEnv("AI_BACKEND_SECRET", ""),
        ModelPubkeyPath: getEnv("MODEL_PUBKEY_PATH", ""),
        BackendURL:      getEnv("BACKEND_URL", "https://localhost:8080"),
        Env:             getEnv("ENV", "development"),
        DebugMode:       getEnv("DEBUG_MODE", "false") == "true",
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        return value == "true"
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if i, err := strconv.Atoi(value); err == nil {
            return i
        }
    }
    return defaultValue
}
