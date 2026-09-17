// internal/handlers/messages_secure.go
// CHAIMAN IoT — Secure Message Handler
// VERSION CORRIGÉE v6.0 — MIGRATION VERS SCHNORR ZKP
//
// CHANGEMENTS v5.4 → v6.0 :
//
//   [SCHNORR-2] Remplacement complet de la vérification ZKP Groth16
//               (zkp.GetSecureZKPManager().VerifyProof) par la vérification
//               Schnorr (zkp.VerifyProof avec SchnorrPublicKey/SchnorrProof).
//
//   [SCHNORR-3] Le flux "challenge pré-enregistré" (ChallengeStore,
//               StoreExpectedChallenge, VerifyAndConsumeChallenge) est
//               SUPPRIMÉ. Avec Schnorr + Fiat-Shamir, le challenge est
//               dérivé DÉTERMINISTIQUEMENT par la fonction pure
//               zkp.HashChallenge(address, T, timestamp, sequence, nonce) —
//               calculée indépendamment par le device ET le serveur à
//               partir des mêmes données publiques. Plus besoin d'un
//               aller-retour réseau ni d'un store partagé pour "convenir"
//               d'un challenge à l'avance. La protection anti-rejeu reste
//               garantie par UpdateSequence() (déjà en place, INCHANGÉ) —
//               rejouer un (T,s) avec une sequence déjà vue est rejeté par
//               UpdateSequence AVANT même d'atteindre la vérification Schnorr.
//
//   [SCHNORR-4] HandleSecureZKPGenerate (ancien endpoint /api/zkp/generate-secure)
//               est SUPPRIMÉ. Il n'a plus de raison d'être : c'est
//               maintenant le DEVICE qui génère sa propre preuve localement
//               (pkg/zkp/schnorr_device.go, GenerateProof) — le serveur n'a
//               plus jamais besoin de calculer une preuve pour le compte
//               d'un device. Ceci élimine aussi un aller-retour réseau complet
//               par authentification (voir iot_simulator.go v6.0).
//
//   Toutes les autres corrections (rate-limiting adaptatif, circuit breakers,
//   nonce anti-replay, ordre de validation) sont CONSERVÉES à l'identique.

package handlers

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/time/rate"

	"iot-backend/internal/blockchain"
	"iot-backend/internal/models"
	"iot-backend/internal/utils"
	"iot-backend/pkg/zkp"
)

// ============================================
// MÉTRIQUES PROMETHEUS — TEMPS RÉELS
// ============================================

var (
	ecdsaVerifyDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "chaiman_ecdsa_verify_duration_ms",
		Help:    "Temps réel de vérification ECDSA en millisecondes",
		Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 20, 50, 100},
	})

	// [SCHNORR-2] Buckets réduits — Schnorr vérifie en ~0.1ms (mesuré),
	// contre 10-2000ms pour l'ancien Groth16. Les anciens buckets (jusqu'à
	// 5000ms) n'avaient plus de sens et auraient rendu l'histogramme illisible.
	zkpVerifyDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "chaiman_zkp_verify_duration_ms",
		Help:    "Temps réel de vérification ZKP Schnorr en millisecondes",
		Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 5, 10, 50},
	})

	blockchainRecordDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "chaiman_blockchain_record_duration_ms",
		Help:    "Temps d'enregistrement blockchain en millisecondes",
		Buckets: []float64{100, 500, 1000, 3000, 5000, 10000, 30000},
	})

	authTotalDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "chaiman_auth_total_duration_ms",
		Help:    "Temps total d'authentification ECDSA+ZKP en millisecondes",
		Buckets: []float64{10, 50, 100, 500, 1000, 2000, 5000, 10000},
	})
)

// ============================================
// CONSTANTES
// ============================================

