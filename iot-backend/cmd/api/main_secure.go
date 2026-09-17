// cmd/server/main_secure.go
// VERSION CORRIGÉE v2.8.0 — MIGRATION SCHNORR ZKP
//
// CHANGEMENTS v2.7.5 → v2.8.0 :
//
//   [SCHNORR-10] SUPPRESSION de toute la logique liée au "ZKP master secret"
//                (Groth16) qui n'existe plus depuis la migration Schnorr :
//                  - Chargement de masterSecretHex (Vault + fallback env) retiré
//                  - Validation de longueur du master secret retirée
//                  - zkp.InitMasterSecret(...) retiré
//                  - zkp.GetSecureZKPManager() (setup Groth16) retiré
//                Le nouveau système Schnorr ne nécessite AUCUNE initialisation
//                globale : chaque device gère sa propre clé, le serveur ne
//                fait qu'enregistrer/lire des clés publiques à la demande
//                (voir internal/handlers/schnorr_store.go, initialisé dans
//                handlers.InitSecureStores — INCHANGÉ, appelé plus bas).
//
//   [SCHNORR-11] checkZKP() (health check) simplifié — Schnorr n'a pas de
//                "service" à initialiser au démarrage, donc rien à vérifier
//                qui puisse être "down". On reporte simplement le mode actif.
//
//   [SCHNORR-12] Routes retirées : POST /api/auth/zkp (AuthenticateWithZKP)
//                et POST /api/zkp/generate-secure (HandleSecureZKPGenerate)
//                — les deux dépendaient du système Groth16 supprimé. Le
//                device génère maintenant sa preuve Schnorr localement et
//                l'envoie directement dans POST /api/node/message-secure
//                (route conservée, INCHANGÉE).
//
//   [SCHNORR-13] Import "iot-backend/pkg/zkp" retiré de ce fichier — plus
//                aucune fonction de ce package n'est appelée directement ici
//                après les suppressions ci-dessus (VerifyECDSASignature etc.
//                restent utilisées, mais uniquement dans les handlers).
//
//   Toutes les autres corrections (CORS dev, TLS split HTTP/HTTPS, rate
//   limiting, circuit breakers, graceful shutdown) sont CONSERVÉES à
//   l'identique — AUCUN changement en dehors de ce qui touche au ZKP.

package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"

	"iot-backend/internal/blockchain"
	"iot-backend/internal/config"
	"iot-backend/internal/handlers"
	"iot-backend/internal/utils"
)

var trustedProxies map[string]bool
var aiBackendSecret string

type rateLimitConfig struct {
	rps                float64
	burst              int
	violationThreshold int
	blockDuration      time.Duration
}

var globalRateLimitCfg rateLimitConfig

func loadRateLimitConfig() rateLimitConfig {
	benchmarkMode := os.Getenv("BENCHMARK_MODE") == "true"
	cfg := rateLimitConfig{
		rps:                10,
		burst:              20,
		violationThreshold: 10,
		blockDuration:      5 * time.Minute,
	}
	if benchmarkMode {
		cfg = rateLimitConfig{
			rps:                1000,
			burst:              2000,
			violationThreshold: 10000,
			blockDuration:      0,
		}
	}
	if v := os.Getenv("RATE_LIMIT_RPS"); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil && n > 0 {
			cfg.rps = n
		}
	}
	if v := os.Getenv("RATE_LIMIT_BURST"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.burst = n
		}
	}
	if v := os.Getenv("RATE_LIMIT_VIOLATIONS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.violationThreshold = n
		}
	}
	if v := os.Getenv("RATE_LIMIT_BLOCK_MIN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			cfg.blockDuration = time.Duration(n) * time.Minute
		}
	}
	return cfg
}

var (
	authAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "chaiman_auth_attempts_total", Help: "Total authentication attempts"},
		[]string{"status", "device_type"},
	)
	authLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "chaiman_auth_duration_seconds", Help: "Authentication latency", Buckets: prometheus.DefBuckets},
		[]string{"status"},
	)
	zkpProofGenTime = prometheus.NewHistogram(
		prometheus.HistogramOpts{Name: "chaiman_zkp_proof_generation_seconds", Help: "ZKP proof generation time", Buckets: []float64{0.1, 0.5, 1, 2, 5, 10}},
	)
	blockchainTxLatency = prometheus.NewHistogram(
		prometheus.HistogramOpts{Name: "chaiman_blockchain_tx_latency_seconds", Help: "Blockchain transaction latency", Buckets: []float64{1, 5, 10, 30, 60, 120}},
	)
	mlServiceHealth = prometheus.NewGauge(
		prometheus.GaugeOpts{Name: "chaiman_ml_service_health", Help: "ML service health (1=up, 0=down)"},
	)
	securityEvents = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "chaiman_security_events_total", Help: "Security events"},
		[]string{"event_type", "severity"},
	)
	circuitBreakerState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "chaiman_circuit_breaker_state", Help: "Circuit breaker state"},
		[]string{"service"},
	)
)

