// iot_simulator.go
// CHAIMAN IoT — Benchmark Scientifique avec Mesures Blockchain Réelles
// Version 6.4.1 — CORRECTIONS wg.Wait + authWindowDuration + stopChan + State Persistence
//
// CORRECTIONS v6.4 → v6.4.1 :
//
//   [FIX-SIM-21] Persistance de l'état simulateur (séquences + nonces).
//                PROBLÈME : à chaque redémarrage, Sequence: 0 provoquait
//                "Séquence invalide" côté backend car sequences.json persiste 24h.
//                SOLUTION : sauvegarde/restauration de simulator_state.json
//                avec séquence basée sur timestamp Unix si fichier absent.
//
//   [FIX-SIM-22] Reset conditionnel de l'état backend en début de benchmark.
//                PROBLÈME : relances successives échouaient à 100% à cause
//                des nonces et séquences déjà consommés côté backend.
//                SOLUTION : appel POST /api/debug/reset-state au démarrage
//                si BENCHMARK_MODE=true côté backend.
//
//   [FIX-SIM-23] Version affichée corrigée v6.3 → v6.4.1-stable.
//
//   Toutes les corrections v6.3/v6.4 sont conservées.
package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ============================================
// CONFIGURATION
// ============================================

var (
	backendURL   = flag.String("backend", "", "URL du backend")
	nodeURL       = flag.String("node", "", "URL du nœud Ethereum (pour confirmation)")
	outputDir     = flag.String("output", "./scientific_results", "Répertoire de sortie")
	waitFinal     = flag.Bool("finality", false, "Attendre 12 confirmations (finalité)")
	// [FIX-SIM-14] Intervalle entre deux cycles d'auth d'un même device.
	authInterval  = flag.Duration("auth-interval", 15*time.Second, "Intervalle entre deux authentifications d'un même device")
	// [FIX-SIM-22] Reset automatique de l'état backend en mode benchmark.
	resetBackend  = flag.Bool("reset-backend", true, "Réinitialiser l'état backend (sequences/nonces) au démarrage")
)

// ============================================
// STRUCTURES DE DONNÉES
// ============================================

// ScientificStats contient TOUTES les métriques d'une opération individuelle.
type ScientificStats struct {
	DeviceID  string `json:"device_id"`
	Address   string `json:"address"`
	Phase     int    `json:"phase"`
	Timestamp int64  `json:"timestamp_unix"`
	ISO8601   string `json:"timestamp_iso"`

	// ── Enregistrement ──────────────────────────────────────────────────────
	RegHTTP_ms    int64 `json:"reg_t1_http_ms"`
	RegConfirm_ms int64 `json:"reg_t2_confirm_ms"`
	RegFinal_ms   int64 `json:"reg_t3_final_ms"`

	RegTxHash      string `json:"reg_tx_hash"`
	RegBlockNumber uint64 `json:"reg_block_number"`
	RegGasUsed     uint64 `json:"reg_gas_used"`
	RegBlockHash   string `json:"reg_block_hash"`
	RegConfirmed   bool   `json:"reg_confirmed"`
	RegFinalized   bool   `json:"reg_finalized"`
	Registered     bool   `json:"registered"`

	// ── Authentification — décomposition fine ───────────────────────────────
	AuthECDSA_ms   float64 `json:"auth_t1_ecdsa_ms"`
	AuthZKP_ms     int64   `json:"auth_t2_zkp_ms"`
	AuthHTTP_ms    int64   `json:"auth_t3_http_ms"`
	AuthBackend_ms int64   `json:"auth_t4_backend_est_ms"`
	AuthApp_ms     int64   `json:"auth_t5_app_ms"`
	AuthConfirm_ms int64   `json:"auth_t6_confirm_ms"`
	AuthFinal_ms   int64   `json:"auth_t7_final_ms"`
	AuthTotal_ms   int64   `json:"auth_total_ms"`

	AuthTxHash      string `json:"auth_tx_hash"`
	AuthBlockNumber uint64 `json:"auth_block_number"`
	AuthGasUsed     uint64 `json:"auth_gas_used"`
	AuthBlockHash   string `json:"auth_block_hash"`
	AuthConfirmed   bool   `json:"auth_confirmed"`
	AuthFinalized   bool   `json:"auth_finalized"`
	AuthSuccess     bool   `json:"auth_success"`

	PayloadBytes int `json:"payload_bytes"`
}

// PhaseResult agrège les métriques d'une phase avec statistiques complètes.
type PhaseResult struct {
	PhaseNum     int `json:"phase"`
	NewDevices   int `json:"new_devices"`
	TotalDevices int `json:"total_devices"`

	RegHTTP_mean      float64 `json:"reg_http_mean_ms"`
	RegHTTP_median    float64 `json:"reg_http_median_ms"`
	RegHTTP_p95       float64 `json:"reg_http_p95_ms"`
	RegHTTP_p99       float64 `json:"reg_http_p99_ms"`
	RegHTTP_stddev    float64 `json:"reg_http_stddev_ms"`
	RegHTTP_ic95      float64 `json:"reg_http_ic95_ms"`
	RegConfirm_mean   float64 `json:"reg_confirm_mean_ms"`
	RegConfirm_median float64 `json:"reg_confirm_median_ms"`
	RegConfirm_stddev float64 `json:"reg_confirm_stddev_ms"`
	RegSuccessRate    float64 `json:"reg_success_rate"`

	AuthECDSA_mean   float64 `json:"auth_ecdsa_mean_ms"`
	AuthECDSA_stddev float64 `json:"auth_ecdsa_stddev_ms"`
	AuthZKP_mean     float64 `json:"auth_zkp_mean_ms"`
	AuthZKP_stddev   float64 `json:"auth_zkp_stddev_ms"`

	AuthHTTP_mean   float64 `json:"auth_http_mean_ms"`
	AuthHTTP_median float64 `json:"auth_http_median_ms"`
	AuthHTTP_p95    float64 `json:"auth_http_p95_ms"`
	AuthHTTP_stddev float64 `json:"auth_http_stddev_ms"`

	AuthApp_mean   float64 `json:"auth_app_mean_ms"`
	AuthApp_median float64 `json:"auth_app_median_ms"`
	AuthApp_p95    float64 `json:"auth_app_p95_ms"`
	AuthApp_p99    float64 `json:"auth_app_p99_ms"`
	AuthApp_stddev float64 `json:"auth_app_stddev_ms"`
	AuthApp_ic95   float64 `json:"auth_app_ic95_ms"`

	AuthConfirm_mean   float64 `json:"auth_confirm_mean_ms"`
	AuthConfirm_median float64 `json:"auth_confirm_median_ms"`
	AuthConfirm_stddev float64 `json:"auth_confirm_stddev_ms"`

	AuthTotal_mean   float64 `json:"auth_total_mean_ms"`
	AuthTotal_median float64 `json:"auth_total_median_ms"`
	AuthTotal_p95    float64 `json:"auth_total_p95_ms"`
	AuthTotal_stddev float64 `json:"auth_total_stddev_ms"`
	AuthTotal_ic95   float64 `json:"auth_total_ic95_ms"`

	AuthSuccessRate float64 `json:"auth_success_rate"`

	AuthThroughput     float64 `json:"auth_throughput_per_sec"`
	AuthWindowDuration float64 `json:"auth_window_duration_sec"`
}

// ============================================
// VARIABLES GLOBALES
// ============================================

var (
	secureHTTPClient *http.Client
	ethClient        *ethclient.Client

	allDevices []*IoTDevice
	allStats   []ScientificStats
	statsMu    sync.Mutex

	totalRegAttempts   int64
	totalRegConfirmed  int64
	totalAuthAttempts  int64
	totalAuthConfirmed int64
)

type IoTDevice struct {
	ID         string
	Name       string
	Address    string
	Type       string
	PrivateKey *ecdsa.PrivateKey
	Sequence   uint64
	mu         sync.Mutex
}