const (
	MaxSequenceAge      = 300
	MaxFutureTimestamp  = 30
	MaxRetryAttempts    = 10
	RetryDelay          = 3 * time.Second
	TxConfirmationWait  = 30
	NonceExpiryHours    = 24
	RateLimitPerSecondProd  = 50
	RateLimitBurstProd      = 100
	RateLimitPerSecondBench = 1000
	RateLimitBurstBench     = 2000
	MaxBodySize             = 10 * 1024 * 1024
	SequenceRetryWindow     = 5 * time.Second
	MinNonceLength          = 8

	CircuitBreakerThreshold   = 5
	CircuitBreakerTimeout     = 30 * time.Second
	CircuitBreakerHalfOpenMax = 3
)

func getRateLimits() (perSecond, burst int) {
	if os.Getenv("BENCHMARK_MODE") == "true" {
		return RateLimitPerSecondBench, RateLimitBurstBench
	}
	return RateLimitPerSecondProd, RateLimitBurstProd
}

// ============================================
// ALIAS TIMINGS
// ============================================

type AuthTimings = models.AuthTimings

// ============================================
// CIRCUIT BREAKER — [INCHANGÉ]
// ============================================

type CircuitState string

const (
	CircuitClosed   CircuitState = "closed"
	CircuitOpen     CircuitState = "open"
	CircuitHalfOpen CircuitState = "half-open"
)

type CircuitBreaker struct {
	mu          sync.RWMutex
	failures    int
	successes   int
	lastFailure time.Time
	state       CircuitState
	threshold   int
	timeout     time.Duration
	halfOpenMax int
}

func NewCircuitBreaker(threshold int, timeout time.Duration, halfOpenMax int) *CircuitBreaker {
	return &CircuitBreaker{
		state:       CircuitClosed,
		threshold:   threshold,
		timeout:     timeout,
		halfOpenMax: halfOpenMax,
	}
}

func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()

	if cb.state == CircuitOpen && time.Since(cb.lastFailure) > cb.timeout {
		cb.state = CircuitHalfOpen
		cb.failures = 0
		cb.successes = 0
	}

	if cb.state == CircuitOpen {
		cb.mu.Unlock()
		return fmt.Errorf("circuit breaker open: service indisponible")
	}

	if cb.state == CircuitHalfOpen && cb.successes >= cb.halfOpenMax {
		cb.state = CircuitClosed
		cb.failures = 0
		cb.successes = 0
	}

	cb.mu.Unlock()

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()
		if cb.failures >= cb.threshold {
			cb.state = CircuitOpen
		}
		return err
	}

	if cb.state == CircuitHalfOpen {
		cb.successes++
	} else {
		cb.failures = 0
	}

	return nil
}

func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// ============================================
// STRUCTURES PERSISTANTES
// ============================================

type NonceEntry struct {
	Timestamp time.Time `json:"timestamp"`
}

type NonceStore struct {
	mu       sync.RWMutex
	nonces   map[string]NonceEntry
	filePath string
}

type SequenceEntry struct {
	Sequence  uint64    `json:"sequence"`
	Timestamp time.Time `json:"timestamp"`
}

type SequenceStore struct {
	mu        sync.RWMutex
	sequences map[string]SequenceEntry
	filePath  string
}

type RateLimiterStore struct {
	mu       sync.RWMutex
	limiters map[string]*rate.Limiter
}

// [SCHNORR-3] ChallengeStore SUPPRIMÉ — plus nécessaire, le challenge est
// dérivé déterministiquement (Fiat-Shamir), pas de pré-enregistrement requis.

var (
	nonceStore       *NonceStore
	sequenceStore    *SequenceStore
	rateLimiterStore *RateLimiterStore

	memoryNonces      = make(map[string]bool)
	memoryNoncesMu    sync.RWMutex
	memorySequences   = make(map[string]SequenceEntry)
	memorySequencesMu sync.RWMutex

	mlCircuitBreaker         = NewCircuitBreaker(CircuitBreakerThreshold, CircuitBreakerTimeout, CircuitBreakerHalfOpenMax)
	blockchainCircuitBreaker = NewCircuitBreaker(3, 60*time.Second, 2)

	prevMLState CircuitState = CircuitClosed
	prevBCState CircuitState = CircuitClosed
	cbStateMu   sync.Mutex
)

// ============================================
// INITIALISATION
// ============================================

