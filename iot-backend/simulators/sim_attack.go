// iot_anomaly_simulator.go
// CHAIMAN IoT — Simulateur de Trafic ANORMAL
//
// Ce simulateur crée 3 devices IoT qui envoient des données réseau
// ANORMALES conçues pour déclencher le moteur ML de détection d'anomalies.
//
// Chaque device simule un profil d'attaque différent :
//
//   ANOMALY_001 — "DDoS / Flood"
//     → Volume massif de paquets, bytes_total extrême, burst_indicator max,
//       high_volume_flag = 1.0. Cible severity CRITICAL (score > 90).
//
//   ANOMALY_002 — "Data Exfiltration"
//     → bytes_sent >> bytes_received (bytes_sent_ratio → 1.0),
//       bytes_ratio déséquilibré, durée longue, gros paquets irréguliers.
//       Cible severity HIGH (score > 70).
//
//   ANOMALY_003 — "Port Scan / Reconnaissance"
//     → Nombreux paquets très courts (bytes_per_pkt faible),
//       haute variance de taille, pkts_per_sec élevé mais bytes_total faible.
//       Cible severity MEDIUM (score > 40).
//
// Usage:
//   go run iot_anomaly_simulator.go -backend=http://localhost:8080
//
// Variables d'environnement:
//   BACKEND_URL        : URL du backend (défaut: http://localhost:8080)
//   TLS_INSECURE_DEV   : "true" pour désactiver la vérification TLS
//   BACKEND_CA_CERT    : chemin vers le certificat CA du backend
//   CLIENT_CERT_PATH   : chemin vers le certificat client mTLS
//   CLIENT_KEY_PATH    : chemin vers la clé privée client mTLS
//   ANOMALY_INTERVAL   : intervalle entre envois (défaut: 5s)

package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
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
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

// ── Flags ──────────────────────────────────────────────────────────────────────

var (
	backendFlag  = flag.String("backend", "", "URL du backend IoT")
	intervalFlag = flag.Duration("interval", 0, "Intervalle fixe entre envois (ex: 5s)")
)

// ── AnomalyProfile — profil de trafic anormal ─────────────────────────────────

// AnomalyProfile décrit comment générer des features ML anormales.
type AnomalyProfile struct {
	Name        string
	Description string
	// Chaque champ est une fonction qui retourne une valeur anormale
	// (avec une légère variation aléatoire pour simuler de vraies données).
	genFeatures func(rng *rand.Rand) mlFeatures
}