func (d *IoTDevice) nextSequence() uint64 {
	return atomic.AddUint64(&d.Sequence, 1)
}

// ============================================
// UTILITAIRES
// ============================================

func safeTxHash(txHash string, n int) string {
	if len(txHash) < n {
		if txHash == "" {
			return "N/A"
		}
		return txHash
	}
	return txHash[:n]
}

// ============================================
// INITIALISATION
// ============================================

func initHTTPClient() {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		},
	}

	if caCertPath := os.Getenv("BACKEND_CA_CERT"); caCertPath != "" {
		caCert, err := os.ReadFile(caCertPath)
		if err != nil {
			log.Fatalf("❌ BACKEND_CA_CERT introuvable : %s → %v", caCertPath, err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			log.Fatalf("❌ BACKEND_CA_CERT invalide : %s", caCertPath)
		}
		tlsConfig.RootCAs = caCertPool
		log.Printf("✅ CA chargée : %s", caCertPath)
	}

	clientCertPath := os.Getenv("CLIENT_CERT_PATH")
	clientKeyPath := os.Getenv("CLIENT_KEY_PATH")
	if clientCertPath != "" && clientKeyPath != "" {
		clientCert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
		if err != nil {
			log.Fatalf("❌ Certificat client mTLS invalide : %v", err)
		}
		tlsConfig.Certificates = []tls.Certificate{clientCert}
		log.Printf("✅ Certificat client mTLS chargé : %s", clientCertPath)
	} else if os.Getenv("TLS_INSECURE_DEV") != "true" {
		log.Printf("⚠️  CLIENT_CERT_PATH / CLIENT_KEY_PATH non définis — mTLS désactivé côté client")
	}

	if os.Getenv("TLS_INSECURE_DEV") == "true" {
		tlsConfig.InsecureSkipVerify = true
		log.Printf("⚠️  TLS_INSECURE_DEV=true — vérification certificat désactivée")
	}

	secureHTTPClient = &http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:     tlsConfig,
			MaxIdleConns:        200,
			MaxIdleConnsPerHost: 50,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}

