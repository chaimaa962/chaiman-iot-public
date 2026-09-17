package zkp

// schnorr.go — NOUVEAU : preuve de connaissance Schnorr (secp256k1)
// REMPLACE Groth16 (service_secure.go / auth_circuit_secure.go) pour
// l'authentification "ZKP" des devices IoT.
//
// POURQUOI CE CHANGEMENT :
//
//   ANCIEN MODÈLE (Groth16 + master secret) :
//     - Le secret de CHAQUE device était dérivé côté BACKEND depuis un
//       masterSecret unique (HKDF, voir GenerateSecureSecretFromMaster).
//       Le backend connaissait donc, de fait, le secret de tous les
//       devices — la preuve "zero-knowledge" ne prouvait rien puisque le
//       vérifieur connaissait déjà le témoin.
//     - Single point of failure : fuite du masterSecret = compromission
//       de TOUS les devices simultanément.
//     - Lourd : compilation de circuit R1CS, setup Groth16, pairing BN254,
//       plus un aller-retour réseau séparé (/api/zkp/generate-secure)
//       avant chaque authentification.
//
//   NOUVEAU MODÈLE (Schnorr) :
//     - Chaque device génère sa propre paire (x, Y) LOCALEMENT — x ne
//       quitte JAMAIS le device.
//     - Réutilise le MÊME keypair secp256k1 déjà utilisé pour signer les
//       messages (ECDSA) : Y est exactement la publicKey déjà enregistrée
//       on-chain lors de RegisterDevice. Aucun changement de schéma
//       d'enregistrement nécessaire.
//     - Le backend ne fait QUE vérifier : g^s == T + c·Y. Il ne calcule
//       jamais x, ne génère jamais de preuve.
//     - Le challenge est dérivé de manière déterministe (Fiat-Shamir) du
//       contexte du message (address, timestamp, sequence, nonce, T, Y) :
//       plus besoin d'un aller-retour "generate-secure" séparé ni d'un
//       ChallengeStore avec fenêtre d'expiration — toute la preuve tient
//       en UN SEUL round-trip HTTP (message + preuve envoyés ensemble).
//     - Calcul largement plus léger que Groth16, exécutable sur un
//       microcontrôleur faible (ESP32, etc).
//
// [FIX-BUILD] Champs SchnorrProof renommés Tx/Ty → TX/TY : le reste du
// projet (messages_secure.go, iot_simulator.go, schnorr_keystore_test.go)
// utilise systématiquement TX/TY (majuscules). Cette incohérence de casse
// empêchait la compilation ("undefined: SchnorrProof.TX" etc. remontaient
// en réalité comme des erreurs sur le package zkp tout entier une fois
// combinées aux symboles manquants de schnorr_device.go). Uniformisé ici.

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	schnorrCurve = crypto.S256()
	schnorrOrder = schnorrCurve.Params().N
)

// SchnorrProof est la preuve envoyée par le device : (T, s).
//   T = r*G   (commitment, point sur la courbe)
//   s = r + c*x mod N   (réponse, scalaire)
type SchnorrProof struct {
	TX *big.Int
	TY *big.Int
	S  *big.Int
}

// Encode sérialise la preuve en 3 chaînes hex (32 bytes chacune, sans "0x").
func (p *SchnorrProof) Encode() (txHex, tyHex, sHex string) {
	return hex.EncodeToString(leftPad32(p.TX)),
		hex.EncodeToString(leftPad32(p.TY)),
		hex.EncodeToString(leftPad32(p.S))
}

// DecodeSchnorrProof reconstruit une preuve depuis 3 chaînes hex reçues du device.
func DecodeSchnorrProof(txHex, tyHex, sHex string) (*SchnorrProof, error) {
	tx, err := hexToBigInt(txHex)
	if err != nil {
		return nil, fmt.Errorf("Tx invalide: %w", err)
	}
	ty, err := hexToBigInt(tyHex)
	if err != nil {
		return nil, fmt.Errorf("Ty invalide: %w", err)
	}
	s, err := hexToBigInt(sHex)
	if err != nil {
		return nil, fmt.Errorf("S invalide: %w", err)
	}
	return &SchnorrProof{TX: tx, TY: ty, S: s}, nil
}

