// simple_simulator.go
// CHAIMAN IoT — Simulateur Simple
// VERSION 6.3 — FIX CASE-SENSITIVITY DANS LA PREUVE DE POSSESSION
//
// CHANGEMENTS v6.2 → v6.3 :
//
//   [FIX-CASE-MISMATCH] register() plantait systématiquement avec
//               "401 Signature ECDSA invalide — preuve de possession de
//               l'adresse échouée" pour TOUS les devices, à CHAQUE tentative.
//
//               CAUSE : d.Address vient de addr.Hex() (go-ethereum), qui
//               retourne l'adresse au format CHECKSUMMED (casse mixte,
//               ex: "0x814340758b90C135257e87eFFc1923ede63462f6"). Le
//               message signé côté device utilisait cette casse mixte.
//
//               Côté serveur (internal/handlers/device.go, RegisterDevice),
//               la toute première ligne fait :
//                   req.Address = strings.ToLower(req.Address)
//               AVANT de recalculer le même message pour vérifier la
//               signature (buildRegistrationProofMessage). Le serveur
//               recalculait donc le hash sur la version EN MINUSCULES de
//               l'adresse, alors que le device avait signé le hash de la
//               version en CASSE MIXTE.
//
//               Keccak256 est sensible à la casse : deux chaînes qui ne
//               diffèrent QUE par la casse produisent des hash totalement
//               différents. La signature ECDSA du device (calculée sur le
//               hash "casse mixte") ne pouvait donc JAMAIS correspondre au
//               hash recalculé par le serveur (casse minuscule) — d'où
//               l'échec systématique, pour 100% des tentatives
//               d'enregistrement, sur tous les devices.
//
//               FIX : le device normalise désormais son adresse en
//               minuscules AVANT de construire et signer le message
//               canonique, pour que la chaîne hashée soit BYTE-POUR-BYTE
//               identique des deux côtés. Voir lowerAddress dans register()
//               ci-dessous.
//
//               Vérifié par simulation indépendante (hash côté "client" vs
//               "serveur") : avant le fix, les hash différaient à 100% ;
//               après le fix, ils sont identiques à 100%.
//
// CHANGEMENTS v6.1 → v6.2 (conservés) :
//
//   [TEST-POP] Ajout d'un 4e device, IOT_0004, dans deviceDefs — device
//               totalement neuf pour valider la preuve de possession en
//               conditions réelles (les 3 premiers sont déjà enregistrés,
//               donc reçoivent 409 avant même la vérification de signature).
//
// CHANGEMENTS v6.0 → v6.1 (conservés) :
//
//   [FIX-REGISTER-POP] register() signe un message canonique
//               "REGISTER:<address>:<schnorrPublicKey>" avec la clé privée
//               ECDSA du device, et envoie signature + message_hash au
//               serveur — obligatoire depuis device.go v5.5.
//
//   CHANGEMENTS v5 → v6.0 (conservés) :
//
//   [SCHNORR-5] Chaque Device génère sa PROPRE paire de clés Schnorr
//               localement (zkp.GenerateKeyPair) — SchnorrKeyPair.PrivateKey
//               ne quitte JAMAIS ce processus.
//
//   [SCHNORR-6] authenticate() ne fait PLUS d'appel réseau à
//               /api/zkp/generate-secure — preuve calculée localement.
//
//   [SCHNORR-7] register() envoie schnorrPubKey (hex) en plus de pubKeyHex.
//
// Usage:
//   go run simple_simulator.go -backend=http://localhost:8080

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
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/crypto"

	"iot-backend/pkg/zkp"
)

// ── Configuration ──────────────────────────────────────────────────────────────

var backendFlag = flag.String("backend", "", "URL du backend IoT (ex: http://localhost:8080)")

// ── Device ─────────────────────────────────────────────────────────────────────