// InitSecureStores — [SCHNORR-1] initialise désormais aussi le SchnorrKeyStore
// (schnorr_store.go) en plus des stores existants.
func InitSecureStores(dataDir string) error {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("création répertoire data: %v", err)
	}

	nonceStore = &NonceStore{
		nonces:   make(map[string]NonceEntry),
		filePath: dataDir + "/nonces.json",
	}
	if err := nonceStore.load(); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("chargement nonces.json: %v", err)
	}

	sequenceStore = &SequenceStore{
		sequences: make(map[string]SequenceEntry),
		filePath:  dataDir + "/sequences.json",
	}
	if err := sequenceStore.load(); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("chargement sequences.json: %v", err)
	}

	rateLimiterStore = &RateLimiterStore{
		limiters: make(map[string]*rate.Limiter),
	}

	// [SCHNORR-1] Nouveau store — remplace ChallengeStore.
	if err := InitSchnorrKeyStore(dataDir); err != nil {
		return fmt.Errorf("initialisation schnorr key store: %v", err)
	}

	go nonceStore.cleanupRoutine()
	go sequenceStore.saveRoutine()
	go circuitBreakerMonitor()

	return nil
}

func circuitBreakerMonitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	logger := utils.GetLogger()
	for range ticker.C {
		mlState := mlCircuitBreaker.State()
		bcState := blockchainCircuitBreaker.State()

		cbStateMu.Lock()
		if mlState != prevMLState {
			if mlState != CircuitClosed {
				logger.Warn("Circuit breaker ML changement d'état", "state", string(mlState))
			} else {
				logger.Info("Circuit breaker ML rétabli")
			}
			prevMLState = mlState
		}
		if bcState != prevBCState {
			if bcState != CircuitClosed {
				logger.Warn("Circuit breaker blockchain changement d'état", "state", string(bcState))
			} else {
				logger.Info("Circuit breaker blockchain rétabli")
			}
			prevBCState = bcState
		}
		cbStateMu.Unlock()
	}
}

// ============================================
// NONCE STORE — [INCHANGÉ]
// ============================================

func (ns *NonceStore) load() error {
	data, err := os.ReadFile(ns.filePath)
	if err != nil {
		return err
	}
	ns.mu.Lock()
	defer ns.mu.Unlock()
	return json.Unmarshal(data, &ns.nonces)
}

func (ns *NonceStore) save() error {
	ns.mu.RLock()
	data, err := json.MarshalIndent(ns.nonces, "", "  ")
	ns.mu.RUnlock()
	if err != nil {
		return err
	}
	tmpPath := ns.filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmpPath, ns.filePath)
}

func (ns *NonceStore) cleanupRoutine() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		ns.mu.Lock()
		for key, entry := range ns.nonces {
			if time.Since(entry.Timestamp) > NonceExpiryHours*time.Hour {
				delete(ns.nonces, key)
			}
		}
		ns.mu.Unlock()
		_ = ns.save()
	}
}

func MarkNonceUsed(deviceID, nonce string) bool {
	if nonceStore == nil {
		key := deviceID + ":" + nonce
		memoryNoncesMu.Lock()
		defer memoryNoncesMu.Unlock()
		if memoryNonces[key] {
			return false
		}
		memoryNonces[key] = true
		return true
	}

	key := deviceID + ":" + nonce
	nonceStore.mu.Lock()
	defer nonceStore.mu.Unlock()

	if _, exists := nonceStore.nonces[key]; exists {
		return false
	}
	nonceStore.nonces[key] = NonceEntry{Timestamp: time.Now()}

	if len(nonceStore.nonces)%50 == 0 {
		go nonceStore.save() //nolint:errcheck
	}
	return true
}

// ============================================
// SEQUENCE STORE — [INCHANGÉ]
// ============================================

func (ss *SequenceStore) load() error {
	data, err := os.ReadFile(ss.filePath)
	if err != nil {
		return err
	}
	ss.mu.Lock()
	defer ss.mu.Unlock()
	return json.Unmarshal(data, &ss.sequences)
}

