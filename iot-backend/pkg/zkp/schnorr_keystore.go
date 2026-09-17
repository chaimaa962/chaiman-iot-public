// pkg/zkp/schnorr_keystore.go
// Stockage persistant SÉCURISÉ de la clé privée Schnorr — CÔTÉ DEVICE.
//
// POURQUOI CE FICHIER EST NÉCESSAIRE (finition "professionnelle") :
//
//   Sans lui, chaque redémarrage du device générerait une NOUVELLE paire de
//   clés (voir GenerateKeyPair dans iot_simulator.go main()) — ce qui casse
//   totalement le modèle : le serveur aurait enregistré l'ANCIENNE clé
//   publique, et toute authentification suivante échouerait après un simple
//   redémarrage. Un device réel doit conserver la MÊME identité (donc la
//   même clé privée) sur toute sa durée de vie.
//
// SÉCURITÉ DU FICHIER DE CLÉ :
//   - Permissions 0600 (lecture/écriture propriétaire uniquement)
//   - Le fichier contient la clé privée en clair sur disque — c'est un
//     compromis pragmatique pour un simulateur/prototype. Sur un VRAI
//     microcontrôleur IoT en production, remplacer ce stockage par :
//       • Un secure element matériel (ATECC608A, TPM, Secure Enclave...)
//       • Ou au minimum un chiffrement au repos (clé dérivée d'un secret
//         matériel unique au device, ex: eFuse/OTP sur ESP32)
//   - Ce fichier documente le point d'extension exact où brancher cela
//     (voir SaveKeyPair / LoadOrCreateKeyPair ci-dessous).

package zkp

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
)

// persistedKeyPair est le format de sérialisation sur disque.
type persistedKeyPair struct {
	PrivateKeyHex string `json:"private_key_hex"` // SECRET — voir avertissement ci-dessus
	PublicKeyHex  string `json:"public_key_hex"`
}

// SaveKeyPair persiste une paire de clés Schnorr sur disque, à l'emplacement
// donné. Crée les répertoires parents si nécessaire. Permissions 0600.
//
// POINT D'EXTENSION PRODUCTION : c'est ICI qu'un vrai déploiement
// chiffrerait PrivateKeyHex avant écriture (ex: AES-GCM avec une clé dérivée
// du secure element du device), au lieu de l'écrire en clair.
func SaveKeyPair(kp *SchnorrKeyPair, path string) error {
	if kp == nil || kp.PrivateKey == nil {
		return fmt.Errorf("paire de clés invalide")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("création répertoire clé: %w", err)
	}

	data := persistedKeyPair{
		PrivateKeyHex: kp.PrivateKey.Text(16),
		PublicKeyHex:  kp.PublicKey.Hex(),
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("sérialisation clé: %w", err)
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, raw, 0o600); err != nil {
		return fmt.Errorf("écriture clé: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("finalisation écriture clé: %w", err)
	}
	return nil
}

// LoadKeyPair charge une paire de clés Schnorr précédemment sauvegardée.
// Retourne une erreur si le fichier n'existe pas (utiliser os.IsNotExist
// pour distinguer "premier démarrage" d'une vraie erreur de lecture).
func LoadKeyPair(path string) (*SchnorrKeyPair, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err // propagé tel quel pour permettre os.IsNotExist(err)
	}

	var data persistedKeyPair
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("désérialisation clé: %w", err)
	}

	priv, ok := new(big.Int).SetString(data.PrivateKeyHex, 16)
	if !ok {
		return nil, fmt.Errorf("clé privée corrompue dans %s", path)
	}
	pub, err := PublicKeyFromHex(data.PublicKeyHex)
	if err != nil {
		return nil, fmt.Errorf("clé publique corrompue dans %s: %w", path, err)
	}

	// Vérification d'intégrité : la clé publique stockée doit correspondre
	// à la clé privée stockée (détecte une corruption ou un fichier trafiqué).
	expectedX, expectedY := curve.ScalarBaseMult(priv.Bytes())
	if expectedX.Cmp(pub.X) != 0 || expectedY.Cmp(pub.Y) != 0 {
		return nil, fmt.Errorf("incohérence clé privée/publique dans %s — fichier corrompu ou trafiqué", path)
	}

	return &SchnorrKeyPair{PrivateKey: priv, PublicKey: pub}, nil
}

// LoadOrCreateKeyPair est le point d'entrée recommandé pour un device au
// démarrage : charge la clé existante si présente, sinon en génère une
// nouvelle ET la persiste immédiatement. Garantit qu'un device conserve la
// MÊME identité (même clé, donc même comportement d'authentification) sur
// tous ses redémarrages suivants.
func LoadOrCreateKeyPair(path string) (kp *SchnorrKeyPair, isNew bool, err error) {
	kp, err = LoadKeyPair(path)
	if err == nil {
		return kp, false, nil
	}
	if !os.IsNotExist(err) {
		// Erreur réelle (corruption, permissions...) — NE PAS écraser
		// silencieusement une clé existante potentiellement récupérable.
		return nil, false, fmt.Errorf("chargement clé existante: %w", err)
	}

	// Premier démarrage : génération + sauvegarde immédiate.
	kp, genErr := GenerateKeyPair()
	if genErr != nil {
		return nil, false, fmt.Errorf("génération nouvelle clé: %w", genErr)
	}
	if saveErr := SaveKeyPair(kp, path); saveErr != nil {
		return nil, false, fmt.Errorf("sauvegarde nouvelle clé: %w", saveErr)
	}
	return kp, true, nil
}