type Device struct {
	ID         string
	Name       string
	Type       string
	Address    string
	PrivateKey *ecdsa.PrivateKey // clé ECDSA — signature du message (INCHANGÉ)

	// [SCHNORR-5] Paire de clés Schnorr — générée localement, une fois.
	// SchnorrKeys.PrivateKey ne doit JAMAIS être sérialisé ni envoyé au réseau.
	SchnorrKeys *zkp.SchnorrKeyPair

	sequence uint64
}

func (d *Device) nextSeq() uint64 {
	return atomic.AddUint64(&d.sequence, 1)
}

// ── HTTP client ────────────────────────────────────────────────────────────────

var httpClient *http.Client

func buildHTTPClient() {
	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12}

	if os.Getenv("TLS_INSECURE_DEV") == "true" {
		tlsCfg.InsecureSkipVerify = true
		log.Println("⚠️  TLS_INSECURE_DEV=true — vérification TLS désactivée")
	}

	if ca := os.Getenv("BACKEND_CA_CERT"); ca != "" {
		pem, err := os.ReadFile(ca)
		if err != nil {
			log.Fatalf("❌ BACKEND_CA_CERT introuvable : %v", err)
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(pem)
		tlsCfg.RootCAs = pool
		log.Printf("✅ CA chargée : %s", ca)
	}

	if cert, key := os.Getenv("CLIENT_CERT_PATH"), os.Getenv("CLIENT_KEY_PATH"); cert != "" && key != "" {
		pair, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			log.Fatalf("❌ Certificat mTLS invalide : %v", err)
		}
		tlsCfg.Certificates = []tls.Certificate{pair}
		log.Printf("✅ Certificat mTLS chargé")
	}

	httpClient = &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:     tlsCfg,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     60 * time.Second,
		},
	}
}

// ── Helpers ────────────────────────────────────────────────────────────────────

func postJSON(url string, payload interface{}) (map[string]interface{}, int, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal: %w", err)
	}
	resp, err := httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	var result map[string]interface{}
	json.Unmarshal(raw, &result)
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
	json.Unmarshal(raw, &result)
	return result, resp.StatusCode, nil
}