func (ss *SequenceStore) save() error {
	ss.mu.RLock()
	data, err := json.MarshalIndent(ss.sequences, "", "  ")
	ss.mu.RUnlock()
	if err != nil {
		return err
	}
	tmpPath := ss.filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmpPath, ss.filePath)
}

func (ss *SequenceStore) saveRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		_ = ss.save()
	}
}

func UpdateSequence(deviceAddress string, sequence uint64) bool {
	if sequenceStore == nil {
		memorySequencesMu.Lock()
		defer memorySequencesMu.Unlock()

		entry := memorySequences[deviceAddress]
		if entry.Sequence == 0 {
			memorySequences[deviceAddress] = SequenceEntry{Sequence: sequence, Timestamp: time.Now()}
			return true
		}
		if sequence < entry.Sequence {
			return false
		}
		if sequence == entry.Sequence {
			return time.Since(entry.Timestamp) <= SequenceRetryWindow
		}
		memorySequences[deviceAddress] = SequenceEntry{Sequence: sequence, Timestamp: time.Now()}
		return true
	}

	sequenceStore.mu.Lock()
	defer sequenceStore.mu.Unlock()

	entry := sequenceStore.sequences[deviceAddress]

	if entry.Sequence == 0 {
		sequenceStore.sequences[deviceAddress] = SequenceEntry{Sequence: sequence, Timestamp: time.Now()}
		return true
	}
	if sequence < entry.Sequence {
		return false
	}
	if sequence == entry.Sequence {
		return time.Since(entry.Timestamp) <= SequenceRetryWindow
	}
	sequenceStore.sequences[deviceAddress] = SequenceEntry{Sequence: sequence, Timestamp: time.Now()}
	return true
}

func GetLastSequence(deviceAddress string) uint64 {
	if sequenceStore == nil {
		memorySequencesMu.RLock()
		defer memorySequencesMu.RUnlock()
		return memorySequences[deviceAddress].Sequence
	}
	sequenceStore.mu.RLock()
	defer sequenceStore.mu.RUnlock()
	return sequenceStore.sequences[deviceAddress].Sequence
}

// ============================================
// RATE LIMITING — [INCHANGÉ]
// ============================================

type RateLimitEntry struct {
	limiter   *rate.Limiter
	lastSeen  time.Time
	violation int
}

var rateLimitEntries = struct {
	mu      sync.RWMutex
	entries map[string]*RateLimitEntry
}{
	entries: make(map[string]*RateLimitEntry),
}

func init() {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			rateLimitEntries.mu.Lock()
			cutoff := time.Now().Add(-1 * time.Hour)
			for key, entry := range rateLimitEntries.entries {
				if entry.lastSeen.Before(cutoff) {
					delete(rateLimitEntries.entries, key)
				}
			}
			rateLimitEntries.mu.Unlock()
		}
	}()
}

func CheckRateLimit(address string) bool {
	rateLimitEntries.mu.Lock()
	defer rateLimitEntries.mu.Unlock()

	perSecond, burst := getRateLimits()

	entry, exists := rateLimitEntries.entries[address]
	if !exists {
		entry = &RateLimitEntry{
			limiter:  rate.NewLimiter(rate.Limit(perSecond), burst),
			lastSeen: time.Now(),
		}
		rateLimitEntries.entries[address] = entry
	}

	entry.lastSeen = time.Now()

	if !entry.limiter.Allow() {
		entry.violation++
		return false
	}

	entry.violation = 0
	return true
}

func IsRateLimited(address string) bool {
	rateLimitEntries.mu.RLock()
	defer rateLimitEntries.mu.RUnlock()

	entry, exists := rateLimitEntries.entries[address]
	if !exists {
		return false
	}
	return entry.violation > 10
}

// ============================================
// TYPES DE REQUÊTE
// ============================================