func hexToBigInt(h string) (*big.Int, error) {
	h = strings.TrimPrefix(h, "0x")
	b, err := hex.DecodeString(h)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, fmt.Errorf("valeur vide")
	}
	return new(big.Int).SetBytes(b), nil
}

func leftPad32(n *big.Int) []byte {
	b := make([]byte, 32)
	if n == nil {
		return b
	}
	nb := n.Bytes()
	if len(nb) > 32 {
		nb = nb[len(nb)-32:]
	}
	copy(b[32-len(nb):], nb)
	return b
}

// ─────────────────────────────────────────────────────────────────────────
// Côté DEVICE — référence uniquement. Le simulateur Go (iot_simulator.go)
// et tout firmware réel doivent implémenter la même logique localement ;
// ce package pkg/zkp est compilé côté BACKEND, pas côté device.
//
// NOTE : GenerateSchnorrProof/VerifySchnorrProof ci-dessous utilisent la
// clé ECDSA du device (x = privateKey.D) et une common.Address comme
// contexte. C'est un modèle alternatif à SchnorrKeyPair (schnorr_device.go),
// qui lui utilise une identité Schnorr dédiée et indépendante de l'ECDSA,
// persistée via schnorr_keystore.go. Le projet utilise ce second modèle
// (SchnorrKeyPair / GenerateProof / VerifyProof) en pratique — ces
// fonctions-ci restent disponibles comme référence / utilitaires de test.
// ─────────────────────────────────────────────────────────────────────────

// GenerateSchnorrProof calcule (T, s) à partir de la clé privée ECDSA du
// device (réutilisée comme scalaire Schnorr x = privateKey.D) et du
// contexte du message. Fournie ici comme référence / pour du code Go
// backend qui aurait besoin de générer une preuve de test.
//
// SÉCURITÉ : r doit être un nonce ALÉATOIRE et JAMAIS réutilisé entre deux
// preuves pour la même clé — sinon x est trivialement récupérable (même
// faille que la réutilisation d'un nonce ECDSA). crypto/rand uniquement,
// jamais math/rand.
func GenerateSchnorrProof(privateKey *ecdsa.PrivateKey, address common.Address, timestamp int64, sequence uint64, nonce string) (*SchnorrProof, error) {
	if privateKey == nil || privateKey.D == nil {
		return nil, fmt.Errorf("clé privée invalide")
	}

	r, err := randScalar()
	if err != nil {
		return nil, fmt.Errorf("génération r: %w", err)
	}

	Tx, Ty := schnorrCurve.ScalarBaseMult(r.Bytes())

	c := computeSchnorrChallenge(Tx, Ty, privateKey.PublicKey.X, privateKey.PublicKey.Y, address, timestamp, sequence, nonce)

	x := new(big.Int).Mod(privateKey.D, schnorrOrder)
	s := new(big.Int).Mul(c, x)
	s.Add(s, r)
	s.Mod(s, schnorrOrder)

	// Effacement best-effort des valeurs sensibles locales.
	r.SetInt64(0)
	x.SetInt64(0)

	return &SchnorrProof{TX: Tx, TY: Ty, S: s}, nil
}

func randScalar() (*big.Int, error) {
	b := make([]byte, 40) // marge > 32 bytes pour réduire le biais modulo
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	r := new(big.Int).SetBytes(b)
	r.Mod(r, schnorrOrder)
	if r.Sign() == 0 {
		r.SetInt64(1)
	}
	return r, nil
}

// ─────────────────────────────────────────────────────────────────────────
// Côté BACKEND — vérification uniquement. Aucun secret n'est jamais
// manipulé ici : Y est une donnée publique lue depuis la blockchain.
// ─────────────────────────────────────────────────────────────────────────