func initEthereumClient() {
	node := *nodeURL
	if node == "" {
		node = os.Getenv("NODE_URL")
	}
	if node == "" {
		log.Fatal("❌ NODE_URL requis. Usage: -node=http://localhost:8545")
	}

	var err error
	ethClient, err = ethclient.Dial(node)
	if err != nil {
		log.Fatalf("❌ Connexion nœud Ethereum %s: %v", node, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	chainID, err := ethClient.NetworkID(ctx)
	if err != nil {
		log.Fatalf("❌ ChainID: %v", err)
	}
	blockNumber, err := ethClient.BlockNumber(ctx)
	if err != nil {
		log.Fatalf("❌ BlockNumber: %v", err)
	}

	log.Printf("✅ Client Ethereum connecté | ChainID: %s | Block: %d", chainID, blockNumber)
}

func getBackendURL() string {
	if *backendURL != "" {
		return *backendURL
	}
	u := os.Getenv("BACKEND_URL")
	if u == "" {
		log.Fatal("❌ BACKEND_URL non défini")
	}
	return u
}

func randomDeviceID(index int) string {
	return fmt.Sprintf("IOT_%04d", index)
}

// ============================================
// [FIX-SIM-22] RESET ÉTAT BACKEND
// ============================================

// resetBackendState appelle l'endpoint de reset côté backend pour vider
// sequences.json et nonces.json. Nécessite BENCHMARK_MODE=true côté backend.
func resetBackendState(backend string) {
	if !*resetBackend {
		log.Println("ℹ️  Reset backend désactivé (flag -reset-backend=false)")
		return
	}

	log.Println("🧹 Réinitialisation de l'état backend (sequences + nonces)...")
	resp, err := secureHTTPClient.Post(
		backend+"/api/debug/reset-state",
		"application/json",
		bytes.NewReader([]byte(`{}`)),
	)
	if err != nil {
		log.Printf("⚠️  Reset backend indisponible: %v (continuer quand même)", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Println("✅ État backend réinitialisé (sequences.json + nonces.json vidés)")
	} else if resp.StatusCode == http.StatusNotFound {
		log.Println("⚠️  Endpoint /api/debug/reset-state non trouvé — le backend doit être lancé avec BENCHMARK_MODE=true")
	} else {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		log.Printf("⚠️  Reset backend HTTP %d: %s", resp.StatusCode, string(body))
	}
}

// ============================================
// [FIX-SIM-21] PERSISTANCE ÉTAT SIMULATEUR
// ============================================

type simulatorState struct {
	Devices []struct {
		ID       string `json:"id"`
		Address  string `json:"address"`
		Sequence uint64 `json:"sequence"`
	} `json:"devices"`
	SavedAt string `json:"saved_at"`
}

func stateFilePath() string {
	return *outputDir + "/simulator_state.json"
}

// saveDeviceState persiste les séquences actuelles des devices pour permettre
// des relances sans collision avec l'état backend.
func saveDeviceState(devices []*IoTDevice) {
	os.MkdirAll(*outputDir, 0755)
	state := simulatorState{SavedAt: time.Now().Format(time.RFC3339)}
	for _, d := range devices {
		state.Devices = append(state.Devices, struct {
			ID       string `json:"id"`
			Address  string `json:"address"`
			Sequence uint64 `json:"sequence"`
		}{ID: d.ID, Address: d.Address, Sequence: atomic.LoadUint64(&d.Sequence)})
	}
	data, _ := json.MarshalIndent(state, "", "  ")
	_ = os.WriteFile(stateFilePath(), data, 0644)
	log.Printf("💾 État simulateur sauvegardé (%d devices)", len(devices))
}

// loadDeviceState restaure les séquences depuis le fichier d'état.
// Si le fichier est absent, initialise avec un timestamp Unix pour éviter
// toute collision avec un état backend persistant.
func loadDeviceState(devices []*IoTDevice) {
	data, err := os.ReadFile(stateFilePath())
	if err != nil {
		log.Println("ℹ️  Pas d'état précédent — initialisation avec séquences temporelles")
		baseSeq := uint64(time.Now().Unix())
		for _, d := range devices {
			atomic.StoreUint64(&d.Sequence, baseSeq)
		}
		return
	}

	var state simulatorState
	if err := json.Unmarshal(data, &state); err != nil {
		log.Printf("⚠️  État corrompu — réinitialisation temporelle: %v", err)
		baseSeq := uint64(time.Now().Unix())
		for _, d := range devices {
			atomic.StoreUint64(&d.Sequence, baseSeq)
		}
		return
	}

	seqMap := make(map[string]uint64)
	for _, sd := range state.Devices {
		seqMap[sd.Address] = sd.Sequence
	}

	restored := 0
	for _, d := range devices {
		if seq, ok := seqMap[d.Address]; ok {
			atomic.StoreUint64(&d.Sequence, seq)
			restored++
		} else {
			atomic.StoreUint64(&d.Sequence, uint64(time.Now().Unix()))
		}
	}
	log.Printf("📂 État simulateur restauré (%d/%d devices depuis sauvegarde)", restored, len(devices))
}

// ============================================
// CONFIRMATION ON-CHAIN
// ============================================

func waitForConfirmation(txHash string, timeout time.Duration, waitFinality bool) (confirmMs, finalMs int64, receipt *types.Receipt, ok bool) {
	if ethClient == nil {
		return 0, 0, nil, false
	}

	start := time.Now()
	hash := common.HexToHash(txHash)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var firstReceipt *types.Receipt
	var confirmTime time.Time

	for {
		select {
		case <-ctx.Done():
			elapsed := time.Since(start).Milliseconds()
			if firstReceipt != nil {
				return time.Since(confirmTime).Milliseconds(), elapsed, firstReceipt, true
			}
			return elapsed, 0, nil, false

		case <-ticker.C:
			receipt, err := ethClient.TransactionReceipt(context.Background(), hash)
			if err == nil && receipt != nil {
				if firstReceipt == nil {
					firstReceipt = receipt
					confirmTime = time.Now()
					confirmMs = time.Since(start).Milliseconds()
					log.Printf("   ⛏️  Confirmée bloc #%d | Gas: %d | Temps: %d ms",
						receipt.BlockNumber, receipt.GasUsed, confirmMs)
					if !waitFinality {
						return confirmMs, 0, receipt, true
					}
				}
				if waitFinality && firstReceipt != nil {
					currentBlock, err := ethClient.BlockNumber(context.Background())
					if err == nil {
						confirmations := currentBlock - receipt.BlockNumber.Uint64()
						if confirmations >= 12 {
							finalMs = time.Since(start).Milliseconds()
							log.Printf("   🔒 Finalisée (%d confirmations) | %d ms", confirmations, finalMs)
							return confirmMs, finalMs, receipt, true
						}
					}
				}
			}
		}
	}
}

// ============================================
// ENREGISTREMENT
// ============================================

func (d *IoTDevice) RegisterScientific(backend string, phase int) ScientificStats {
	now := time.Now()
	stats := ScientificStats{
		DeviceID:  d.ID,
		Address:   d.Address,
		Phase:     phase,
		Timestamp: now.Unix(),
		ISO8601:   now.Format(time.RFC3339Nano),
	}

	payload := map[string]interface{}{
		"address":    d.Address,
		"entityType": "iot_device",
		"publicKey":  hex.EncodeToString(crypto.FromECDSAPub(&d.PrivateKey.PublicKey)),
		"metadata":   fmt.Sprintf(`{"type":%q,"name":%q,"phase":%d}`, d.Type, d.Name, phase),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("❌ %s: marshal: %v", d.ID, err)
		return stats
	}

	startHTTP := time.Now()
	resp, err := secureHTTPClient.Post(
		backend+"/api/device/register",
		"application/json",
		bytes.NewReader(body),
	)
	stats.RegHTTP_ms = time.Since(startHTTP).Milliseconds()

	if err != nil {
		log.Printf("❌ %s: réseau: %v", d.ID, err)
		return stats
	}
	defer resp.Body.Close()

	bodyResp, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var result map[string]interface{}
	json.Unmarshal(bodyResp, &result)

	if th, ok := result["txHash"].(string); ok && th != "" {
		stats.RegTxHash = th
	}

	if resp.StatusCode == http.StatusOK {
		stats.Registered = true
		atomic.AddInt64(&totalRegAttempts, 1)

		if stats.RegTxHash != "" {
			log.Printf("⏳ %s: Attente confirmation (tx: %s)...", d.ID, safeTxHash(stats.RegTxHash, 16))

			// [FIX-SIM-6] Timeout augmenté 30s → 60s pour absorber la congestion PoA
			// quand le scanner Geth consomme des nonces en parallèle.
			confirmMs, finalMs, receipt, confirmed := waitForConfirmation(
				stats.RegTxHash, 60*time.Second, *waitFinal,
			)
			stats.RegConfirm_ms = confirmMs
			stats.RegFinal_ms = finalMs
			stats.RegConfirmed = confirmed
			if receipt != nil {
				stats.RegBlockNumber = receipt.BlockNumber.Uint64()
				stats.RegGasUsed = receipt.GasUsed
				stats.RegBlockHash = receipt.BlockHash.Hex()

				if receipt.Status == 0 {
					log.Printf("❌ %s: TX REVERTÉE on-chain (status=0) | Gas:%d — device NON enregistré",
						d.ID, receipt.GasUsed)
					stats.Registered = false
					stats.RegConfirmed = false
					return stats
				}
			}
			if confirmed {
				atomic.AddInt64(&totalRegConfirmed, 1)
				if finalMs > 0 {
					stats.RegFinalized = true
				}

				log.Printf("🔓 %s: Activation du device...", d.ID)
				if activateErr := d.activateDevice(backend); activateErr != nil {
					log.Printf("⚠️  %s: Activation échouée: %v (on continue)", d.ID, activateErr)
				} else {
					log.Printf("✅ %s: Device activé avec succès", d.ID)
				}
			}
		}

		log.Printf("✅ %s ENREGISTRÉ | HTTP:%4dms Confirm:%5dms Gas:%d Block:#%d",
			d.ID, stats.RegHTTP_ms, stats.RegConfirm_ms, stats.RegGasUsed, stats.RegBlockNumber)

	} else if resp.StatusCode == http.StatusConflict {
		stats.Registered = true
		log.Printf("ℹ️  %s: Déjà enregistré (409) — tentative d'activation", d.ID)
		if activateErr := d.activateDevice(backend); activateErr != nil {
			log.Printf("⚠️  %s: Activation échouée: %v", d.ID, activateErr)
		} else {
			log.Printf("✅ %s: Device activé (déjà enregistré)", d.ID)
		}
	} else {
		log.Printf("⚠️  %s: HTTP %d | Body: %s", d.ID, resp.StatusCode, string(bodyResp))
	}

	return stats
}

func (d *IoTDevice) activateDevice(backend string) error {
	payload := map[string]interface{}{
		"deviceAddress": d.Address,
	}
	body, _ := json.Marshal(payload)

	maxRetries := 3
	backoff := 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := secureHTTPClient.Post(
			backend+"/api/device/activate",
			"application/json",
			bytes.NewReader(body),
		)
		if err != nil {
			if attempt < maxRetries {
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return fmt.Errorf("réseau après %d tentatives: %w", maxRetries, err)
		}

		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		resp.Body.Close()
		bodyStr := string(respBody)

		if resp.StatusCode == http.StatusOK {
			return nil
		}

		if resp.StatusCode == http.StatusBadRequest {
			if strings.Contains(bodyStr, "déjà actif") ||
				strings.Contains(bodyStr, "already active") ||
				strings.Contains(bodyStr, "deja actif") {
				return nil
			}
		}

		if attempt < maxRetries {
			log.Printf("   ⏳ %s: activation HTTP %d (tentative %d/%d)...",
				d.ID, resp.StatusCode, attempt, maxRetries)
			time.Sleep(backoff)
			backoff *= 2
		} else {
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, bodyStr)
		}
	}
	return fmt.Errorf("activation échouée après %d tentatives", maxRetries)
}

// ============================================
// VÉRIFICATION PRE-AUTH
// ============================================

func (d *IoTDevice) verifyDeviceRegistered(backend string) bool {
	maxRetries := 8
	backoff := 3 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := secureHTTPClient.Get(
			fmt.Sprintf("%s/api/device/%s/status", backend, d.Address),
		)
		if err == nil {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			resp.Body.Close()

			switch resp.StatusCode {
			case http.StatusOK:
				var result map[string]interface{}
				if json.Unmarshal(body, &result) == nil {
					if active, ok := result["isActive"].(bool); ok && active {
						log.Printf("   ✅ %s: visible et actif sur le backend", d.ID)
						return true
					}
					log.Printf("   ⚠️  %s: visible mais inactif (isActive=false) — tentative activation (essai %d/%d)",
						d.ID, attempt, maxRetries)
					if activateErr := d.activateDevice(backend); activateErr != nil {
						log.Printf("   ❌ %s: activation échouée: %v", d.ID, activateErr)
					} else {
						time.Sleep(1 * time.Second)
						respCheck, errCheck := secureHTTPClient.Get(
							fmt.Sprintf("%s/api/device/%s/status", backend, d.Address),
						)
						if errCheck == nil {
							checkBody, _ := io.ReadAll(io.LimitReader(respCheck.Body, 512))
							respCheck.Body.Close()
							var checkResult map[string]interface{}
							if json.Unmarshal(checkBody, &checkResult) == nil {
								if active, ok := checkResult["isActive"].(bool); ok && active {
									log.Printf("   ✅ %s: activé avec succès", d.ID)
									return true
								}
							}
						}
					}
				}

			case http.StatusNotFound:
				log.Printf("   ❌ %s: backend 404 — device non indexé (essai %d/%d)",
					d.ID, attempt, maxRetries)

			default:
				log.Printf("   ⚠️  %s: backend HTTP %d (essai %d/%d)",
					d.ID, resp.StatusCode, attempt, maxRetries)
			}
		} else {
			if ethClient != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				_, blockErr := ethClient.BlockNumber(ctx)
				cancel()
				if blockErr == nil && attempt >= 4 {
					log.Printf("   ⡠ %s: backend inaccessible mais chaîne OK — inclus (essai %d)",
						d.ID, attempt)
					return true
				}
			}
		}

		if attempt < maxRetries {
			jitter := time.Duration(rand.Int63n(1000)) * time.Millisecond
			waitDur := backoff + jitter
			log.Printf("   ⏳ %s: retry %d/%d dans %v...", d.ID, attempt, maxRetries, waitDur)
			time.Sleep(waitDur)
			if backoff < 15*time.Second {
				backoff = backoff * 3 / 2
			}
		}
	}

	log.Printf("   ⚠️  %s: non actif sur le backend après %d tentatives — exclu", d.ID, maxRetries)
	return false
}