func init() {
	trustedProxies = make(map[string]bool)
	proxiesEnv := os.Getenv("TRUSTED_PROXIES")
	if proxiesEnv != "" {
		for _, ip := range strings.Split(proxiesEnv, ",") {
			ip = strings.TrimSpace(ip)
			if ip != "" {
				trustedProxies[ip] = true
			}
		}
	}
	prometheus.MustRegister(authAttempts, authLatency, zkpProofGenTime, blockchainTxLatency, mlServiceHealth, securityEvents, circuitBreakerState)
}

// ─────────────────────────────────────────────────────────────
// Security headers - PAS de CSP en dev (bloque React inline)
// ─────────────────────────────────────────────────────────────

func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		if os.Getenv("ENV") == "production" {
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'")
		}
		if os.Getenv("ENV") == "production" || r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		next.ServeHTTP(w, r)
	})
}

func contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions || r.Method == http.MethodGet || r.Method == http.MethodDelete || r.Method == http.MethodHead {
			next.ServeHTTP(w, r)
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			ct := r.Header.Get("Content-Type")
			if ct != "" && !strings.Contains(ct, "application/json") &&
				!strings.Contains(ct, "multipart/form-data") &&
				!strings.Contains(ct, "application/x-www-form-urlencoded") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnsupportedMediaType)
				w.Write([]byte(`{"error":"Content-Type must be application/json"}`)) //nolint:errcheck
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func maxBodySizeMiddleware(maxSize int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)
			next.ServeHTTP(w, r)
		})
	}
}

// ─────────────────────────────────────────────────────────────
// Rate-limiter
// ─────────────────────────────────────────────────────────────

type ipRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*ipEntry
}

type ipEntry struct {
	limiter      *rate.Limiter
	lastSeen     time.Time
	violations   int
	blocked      bool
	blockedUntil time.Time
}

var globalRateLimiter = &ipRateLimiter{limiters: make(map[string]*ipEntry)}

func init() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			globalRateLimiter.mu.Lock()
			cutoff := time.Now().Add(-10 * time.Minute)
			for ip, entry := range globalRateLimiter.limiters {
				if entry.lastSeen.Before(cutoff) && !entry.blocked {
					delete(globalRateLimiter.limiters, ip)
				}
			}
			globalRateLimiter.mu.Unlock()
		}
	}()
}

func (rl *ipRateLimiter) getLimiter(ip string) (*rate.Limiter, *rate.Reservation) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	entry, ok := rl.limiters[ip]
	if !ok {
		entry = &ipEntry{limiter: rate.NewLimiter(rate.Limit(globalRateLimitCfg.rps), globalRateLimitCfg.burst)}
		rl.limiters[ip] = entry
	}
	entry.lastSeen = time.Now()
	r := entry.limiter.Reserve()
	return entry.limiter, r
}

func (rl *ipRateLimiter) isBlocked(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	entry, exists := rl.limiters[ip]
	if !exists {
		return false
	}
	if entry.blocked && time.Now().Before(entry.blockedUntil) {
		return true
	}
	if entry.blocked && time.Now().After(entry.blockedUntil) {
		entry.blocked = false
		entry.violations = 0
	}
	return false
}

func (rl *ipRateLimiter) recordViolation(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	entry, exists := rl.limiters[ip]
	if !exists {
		return
	}
	entry.violations++
	if entry.violations > globalRateLimitCfg.violationThreshold {
		if globalRateLimitCfg.blockDuration > 0 {
			entry.blocked = true
			entry.blockedUntil = time.Now().Add(globalRateLimitCfg.blockDuration)
		}
		securityEvents.WithLabelValues("ip_blocked", "high").Inc()
	}
}

func globalRateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := realIP(r)
		if globalRateLimiter.isBlocked(ip) {
			w.Header().Set("Content-Type", "application/json")
			remaining := globalRateLimiter.blockedRemaining(ip)
			retryAfter := int(remaining.Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"IP blocked due to excessive requests"}`)) //nolint:errcheck
			securityEvents.WithLabelValues("rate_limit_blocked", "high").Inc()
			return
		}
		_, reservation := globalRateLimiter.getLimiter(ip)
		delay := reservation.Delay()
		if delay > 0 {
			reservation.Cancel()
			globalRateLimiter.recordViolation(ip)
			retryAfterSec := int(delay.Seconds()) + 1
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", strconv.Itoa(retryAfterSec))
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"Too many requests"}`)) //nolint:errcheck
			securityEvents.WithLabelValues("rate_limit_exceeded", "medium").Inc()
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *ipRateLimiter) blockedRemaining(ip string) time.Duration {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	entry, exists := rl.limiters[ip]
	if !exists || !entry.blocked {
		return 0
	}
	return time.Until(entry.blockedUntil)
}

func realIP(r *http.Request) string {
	remoteIP, _, _ := strings.Cut(r.RemoteAddr, ":")
	if trustedProxies[remoteIP] {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			if idx := strings.Index(xff, ","); idx != -1 {
				return strings.TrimSpace(xff[:idx])
			}
			return strings.TrimSpace(xff)
		}
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			return xri
		}
	}
	return remoteIP
}

// ─────────────────────────────────────────────────────────────
// Admin auth
// ─────────────────────────────────────────────────────────────

func requireAdminAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vault, err := config.GetVaultManager()
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"success":false,"message":"Vault unavailable"}`)) //nolint:errcheck
			return
		}
		adminKey, err := vault.GetAdminAPIKey()
		if err != nil || adminKey == "" {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"success":false,"message":"Configuration admin manquante"}`)) //nolint:errcheck
			return
		}
		providedKey := r.Header.Get("X-Admin-Key")
		if subtle.ConstantTimeCompare([]byte(providedKey), []byte(adminKey)) != 1 {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"success":false,"message":"Authentification admin requise"}`)) //nolint:errcheck
			securityEvents.WithLabelValues("admin_auth_failure", "high").Inc()
			return
		}
		next(w, r)
	}
}

// ─────────────────────────────────────────────────────────────
// AI Signature
// ─────────────────────────────────────────────────────────────

func requireAISignature(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		aiSecret := aiBackendSecret
		if aiSecret == "" {
			vault, err := config.GetVaultManager()
			if err == nil {
				aiSecret, _ = vault.GetAIBackendSecret()
			}
		}
		if aiSecret == "" {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"success":false,"message":"AI_BACKEND_SECRET non configuré"}`)) //nolint:errcheck
			return
		}
		providedSig := r.Header.Get("X-AI-Signature")
		if providedSig == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"success":false,"message":"Signature AI manquante"}`)) //nolint:errcheck
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 1*1024*1024))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"success":false,"message":"Erreur lecture body"}`)) //nolint:errcheck
			return
		}
		r.Body = io.NopCloser(strings.NewReader(string(body)))
		provided, err := hex.DecodeString(providedSig)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"success":false,"message":"Signature AI invalide"}`)) //nolint:errcheck
			return
		}
		mac2 := hmac.New(sha256.New, []byte(aiSecret))
		mac2.Write(body)
		if subtle.ConstantTimeCompare(provided, mac2.Sum(nil)) != 1 {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"success":false,"message":"Signature AI invalide"}`)) //nolint:errcheck
			securityEvents.WithLabelValues("ai_signature_invalid", "high").Inc()
			return
		}
		next(w, r)
	}
}

// ─────────────────────────────────────────────────────────────
// Health check
// ─────────────────────────────────────────────────────────────

func advancedHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	checks := map[string]interface{}{
		"blockchain": checkBlockchain(),
		"zkp":        checkZKP(),
		"vault":      checkVault(),
		"ml":         checkML(),
	}
	status := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "2.8.0-secure-schnorr",
		"checks":    checks,
	}
	allHealthy := true
	isProd := os.Getenv("ENV") == "production"
	for key, check := range checks {
		if c, ok := check.(map[string]interface{}); ok {
			s, _ := c["status"].(string)
			if s != "up" && s != "env_vars" {
				if isProd {
					allHealthy = false
				} else if key == "blockchain" || key == "zkp" {
					allHealthy = false
				}
			}
		}
	}
	if !allHealthy {
		if isProd {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		status["status"] = "degraded"
	}
	json.NewEncoder(w).Encode(status) //nolint:errcheck
}