// VerifySchnorrProof vérifie g^s == T + c·Y en utilisant la publicKey Y
// DÉJÀ enregistrée on-chain pour ce device (voir blockchain.DeviceGetInfo
// + ParseRegisteredPublicKey ci-dessous).
func VerifySchnorrProof(Y *ecdsa.PublicKey, proof *SchnorrProof, address common.Address, timestamp int64, sequence uint64, nonce string) (bool, error) {
	if Y == nil || Y.X == nil || Y.Y == nil {
		return false, fmt.Errorf("clé publique invalide")
	}
	if proof == nil || proof.TX == nil || proof.TY == nil || proof.S == nil {
		return false, fmt.Errorf("preuve incomplète")
	}
	if !schnorrCurve.IsOnCurve(Y.X, Y.Y) {
		return false, fmt.Errorf("clé publique hors courbe")
	}
	if !schnorrCurve.IsOnCurve(proof.TX, proof.TY) {
		return false, fmt.Errorf("commitment T hors courbe")
	}
	if proof.S.Sign() <= 0 || proof.S.Cmp(schnorrOrder) >= 0 {
		return false, fmt.Errorf("réponse S hors intervalle valide")
	}

	c := computeSchnorrChallenge(proof.TX, proof.TY, Y.X, Y.Y, address, timestamp, sequence, nonce)

	// left = s*G
	leftX, leftY := schnorrCurve.ScalarBaseMult(proof.S.Bytes())

	// right = T + c*Y
	cYx, cYy := schnorrCurve.ScalarMult(Y.X, Y.Y, c.Bytes())
	rightX, rightY := schnorrCurve.Add(proof.TX, proof.TY, cYx, cYy)

	if leftX.Cmp(rightX) != 0 || leftY.Cmp(rightY) != 0 {
		return false, fmt.Errorf("équation Schnorr non vérifiée")
	}
	return true, nil
}

// computeSchnorrChallenge : Fiat-Shamir — lie le commitment T, la clé
// publique Y ET le contexte applicatif (address, timestamp, sequence,
// nonce) dans un seul hash Keccak256. Inclure T et Y (pas seulement le
// contexte) empêche les attaques "rogue-key" et la réutilisation d'une
// preuve pour une autre clé ou un autre device.
func computeSchnorrChallenge(Tx, Ty, Yx, Yy *big.Int, address common.Address, timestamp int64, sequence uint64, nonce string) *big.Int {
	data := make([]byte, 0, 4*32+20+8+8+len(nonce))
	data = append(data, leftPad32(Tx)...)
	data = append(data, leftPad32(Ty)...)
	data = append(data, leftPad32(Yx)...)
	data = append(data, leftPad32(Yy)...)
	data = append(data, address.Bytes()...)
	data = append(data, big.NewInt(timestamp).Bytes()...)
	data = append(data, new(big.Int).SetUint64(sequence).Bytes()...)
	data = append(data, []byte(nonce)...)

	hash := crypto.Keccak256Hash(data)
	c := new(big.Int).SetBytes(hash.Bytes())
	c.Mod(c, schnorrOrder)
	if c.Sign() == 0 {
		c.SetInt64(1)
	}
	return c
}

// ParseRegisteredPublicKey décode la publicKey hex (format ECDSA non
// compressé, même format que crypto.FromECDSAPub, tel que stocké on-chain
// lors du register) en *ecdsa.PublicKey utilisable par VerifySchnorrProof.
//
// NOTE : les nœuds Geth internes (validateurs/fullnodes) utilisent une
// "clé publique synthétique" (voir client.go: generateSyntheticPublicKeyFromAddress)
// qui n'est PAS un point EC valide — ParseRegisteredPublicKey échouera pour
// ces entités, ce qui est attendu : Schnorr ne s'applique qu'aux devices IoT.
func ParseRegisteredPublicKey(pubKeyHex string) (*ecdsa.PublicKey, error) {
	cleaned := strings.TrimPrefix(pubKeyHex, "0x")
	raw, err := hex.DecodeString(cleaned)
	if err != nil {
		return nil, fmt.Errorf("clé publique hex invalide: %w", err)
	}
	pub, err := crypto.UnmarshalPubkey(raw)
	if err != nil {
		return nil, fmt.Errorf("clé publique EC invalide (device non-Schnorr ou entité non-IoT ?): %w", err)
	}
	return pub, nil
}
