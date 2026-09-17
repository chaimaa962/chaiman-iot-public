package zkp

// service_secure.go — VERSION 3.0 — NETTOYAGE POST-MIGRATION SCHNORR
//
// [SCHNORR-9] SUPPRESSION COMPLÈTE du système Groth16 + master secret :
//
//   SUPPRIMÉ définitivement (c'était la faille de sécurité corrigée) :
//     - masterSecret, InitMasterSecret, InitMasterSecretBytes
//     - GenerateSecureSecretFromMaster()  ← permettait au serveur de
//       recalculer le secret de N'IMPORTE QUEL device
//     - GenerateSecureSecretFromPrivateKey(), GenerateRandomSecret()
//     - ComputePublicHash()
//     - SecureZKPManager, GetSecureZKPManager(), GenerateProof(), VerifyProof()
//       (Groth16 — remplacés par pkg/zkp/schnorr_server.go)
//     - GenerateChallenge()  ← remplacé par HashChallenge() dans schnorr.go
//       (même principe Fiat-Shamir, mais lié à l'engagement Schnorr T)
//     - Alias GetZKPService(), GenerateSecureSecret()
//
//   CONSERVÉ (toujours utilisé ailleurs dans le projet — device.go,
//   messages_secure.go — pour la vérification de signature ECDSA classique,
//   INDÉPENDANTE du système ZKP) :
//     - VerifyECDSASignature()
//     - VerifyECDSASignatureWithRecovery()
//     - GenerateSecureNonce()
//     - cleanHexPrefix()
//
// RAISON DE CE NETTOYAGE :
//   Laisser le code Groth16/master-secret présent mais "non appelé" dans le
//   projet est un risque en soi : un futur développeur pourrait le
//   réactiver par erreur (ou un attaquant l'exploiter s'il trouve un chemin
//   de code oublié qui l'invoque encore). Une migration de sécurité n'est
//   complète que lorsque l'ancien système dangereux est PHYSIQUEMENT retiré
//   du code, pas seulement désactivé.
//
// SI VOUS AVEZ BESOIN DE L'ANCIEN CODE (référence historique uniquement) :
//   Il reste consultable dans l'historique Git avant ce commit de nettoyage.
//   auth_circuit_secure.go (le circuit gnark) doit également être SUPPRIMÉ
//   du projet — il n'est plus référencé par aucun fichier après ce nettoyage.

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// ─────────────────────────────────────────────────────────────────────────────
// GenerateSecureNonce génère un nonce cryptographiquement aléatoire (32 octets).
// ─────────────────────────────────────────────────────────────────────────────

func GenerateSecureNonce() (string, error) {
	nonceBytes := make([]byte, 32)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", fmt.Errorf("erreur génération nonce: %v", err)
	}
	return hex.EncodeToString(nonceBytes), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Vérification ECDSA en temps constant — [INCHANGÉ, toujours nécessaire]
// ─────────────────────────────────────────────────────────────────────────────

// VerifyECDSASignature vérifie qu'une signature ECDSA correspond à l'adresse
// donnée. Utilisée pour authentifier l'IDENTITÉ Ethereum du device (signature
// du message applicatif), INDÉPENDAMMENT de la preuve Schnorr ZKP (qui, elle,
// authentifie la possession du secret Schnorr enregistré). Les deux mécanismes
// sont complémentaires : ECDSA lie le message à l'adresse on-chain, Schnorr
// prouve la possession du secret sans le révéler.
func VerifyECDSASignature(
	address common.Address,
	messageHash string,
	signatureHex string,
) (bool, error) {
	signatureHex = cleanHexPrefix(signatureHex)
	messageHash = cleanHexPrefix(messageHash)

	sigBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false, fmt.Errorf("signature hex invalide")
	}
	if len(sigBytes) != 65 {
		return false, fmt.Errorf("longueur signature invalide: %d (attendu 65)", len(sigBytes))
	}

	msgHashBytes, err := hex.DecodeString(messageHash)
	if err != nil {
		return false, fmt.Errorf("message hash invalide")
	}
	if len(msgHashBytes) != 32 {
		return false, fmt.Errorf("hash doit être 32 bytes, reçu %d", len(msgHashBytes))
	}

	sigCopy := make([]byte, 65)
	copy(sigCopy, sigBytes)

	v := sigCopy[64]
	if v >= 27 {
		v -= 27
		sigCopy[64] = v
	}
	if v > 1 {
		return false, fmt.Errorf("recovery id invalide: %d", v)
	}

	pubKey, err := crypto.SigToPub(msgHashBytes, sigCopy)
	if err != nil {
		return false, fmt.Errorf("authentification échouée")
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)

	addr1 := []byte(strings.ToLower(address.Hex()))
	addr2 := []byte(strings.ToLower(recoveredAddr.Hex()))

	if subtle.ConstantTimeCompare(addr1, addr2) != 1 {
		return false, fmt.Errorf("authentification échouée")
	}

	return true, nil
}

// VerifyECDSASignatureWithRecovery récupère l'adresse depuis une signature ECDSA.
func VerifyECDSASignatureWithRecovery(
	messageHash string,
	signatureHex string,
) (common.Address, error) {
	signatureHex = cleanHexPrefix(signatureHex)
	messageHash = cleanHexPrefix(messageHash)

	sigBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return common.Address{}, fmt.Errorf("signature hex invalide")
	}
	if len(sigBytes) != 65 {
		return common.Address{}, fmt.Errorf("longueur signature invalide: %d", len(sigBytes))
	}

	msgHashBytes, err := hex.DecodeString(messageHash)
	if err != nil {
		return common.Address{}, fmt.Errorf("message hash invalide")
	}
	if len(msgHashBytes) != 32 {
		return common.Address{}, fmt.Errorf("hash doit être 32 bytes")
	}

	sigCopy := make([]byte, 65)
	copy(sigCopy, sigBytes)

	v := sigCopy[64]
	if v >= 27 {
		v -= 27
		sigCopy[64] = v
	}
	if v > 1 {
		return common.Address{}, fmt.Errorf("recovery id invalide: %d", v)
	}

	pubKey, err := crypto.SigToPub(msgHashBytes, sigCopy)
	if err != nil {
		return common.Address{}, fmt.Errorf("authentification échouée")
	}

	return crypto.PubkeyToAddress(*pubKey), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Utilitaires internes
// ─────────────────────────────────────────────────────────────────────────────

// cleanHexPrefix supprime le préfixe "0x" ou "0X" d'une chaîne hex.
func cleanHexPrefix(s string) string {
	if len(s) >= 2 && (s[0] == '0') && (s[1] == 'x' || s[1] == 'X') {
		return s[2:]
	}
	return s
}