// mlFeatures contient les 14 features attendues par le service ML.
type mlFeatures struct {
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

// ── 3 profils d'anomalie ──────────────────────────────────────────────────────

var anomalyProfiles = []AnomalyProfile{
	{
		Name:        "DDoS_FLOOD",
		Description: "Inondation massive — volume extrême, burst maximal → CRITICAL",
		genFeatures: func(rng *rand.Rand) mlFeatures {
			// Volume extrême : des centaines de milliers de paquets à haute fréquence
			pktCount := 50000.0 + rng.Float64()*30000.0   // 50k–80k paquets
			duration := 0.5 + rng.Float64()*0.5            // 0.5–1s (très court = flood)
			bytesSent := pktCount * (64.0 + rng.Float64()*32.0)
			bytesRcv := 512.0 + rng.Float64()*256.0        // presque rien en retour
			bytesTotal := bytesSent + bytesRcv
			return mlFeatures{
				Duration:        duration,
				BytesSent:       bytesSent,
				BytesReceived:   bytesRcv,
				PacketCount:     pktCount,
				BytesPerSec:     bytesTotal / duration,
				PktsPerSec:      pktCount / duration,
				BytesPerPkt:     bytesSent / pktCount,
				BytesRatio:      clamp01(bytesRcv / bytesTotal),
				BytesSentRatio:  clamp01(bytesSent / bytesTotal),
				BytesTotal:      bytesTotal,
				LogBytesTotal:   math.Log1p(bytesTotal),
				PktSizeVariance: 8.0 + rng.Float64()*16.0,  // variance faible (paquets uniformes)
				HighVolumeFlag:  1.0,                         // flag activé
				BurstIndicator:  95.0 + rng.Float64()*5.0,   // quasi-maximal (95–100)
			}
		},
	},
	{
		Name:        "DATA_EXFIL",
		Description: "Exfiltration de données — envoi massif asymétrique → HIGH",
		genFeatures: func(rng *rand.Rand) mlFeatures {
			// Envoie beaucoup, reçoit très peu : ratio très déséquilibré
			duration := 30.0 + rng.Float64()*60.0           // 30–90s (transfert long)
			bytesSent := 5e8 + rng.Float64()*5e8             // 500MB–1GB envoyés
			bytesRcv := 1024.0 + rng.Float64()*4096.0        // 1–5 KB reçus (ACKs seulement)
			pktCount := 350000.0 + rng.Float64()*150000.0
			bytesTotal := bytesSent + bytesRcv
			return mlFeatures{
				Duration:        duration,
				BytesSent:       bytesSent,
				BytesReceived:   bytesRcv,
				PacketCount:     pktCount,
				BytesPerSec:     bytesTotal / duration,
				PktsPerSec:      pktCount / duration,
				BytesPerPkt:     bytesSent / pktCount,
				BytesRatio:      clamp01(bytesRcv / bytesTotal),  // ≈ 0 (quasi rien reçu)
				BytesSentRatio:  clamp01(bytesSent / bytesTotal), // ≈ 1.0
				BytesTotal:      bytesTotal,
				LogBytesTotal:   math.Log1p(bytesTotal),
				PktSizeVariance: 250.0 + rng.Float64()*500.0,     // variance élevée (gros paquets irréguliers)
				HighVolumeFlag:  1.0,
				BurstIndicator:  75.0 + rng.Float64()*20.0,       // 75–95
			}
		},
	},
	{
		Name:        "PORT_SCAN",
		Description: "Scan de ports / reconnaissance — nombreux micro-paquets → MEDIUM",
		genFeatures: func(rng *rand.Rand) mlFeatures {
			// Beaucoup de paquets très courts vers des destinations variées
			pktCount := 8000.0 + rng.Float64()*4000.0       // 8k–12k paquets
			duration := 2.0 + rng.Float64()*3.0              // 2–5s
			bytesSent := pktCount * (40.0 + rng.Float64()*20.0) // ~40–60 bytes/pkt (SYN only)
			bytesRcv := pktCount * (20.0 + rng.Float64()*20.0)  // RST/ACK courts
			bytesTotal := bytesSent + bytesRcv
			return mlFeatures{
				Duration:        duration,
				BytesSent:       bytesSent,
				BytesReceived:   bytesRcv,
				PacketCount:     pktCount,
				BytesPerSec:     bytesTotal / duration,
				PktsPerSec:      pktCount / duration,            // très élevé
				BytesPerPkt:     bytesSent / pktCount,           // très faible (40–60 bytes)
				BytesRatio:      clamp01(bytesRcv / bytesTotal),
				BytesSentRatio:  clamp01(bytesSent / bytesTotal),
				BytesTotal:      bytesTotal,
				LogBytesTotal:   math.Log1p(bytesTotal),
				PktSizeVariance: 1800.0 + rng.Float64()*1200.0, // variance très élevée
				HighVolumeFlag:  0.0,                            // volume total faible
				BurstIndicator:  45.0 + rng.Float64()*30.0,     // 45–75
			}
		},
	},
}

// ── Device ─────────────────────────────────────────────────────────────────────

// AnomalyDevice est un device IoT qui envoie des données anormales.
type AnomalyDevice struct {
	ID         string
	Name       string
	Type       string
	Address    string
	PrivateKey *ecdsa.PrivateKey
	Profile    AnomalyProfile
	rng        *rand.Rand
	sequence   uint64
	mu         sync.Mutex
}

func (d *AnomalyDevice) nextSeq() uint64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.sequence++
	return d.sequence
}

// ── HTTP client ────────────────────────────────────────────────────────────────

var httpClient *http.Client

func buildHTTPClient() {
	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12}

	if os.Getenv("TLS_INSECURE_DEV") == "true" {
		tlsCfg.InsecureSkipVerify = true //nolint:gosec
		log.Println("⚠️  TLS_INSECURE_DEV=true — vérification TLS désactivée")
	}

	if ca := os.Getenv("BACKEND_CA_CERT"); ca != "" {
		pem, err := os.ReadFile(ca)
		if err != nil {
			log.Fatalf("BACKEND_CA_CERT introuvable : %v", err)
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(pem)
		tlsCfg.RootCAs = pool
	}

	if cert, key := os.Getenv("CLIENT_CERT_PATH"), os.Getenv("CLIENT_KEY_PATH"); cert != "" && key != "" {
		pair, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			log.Fatalf("Certificat mTLS invalide : %v", err)
		}
		tlsCfg.Certificates = []tls.Certificate{pair}
	}

	httpClient = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:     tlsCfg,
			MaxIdleConnsPerHost: 10,
		},
	}
}