// ============================================
// AUTHENTIFICATION
// ============================================

func (d *IoTDevice) AuthenticateScientific(backend string, phase int) ScientificStats {
	now := time.Now()
	stats := ScientificStats{
		DeviceID:  d.ID,
		Address:   d.Address,
		Phase:     phase,
		Timestamp: now.Unix(),
		ISO8601:   now.Format(time.RFC3339Nano),
	}

	tAppStart := time.Now()

	timestamp := time.Now().Unix()
	seq := d.nextSequence()
	nonce := fmt.Sprintf("%s-%d-%d", d.ID, timestamp, seq)

	data := map[string]interface{}{
		"temperature": 20.0 + float64(rand.Intn(1000))/100.0,
		"humidity":    50.0 + float64(rand.Intn(3000))/100.0,
		"device_id":   d.ID,
		"timestamp":   timestamp,
	}
	message, _ := json.Marshal(data)
	messageStr := string(message)
	stats.PayloadBytes = len(messageStr)

	msgHash := crypto.Keccak256Hash([]byte(messageStr))
	msgHashHex := hex.EncodeToString(msgHash.Bytes())

	startECDSA := time.Now()
	signature, err := crypto.Sign(msgHash.Bytes(), d.PrivateKey)
	stats.AuthECDSA_ms = float64(time.Since(startECDSA).Microseconds()) / 1000.0
	if err != nil {
		log.Printf("❌ %s: ECDSA: %v", d.ID, err)
		return stats
	}
	sigHex := hex.EncodeToString(signature)

	startZKP := time.Now()
	zkpProof, zkpExpectedHash, zkpChallenge, err := d.requestZKPProofWithRetry(
		backend, msgHashHex, sigHex, timestamp, seq, nonce,
	)
	stats.AuthZKP_ms = time.Since(startZKP).Milliseconds()

	if err != nil {
		log.Printf("❌ %s: ZKP: %v", d.ID, err)
		return stats
	}

	payload := map[string]interface{}{
		"device_id":     d.ID,
		"address":       d.Address,
		"message":       messageStr,
		"message_hash":  msgHashHex,
		"signature":     sigHex,
		"proof":         zkpProof,
		"expected_hash": zkpExpectedHash,
		"challenge":     zkpChallenge,
		"timestamp":     timestamp,
		"sequence":      seq,
		"nonce":         nonce,
		"duration":          2.5,
		"bytes_sent":        float64(len(messageStr)),
		"bytes_received":    256.0,
		"packet_count":      4.0,
		"bytes_per_sec":     256.0,
		"pkts_per_sec":      2.0,
		"bytes_per_pkt":     128.0,
		"bytes_ratio":       0.5,
		"bytes_sent_ratio":  0.67,
		"bytes_total":       float64(len(messageStr) + 256),
		"log_bytes_total":   6.64,
		"pkt_size_variance": 64.0,
		"high_volume_flag":  0.0,
		"burst_indicator":   1.0,
	}
	body, _ := json.Marshal(payload)

	startHTTP := time.Now()
	resp, err := secureHTTPClient.Post(
		backend+"/api/node/message-secure",
		"application/json",
		bytes.NewReader(body),
	)
	stats.AuthHTTP_ms = time.Since(startHTTP).Milliseconds()

	stats.AuthApp_ms = int64(stats.AuthECDSA_ms) + stats.AuthZKP_ms + stats.AuthHTTP_ms

	backendEst := stats.AuthHTTP_ms - stats.AuthZKP_ms - int64(stats.AuthECDSA_ms)
	if backendEst < 0 {
		backendEst = 0
	}
	stats.AuthBackend_ms = backendEst

	if err != nil {
		log.Printf("❌ %s: auth réseau: %v", d.ID, err)
		return stats
	}
	defer resp.Body.Close()

	bodyResp, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var result map[string]interface{}
	json.Unmarshal(bodyResp, &result)

	if th, ok := result["txHash"].(string); ok && th != "" {
		stats.AuthTxHash = th
	}
	if s, ok := result["success"].(bool); ok {
		stats.AuthSuccess = s
	}

	// [FIX-SIM-9] Timeout auth confirmation 30s → 60s
	if stats.AuthSuccess && stats.AuthTxHash != "" {
		atomic.AddInt64(&totalAuthAttempts, 1)
		log.Printf("⏳ %s: Attente confirmation auth (tx: %s)...", d.ID, safeTxHash(stats.AuthTxHash, 16))
		confirmMs, finalMs, receipt, confirmed := waitForConfirmation(
			stats.AuthTxHash, 60*time.Second, *waitFinal,
		)
		stats.AuthConfirm_ms = confirmMs
		stats.AuthFinal_ms = finalMs
		stats.AuthConfirmed = confirmed

		if receipt != nil {
			stats.AuthBlockNumber = receipt.BlockNumber.Uint64()
			stats.AuthGasUsed = receipt.GasUsed
			stats.AuthBlockHash = receipt.BlockHash.Hex()
		}
		if confirmed {
			atomic.AddInt64(&totalAuthConfirmed, 1)
			if finalMs > 0 {
				stats.AuthFinalized = true
			}
		}
	}

	stats.AuthTotal_ms = stats.AuthApp_ms + stats.AuthConfirm_ms

	_ = tAppStart
	if stats.AuthConfirmed {
		log.Printf(
			"📡 %s AUTH OK | ECDSA:%7.3fms ZKP:%4dms HTTP:%5dms App:%5dms Confirm:%5dms Total:%6dms Gas:%d Bloc:#%d",
			d.ID, stats.AuthECDSA_ms, stats.AuthZKP_ms, stats.AuthHTTP_ms,
			stats.AuthApp_ms, stats.AuthConfirm_ms, stats.AuthTotal_ms,
			stats.AuthGasUsed, stats.AuthBlockNumber,
		)
	} else if stats.AuthSuccess {
		log.Printf(
			"⚠️  %s AUTH HTTP OK (non confirmé) | ECDSA:%7.3fms ZKP:%4dms HTTP:%5dms App:%5dms",
			d.ID, stats.AuthECDSA_ms, stats.AuthZKP_ms, stats.AuthHTTP_ms, stats.AuthApp_ms,
		)
	} else {
		log.Printf(
			"❌ %s AUTH ÉCHEC | HTTP %d | ECDSA:%7.3fms ZKP:%4dms HTTP:%5dms",
			d.ID, resp.StatusCode, stats.AuthECDSA_ms, stats.AuthZKP_ms, stats.AuthHTTP_ms,
		)
	}

	return stats
}