// SecureNodeMessageRequest — [MODIFIÉ v6.0]
//
// SUPPRIMÉ : Proof, ExpectedHash (format Groth16 — hex de bytes sérialisés).
// AJOUTÉ    : SchnorrTX, SchnorrTY, SchnorrS — les 3 composantes d'une
//             preuve Schnorr (voir pkg/zkp/schnorr.go, type SchnorrProof).
//             Envoyées en decimal string (big.Int.String()) pour rester
//             lisibles en JSON sans ambiguïté d'encodage.
type SecureNodeMessageRequest struct {
	DeviceID     string  `json:"device_id"`
	Address      string  `json:"address"`
	Message      string  `json:"message"`
	MessageHash  string  `json:"message_hash"`
	Signature    string  `json:"signature"`
	Timestamp    int64   `json:"timestamp"`
	Sequence     uint64  `json:"sequence"`
	Nonce        string  `json:"nonce"`

	// [NOUVEAU v6.0] Preuve Schnorr (remplace Proof/ExpectedHash Groth16)
	SchnorrTX string `json:"schnorr_tx"`
	SchnorrTY string `json:"schnorr_ty"`
	SchnorrS  string `json:"schnorr_s"`

	Duration        float64 `json:"duration"`
	BytesSent       float64 `json:"bytes_sent"`
	BytesReceived   float64 `json:"bytes_received"`
	PacketCount     float64 `json:"packet_count"`
	BytesPerSec     float64 `json:"bytes_per_sec"`
	PktsPerSec      float64 `json:"pkts_per_sec"`
	BytesPerPkt     float64 `json:"bytes_per_pkt"`
	BytesRatio      float64 `json:"bytes_ratio"`
	BytesSentRatio  float64 `json:"bytes_sent_ratio"`
	BytesTotal      float64 `json:"bytes_total"`
	LogBytesTotal   float64 `json:"log_bytes_total"`
	PktSizeVariance float64 `json:"pkt_size_variance"`
	HighVolumeFlag  float64 `json:"high_volume_flag"`
	BurstIndicator  float64 `json:"burst_indicator"`
}

// parseSchnorrProof convertit les champs string du JSON en SchnorrProof.
func parseSchnorrProof(txStr, tyStr, sStr string) (*zkp.SchnorrProof, error) {
	tx, ok := new(big.Int).SetString(txStr, 10)
	if !ok {
		return nil, fmt.Errorf("schnorr_tx invalide")
	}
	ty, ok := new(big.Int).SetString(tyStr, 10)
	if !ok {
		return nil, fmt.Errorf("schnorr_ty invalide")
	}
	s, ok := new(big.Int).SetString(sStr, 10)
	if !ok {
		return nil, fmt.Errorf("schnorr_s invalide")
	}
	return &zkp.SchnorrProof{TX: tx, TY: ty, S: s}, nil
}

// ============================================
// HANDLER PRINCIPAL — [MODIFIÉ v6.0]
//
// Ordre de validation :
//   1. Rate limit + validation champs
//   2. Vérification enregistrement blockchain
//   3. Vérification hash du message
//   4. Vérification ECDSA         ← mesure ecdsaMs (INCHANGÉ)
//   5. UpdateSequence             ← anti-replay (INCHANGÉ, protège aussi Schnorr)
//   6. Vérification Schnorr ZKP   ← [MODIFIÉ] mesure zkpMs, remplace Groth16
//   7. MarkNonceUsed              ← APRÈS Schnorr, jamais avant (INCHANGÉ)
//   8. Envoi ML (fire-and-forget) (INCHANGÉ)
//   9. Enregistrement blockchain   ← mesure blockchainMs (INCHANGÉ)
//  10. RecordTxTimings             ← alimente le dashboard (INCHANGÉ)
// ============================================