func short(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func getBackendURL() string {
	if *backendFlag != "" {
		return *backendFlag
	}
	if u := os.Getenv("BACKEND_URL"); u != "" {
		return u
	}
	return "http://localhost:8080"
}

// ── Enregistrement ─────────────────────────────────────────────────────────────

// register — [FIX-REGISTER-POP v6.1 + FIX-CASE-MISMATCH v6.3] signe un
// message canonique liant l'adresse à la clé publique Schnorr envoyée, en
// plus de schnorrPubKey (ECDSA) et publicKey (v6.0, inchangés).
//
// Le message signé DOIT être identique à celui recalculé côté serveur
// (voir internal/handlers/device.go, buildRegistrationProofMessage) :
//
//	"REGISTER:<address minuscule>:<schnorrPublicKeyHex>"
//
// [FIX-CASE-MISMATCH v6.3] : le serveur applique strings.ToLower(req.Address)
// AVANT de recalculer ce message. d.Address (issu de addr.Hex()) est au
// format checksummed (casse mixte) — on DOIT donc le passer en minuscules
// ICI, sinon le hash signé et le hash recalculé côté serveur diffèrent
// (Keccak256 est sensible à la casse) et la vérification échoue à 100% des
// tentatives, pour tous les devices.
func (d *Device) register(backend string) bool {
	pubKeyHex := hex.EncodeToString(crypto.FromECDSAPub(&d.PrivateKey.PublicKey))
	schnorrPubKeyHex := d.SchnorrKeys.PublicKey.Hex() // [SCHNORR-7]
	metadata := fmt.Sprintf(`{"device_id":%q,"type":%q,"name":%q}`, d.ID, d.Type, d.Name)

	// [FIX-CASE-MISMATCH] Normalisation en minuscules — DOIT correspondre
	// EXACTEMENT à ce que le serveur utilise pour recalculer le hash.
	lowerAddress := strings.ToLower(d.Address)

	// [FIX-REGISTER-POP] Preuve de possession — signature ECDSA du message
	// canonique liant cette adresse à CETTE clé Schnorr précise.
	registerMsg := fmt.Sprintf("REGISTER:%s:%s", lowerAddress, schnorrPubKeyHex)
	msgHash := crypto.Keccak256Hash([]byte(registerMsg))
	sig, err := crypto.Sign(msgHash.Bytes(), d.PrivateKey)
	if err != nil {
		log.Printf("[%s] ❌ Signature de la preuve d'enregistrement : %v", d.ID, err)
		return false
	}
	sigHex := hex.EncodeToString(sig)
	msgHashHex := hex.EncodeToString(msgHash.Bytes())

	result, code, err := postJSON(backend+"/api/device/register", map[string]interface{}{
		"address":          d.Address,
		"entityType":       "iot_device",
		"publicKey":        pubKeyHex,
		"schnorrPublicKey": schnorrPubKeyHex, // [SCHNORR-7]
		"metadata":         metadata,
		"signature":        sigHex,     // [FIX-REGISTER-POP] preuve de possession
		"message_hash":     msgHashHex, // [FIX-REGISTER-POP] informatif — recalculé côté serveur de toute façon
	})
	if err != nil {
		log.Printf("[%s] ❌ Enregistrement réseau : %v", d.ID, err)
		return false
	}

	switch code {
	case 200:
		txHash, _ := result["txHash"].(string)
		log.Printf("[%s] ✅ Enregistré | tx: %s | clé Schnorr: %s", d.ID, short(txHash, 20), short(schnorrPubKeyHex, 16))
		return true
	case 409:
		log.Printf("[%s] ℹ️  Déjà enregistré", d.ID)
		return true
	default:
		msg, _ := result["message"].(string)
		log.Printf("[%s] ❌ Enregistrement HTTP %d : %s", d.ID, code, msg)
		return false
	}
}

// ── Activation ─────────────────────────────────────────────────────────────────

func (d *Device) activate(backend string) {
	result, code, err := postJSON(backend+"/api/device/activate", map[string]interface{}{
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

func (d *Device) waitUntilActive(backend string) bool {
	for attempt := 1; attempt <= 10; attempt++ {
		result, code, err := getJSON(fmt.Sprintf("%s/api/device/%s/status", backend, d.Address))
		if err == nil && code == 200 {
			if active, ok := result["isActive"].(bool); ok && active {
				log.Printf("[%s] ✅ Device actif et prêt", d.ID)
				return true
			}
			log.Printf("[%s] ⏳ Device inactif — tentative activation (%d/10)", d.ID, attempt)
			d.activate(backend)
		} else {
			log.Printf("[%s] ⏳ En attente d'indexation (%d/10)...", d.ID, attempt)
		}
		time.Sleep(3 * time.Second)
	}
	log.Printf("[%s] ❌ Device toujours inactif après 10 tentatives", d.ID)
	return false
}

// ── Authentification — [INCHANGÉ v6.0] ────────────────────────────────────────
//
// [SCHNORR-6] SUPPRIMÉ : l'ancien appel réseau à getZKPProof()
// (/api/zkp/generate-secure, avec 3 tentatives et 2s de délai entre chaque)
// est entièrement retiré. La preuve Schnorr est calculée LOCALEMENT en
// microsecondes (zkp.GenerateProof) — un seul aller-retour réseau reste
// nécessaire pour toute l'authentification (POST /api/node/message-secure),
// au lieu de deux dans l'ancien flux Groth16.
func (d *Device) authenticate(backend string) {
	tStart := time.Now()

	timestamp := time.Now().Unix()
	seq := d.nextSeq()
	nonce := fmt.Sprintf("%s-%d-%d", d.ID, timestamp, seq)

	// Données IoT simulées
	data := map[string]interface{}{
		"device_id":   d.ID,
		"type":        d.Type,
		"timestamp":   timestamp,
		"temperature": 18.0 + rand.Float64()*12.0,
		"humidity":    40.0 + rand.Float64()*30.0,
		"pressure":    1000.0 + rand.Float64()*30.0,
	}
	message, _ := json.Marshal(data)
	msgStr := string(message)

	// Signature ECDSA — [INCHANGÉ]
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

	// [SCHNORR-6] Preuve ZKP générée LOCALEMENT — plus d'appel réseau, plus de retry.
	tZKP := time.Now()
	proof, err := d.SchnorrKeys.GenerateProof(d.Address, timestamp, seq, nonce)
	zkpMs := time.Since(tZKP).Milliseconds()
	if err != nil {
		log.Printf("[%s] ❌ Génération preuve Schnorr : %v", d.ID, err)
		return
	}

	// Envoi message sécurisé
	tHTTP := time.Now()
	result, code, err := postJSON(backend+"/api/node/message-secure", map[string]interface{}{
		"device_id":    d.ID,
		"address":      d.Address,
		"message":      msgStr,
		"message_hash": msgHashHex,
		"signature":    sigHex,
		"timestamp":    timestamp,
		"sequence":     seq,
		"nonce":        nonce,

		// [SCHNORR-6] Preuve Schnorr — remplace proof/expected_hash Groth16
		"schnorr_tx": proof.TX.String(),
		"schnorr_ty": proof.TY.String(),
		"schnorr_s":  proof.S.String(),

		// Champs ML (valeurs synthétiques) — [INCHANGÉ]
		"duration":          2.0 + rand.Float64()*3.0,
		"bytes_sent":        float64(len(msgStr)),
		"bytes_received":    256.0,
		"packet_count":      4.0,
		"bytes_per_sec":     128.0 + rand.Float64()*128.0,
		"pkts_per_sec":      2.0,
		"bytes_per_pkt":     64.0 + rand.Float64()*64.0,
		"bytes_ratio":       0.4 + rand.Float64()*0.2,
		"bytes_sent_ratio":  0.5 + rand.Float64()*0.2,
		"bytes_total":       float64(len(msgStr)) + 256.0,
		"log_bytes_total":   6.5,
		"pkt_size_variance": 32.0 + rand.Float64()*32.0,
		"high_volume_flag":  0.0,
		"burst_indicator":   rand.Float64(),
	})
	httpMs := time.Since(tHTTP).Milliseconds()
	totalMs := time.Since(tStart).Milliseconds()

	if err != nil {
		log.Printf("[%s] ❌ Auth réseau : %v", d.ID, err)
		return
	}

	if code == 200 {
		txHash, _ := result["txHash"].(string)
		log.Printf("[%s] ✅ Auth OK | ECDSA:%dms Schnorr:%dms HTTP:%dms Total:%dms | tx:%s",
			d.ID, ecdsaMs, zkpMs, httpMs, totalMs, short(txHash, 18))
	} else {
		msg, _ := result["message"].(string)
		if msg == "" {
			msg, _ = result["error"].(string)
		}
		log.Printf("[%s] ❌ Auth HTTP %d : %s | ECDSA:%dms Schnorr:%dms HTTP:%dms",
			d.ID, code, msg, ecdsaMs, zkpMs, httpMs)
	}
}

// ── Boucle principale d'un device ─────────────────────────────────────────────

func (d *Device) run(backend string, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("[%s] 🚀 Démarrage | Adresse: %s | Type: %s", d.ID, short(d.Address, 20), d.Type)

	if !d.register(backend) {
		log.Printf("[%s] ❌ Enregistrement échoué — device arrêté", d.ID)
		return
	}

	time.Sleep(3 * time.Second)
	if !d.waitUntilActive(backend) {
		log.Printf("[%s] ❌ Device inactif — arrêt", d.ID)
		return
	}

	log.Printf("[%s] 📡 Début des envois (intervalle aléatoire 10-30s)", d.ID)

	for {
		d.authenticate(backend)

		wait := time.Duration(10000+rand.Intn(20001)) * time.Millisecond
		log.Printf("[%s] ⏳ Prochain envoi dans %.1fs", d.ID, wait.Seconds())
		time.Sleep(wait)
	}
}

// ── Main ───────────────────────────────────────────────────────────────────────

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	buildHTTPClient()

	backend := getBackendURL()

	log.Printf("════════════════════════════════════════════")
	log.Printf("🌐 CHAIMAN IoT — Simulateur Simple v6.3 (Schnorr ZKP + preuve de possession, fix casse)")
	log.Printf("   Backend  : %s", backend)
	log.Printf("   Devices  : 4 (IOT_0004 = nouveau, teste la preuve de possession)")
	log.Printf("   Intervalle : 10 à 30 secondes (aléatoire)")
	log.Printf("════════════════════════════════════════════")

	deviceDefs := []struct{ id, name, devType string }{
		{"IOT_0001", "capteur_temperature_salon", "temperature_sensor"},
		{"IOT_0002", "capteur_humidite_cuisine", "humidity_sensor"},
		{"IOT_0003", "detecteur_mouvement_entree", "motion_sensor"},
		// [TEST-POP] Device neuf — jamais enregistré, sert à valider en
		// conditions réelles la preuve de possession ECDSA obligatoire à
		// l'enregistrement (device.go v5.5). Voir commentaire d'en-tête.
		{"IOT_0004", "capteur_test_preuve_possession", "test_sensor"},
	}

	devices := make([]*Device, 0, len(deviceDefs))
	for _, def := range deviceDefs {
		pk, err := crypto.GenerateKey()
		if err != nil {
			log.Fatalf("❌ Génération clé ECDSA : %v", err)
		}
		addr := crypto.PubkeyToAddress(pk.PublicKey)

		// [SCHNORR-8] Charge la clé Schnorr existante si le device a déjà
		// démarré une fois (fichier persisté), sinon en génère une nouvelle
		// ET la sauvegarde immédiatement. Garantit que ce device conserve la
		// MÊME identité (donc les mêmes droits d'accès enregistrés côté
		// serveur) à travers tous ses redémarrages — testé dans
		// pkg/zkp/schnorr_keystore_test.go (TestKeystore_LoadOrCreate_FirstBoot).
		//
		// Chemin : ./data/keys/<device_id>_schnorr.json (créé automatiquement).
		// Permissions 0600 — voir avertissement sécurité dans schnorr_keystore.go
		// pour la migration vers un secure element en production.
		keyPath := fmt.Sprintf("./data/keys/%s_schnorr.json", def.id)
		schnorrKP, isNewKey, err := zkp.LoadOrCreateKeyPair(keyPath)
		if err != nil {
			log.Fatalf("❌ Chargement/génération clé Schnorr pour %s : %v", def.id, err)
		}
		if isNewKey {
			log.Printf("🆕 [%s] Nouvelle identité Schnorr créée : %s", def.id, keyPath)
		} else {
			log.Printf("🔑 [%s] Identité Schnorr existante chargée : %s", def.id, keyPath)
		}

		devices = append(devices, &Device{
			ID:          def.id,
			Name:        def.name,
			Type:        def.devType,
			Address:     addr.Hex(),
			PrivateKey:  pk,
			SchnorrKeys: schnorrKP,
			sequence:    uint64(time.Now().Unix()),
		})
	}

	log.Printf("\n📋 Appareils créés :")
	for _, d := range devices {
		log.Printf("   • %s (%s) → %s | Schnorr: %s", d.ID, d.Type, d.Address, short(d.SchnorrKeys.PublicKey.Hex(), 16))
	}
	log.Println()

	var wg sync.WaitGroup
	for i, d := range devices {
		wg.Add(1)
		go func(dev *Device, delay time.Duration) {
			time.Sleep(delay)
			dev.run(backend, &wg)
		}(d, time.Duration(i)*2*time.Second)
	}

	wg.Wait()
	log.Println("Tous les devices se sont arrêtés.")
}