func (d *IoTDevice) requestZKPProofWithRetry(
	backend, messageHash, signature string,
	timestamp int64, sequence uint64, nonce string,
) (proof, expectedHash, challenge string, err error) {

	maxRetries := 3
	backoff := 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		proof, expectedHash, challenge, err = d.requestZKPProof(
			backend, messageHash, signature, timestamp, sequence, nonce,
		)
		if err == nil {
			return proof, expectedHash, challenge, nil
		}

		if attempt < maxRetries {
			// [FIX-SIM-15] Respecter le header Retry-After si le backend le fournit.
			wait := backoff + time.Duration(rand.Int63n(1000))*time.Millisecond
			if ze, ok := err.(*zkpError); ok && ze.retryAfter > wait {
				wait = ze.retryAfter + time.Duration(rand.Int63n(500))*time.Millisecond
			}
			log.Printf("   🔄 %s: ZKP retry %d/%d (attente %v): %v",
				d.ID, attempt, maxRetries, wait, err)
			time.Sleep(wait)
			backoff *= 2
		}
	}
	return "", "", "", fmt.Errorf("ZKP échoué après %d tentatives: %w", maxRetries, err)
}


func (d *IoTDevice) requestZKPProof(
	backend, messageHash, signature string,
	timestamp int64, sequence uint64, nonce string,
) (proof, expectedHash, challenge string, err error) {

	payload := map[string]interface{}{
		"signature":    signature,
		"message_hash": messageHash,
		"address":      d.Address,
		"sequence":     sequence,
		"timestamp":    timestamp,
		"nonce":        nonce,
	}
	body, _ := json.Marshal(payload)

	resp, err := secureHTTPClient.Post(
		backend+"/api/zkp/generate-secure",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", "", "", fmt.Errorf("réseau: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		// [FIX-SIM-15] Extraire le délai Retry-After si le backend l'indique (HTTP 429).
		var retryAfterHint time.Duration
		if resp.StatusCode == http.StatusTooManyRequests {
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if secs, parseErr := strconv.Atoi(ra); parseErr == nil && secs > 0 {
					retryAfterHint = time.Duration(secs) * time.Second
				}
			}
		}
		return "", "", "", &zkpError{
			msg:        fmt.Sprintf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes))),
			retryAfter: retryAfterHint,
		}
	}

	var zkpResp struct {
		Proof        string `json:"proof"`
		ExpectedHash string `json:"expected_hash"`
		Challenge    string `json:"challenge"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&zkpResp); err != nil {
		return "", "", "", fmt.Errorf("décodage réponse ZKP: %w", err)
	}

	return zkpResp.Proof, zkpResp.ExpectedHash, zkpResp.Challenge, nil
}

// zkpError transporte un message d'erreur ZKP et un délai Retry-After optionnel.
type zkpError struct {
	msg        string
	retryAfter time.Duration
}

func (e *zkpError) Error() string { return e.msg }


// ============================================
// GÉNÉRATION D'APPAREILS
// ============================================

func generateDevices(count int, startIndex int) []*IoTDevice {
	deviceTypes := []string{"temperature_sensor", "humidity_sensor", "motion_sensor", "door_sensor", "energy_meter"}
	locations := []string{"salon", "cuisine", "chambre", "garage", "bureau", "jardin", "cave", "grenier"}

	devices := make([]*IoTDevice, count)
	for i := 0; i < count; i++ {
		privateKey, err := crypto.GenerateKey()
		if err != nil {
			log.Fatalf("❌ Génération clé: %v", err)
		}
		addr := crypto.PubkeyToAddress(privateKey.PublicKey)
		deviceIndex := startIndex + i + 1

		devices[i] = &IoTDevice{
			ID:         randomDeviceID(deviceIndex),
			Name:       fmt.Sprintf("%s_%s", deviceTypes[i%len(deviceTypes)], locations[i%len(locations)]),
			Address:    addr.Hex(),
			Type:       deviceTypes[i%len(deviceTypes)],
			PrivateKey: privateKey,
			Sequence:   0,
		}
	}
	return devices
}

// ============================================
// STATISTIQUES — CALCULS SCIENTIFIQUES
// ============================================

func stddev(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	m := mean(values)
	var sumSq float64
	for _, v := range values {
		d := v - m
		sumSq += d * d
	}
	return math.Sqrt(sumSq / float64(len(values)))
}

func ic95(values []float64) float64 {
	n := float64(len(values))
	if n < 2 {
		return 0
	}
	return 1.96 * stddev(values) / math.Sqrt(n)
}

func computePhaseResult(
	phaseNum, newDevices, totalDevices int,
	regStats, authStats []ScientificStats,
	authWindowDuration time.Duration,
) PhaseResult {
	result := PhaseResult{
		PhaseNum:     phaseNum,
		NewDevices:   newDevices,
		TotalDevices: totalDevices,
	}

	var regHTTPs, regConfirms []float64
	var regSuccess int
	for _, s := range regStats {
		if s.Registered {
			regHTTPs = append(regHTTPs, float64(s.RegHTTP_ms))
			if s.RegConfirmed {
				regConfirms = append(regConfirms, float64(s.RegConfirm_ms))
			}
			regSuccess++
		}
	}
	if len(regHTTPs) > 0 {
		result.RegHTTP_mean   = mean(regHTTPs)
		result.RegHTTP_median = median(regHTTPs)
		result.RegHTTP_p95    = percentile(regHTTPs, 95)
		result.RegHTTP_p99    = percentile(regHTTPs, 99)
		result.RegHTTP_stddev = stddev(regHTTPs)
		result.RegHTTP_ic95   = ic95(regHTTPs)
		result.RegSuccessRate = float64(regSuccess) / float64(len(regStats)) * 100
	}
	if len(regConfirms) > 0 {
		result.RegConfirm_mean   = mean(regConfirms)
		result.RegConfirm_median = median(regConfirms)
		result.RegConfirm_stddev = stddev(regConfirms)
	}

	var authECDSAs, authZKPs, authHTTPs, authApps, authConfirms, authTotals []float64
	var authSuccess int

	for _, s := range authStats {
		if s.AuthSuccess {
			authECDSAs  = append(authECDSAs, s.AuthECDSA_ms)
			authZKPs    = append(authZKPs, float64(s.AuthZKP_ms))
			authHTTPs   = append(authHTTPs, float64(s.AuthHTTP_ms))
			authApps    = append(authApps, float64(s.AuthApp_ms))
			authTotals  = append(authTotals, float64(s.AuthTotal_ms))
			if s.AuthConfirmed {
				authConfirms = append(authConfirms, float64(s.AuthConfirm_ms))
			}
			authSuccess++
		}
	}

	if len(authHTTPs) > 0 {
		result.AuthECDSA_mean   = mean(authECDSAs)
		result.AuthECDSA_stddev = stddev(authECDSAs)
		result.AuthZKP_mean     = mean(authZKPs)
		result.AuthZKP_stddev   = stddev(authZKPs)

		result.AuthHTTP_mean   = mean(authHTTPs)
		result.AuthHTTP_median = median(authHTTPs)
		result.AuthHTTP_p95    = percentile(authHTTPs, 95)
		result.AuthHTTP_stddev = stddev(authHTTPs)

		result.AuthApp_mean   = mean(authApps)
		result.AuthApp_median = median(authApps)
		result.AuthApp_p95    = percentile(authApps, 95)
		result.AuthApp_p99    = percentile(authApps, 99)
		result.AuthApp_stddev = stddev(authApps)
		result.AuthApp_ic95   = ic95(authApps)

		result.AuthTotal_mean   = mean(authTotals)
		result.AuthTotal_median = median(authTotals)
		result.AuthTotal_p95    = percentile(authTotals, 95)
		result.AuthTotal_stddev = stddev(authTotals)
		result.AuthTotal_ic95   = ic95(authTotals)

		result.AuthSuccessRate = float64(authSuccess) / float64(len(authStats)) * 100
	}

	if len(authConfirms) > 0 {
		result.AuthConfirm_mean   = mean(authConfirms)
		result.AuthConfirm_median = median(authConfirms)
		result.AuthConfirm_stddev = stddev(authConfirms)
	}

	result.AuthWindowDuration = authWindowDuration.Seconds()
	if authSuccess > 0 && authWindowDuration.Seconds() > 0 {
		result.AuthThroughput = float64(authSuccess) / authWindowDuration.Seconds()
	}

	return result
}

func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var s float64
	for _, v := range values {
		s += v
	}
	return s / float64(len(values))
}

func sum(values []float64) float64 {
	var s float64
	for _, v := range values {
		s += v
	}
	return s
}

func median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)
	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)
	index := int(float64(len(sorted)-1) * p / 100.0)
	return sorted[index]
}

// ============================================
// EXPORT DES RÉSULTATS
// ============================================

func saveResults(phaseNum, totalDevices int, regStats, authStats []ScientificStats, phaseResult PhaseResult) {
	os.MkdirAll(*outputDir, 0755)

	detailed := map[string]interface{}{
		"phase":           phaseNum,
		"total_devices":   totalDevices,
		"timestamp":       time.Now().Format(time.RFC3339),
		"registrations":   regStats,
		"authentications": authStats,
		"summary":         phaseResult,
	}
	jsonData, _ := json.MarshalIndent(detailed, "", "  ")
	jsonFile := fmt.Sprintf("%s/phase%d_%ddevices_detailed.json", *outputDir, phaseNum, totalDevices)
	os.WriteFile(jsonFile, jsonData, 0644)

	csvFile := fmt.Sprintf("%s/phase%d_%ddevices.csv", *outputDir, phaseNum, totalDevices)
	f, _ := os.Create(csvFile)
	defer f.Close()
	writer := csv.NewWriter(f)
	writer.Write([]string{
		"device_id", "phase", "type",
		"reg_http_ms", "reg_confirm_ms", "reg_final_ms", "reg_gas", "reg_block",
		"auth_ecdsa_ms", "auth_zkp_ms", "auth_http_ms", "auth_backend_ms",
		"auth_app_ms",
		"auth_confirm_ms", "auth_final_ms",
		"auth_total_ms",
		"auth_gas", "auth_block", "payload_bytes",
	})
	for _, s := range regStats {
		writer.Write([]string{
			s.DeviceID, fmt.Sprintf("%d", s.Phase), "registration",
			fmt.Sprintf("%d", s.RegHTTP_ms), fmt.Sprintf("%d", s.RegConfirm_ms),
			fmt.Sprintf("%d", s.RegFinal_ms), fmt.Sprintf("%d", s.RegGasUsed),
			fmt.Sprintf("%d", s.RegBlockNumber),
			"", "", "", "", "", "", "", "", "", "", "",
		})
	}
	for _, s := range authStats {
		writer.Write([]string{
			s.DeviceID, fmt.Sprintf("%d", s.Phase), "authentication",
			"", "", "", "", "",
			fmt.Sprintf("%.3f", s.AuthECDSA_ms),
			fmt.Sprintf("%d", s.AuthZKP_ms),
			fmt.Sprintf("%d", s.AuthHTTP_ms),
			fmt.Sprintf("%d", s.AuthBackend_ms),
			fmt.Sprintf("%d", s.AuthApp_ms),
			fmt.Sprintf("%d", s.AuthConfirm_ms),
			fmt.Sprintf("%d", s.AuthFinal_ms),
			fmt.Sprintf("%d", s.AuthTotal_ms),
			fmt.Sprintf("%d", s.AuthGasUsed),
			fmt.Sprintf("%d", s.AuthBlockNumber),
			fmt.Sprintf("%d", s.PayloadBytes),
		})
	}
	writer.Flush()

	latexFile := fmt.Sprintf("%s/phase%d_%ddevices_table.tex", *outputDir, phaseNum, totalDevices)
	os.WriteFile(latexFile, []byte(generateLaTeXTable(phaseResult)), 0644)

	log.Printf("💾 Résultats:\n   JSON: %s\n   CSV:  %s\n   LaTeX:%s",
		jsonFile, csvFile, latexFile)
}

func generateLaTeXTable(result PhaseResult) string {
	return fmt.Sprintf(`%%
%% Tableau généré par CHAIMAN IoT Benchmark v6.4.1
%% Phase %d — %d appareils
%%
\begin{table}[htbp]
\centering
\caption{Performance Results --- Phase %d (%d Devices)}
\label{tab:phase%d}
\begin{tabular}{lccccc}
\toprule
\textbf{Metric} & \textbf{Mean} & \textbf{Median} & \textbf{StdDev} & \textbf{P95} & \textbf{IC95} \\
                & \multicolumn{5}{c}{(ms)} \\
\midrule
\multicolumn{6}{l}{\textit{Registration}} \\
HTTP Latency      & %.1f & %.1f & %.1f & %.1f & ±%.1f \\
Confirm Time      & %.1f & %.1f & %.1f & ---  & ---   \\
Success Rate (\%%) & %.1f & ---  & ---  & ---  & ---   \\
\midrule
\multicolumn{6}{l}{\textit{Authentication (Applicative Layer: $T_{\mathrm{app}}$)}} \\
ECDSA Signature   & %.3f & ---  & %.3f & ---  & ---   \\
ZKP Generation    & %.1f & ---  & %.1f & ---  & ---   \\
HTTP Latency      & %.1f & %.1f & %.1f & %.1f & ---   \\
$T_{\mathrm{app}}$ (ECDSA+ZKP+HTTP) & %.1f & %.1f & %.1f & %.1f & ±%.1f \\
\midrule
\multicolumn{6}{l}{\textit{Blockchain Layer}} \\
Confirm Time      & %.1f & %.1f & %.1f & ---  & ---   \\
\midrule
\multicolumn{6}{l}{\textit{End-to-End ($T_{\mathrm{total}} = T_{\mathrm{app}} + T_{\mathrm{conf}}$)}} \\
Total Latency     & %.1f & %.1f & %.1f & %.1f & ±%.1f \\
Success Rate (\%%) & %.1f & ---  & ---  & ---  & ---   \\
Throughput (tx/s) & %.2f & ---  & ---  & ---  & ---   \\
\bottomrule
\end{tabular}
\end{table}
`,
		result.PhaseNum, result.TotalDevices, result.PhaseNum, result.TotalDevices, result.PhaseNum,
		result.RegHTTP_mean, result.RegHTTP_median, result.RegHTTP_stddev,
		result.RegHTTP_p95, result.RegHTTP_ic95,
		result.RegConfirm_mean, result.RegConfirm_median, result.RegConfirm_stddev,
		result.RegSuccessRate,
		result.AuthECDSA_mean, result.AuthECDSA_stddev,
		result.AuthZKP_mean, result.AuthZKP_stddev,
		result.AuthHTTP_mean, result.AuthHTTP_median, result.AuthHTTP_stddev, result.AuthHTTP_p95,
		result.AuthApp_mean, result.AuthApp_median, result.AuthApp_stddev,
		result.AuthApp_p95, result.AuthApp_ic95,
		result.AuthConfirm_mean, result.AuthConfirm_median, result.AuthConfirm_stddev,
		result.AuthTotal_mean, result.AuthTotal_median, result.AuthTotal_stddev,
		result.AuthTotal_p95, result.AuthTotal_ic95,
		result.AuthSuccessRate,
		result.AuthThroughput,
	)
}

// ============================================
// AFFICHAGE RÉSULTATS
// ============================================

func printScientificResults(phases []PhaseResult) {
	fmt.Println()
	fmt.Println(strings.Repeat("█", 130))
	fmt.Println("📊 RÉSULTATS SCIENTIFIQUES — MESURES BLOCKCHAIN RÉELLES v6.4.1")
	fmt.Println(strings.Repeat("█", 130))
	fmt.Println()
	fmt.Printf("%-5s | %-7s | %-10s | %-10s | %-7s | %-8s | %-10s | %-10s | %-10s | %-10s | %-10s | %-10s\n",
		"Phase", "Devices",
		"Reg HTTP", "Reg Conf",
		"ECDSA", "ZKP",
		"HTTP", "T_app", "T_app±σ", "T_conf",
		"T_total", "Tput tx/s")
	fmt.Println(strings.Repeat("─", 130))
	for _, p := range phases {
		fmt.Printf("%-5d | %-7d | %-10.1f | %-10.1f | %-7.3f | %-8.1f | %-10.1f | %-10.1f | ±%-9.1f | %-10.1f | %-10.1f | %-10.2f\n",
			p.PhaseNum, p.TotalDevices,
			p.RegHTTP_mean, p.RegConfirm_mean,
			p.AuthECDSA_mean, p.AuthZKP_mean,
			p.AuthHTTP_mean,
			p.AuthApp_mean, p.AuthApp_stddev,
			p.AuthConfirm_mean,
			p.AuthTotal_mean,
			p.AuthThroughput,
		)
	}
	fmt.Println(strings.Repeat("█", 130))
	fmt.Println()
	fmt.Println("Légende :")
	fmt.Println("  T_app    = ECDSA + ZKP + HTTP  (latence applicative pure, sans blockchain)")
	fmt.Println("  T_conf   = attente confirmation on-chain")
	fmt.Println("  T_total  = T_app + T_conf       (latence bout-en-bout)")
	fmt.Println("  Tput     = authSuccès / durée_fenêtre_wall-clock")
	fmt.Println()
	fmt.Println("ℹ️  CONSEIL BENCHMARK : lancer le serveur avec DISABLE_GETH_SCANNER=true")
	fmt.Println("   pour éviter la compétition de nonces entre scanner Geth et enregistrements IoT.")
	fmt.Println()
}

// ============================================
// MAIN
// ============================================

func main() {
	flag.Parse()
	initHTTPClient()
	initEthereumClient()
	defer ethClient.Close()

	backend := getBackendURL()

	// [FIX-SIM-22] Réinitialiser l'état backend avant le benchmark
	resetBackendState(backend)

	fmt.Println(strings.Repeat("█", 100))
	fmt.Println("🚀 CHAIMAN IoT — BENCHMARK SCIENTIFIQUE v6.4.1 (State Persistence + Backend Reset)")
	fmt.Println("   Timeouts confirmation: 60s | Délais inter-registration augmentés | Reset backend auto")
	fmt.Println(strings.Repeat("█", 100))
	fmt.Printf("   Backend:    %s\n", backend)
	fmt.Printf("   Node ETH:   %s\n", *nodeURL)
	fmt.Printf("   Output:     %s\n", *outputDir)
	fmt.Printf("   Finalité:   %v (12 confirmations)\n", *waitFinal)
	fmt.Printf("   Reset:      %v (sequences + nonces backend)\n", *resetBackend)
	fmt.Println()
	fmt.Println("⚠️  IMPORTANT: Pour un benchmark propre, lancer le backend avec:")
	fmt.Println("   DISABLE_GETH_SCANNER=true BENCHMARK_MODE=true ./iot-backend")
	fmt.Println("   Cela évite les conflits de nonces entre le scanner Geth et les devices IoT.")
	fmt.Println(strings.Repeat("█", 100))

	phases := []struct {
		name         string
		newDevices   int
		totalAfter   int
		waitAfterReg int
		authDuration time.Duration
	}{
		{"Phase 1: Initialisation", 20, 20, 30, 60 * time.Second},
		{"Phase 2: Échelle 50",     30, 50, 25, 60 * time.Second},
		{"Phase 3: Échelle 100",    50, 100, 20, 75 * time.Second},
		{"Phase 4: Échelle 150",    50, 150, 20, 75 * time.Second},
		{"Phase 5: Échelle 200",    50, 200, 20, 90 * time.Second},
	}

	var phaseResults []PhaseResult
	currentDeviceIndex := 0

	for phaseIdx, phase := range phases {
		fmt.Println()
		fmt.Println(strings.Repeat("🔷", 50))
		fmt.Printf("🔷 %s\n", phase.name)
		fmt.Printf("   ➕ %d nouveaux appareils → Total: %d\n", phase.newDevices, phase.totalAfter)
		fmt.Println(strings.Repeat("🔷", 50))

		newDevices := generateDevices(phase.newDevices, currentDeviceIndex)
		currentDeviceIndex += phase.newDevices

		// [FIX-SIM-21] Charger l'état précédent (séquences) pour éviter les collisions
		loadDeviceState(newDevices)

		fmt.Printf("\n📝 Enregistrement de %d appareils...\n", len(newDevices))
		phaseRegStats := make([]ScientificStats, 0, len(newDevices))

		for _, device := range newDevices {
			stats := device.RegisterScientific(backend, phaseIdx+1)
			phaseRegStats = append(phaseRegStats, stats)

			statsMu.Lock()
			allStats = append(allStats, stats)
			statsMu.Unlock()

			// [FIX-SIM-7/8] Délais augmentés pour laisser le backend stabiliser
			// son nonce pool après chaque transaction blockchain.
			var delay time.Duration
			if stats.RegConfirmed {
				// Transaction confirmée → délai minimal
				delay = 500 * time.Millisecond
			} else if stats.Registered && stats.RegTxHash != "" {
				// [FIX-SIM-7] Non confirmée dans 60s (mempool congestionné)
				// → 15s pour laisser le backend terminer ses retries internes
				// (backoff 4s+8s + 500ms stabilisation = ~13s minimum)
				delay = 15 * time.Second
				log.Printf("   ⏳ Délai 15s (tx non confirmée dans 60s) avant prochain device...")
			} else {
				// [FIX-SIM-8] Échec total → 20s pour reset complet du nonce pool
				delay = 20 * time.Second
				log.Printf("   ⏳ Délai 20s (enregistrement échoué) avant prochain device...")
			}
			time.Sleep(delay)
		}

		fmt.Printf("\n⏳ Stabilisation blockchain (%ds)...\n", phase.waitAfterReg)
		time.Sleep(time.Duration(phase.waitAfterReg) * time.Second)

		fmt.Printf("\n🔍 Vérification visibilité des %d appareils...\n", len(newDevices))
		registeredDevices := make([]*IoTDevice, 0, len(newDevices))
		for _, device := range newDevices {
			if device.verifyDeviceRegistered(backend) {
				registeredDevices = append(registeredDevices, device)
			} else {
				log.Printf("⚠️  %s exclu de la phase auth (non visible)", device.ID)
			}
		}
		log.Printf("✅ %d/%d appareils prêts pour l'auth",
			len(registeredDevices), len(newDevices))

		allDevicesList := append([]*IoTDevice{}, allDevices...)
		allDevicesList = append(allDevicesList, registeredDevices...)
		allDevices = allDevicesList

		fmt.Printf("\n🔐 Authentification de %d appareils pendant %v...\n",
			len(allDevicesList), phase.authDuration)

		stopChan := make(chan struct{})
		var wg sync.WaitGroup
		phaseAuthStats := make([]ScientificStats, 0)
		var authMu sync.Mutex

		authWindowStart := time.Now()

		// [FIX-SIM-13] Étalement temporel : interDeviceDelay = authInterval / N
		// Chaque device démarre à i * interDeviceDelay pour éviter la rafale initiale.
		nDevices := len(allDevicesList)
		var interDeviceDelay time.Duration
		if nDevices > 1 {
			interDeviceDelay = *authInterval / time.Duration(nDevices)
		}
		if interDeviceDelay > 2*time.Second {
			interDeviceDelay = 2 * time.Second
		}

		for i, device := range allDevicesList {
			wg.Add(1)
			go func(d *IoTDevice, deviceIdx int) {
				defer wg.Done()

				// Délai de démarrage déterministe + petit jitter pour éviter les collisions exactes.
				startDelay := time.Duration(deviceIdx) * interDeviceDelay
				if interDeviceDelay > 0 {
					jitter := time.Duration(rand.Int63n(int64(interDeviceDelay)/10 + 1))
					startDelay += jitter
				}
				select {
				case <-stopChan:
					return
				case <-time.After(startDelay):
				}

				ticker := time.NewTicker(*authInterval)
				defer ticker.Stop()
				timeout := time.After(phase.authDuration)

				// Première auth immédiate après le délai de démarrage.
				stats := d.AuthenticateScientific(backend, phaseIdx+1)
				authMu.Lock()
				phaseAuthStats = append(phaseAuthStats, stats)
				authMu.Unlock()
				statsMu.Lock()
				allStats = append(allStats, stats)
				statsMu.Unlock()

				for {
					select {
					case <-timeout:
						return
					case <-stopChan:
						return
					case <-ticker.C:
						stats := d.AuthenticateScientific(backend, phaseIdx+1)
						authMu.Lock()
						phaseAuthStats = append(phaseAuthStats, stats)
						authMu.Unlock()
						statsMu.Lock()
						allStats = append(allStats, stats)
						statsMu.Unlock()
					}
				}
			}(device, i)
		}

		// [FIX-SIM-17] Attendre la fin de TOUTES les goroutines d'auth avant
		// de calculer les stats. Sans wg.Wait(), phaseAuthStats est vide au
		// moment de computePhaseResult → Success rate = 0%, throughput = 0.
		// On mesure authWindowDuration APRÈS wg.Wait() pour avoir la vraie
		// durée de la fenêtre (pas 0.1s comme avant).
		wg.Wait()
		close(stopChan)

		authWindowDuration := time.Since(authWindowStart)

		authMu.Lock()
		finalAuthStats := make([]ScientificStats, len(phaseAuthStats))
		copy(finalAuthStats, phaseAuthStats)
		authMu.Unlock()

		result := computePhaseResult(
			phaseIdx+1, phase.newDevices, phase.totalAfter,
			phaseRegStats, finalAuthStats,
			authWindowDuration,
		)
		phaseResults = append(phaseResults, result)

		fmt.Printf("\n📊 RÉSULTATS PHASE %d (%d devices | fenêtre: %.1fs):\n",
			phaseIdx+1, phase.totalAfter, authWindowDuration.Seconds())
		fmt.Printf("   📝 Enregistrement   : HTTP=%.1fms±%.1f Confirm=%.1fms Success=%.1f%%\n",
			result.RegHTTP_mean, result.RegHTTP_stddev,
			result.RegConfirm_mean, result.RegSuccessRate)
		fmt.Printf("   🔐 Auth applicative : ECDSA=%.3f±%.3fms ZKP=%.1f±%.1fms HTTP=%.1fms\n",
			result.AuthECDSA_mean, result.AuthECDSA_stddev,
			result.AuthZKP_mean, result.AuthZKP_stddev,
			result.AuthHTTP_mean)
		fmt.Printf("   📡 T_app (pure)     : mean=%.1f median=%.1f p95=%.1f σ=%.1f IC95=±%.1f ms\n",
			result.AuthApp_mean, result.AuthApp_median,
			result.AuthApp_p95, result.AuthApp_stddev, result.AuthApp_ic95)
		fmt.Printf("   ⛓️  Blockchain       : Confirm=%.1f±%.1fms\n",
			result.AuthConfirm_mean, result.AuthConfirm_stddev)
		fmt.Printf("   🏁 T_total (E2E)    : mean=%.1f median=%.1f p95=%.1f σ=%.1f IC95=±%.1f ms\n",
			result.AuthTotal_mean, result.AuthTotal_median,
			result.AuthTotal_p95, result.AuthTotal_stddev, result.AuthTotal_ic95)
		fmt.Printf("   📈 Throughput       : %.2f tx/s (wall-clock)\n", result.AuthThroughput)
		fmt.Printf("   ✅ Success rate     : %.1f%%\n", result.AuthSuccessRate)

		saveResults(phaseIdx+1, phase.totalAfter, phaseRegStats, finalAuthStats, result)

		if phaseIdx < len(phases)-1 {
			fmt.Printf("\n⏸️  Pause de 10s avant phase suivante...\n")
			time.Sleep(10 * time.Second)
		}
	}

	// [FIX-SIM-21] Sauvegarder l'état final pour les prochaines exécutions
	saveDeviceState(allDevices)

	printScientificResults(phaseResults)
	exportFullScientificReport(phaseResults, allStats)

	// [FIX-SIM-23] Version corrigée v6.3 → v6.4.1
	fmt.Println("\n✅ BENCHMARK SCIENTIFIQUE v6.4.1 TERMINÉ!")
	fmt.Printf("   Enregistrements: %d (confirmés: %d)\n",
		atomic.LoadInt64(&totalRegAttempts), atomic.LoadInt64(&totalRegConfirmed))
	fmt.Printf("   Authentifications: %d (confirmées: %d)\n",
		atomic.LoadInt64(&totalAuthAttempts), atomic.LoadInt64(&totalAuthConfirmed))
}

func exportFullScientificReport(phases []PhaseResult, allStats []ScientificStats) {
	report := map[string]interface{}{
		"test_metadata": map[string]interface{}{
			"version":       "6.4.1-stable",
			"date":          time.Now().Format(time.RFC3339),
			"backend_url":   getBackendURL(),
			"node_url":      *nodeURL,
			"finality_wait": *waitFinal,
			"fixes_v6_4_1": []string{
				"[FIX-SIM-21] Persistance de l'état simulateur (séquences) — évite 'Séquence invalide' aux relances",
				"[FIX-SIM-22] Reset automatique de l'état backend (sequences + nonces) au démarrage",
				"[FIX-SIM-23] Version affichée corrigée v6.3 → v6.4.1-stable",
			},
			"fixes_v6_4": []string{
				"[FIX-SIM-17] wg.Wait() ajouté avant computePhaseResult — stats ne sont plus toutes à zéro",
				"[FIX-SIM-18] stopChan fermé après wg.Wait() — pas de goroutines orphelines",
				"[FIX-SIM-19] finalAuthStats copie thread-safe pour saveResults",
				"[FIX-SIM-20] Version 6.4.0-stable",
			},
			"fixes_v6_3": []string{
				"[FIX-SIM-13] Étalement temporel des auth via interDeviceDelay = authInterval/N",
				"[FIX-SIM-14] Flag -auth-interval configurable",
				"[FIX-SIM-15] Retry ZKP respecte le header Retry-After du backend",
				"[FIX-SIM-16] Version 6.3.0-stable",
			},
			"fixes_v6_2": []string{
				"[FIX-SIM-11] Cause racine nonce gap corrigée dans backend device.go v5.3 et device_ops.go v5.3",
				"[FIX-SIM-12] Version mise à jour 6.2.0-stable",
				"[FIX-BACKEND-13] ActivateDevice acquiert TxSemaphore + ReleaseNonceSync (device.go v5.3)",
				"[FIX-BACKEND-14] DeactivateDevice acquiert TxSemaphore + ReleaseNonceSync (device.go v5.3)",
				"[FIX-BACKEND-15] DeviceRegisterWithType : ReleaseNonceSync pour nonce too low (device_ops.go v5.3)",
			},
			"fixes_v6_1": []string{
				"[FIX-SIM-6] Registration confirmation timeout 30s → 60s",
				"[FIX-SIM-7] Delay after unconfirmed registration 5s → 15s",
				"[FIX-SIM-8] Delay after failed registration 6s → 20s",
				"[FIX-SIM-9] Auth confirmation timeout 30s → 60s",
				"[FIX-SIM-10] DISABLE_GETH_SCANNER env var documented",
			},
		},
		"phases":           phases,
		"all_measurements": allStats,
	}

	jsonData, _ := json.MarshalIndent(report, "", "  ")
	path := fmt.Sprintf("%s/full_scientific_report_v6_4_1.json", *outputDir)
	os.WriteFile(path, jsonData, 0644)
	fmt.Printf("\n📄 Rapport scientifique complet: %s\n", path)
}