func checkBlockchain() map[string]interface{} {
	if blockchain.Client == nil {
		return map[string]interface{}{"status": "down", "error": "client not initialized"}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := blockchain.Client.NetworkID(ctx)
	if err != nil {
		return map[string]interface{}{"status": "down", "error": err.Error()}
	}
	return map[string]interface{}{"status": "up"}
}

// checkZKP — [SCHNORR-11] simplifié. Le système Schnorr n'a pas de "manager"
// global à initialiser (contrairement à l'ancien Groth16/SecureZKPManager) :
// chaque vérification est une opération pure et sans état partagé au-delà du
// store de clés publiques (déjà couvert par son propre fonctionnement normal
// — s'il était down, RegisterDevice/HandleSecureNodeMessage échoueraient
// directement et de façon visible, pas besoin d'un check séparé ici).
func checkZKP() map[string]interface{} {
	return map[string]interface{}{"status": "up", "mode": "schnorr"}
}

func checkVault() map[string]interface{} {
	if os.Getenv("VAULT_DISABLE") == "true" || os.Getenv("ENV") != "production" {
		return map[string]interface{}{"status": "up", "mode": "env_vars"}
	}
	vault, err := config.GetVaultManager()
	if err != nil || vault == nil {
		return map[string]interface{}{"status": "down", "error": fmt.Sprintf("%v", err)}
	}
	// [SCHNORR-10] vault.GetZKPMasterSecret() n'est plus appelé pour valider
	// Vault (le master secret ZKP n'existe plus) — on vérifie juste que le
	// client Vault répond en lisant la clé privée blockchain, déjà nécessaire.
	_, err = vault.GetPrivateKey()
	if err != nil {
		return map[string]interface{}{"status": "down", "error": err.Error()}
	}
	return map[string]interface{}{"status": "up"}
}

func checkML() map[string]interface{} {
	mlURL := os.Getenv("ML_SERVICE_URL")
	if mlURL == "" {
		mlURL = "http://localhost:5001"
	}
	if os.Getenv("ENV") == "production" && strings.HasPrefix(mlURL, "http://") {
		mlServiceHealth.Set(0)
		return map[string]interface{}{"status": "config_error", "error": fmt.Sprintf("ML_SERVICE_URL doit utiliser https:// en production (%s)", mlURL)}
	}
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(mlURL + "/health")
	if err != nil {
		mlServiceHealth.Set(0)
		if os.Getenv("ENV") != "production" {
			return map[string]interface{}{"status": "up", "note": "ML service optional in dev mode"}
		}
		return map[string]interface{}{"status": "down", "error": err.Error()}
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, io.LimitReader(resp.Body, 4096)) //nolint:errcheck
	mlServiceHealth.Set(1)
	return map[string]interface{}{"status": "up"}
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("PANIC recovered: %v", rec)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"Internal server error"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ─────────────────────────────────────────────────────────────
// Main
// ─────────────────────────────────────────────────────────────

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found, using environment variables")
	}

	utils.InitLogger()
	logger := utils.GetLogger()
	cfg := utils.LoadConfig()
	globalRateLimitCfg = loadRateLimitConfig()

	logger.Info("Starting secure IoT server v2.8.0 (Schnorr ZKP)")

	if os.Getenv("ENV") == "production" && os.Getenv("TLS_INSECURE_DEV") == "true" {
		logger.Error("TLS_INSECURE_DEV=true en production - ARRET IMMEDIAT")
		securityEvents.WithLabelValues("tls_insecure_prod", "critical").Inc()
		os.Exit(1)
	}

	if os.Getenv("DISABLE_GETH_SCANNER") == "false" {
		logger.Warn("DISABLE_GETH_SCANNER=true — Scanner Geth DÉSACTIVÉ (mode benchmark)")
		logger.Warn("Les nœuds Geth existants ne seront PAS enregistrés automatiquement.")
		logger.Warn("Relancer sans DISABLE_GETH_SCANNER pour réactiver en production.")
	} else {
		logger.Info("Scanner Geth ACTIVÉ (mode production). Utiliser DISABLE_GETH_SCANNER=true pour les benchmarks.")
	}

	// [SCHNORR-10] Chargement des secrets — masterSecretHex RETIRÉ, le
	// système Schnorr n'a besoin d'aucun secret maître global. Seuls
	// privateKeyHex (clé blockchain du serveur) et adminKey/aiBackendSecret
	// (auth des routes admin/AI, INCHANGÉS) restent nécessaires.
	var vault *config.VaultManager
	var privateKeyHex, adminKey string
	var vaultErr error

	logger.Info("Connecting to Vault...")
	vault, vaultErr = config.GetVaultManager()
	if vaultErr != nil {
		if os.Getenv("ENV") == "production" {
			logger.Error("Vault connection failed", "error", vaultErr)
			os.Exit(1)
		}
		logger.Warn("Vault unavailable - using environment variables for secrets (dev mode)", "error", vaultErr)
		privateKeyHex = os.Getenv("BLOCKCHAIN_PRIVATE_KEY")
		adminKey = os.Getenv("ADMIN_API_KEY")
		aiBackendSecret = os.Getenv("AI_BACKEND_SECRET")
		if privateKeyHex == "" {
			logger.Error("BLOCKCHAIN_PRIVATE_KEY required (env var or Vault)")
			os.Exit(1)
		}
	} else {
		logger.Info("Vault connected securely")
		var err error
		privateKeyHex, err = vault.GetPrivateKey()
		if err != nil {
			logger.Error("Blockchain private key not retrievable", "error", err)
			os.Exit(1)
		}
		adminKey, _ = vault.GetAdminAPIKey()
		aiBackendSecret, _ = vault.GetAIBackendSecret()
	}

	if adminKey == "" {
		logger.Warn("ADMIN_API_KEY not configured - admin routes disabled")
	}
	if aiBackendSecret == "" {
		logger.Warn("AI_BACKEND_SECRET not configured - AI revocation routes disabled")
	} else {
		logger.Info("AI_BACKEND_SECRET configured", "length", len(aiBackendSecret))
	}

	// [SCHNORR-10] zkp.InitMasterSecret(...) SUPPRIMÉ — plus de secret maître
	// à initialiser. Chaque device Schnorr est autonome (voir pkg/zkp/schnorr_device.go).

	for _, v := range []struct{ name, val string }{
		{"CONTRACT_ADDRESS", cfg.ContractAddress},
		{"NODE_URL", cfg.NodeURL},
	} {
		if v.val == "" {
			logger.Error(v.name + " not defined in .env")
			os.Exit(1)
		}
	}

	logger.Info("Connecting to blockchain...", "node_url", cfg.NodeURL)
	if err := blockchain.InitBlockchain(cfg.NodeURL, privateKeyHex, cfg.ContractAddress); err != nil {
		logger.Error("Blockchain error", "error", err)
		os.Exit(1)
	}
	logger.Info("Blockchain initialized")

	if err := blockchain.InitNoncePool(); err != nil {
		logger.Error("Error initializing nonce pool", "error", err)
		os.Exit(1)
	}
	logger.Info("Nonce pool initialized")

	if os.Getenv("DISABLE_GETH_SCANNER") != "true" {
		blockchain.StartGethNodeScanner()
		logger.Info("Geth node scanner started")
	} else {
		logger.Warn("Geth node scanner DISABLED (DISABLE_GETH_SCANNER=true)")
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		logger.Error("Error creating data directory", "error", err)
		os.Exit(1)
	}
	// [SCHNORR-1] InitSecureStores initialise désormais AUSSI le
	// SchnorrKeyStore en interne (voir internal/handlers/messages_secure.go,
	// InitSecureStores — appelle handlers.InitSchnorrKeyStore(dataDir)).
	if err := handlers.InitSecureStores(dataDir); err != nil {
		logger.Error("Error initializing stores", "error", err)
		os.Exit(1)
	}
	logger.Info("Stores initialized (nonces, sequences, Schnorr keys)", "data_dir", dataDir)

	// [SCHNORR-10] Bloc d'initialisation zkpService (Groth16) SUPPRIMÉ.
	logger.Info("ZKP: Schnorr mode active — no global service to initialize")

	go func() {
		time.Sleep(2 * time.Second)
		logger.Info("Warmup: chargement historique transactions depuis la blockchain...")
		records, warmupErr := blockchain.GetRecentAuthentications(10000)
		if warmupErr != nil {
			logger.Warn("Warmup: erreur chargement historique", "error", warmupErr)
			return
		}
		logger.Info("Warmup: historique chargé", "transactions", len(records))
	}()

	// Routes
	router := mux.NewRouter()

	router.HandleFunc("/api/health", advancedHealthCheck).Methods("GET", "OPTIONS")
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	router.HandleFunc("/api/device/register", handlers.RegisterDevice).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/devices", handlers.GetAllDevices).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/device/{address}", handlers.GetDeviceInfo).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/device/{address}/status", handlers.GetDeviceStatus).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/device/deactivate", requireAISignature(handlers.DeactivateDevice)).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/device/activate", handlers.ActivateDevice).Methods("POST", "OPTIONS")

	// [SCHNORR-12] Route retirée : POST /api/auth/zkp (handlers.AuthenticateWithZKP)
	// — dépendait du Groth16 manager supprimé.

	router.HandleFunc("/api/node/message-secure", handlers.HandleSecureNodeMessage).Methods("POST", "OPTIONS")
	// [SCHNORR-12] Route retirée : POST /api/zkp/generate-secure
	// (handlers.HandleSecureZKPGenerate) — le device génère maintenant sa
	// preuve Schnorr localement (pkg/zkp/schnorr_device.go) et l'envoie
	// directement dans /api/node/message-secure ci-dessus.

	router.HandleFunc("/api/geth-nodes", handlers.GetGethNodes).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/geth-nodes/add", requireAdminAuth(handlers.AddGethNodes)).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/geth-nodes/status", requireAdminAuth(handlers.GetAddGethNodesStatus)).Methods("GET", "OPTIONS")

	router.HandleFunc("/api/stats/global", handlers.HandleGlobalStats).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/transactions/recent", handlers.HandleRecentTransactions).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/device/{address}/history", handlers.HandleDeviceHistory).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/blockchain/info", handlers.HandleBlockchainInfo).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/transaction/{hash}", handlers.HandleTransactionByHash).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/blockchain/metrics", handlers.HandleBlockchainMetrics).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/ai/metrics", handlers.HandleAIMetrics).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/ai/anomalies", handlers.HandleAIAnomalies).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/ai/stats", handlers.HandleAIStats).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/ai/actions", handlers.HandleAIActions).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/ai/devices", handlers.HandleAIDevices).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/device/registration-history", handlers.HandleRegistrationHistory).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/device/registration-stats", handlers.HandleRegistrationStats).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/device/{address}/registration-history", handlers.GetRegistrationHistoryByDevice).Methods("GET", "OPTIONS")

	if os.Getenv("DEBUG_MODE") == "true" {
		logger.Warn("DEBUG MODE ENABLED")
		router.HandleFunc("/api/debug/reset-sequences", func(w http.ResponseWriter, r *http.Request) {
			handlers.ResetSequences()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok","message":"Sequences reset"}`)) //nolint:errcheck
		}).Methods("POST", "OPTIONS")
		router.HandleFunc("/api/debug/reset-nonces", func(w http.ResponseWriter, r *http.Request) {
			handlers.ResetNonces()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok","message":"Nonces reset"}`)) //nolint:errcheck
		}).Methods("POST", "OPTIONS")
	}

	if os.Getenv("BENCHMARK_MODE") == "true" {
		router.HandleFunc("/api/debug/reset-state", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			handlers.ResetSequences()
			handlers.ResetNonces()
			globalRateLimiter.mu.Lock()
			globalRateLimiter.limiters = make(map[string]*ipEntry)
			globalRateLimiter.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"reset":"ok","sequences":true,"nonces":true}`))
		}).Methods("POST", "OPTIONS")
	}

	// ─────────────────────────────────────────────────────────────
	// [FIX-CORS-DEV] CORS complètement ouvert en développement
	// ─────────────────────────────────────────────────────────────
	var handler http.Handler = router
	handler = securityHeadersMiddleware(handler)
	handler = contentTypeMiddleware(handler)
	handler = maxBodySizeMiddleware(10 * 1024 * 1024)(handler)
	handler = globalRateLimitMiddleware(handler)
	handler = recoveryMiddleware(handler)
	handler = http.TimeoutHandler(handler, 60*time.Second, `{"error":"Request timeout"}`)

	corsHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		origin := req.Header.Get("Origin")
		isDev := os.Getenv("ENV") != "production"

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, Authorization, X-Admin-Key, X-AI-Signature, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.Header().Set("Vary", "Origin")

		if isDev {
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
		} else {
			originsList := make(map[string]bool)
			for _, o := range strings.Split(cfg.AllowedOrigins, ",") {
				if t := strings.TrimSpace(o); t != "" {
					originsList[t] = true
				}
			}
			if origin != "" && !originsList[origin] {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"error":"Origin not authorized"}`)) //nolint:errcheck
				return
			}
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}

		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		handler.ServeHTTP(w, req)
	})

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = port
	}
	httpsPort := os.Getenv("HTTPS_PORT")
	if httpsPort == "" {
		httpsPort = "8443"
	}

	tlsCert := os.Getenv("TLS_CERT_PATH")
	tlsKey := os.Getenv("TLS_KEY_PATH")
	useTLS := tlsCert != "" && tlsKey != ""

	if os.Getenv("ENV") == "production" && !useTLS {
		logger.Error("TLS required in production for HTTPS server. Set TLS_CERT_PATH and TLS_KEY_PATH")
		os.Exit(1)
	}

	// ─────────────────────────────────────────────────────────────
	// TLS CONFIG - UNIQUEMENT pour le port HTTPS (simulator + AI)
	// ─────────────────────────────────────────────────────────────
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
		PreferServerCipherSuites: true,
	}

	if clientCA := os.Getenv("TLS_CLIENT_CA_PATH"); clientCA != "" {
		caCert, err := os.ReadFile(clientCA)
		if err != nil {
			logger.Error("Failed to load client CA", "error", err)
			os.Exit(1)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.ClientCAs = caCertPool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		logger.Info("mTLS enabled for client authentication")
	}

	httpServer := &http.Server{
		Addr:         ":" + httpPort,
		Handler:      corsHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	httpsServer := &http.Server{
		Addr:         ":" + httpsPort,
		Handler:      corsHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
		TLSConfig:    tlsConfig,
	}

	logger.Info("========================================")
	logger.Info("Secure server starting (ECDSA + Schnorr-ZKP + HMAC)")
	logger.Info("Version: 2.8.0-secure-schnorr")
	logger.Info(fmt.Sprintf("HTTP  port (dashboard/React)   : %s  [NO TLS]", httpPort))
	if useTLS {
		logger.Info(fmt.Sprintf("HTTPS port (simulator + AI)    : %s  [TLS 1.3]", httpsPort))
	} else {
		logger.Warn(fmt.Sprintf("HTTPS port (simulator + AI)    : %s  [TLS DISABLED]", httpsPort))
	}
	logger.Info(fmt.Sprintf("Blockchain: %s", cfg.NodeURL))
	logger.Info("Vault: Connected")
	logger.Info("Prometheus: /metrics")
	if os.Getenv("DISABLE_GETH_SCANNER") == "true" {
		logger.Warn("Geth Scanner: DISABLED (benchmark mode)")
	} else {
		logger.Info("Geth Scanner: ENABLED (production mode)")
	}
	if os.Getenv("BENCHMARK_MODE") == "true" {
		logger.Warn("Rate limiter: BENCHMARK MODE",
			"rps", globalRateLimitCfg.rps,
			"burst", globalRateLimitCfg.burst,
			"violation_threshold", globalRateLimitCfg.violationThreshold,
			"block_duration", globalRateLimitCfg.blockDuration,
		)
	} else {
		logger.Info("Rate limiter: PRODUCTION MODE",
			"rps", globalRateLimitCfg.rps,
			"burst", globalRateLimitCfg.burst,
			"violation_threshold", globalRateLimitCfg.violationThreshold,
			"block_duration", globalRateLimitCfg.blockDuration,
		)
	}
	logger.Info("========================================")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("HTTP server listening (dashboard/React - NO TLS)", "addr", ":"+httpPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	go func() {
		if !useTLS {
			logger.Warn("HTTPS server not started: TLS_CERT_PATH or TLS_KEY_PATH missing")
			return
		}
		logger.Info("HTTPS server listening (simulator + AI - TLS 1.3)", "addr", ":"+httpsPort)
		if err := httpsServer.ListenAndServeTLS(tlsCert, tlsKey); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTPS server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	logger.Info("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	blockchain.StopBlockchain()
	blockchain.StopNoncePool()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("HTTP server graceful shutdown error", "error", err)
	}
	if useTLS {
		if err := httpsServer.Shutdown(ctx); err != nil {
			logger.Error("HTTPS server graceful shutdown error", "error", err)
		}
	}
	logger.Info("Servers stopped cleanly")
}