func HandleSecureNodeMessage(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()

	tTotal := time.Now()

	r.Body = http.MaxBytesReader(w, r.Body, MaxBodySize)

	var req SecureNodeMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Format JSON invalide", http.StatusBadRequest)
		return
	}

	req.Address = strings.ToLower(req.Address)
	clientIP := getClientIP(r)

	// ── 1. Rate limiting ────────────────────────────────────────────────────
	if !CheckRateLimit(req.Address) || !CheckRateLimit(clientIP) {
		logger.Warn("Rate limit dépassé", "ip", clientIP, "address", req.Address)
		sendError(w, "Trop de requêtes", http.StatusTooManyRequests)
		return
	}

	if !common.IsHexAddress(req.Address) {
		sendError(w, "Adresse invalide", http.StatusBadRequest)
		return
	}
	if req.DeviceID == "" || len(req.DeviceID) > 256 {
		sendError(w, "DeviceID invalide", http.StatusBadRequest)
		return
	}
	if len(req.Nonce) < MinNonceLength {
		sendError(w, "Nonce trop court", http.StatusBadRequest)
		return
	}

	now := time.Now().Unix()
	if now-req.Timestamp > MaxSequenceAge {
		sendError(w, "Message trop ancien (max 300s)", http.StatusBadRequest)
		return
	}
	if req.Timestamp > now+MaxFutureTimestamp {
		sendError(w, "Timestamp dans le futur", http.StatusBadRequest)
		return
	}

	deviceAddr := common.HexToAddress(req.Address)

	// ── 2. Vérification enregistrement blockchain ───────────────────────────
	var exists bool
	var err error
	for attempt := 1; attempt <= MaxRetryAttempts; attempt++ {
		err = blockchainCircuitBreaker.Call(func() error {
			var innerErr error
			exists, innerErr = blockchain.DeviceIsRegistered(deviceAddr)
			return innerErr
		})
		if err == nil && exists {
			break
		}
		if attempt < MaxRetryAttempts {
			time.Sleep(RetryDelay)
		}
	}

	if err != nil || !exists {
		logger.Warn("Appareil non enregistré", "address", req.Address)
		sendError(w, "Appareil non enregistré", http.StatusNotFound)
		return
	}

	isActive, err := blockchain.IsDeviceActive(deviceAddr)
	if err != nil {
		logger.Warn("Erreur vérification actif", "error", err)
	} else if !isActive {
		sendError(w, "Appareil désactivé", http.StatusForbidden)
		return
	}

	// ── 3. Vérification hash du message ─────────────────────────────────────
	if req.Signature == "" || req.MessageHash == "" {
		sendError(w, "Signature ou hash manquant", http.StatusBadRequest)
		return
	}

	computedHash    := crypto.Keccak256Hash([]byte(req.Message))
	computedHashHex := hex.EncodeToString(computedHash.Bytes())
	providedHash    := strings.TrimPrefix(strings.ToLower(req.MessageHash), "0x")

	if !strings.EqualFold(computedHashHex, providedHash) {
		logger.Warn("Message hash invalide", "device", req.DeviceID)
		sendError(w, "Message hash invalide", http.StatusBadRequest)
		return
	}

	// ── 4. Vérification ECDSA — mesure du temps réel ────────────────────────
	tECDSA := time.Now()
	validECDSA, err := zkp.VerifyECDSASignature(deviceAddr, req.MessageHash, req.Signature)
	ecdsaMs := time.Since(tECDSA).Milliseconds()
	ecdsaVerifyDuration.Observe(float64(ecdsaMs))

	if err != nil || !validECDSA {
		logger.Warn("Signature ECDSA invalide",
			"device", req.DeviceID,
			"ecdsa_ms", ecdsaMs,
		)
		sendError(w, "Authentification ECDSA échouée", http.StatusUnauthorized)
		return
	}

	// ── 5. Anti-replay séquence — [DÉPLACÉ AVANT Schnorr] ────────────────────
	// [SCHNORR-3] Contrairement à l'ancien flux (challenge pré-enregistré puis
	// consommé), ici UpdateSequence() est LA protection anti-rejeu principale :
	// rejouer une preuve Schnorr valide avec une sequence déjà vue est rejeté
	// ici, avant même de faire le calcul cryptographique Schnorr (plus rapide
	// à rejeter, moins de calcul gaspillé sur des requêtes rejouées).
	if !UpdateSequence(req.Address, req.Sequence) {
		sendError(w, "Séquence invalide (possible replay attack)", http.StatusBadRequest)
		return
	}

	// ── 6. Vérification Schnorr ZKP — [MODIFIÉ v6.0] ─────────────────────────
	devicePubKey, err := GetSchnorrPublicKey(req.Address)
	if err != nil {
		logger.Warn("Clé publique Schnorr introuvable",
			"device", req.DeviceID, "address", req.Address, "error", err)
		sendError(w, "Device non migré vers Schnorr — ré-enregistrement requis", http.StatusUnauthorized)
		return
	}

	proof, err := parseSchnorrProof(req.SchnorrTX, req.SchnorrTY, req.SchnorrS)
	if err != nil {
		sendError(w, "Preuve Schnorr malformée: "+err.Error(), http.StatusBadRequest)
		return
	}

	tZKP := time.Now()
	validZKP, err := zkp.VerifyProof(devicePubKey, proof, req.Address, req.Timestamp, req.Sequence, req.Nonce)
	zkpMs := time.Since(tZKP).Milliseconds()
	zkpVerifyDuration.Observe(float64(zkpMs))

	if err != nil || !validZKP {
		logger.Warn("Preuve Schnorr invalide",
			"device", req.DeviceID,
			"error", err,
			"zkp_ms", zkpMs,
		)
		sendError(w, "Authentification ZKP échouée", http.StatusUnauthorized)
		return
	}

	// ── 7. Marquer le nonce — UNIQUEMENT après ECDSA + Schnorr réussis ──────
	if !MarkNonceUsed(req.DeviceID, req.Nonce) {
		logger.Warn("Nonce déjà utilisé (replay attack?)",
			"device", req.DeviceID,
		)
		sendError(w, "Nonce déjà utilisé", http.StatusBadRequest)
		return
	}

	logger.Info("Double authentification réussie (ECDSA+Schnorr-ZKP)",
		"device", req.DeviceID,
		"ecdsa_ms", ecdsaMs,
		"zkp_ms", zkpMs,
	)

	// ── 8. Envoi ML — fire-and-forget avec circuit breaker ───────────────────
	go func() {
		if err := mlCircuitBreaker.Call(func() error {
			return sendToMLService(req)
		}); err != nil {
			logger.Warn("ML service indisponible", "error", err)
		}
	}()

	// ── 9. Enregistrement blockchain — mesure du temps réel ─────────────────
	// [SCHNORR-2] proofHash est maintenant dérivé du hash de la preuve Schnorr
	// elle-même (au lieu de expectedHashBytes qui n'existe plus), pour garder
	// une empreinte unique et vérifiable de chaque authentification on-chain.
	proofDigest := crypto.Keccak256(
		proof.TX.Bytes(), proof.TY.Bytes(), proof.S.Bytes(),
	)
	var proofHash [32]byte
	copy(proofHash[:], proofDigest)

	tBlockchain := time.Now()
	var txHash string
	err = blockchainCircuitBreaker.Call(func() error {
		tx, innerErr := blockchain.AuthRecordWithMeta(
			deviceAddr,
			req.DeviceID,
			req.Message,
			req.Sequence,
			true,
			proofHash,
		)
		if innerErr == nil {
			txHash = tx.Hex()
		}
		return innerErr
	})
	blockchainMs := time.Since(tBlockchain).Milliseconds()
	blockchainRecordDuration.Observe(float64(blockchainMs))

	if err != nil {
		logger.Error("Erreur enregistrement blockchain",
			"error", err,
			"blockchain_ms", blockchainMs,
		)
		sendError(w, "Erreur blockchain: "+err.Error(), http.StatusInternalServerError)
		return
	}

	totalMs := time.Since(tTotal).Milliseconds()
	authTotalDuration.Observe(float64(totalMs))

	if txHash != "" {
		RecordTxTimings(txHash, ecdsaMs, zkpMs, blockchainMs, totalMs)
	}

	logger.Info("Authentification complète",
		"device", req.DeviceID,
		"address", req.Address,
		"tx_hash", txHash,
		"ecdsa_ms", ecdsaMs,
		"zkp_ms", zkpMs,
		"blockchain_ms", blockchainMs,
		"total_ms", totalMs,
	)

	timings := AuthTimings{
		ECDSAMs:      ecdsaMs,
		ZKPMs:        zkpMs,
		BlockchainMs: blockchainMs,
		TotalMs:      totalMs,
	}

	response := models.NodeMessageResponse{
		Success: true,
		Message: "Authentifié avec succès (ECDSA + Schnorr-ZKP)",
		TxHash:  txHash,
		Block:   0,
		Timings: timings,
	}

	sendJSON(w, response, http.StatusOK)
}

