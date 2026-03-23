package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"iot-backend/internal/blockchain"
	"iot-backend/internal/handlers"
	"iot-backend/internal/utils"
)

func main() {
	godotenv.Load()
	utils.InitLogger()
	logger := utils.GetLogger()
	cfg := utils.LoadConfig()

	if err := blockchain.InitBlockchain(cfg.NodeURL, cfg.PrivateKey, cfg.ContractAddress); err != nil {
		logger.Error("Erreur blockchain", "error", err)
		os.Exit(1)
	}

	router := mux.NewRouter()

	// ============================================
	// ROUTES DE SANTÉ
	// ============================================
	router.HandleFunc("/api/health", handlers.HealthCheck).Methods("GET", "OPTIONS")

	// ============================================
	// GESTION DES APPAREILS
	// ============================================
	router.HandleFunc("/api/device/register", handlers.RegisterDevice).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/device/{address}", handlers.GetDeviceInfo).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/devices", handlers.GetAllDevices).Methods("GET", "OPTIONS")

	// ============================================
	// ROUTES SÉCURISÉES (ECDSA + ZKP)
	// ============================================
	router.HandleFunc("/api/node/message-secure", handlers.HandleSecureNodeMessage).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/zkp/generate-secure", handlers.HandleSecureZKPGenerate).Methods("POST", "OPTIONS")


	// ============================================
	// NŒUDS GETH
	// ============================================
	router.HandleFunc("/api/geth-nodes", handlers.GetGethNodes).Methods("GET", "OPTIONS")

	// ============================================
	// ROUTES POUR LE DASHBOARD (TEMPS RÉEL)
	// ============================================
	router.HandleFunc("/api/stats/global", handlers.HandleGlobalStats).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/transactions/recent", handlers.HandleRecentTransactions).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/device/{address}/history", handlers.HandleDeviceHistory).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/blockchain/info", handlers.HandleBlockchainInfo).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/transaction/{hash}", handlers.HandleTransactionByHash).Methods("GET", "OPTIONS")
        router.HandleFunc("/api/blockchain/metrics", handlers.HandleBlockchainMetrics).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/ai/metrics", handlers.HandleAIMetrics).Methods("GET", "OPTIONS")
	// ============================================
	// MIDDLEWARE CORS
	// ============================================
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin, Accept")

		if req.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		router.ServeHTTP(w, req)
	})

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	logger.Info("🚀 Serveur SÉCURISÉ démarré (ECDSA + ZKP)", "port", port)
	logger.Info("📋 Routes disponibles:")
	logger.Info("   POST /api/node/message-secure    - Envoyer message (ECDSA+ZKP)")
	logger.Info("   POST /api/zkp/generate-secure    - Générer preuve ZKP")
	logger.Info("   GET  /api/stats/global           - Statistiques globales")
	logger.Info("   GET  /api/transactions/recent    - Transactions récentes")
	logger.Info("   GET  /api/blockchain/info        - Info blockchain")
	logger.Info("   GET  /api/device/{address}/history - Historique appareil")
	logger.Info("   GET  /api/transaction/{hash}     - Détail transaction")

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		logger.Error("Erreur serveur", "error", err)
		os.Exit(1)
	}
}
