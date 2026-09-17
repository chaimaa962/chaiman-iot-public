package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

type SimpleDevice struct {
	ID         string
	Address    string
	PrivateKey *ecdsa.PrivateKey
}

func generateDevice(id string) *SimpleDevice {
	privateKey, _ := crypto.GenerateKey()
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	return &SimpleDevice{
		ID:         id,
		Address:    address.Hex(),
		PrivateKey: privateKey,
	}
}

func (d *SimpleDevice) Authenticate(backendURL string) {
	// Données du capteur
	data := map[string]interface{}{
		"temperature": 20.5,
		"humidity":    65,
		"device_id":   d.ID,
		"timestamp":   time.Now().Unix(),
	}
	message, _ := json.Marshal(data)
	messageStr := string(message)

	// Signer le message
	msgHash := crypto.Keccak256Hash([]byte(messageStr))
	signature, _ := crypto.Sign(msgHash.Bytes(), d.PrivateKey)
	sigHex := hex.EncodeToString(signature)

	// Payload pour le backend
	payload := map[string]interface{}{
		"device_id":    d.ID,
		"address":      d.Address,
		"message":      messageStr,
		"message_hash": hex.EncodeToString(msgHash.Bytes()),
		"signature":    sigHex,
		"timestamp":    time.Now().Unix(),
		"sequence":     1,
		"nonce":        fmt.Sprintf("%s-%d", d.ID, time.Now().Unix()),
	}

	body, _ := json.Marshal(payload)

	// Envoyer la requête HTTPS
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Post(backendURL+"/api/node/message-secure", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("❌ %s: Erreur: %v", d.ID, err)
		return
	}
	defer resp.Body.Close()

	bodyResp, _ := io.ReadAll(resp.Body)
	log.Printf("✅ %s: Réponse %d: %s", d.ID, resp.StatusCode, string(bodyResp))
}

func main() {
	backendURL := "https://localhost:8443"
	
	fmt.Println("🚀 Simulateur IoT Simplifié")
	fmt.Println("==========================")
	fmt.Printf("Backend: %s\n", backendURL)
	fmt.Println()

	// Créer 5 appareils
	devices := make([]*SimpleDevice, 5)
	for i := 1; i <= 5; i++ {
		devices[i-1] = generateDevice(fmt.Sprintf("SIMPLE_%03d", i))
		fmt.Printf("📱 Appareil créé: %s (%s)\n", devices[i-1].ID, devices[i-1].Address[:10])
	}

	fmt.Println("\n🔐 Démarrage des authentifications...")
	fmt.Println()

	for _, device := range devices {
		device.Authenticate(backendURL)
		time.Sleep(2 * time.Second)
	}

	fmt.Println("\n✅ Simulation terminée!")
}