// ============================================
// sendToMLService — [INCHANGÉ]
// ============================================

func sendToMLService(req SecureNodeMessageRequest) error {
	mlURL := os.Getenv("ML_SERVICE_URL")
	if mlURL == "" {
		mlURL = "http://localhost:5001"
	}

	if os.Getenv("ENV") == "production" && strings.HasPrefix(mlURL, "http://") {
		return fmt.Errorf("ML_SERVICE_URL doit utiliser https:// en production (actuel: %s)", mlURL)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mlPayload := map[string]interface{}{
		"device_address": req.Address,
		"device_id":      req.DeviceID,
		"entity_type":    "iot_device",
		"features": map[string]interface{}{
			"duration":          req.Duration,
			"bytes_sent":        req.BytesSent,
			"bytes_received":    req.BytesReceived,
			"packet_count":      req.PacketCount,
			"bytes_per_sec":     req.BytesPerSec,
			"pkts_per_sec":      req.PktsPerSec,
			"bytes_per_pkt":     req.BytesPerPkt,
			"bytes_ratio":       req.BytesRatio,
			"bytes_sent_ratio":  req.BytesSentRatio,
			"bytes_total":       req.BytesTotal,
			"log_bytes_total":   req.LogBytesTotal,
			"pkt_size_variance": req.PktSizeVariance,
			"high_volume_flag":  req.HighVolumeFlag,
			"burst_indicator":   req.BurstIndicator,
		},
	}

	mlPayloadBytes, err := json.Marshal(mlPayload)
	if err != nil {
		return fmt.Errorf("marshal payload ML: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		mlURL+"/v4/add_message",
		bytes.NewBuffer(mlPayloadBytes),
	)
	if err != nil {
		return fmt.Errorf("construction requête ML: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("appel ML: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, io.LimitReader(resp.Body, 64*1024)) //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ML service HTTP %d", resp.StatusCode)
	}

	return nil
}

// ============================================
// UTILITAIRES — [INCHANGÉ]
// ============================================

func getClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func safeAddr(addr string, n int) string {
	if len(addr) <= n {
		return addr
	}
	return addr[:n] + "…"
}

// NOTE [SCHNORR-4] : HandleSecureZKPGenerate est SUPPRIMÉ dans cette version.
// L'ancien endpoint POST /api/zkp/generate-secure n'est plus utilisé : le
// device génère maintenant sa preuve Schnorr localement et l'envoie
// directement dans SecureNodeMessageRequest (voir iot_simulator.go v6.0).
// Retirez la route correspondante dans main_secure.go (voir
// INTEGRATION_SCHNORR.md, section "main_secure.go").

// ============================================
// RÉINITIALISATION (DEBUG UNIQUEMENT) — [INCHANGÉ]
// ============================================

func ResetSequences() {
	if sequenceStore != nil {
		sequenceStore.mu.Lock()
		sequenceStore.sequences = make(map[string]SequenceEntry)
		sequenceStore.mu.Unlock()
		_ = sequenceStore.save()
	}
	memorySequencesMu.Lock()
	memorySequences = make(map[string]SequenceEntry)
	memorySequencesMu.Unlock()
}

func ResetNonces() {
	if nonceStore != nil {
		nonceStore.mu.Lock()
		nonceStore.nonces = make(map[string]NonceEntry)
		nonceStore.mu.Unlock()
		_ = nonceStore.save()
	}
	memoryNoncesMu.Lock()
	memoryNonces = make(map[string]bool)
	memoryNoncesMu.Unlock()
}

var _ = safeAddr // conservé pour compatibilité, utilisé ailleurs dans le package
