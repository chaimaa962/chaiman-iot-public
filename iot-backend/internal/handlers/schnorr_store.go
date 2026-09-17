// internal/handlers/schnorr_store.go
// Store persistant pour les clés publiques Schnorr des devices.
//
// Suit exactement le même pattern que NonceStore et SequenceStore
// (messages_secure.go) : map en mémoire + persistance JSON sur disque,
// pour rester cohérent avec le reste du projet.
//
// REMPLACE : la dérivation zkp.GenerateSecureSecretFromMaster() qui
// permettait au serveur de recalculer le secret de N'IMPORTE QUEL device.
// Ici, le serveur stocke UNIQUEMENT ce que chaque device lui a envoyé
// (sa clé publique) — il ne peut rien calculer ni deviner.

package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"iot-backend/pkg/zkp"
)

type SchnorrKeyStore struct {
	mu       sync.RWMutex
	keys     map[string]string // deviceAddress (lowercase) -> clé publique hex
	filePath string
}

var schnorrKeyStore *SchnorrKeyStore

// InitSchnorrKeyStore initialise le store — à appeler depuis InitSecureStores
// (messages_secure.go) au démarrage du serveur, comme nonceStore/sequenceStore.
func InitSchnorrKeyStore(dataDir string) error {
	schnorrKeyStore = &SchnorrKeyStore{
		keys:     make(map[string]string),
		filePath: dataDir + "/schnorr_keys.json",
	}
	if err := schnorrKeyStore.load(); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("chargement schnorr_keys.json: %v", err)
	}
	return nil
}

func (s *SchnorrKeyStore) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return json.Unmarshal(data, &s.keys)
}

func (s *SchnorrKeyStore) save() error {
	s.mu.RLock()
	data, err := json.MarshalIndent(s.keys, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return err
	}
	tmpPath := s.filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmpPath, s.filePath)
}

// RegisterSchnorrPublicKey enregistre la clé publique d'un device lors de
// son inscription. Appelé UNE FOIS, à l'enregistrement (RegisterDevice).
// Retourne une erreur si une clé différente existe déjà pour cette adresse
// (empêche un attaquant de "voler" une adresse en ré-enregistrant une
// nouvelle clé publique par-dessus celle d'un device légitime).
func RegisterSchnorrPublicKey(deviceAddress, publicKeyHex string) error {
	addr := strings.ToLower(deviceAddress)

	schnorrKeyStore.mu.Lock()
	existing, exists := schnorrKeyStore.keys[addr]
	if exists && existing != publicKeyHex {
		schnorrKeyStore.mu.Unlock()
		return fmt.Errorf("une clé Schnorr différente est déjà enregistrée pour cette adresse")
	}
	schnorrKeyStore.keys[addr] = publicKeyHex
	schnorrKeyStore.mu.Unlock()

	return schnorrKeyStore.save()
}

// GetSchnorrPublicKey récupère la clé publique enregistrée d'un device.
// Utilisé à CHAQUE authentification pour vérifier la preuve Schnorr envoyée.
func GetSchnorrPublicKey(deviceAddress string) (*zkp.SchnorrPublicKey, error) {
	addr := strings.ToLower(deviceAddress)

	schnorrKeyStore.mu.RLock()
	hexKey, exists := schnorrKeyStore.keys[addr]
	schnorrKeyStore.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("aucune clé Schnorr enregistrée pour %s — le device doit s'enregistrer d'abord", deviceAddress)
	}

	return zkp.PublicKeyFromHex(hexKey)
}