// ── Helpers ────────────────────────────────────────────────────────────────────

func post(url string, payload interface{}) (map[string]interface{}, int, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}
	resp, err := httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	var result map[string]interface{}
	json.Unmarshal(raw, &result) //nolint:errcheck
	return result, resp.StatusCode, nil
}

func getJSON(url string) (map[string]interface{}, int, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var result map[string]interface{}
	json.Unmarshal(raw, &result) //nolint:errcheck
	return result, resp.StatusCode, nil
}

// clamp01 borne une valeur entre 0 et 1 (pour les ratios).
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// short tronque une string pour les logs.
func short(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// ── Enregistrement ─────────────────────────────────────────────────────────────

func (d *AnomalyDevice) register(backend string) bool {
	pubKeyHex := hex.EncodeToString(crypto.FromECDSAPub(&d.PrivateKey.PublicKey))
	metadata := fmt.Sprintf(
		`{"device_id":%q,"type":%q,"name":%q,"anomaly_profile":%q,"warning":"ANOMALY_SIMULATOR"}`,
		d.ID, d.Type, d.Name, d.Profile.Name,
	)

	result, code, err := post(backend+"/api/device/register", map[string]interface{}{
		"address":    d.Address,
		"entityType": "iot_device",
		"publicKey":  pubKeyHex,
		"metadata":   metadata,
	})
	if err != nil {
		log.Printf("[%s] ❌ Enregistrement réseau : %v", d.ID, err)
		return false
	}

	switch code {
	case 200:
		txHash, _ := result["txHash"].(string)
		log.Printf("[%s] ✅ Enregistré | tx: %s", d.ID, short(txHash, 18))
		return true
	case 409:
		log.Printf("[%s] ℹ️  Déjà enregistré", d.ID)
		return true
	default:
		msg, _ := result["message"].(string)
		log.Printf("[%s] ⚠️  Enregistrement HTTP %d : %s", d.ID, code, msg)
		return false
	}
}

// ── Activation ─────────────────────────────────────────────────────────────────

func (d *AnomalyDevice) activate(backend string) {
	result, code, err := post(backend+"/api/device/activate", map[string]interface{}{
		"deviceAddress": d.Address,
	})
	if err != nil {
		log.Printf("[%s] ⚠️  Activation réseau : %v", d.ID, err)
		return
	}
	if code == 200 {
		msg, _ := result["message"].(string)
		log.Printf("[%s] 🔓 %s", d.ID, msg)
	} else {
		log.Printf("[%s] ⚠️  Activation HTTP %d", d.ID, code)
	}
}

// ── Vérification état ──────────────────────────────────────────────────────────

func (d *AnomalyDevice) waitUntilActive(backend string) bool {
	for attempt := 1; attempt <= 10; attempt++ {
		result, code, err := getJSON(fmt.Sprintf("%s/api/device/%s/status", backend, d.Address))
		if err == nil && code == 200 {
			if active, ok := result["isActive"].(bool); ok && active {
				log.Printf("[%s] ✅ Actif sur le backend", d.ID)
				return true
			}
			log.Printf("[%s] ⏳ Inactif — tentative activation (%d/10)", d.ID, attempt)
			d.activate(backend)
		} else {
			log.Printf("[%s] ⏳ En attente d'indexation (%d/10)…", d.ID, attempt)
		}
		time.Sleep(3 * time.Second)
	}
	log.Printf("[%s] ❌ Toujours inactif après 10 tentatives", d.ID)
	return false
}

// ── Génération preuve ZKP ──────────────────────────────────────────────────────

func (d *AnomalyDevice) getZKPProof(backend, msgHashHex, sigHex string, timestamp int64, seq uint64, nonce string) (proof, expectedHash string, err error) {
	result, code, e := post(backend+"/api/zkp/generate-secure", map[string]interface{}{
		"signature":    sigHex,
		"message_hash": msgHashHex,
		"address":      d.Address,
		"sequence":     seq,
		"timestamp":    timestamp,
		"nonce":        nonce,
	})
	if e != nil {
		return "", "", e
	}
	if code != 200 {
		msg, _ := result["message"].(string)
		return "", "", fmt.Errorf("ZKP HTTP %d : %s", code, msg)
	}
	proof, _ = result["proof"].(string)
	expectedHash, _ = result["expected_hash"].(string)
	return proof, expectedHash, nil
}

// ── Envoi d'un message anormal ─────────────────────────────────────────────────

func (d *AnomalyDevice) sendAnomaly(backend string) {
	tStart := time.Now()

	timestamp := time.Now().Unix()
	seq := d.nextSeq()
	nonce := fmt.Sprintf("%s-%d-%d", d.ID, timestamp, seq)

	// Générer les features anormales selon le profil
	feat := d.Profile.genFeatures(d.rng)

	// Message IoT avec marqueur d'anomalie explicite
	data := map[string]interface{}{
		"device_id":       d.ID,
		"type":            d.Type,
		"timestamp":       timestamp,
		"anomaly_profile": d.Profile.Name,
		"alert":           fmt.Sprintf("⚠️ ANOMALY: %s", d.Profile.Description),
		// Quelques valeurs capteurs cohérentes avec l'attaque
		"bytes_sent":    fmt.Sprintf("%.0f", feat.BytesSent),
		"packet_count":  fmt.Sprintf("%.0f", feat.PacketCount),
		"burst_level":   fmt.Sprintf("%.1f%%", feat.BurstIndicator),
	}
	message, _ := json.Marshal(data)
	msgStr := string(message)

	// Signature ECDSA
	tECDSA := time.Now()
	msgHash := crypto.Keccak256Hash([]byte(msgStr))
	sig, err := crypto.Sign(msgHash.Bytes(), d.PrivateKey)
	ecdsaMs := time.Since(tECDSA).Milliseconds()
	if err != nil {
		log.Printf("[%s] ❌ ECDSA : %v", d.ID, err)
		return
	}
	msgHashHex := hex.EncodeToString(msgHash.Bytes())
	sigHex := hex.EncodeToString(sig)

	// Preuve ZKP
	tZKP := time.Now()
	proof, expectedHash, err := d.getZKPProof(backend, msgHashHex, sigHex, timestamp, seq, nonce)
	zkpMs := time.Since(tZKP).Milliseconds()
	if err != nil {
		log.Printf("[%s] ❌ ZKP : %v", d.ID, err)
		return
	}

	// Envoi au backend avec les features ML anormales
	tHTTP := time.Now()
	result, code, err := post(backend+"/api/node/message-secure", map[string]interface{}{
		"device_id":     d.ID,
		"address":       d.Address,
		"message":       msgStr,
		"message_hash":  msgHashHex,
		"signature":     sigHex,
		"proof":         proof,
		"expected_hash": expectedHash,
		"timestamp":     timestamp,
		"sequence":      seq,
		"nonce":         nonce,
		// ── Features ML ANORMALES ──────────────────────────────
		"duration":          feat.Duration,
		"bytes_sent":        feat.BytesSent,
		"bytes_received":    feat.BytesReceived,
		"packet_count":      feat.PacketCount,
		"bytes_per_sec":     feat.BytesPerSec,
		"pkts_per_sec":      feat.PktsPerSec,
		"bytes_per_pkt":     feat.BytesPerPkt,
		"bytes_ratio":       feat.BytesRatio,
		"bytes_sent_ratio":  feat.BytesSentRatio,
		"bytes_total":       feat.BytesTotal,
		"log_bytes_total":   feat.LogBytesTotal,
		"pkt_size_variance": feat.PktSizeVariance,
		"high_volume_flag":  feat.HighVolumeFlag,
		"burst_indicator":   feat.BurstIndicator,
	})
	httpMs := time.Since(tHTTP).Milliseconds()
	totalMs := time.Since(tStart).Milliseconds()

	if err != nil {
		log.Printf("[%s] ❌ Envoi réseau : %v", d.ID, err)
		return
	}

	if code == 200 {
		txHash, _ := result["txHash"].(string)
		log.Printf("[%s] 🚨 ANOMALIE envoyée | profil=%-12s | pkts=%.0f bytes_sent=%.0f burst=%.0f%% | ECDSA:%dms ZKP:%dms HTTP:%dms Total:%dms | tx:%s",
			d.ID, d.Profile.Name,
			feat.PacketCount, feat.BytesSent, feat.BurstIndicator,
			ecdsaMs, zkpMs, httpMs, totalMs,
			short(txHash, 16),
		)
	} else {
		msg, _ := result["message"].(string)
		if msg == "" {
			msg, _ = result["error"].(string)
		}
		log.Printf("[%s] ⚠️  Auth HTTP %d : %s", d.ID, code, msg)
	}
}

// ── Boucle principale du device ────────────────────────────────────────────────

func (d *AnomalyDevice) run(backend string, interval time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("[%s] 🔴 Démarrage simulateur anomalie", d.ID)
	log.Printf("[%s]    Profil   : %s", d.ID, d.Profile.Name)
	log.Printf("[%s]    Desc     : %s", d.ID, d.Profile.Description)
	log.Printf("[%s]    Adresse  : %s", d.ID, d.Address)

	// Enregistrement
	if !d.register(backend) {
		log.Printf("[%s] ❌ Impossible de s'enregistrer — arrêt", d.ID)
		return
	}

	// Attente d'activation
	time.Sleep(2 * time.Second)
	if !d.waitUntilActive(backend) {
		log.Printf("[%s] ❌ Device inactif — arrêt", d.ID)
		return
	}

	log.Printf("[%s] 📡 Début envoi de données ANORMALES (intervalle: %v)", d.ID, interval)

	for {
		d.sendAnomaly(backend)

		// Légère variation autour de l'intervalle de base (±20%) pour réalisme
		jitter := time.Duration(float64(interval) * (0.8 + d.rng.Float64()*0.4))
		log.Printf("[%s] ⏳ Prochain envoi dans %v", d.ID, jitter.Round(time.Millisecond))
		time.Sleep(jitter)
	}
}

// ── Main ───────────────────────────────────────────────────────────────────────

func main() {
	flag.Parse()
	//nolint:staticcheck
	rand.Seed(time.Now().UnixNano())

	buildHTTPClient()

	backend := *backendFlag
	if backend == "" {
		backend = getEnv("BACKEND_URL", "http://localhost:8080")
	}

	// Intervalle entre envois
	interval := *intervalFlag
	if interval == 0 {
		if v := os.Getenv("ANOMALY_INTERVAL"); v != "" {
			if d, err := time.ParseDuration(v); err == nil {
				interval = d
			}
		}
	}
	if interval == 0 {
		interval = 5 * time.Second // défaut : 5s
	}

	log.Printf("╔══════════════════════════════════════════════════╗")
	log.Printf("║  CHAIMAN IoT — Simulateur de Trafic ANORMAL      ║")
	log.Printf("╚══════════════════════════════════════════════════╝")
	log.Printf("  Backend  : %s", backend)
	log.Printf("  Interval : %v (± 20%% jitter)", interval)
	log.Printf("")
	log.Printf("  Profils d'anomalie :")
	log.Printf("    ANOMALY_001 → DDoS / Flood         → CRITICAL")
	log.Printf("    ANOMALY_002 → Data Exfiltration     → HIGH")
	log.Printf("    ANOMALY_003 → Port Scan / Recon     → MEDIUM")
	log.Printf("")

	// Définitions des 3 devices anormaux (un par profil)
	deviceDefs := []struct {
		id, name, devType string
		profileIdx        int
	}{
		{"ANOMALY_001", "ddos_flood_node",        "attack_node",   0},
		{"ANOMALY_002", "exfiltration_node",      "attack_node",   1},
		{"ANOMALY_003", "port_scan_node",         "attack_node",   2},
	}

	devices := make([]*AnomalyDevice, 0, len(deviceDefs))
	for i, def := range deviceDefs {
		pk, err := crypto.GenerateKey()
		if err != nil {
			log.Fatalf("Génération clé ECDSA [%s]: %v", def.id, err)
		}
		addr := crypto.PubkeyToAddress(pk.PublicKey)
		devices = append(devices, &AnomalyDevice{
			ID:         def.id,
			Name:       def.name,
			Type:       def.devType,
			Address:    addr.Hex(),
			PrivateKey: pk,
			Profile:    anomalyProfiles[def.profileIdx],
			rng:        rand.New(rand.NewSource(time.Now().UnixNano() + int64(i*1000))),
			sequence:   uint64(time.Now().Unix()),
		})
	}

	log.Printf("  Devices créés :")
	for _, d := range devices {
		log.Printf("    • %-12s | %-20s → %s", d.ID, d.Profile.Name, d.Address)
	}
	log.Printf("")
	log.Printf("  ⚠️  ATTENTION : ces devices envoient du trafic anormal.")
	log.Printf("     Le service ML devrait déclencher des alertes après")
	log.Printf("     accumulation de messages dans son buffer.")
	log.Printf("")

	// Lancer chaque device avec un décalage pour éviter les collisions de nonce
	var wg sync.WaitGroup
	for i, d := range devices {
		wg.Add(1)
		go func(dev *AnomalyDevice, delay time.Duration) {
			time.Sleep(delay)
			dev.run(backend, interval, &wg)
		}(d, time.Duration(i)*3*time.Second)
	}

	wg.Wait()
	log.Println("Tous les devices anormaux se sont arrêtés.")
}
