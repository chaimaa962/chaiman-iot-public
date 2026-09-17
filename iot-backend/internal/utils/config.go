package utils

import (
	"os"
	"strings"
)

// Config contient la configuration publique du serveur.
//
// SÉCURITÉ :
//   - Les secrets (clé privée blockchain, master ZKP, admin key, AI secret)
//     sont gérés exclusivement par le VaultManager (HashiCorp Vault ou équivalent).
//   - Ce struct ne contient AUCUN secret — uniquement des paramètres de configuration.
//   - La clé privée était précédemment dans ce struct (PrivateKey string) : SUPPRIMÉ.
type Config struct {
	// NodeURL est l'URL du nœud Ethereum (ex: http://localhost:8545)
	NodeURL string

	// ContractAddress est l'adresse du smart contract IoTDeviceManager
	ContractAddress string

	// Port est le port d'écoute du serveur HTTP(S)
	Port string

	// AllowedOrigins est la liste des origines CORS autorisées, séparées par des virgules.
	// Exemple: "https://dashboard.example.com,https://admin.example.com"
	// Une valeur vide désactive les origines CORS (tout bloquer sauf même origine).
	AllowedOrigins string

	// DataDir est le répertoire de stockage des données persistantes (nonces, séquences).
	DataDir string

	// MLServiceURL est l'URL du service ML d'analyse d'anomalies.
	MLServiceURL string
}

// LoadConfig charge la configuration depuis les variables d'environnement.
// Les valeurs par défaut sont conservatrices et sécurisées.
//
// Variables d'environnement:
//   - NODE_URL          : URL du nœud Ethereum         (défaut: http://localhost:8545)
//   - CONTRACT_ADDRESS  : Adresse du smart contract    (défaut: "")
//   - PORT              : Port d'écoute                (défaut: "8080")
//   - ALLOWED_ORIGINS   : Origines CORS autorisées     (défaut: "" — aucune origine externe)
//   - DATA_DIR          : Répertoire de données        (défaut: "./data")
//   - ML_SERVICE_URL    : URL du service ML            (défaut: http://localhost:5001)
func LoadConfig() *Config {
	return &Config{
		NodeURL:         getEnv("NODE_URL", "http://localhost:8545"),
		ContractAddress: getEnv("CONTRACT_ADDRESS", ""),
		Port:            getEnv("PORT", "8080"),
		AllowedOrigins:  getEnv("ALLOWED_ORIGINS", ""),
		DataDir:         getEnv("DATA_DIR", "./data"),
		MLServiceURL:    getEnv("ML_SERVICE_URL", "http://localhost:5001"),
	}
}

// AllowedOriginsMap retourne les origines CORS autorisées sous forme de map
// pour une recherche O(1). Les espaces et entrées vides sont ignorés.
func (c *Config) AllowedOriginsMap() map[string]bool {
	result := make(map[string]bool)
	if c.AllowedOrigins == "" {
		return result
	}
	for _, origin := range strings.Split(c.AllowedOrigins, ",") {
		if t := strings.TrimSpace(origin); t != "" {
			result[t] = true
		}
	}
	return result
}

// Validate vérifie que les champs obligatoires sont définis.
// Retourne une liste de champs manquants.
func (c *Config) Validate() []string {
	var missing []string
	if c.NodeURL == "" {
		missing = append(missing, "NODE_URL")
	}
	if c.ContractAddress == "" {
		missing = append(missing, "CONTRACT_ADDRESS")
	}
	return missing
}

// getEnv retourne la valeur d'une variable d'environnement ou la valeur par défaut.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
