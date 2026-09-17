// pkg/zkp/schnorr_device.go
//
// [FIX-BUILD] Ce fichier était manquant — il fournit l'API "identité Schnorr
// du device" attendue par :
//   - schnorr_keystore.go  (SaveKeyPair/LoadKeyPair/LoadOrCreateKeyPair)
//   - schnorr_keystore_test.go
//   - internal/handlers/schnorr_store.go (RegisterSchnorrPublicKey/GetSchnorrPublicKey)
//   - internal/handlers/messages_secure.go (vérification à l'authentification)
//   - iot_simulator.go (génération de preuve côté device)
//
// Erreurs de compilation corrigées par ce fichier :
//   undefined: SchnorrKeyPair
//   undefined: PublicKeyFromHex
//   undefined: curve
//   undefined: GenerateKeyPair
//   (+ (*SchnorrKeyPair).GenerateProof, VerifyProof)
//
// DIFFÉRENCE avec Schnorr.go :
//   Schnorr.go (GenerateSchnorrProof/VerifySchnorrProof) réutilise la clé
//   ECDSA existante du device et lie le challenge à une common.Address.
//   Ce fichier définit une identité Schnorr INDÉPENDANTE et dédiée
//   (SchnorrKeyPair), persistée sur disque côté device (schnorr_keystore.go)
//   et dont seule la clé publique (SchnorrPublicKey) est envoyée/stockée
//   côté serveur (schnorr_store.go). C'est CE modèle qui est utilisé dans
//   le flux d'authentification réel (messages_secure.go, iot_simulator.go).
//
// Réutilise schnorrCurve / schnorrOrder / leftPad32 déjà déclarés dans
// Schnorr.go (même package) — pas de duplication.

package zkp

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

// [FIX-SCHNORR-CHALLENGE-CASE] Le device (iot_simulator.go) transmet
// l'adresse au format "checksummed" renvoyé par crypto.Address.Hex()
// (ex: 0xBCD3351e42171A8b2211C61f423C3E17A455161a), tandis que le serveur
// (messages_secure.go, HandleSecureNodeMessage) applique
// strings.ToLower(req.Address) AVANT d'appeler VerifyProof. Comme le
// challenge Fiat-Shamir inclut les octets bruts de la chaîne "address"
// dans le hash Keccak256, une différence de casse produit un challenge c
// totalement différent des deux côtés — d'où l'échec systématique
// "équation Schnorr non vérifiée" observé sur tous les devices, à chaque
// tentative. Normaliser ICI (dans la fonction partagée par le device ET
// le serveur) garantit que peu importe la casse fournie par l'appelant,
// le hash calculé est identique des deux côtés.

// curve — alias vers la courbe secp256k1 utilisée dans tout le package.
// Nom attendu tel quel par schnorr_keystore.go (curve.ScalarBaseMult(...)).
var curve = schnorrCurve

// ─────────────────────────────────────────────────────────────────────────
// SchnorrPublicKey — clé publique Y = x·G
// ─────────────────────────────────────────────────────────────────────────

type SchnorrPublicKey struct {
	X *big.Int
	Y *big.Int
}

// Hex sérialise la clé publique au format non compressé standard
// (0x04 || X(32 bytes) || Y(32 bytes)), comme crypto.FromECDSAPub.
func (pk *SchnorrPublicKey) Hex() string {
	if pk == nil || pk.X == nil || pk.Y == nil {
		return ""
	}
	buf := make([]byte, 65)
	buf[0] = 0x04
	copy(buf[1:33], leftPad32(pk.X))
	copy(buf[33:65], leftPad32(pk.Y))
	return hex.EncodeToString(buf)
}

// PublicKeyFromHex reconstruit une SchnorrPublicKey depuis sa forme hex
// (avec ou sans préfixe "0x"). Vérifie que le point est bien sur la courbe.
func PublicKeyFromHex(s string) (*SchnorrPublicKey, error) {
	cleaned := strings.TrimPrefix(s, "0x")
	raw, err := hex.DecodeString(cleaned)
	if err != nil {
		return nil, fmt.Errorf("clé publique Schnorr hex invalide: %w", err)
	}
	if len(raw) != 65 || raw[0] != 0x04 {
		return nil, fmt.Errorf("format clé publique Schnorr invalide (65 octets non compressés attendus, reçu %d)", len(raw))
	}
	x := new(big.Int).SetBytes(raw[1:33])
	y := new(big.Int).SetBytes(raw[33:65])
	if !curve.IsOnCurve(x, y) {
		return nil, fmt.Errorf("clé publique Schnorr hors courbe")
	}
	return &SchnorrPublicKey{X: x, Y: y}, nil
}

// ─────────────────────────────────────────────────────────────────────────
// SchnorrKeyPair — identité complète (secrète + publique) d'un device
// ─────────────────────────────────────────────────────────────────────────

type SchnorrKeyPair struct {
	PrivateKey *big.Int          // x — NE JAMAIS quitter le device / sérialiser côté serveur
	PublicKey  *SchnorrPublicKey // Y = x·G — seule valeur envoyée/stockée côté serveur
}

// GenerateKeyPair génère une nouvelle paire de clés Schnorr aléatoire.
// Utilise crypto/rand exclusivement (jamais math/rand).
func GenerateKeyPair() (*SchnorrKeyPair, error) {
	priv, err := randFieldElement()
	if err != nil {
		return nil, fmt.Errorf("génération clé privée Schnorr: %w", err)
	}
	x, y := curve.ScalarBaseMult(priv.Bytes())
	return &SchnorrKeyPair{
		PrivateKey: priv,
		PublicKey:  &SchnorrPublicKey{X: x, Y: y},
	}, nil
}

func randFieldElement() (*big.Int, error) {
	b := make([]byte, 40) // marge > 32 bytes pour réduire le biais modulo
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	k := new(big.Int).SetBytes(b)
	k.Mod(k, schnorrOrder)
	if k.Sign() == 0 {
		k.SetInt64(1)
	}
	return k, nil
}

// GenerateProof calcule une preuve Schnorr (T, s) liant le commitment à
// l'identité publique du device (PublicKey) et au contexte applicatif
// (address, timestamp, sequence, nonce) via Fiat-Shamir.
//
// NOTE : "address" ici est une chaîne libre (voir iot_simulator.go, qui
// passe d.Address — une adresse Ethereum hex — et les tests, qui passent
// un identifiant arbitraire comme "IOT_0001"). Le hash Keccak256 englobe
// la chaîne telle quelle, donc le format n'a pas besoin d'être une adresse
// valide : seule la cohérence entre device et vérificateur compte.
func (kp *SchnorrKeyPair) GenerateProof(address string, timestamp int64, sequence uint64, nonce string) (*SchnorrProof, error) {
	if kp == nil || kp.PrivateKey == nil || kp.PublicKey == nil {
		return nil, fmt.Errorf("paire de clés Schnorr invalide")
	}

	r, err := randFieldElement()
	if err != nil {
		return nil, fmt.Errorf("génération r: %w", err)
	}

	Tx, Ty := curve.ScalarBaseMult(r.Bytes())

	c := computeDeviceChallenge(Tx, Ty, kp.PublicKey.X, kp.PublicKey.Y, address, timestamp, sequence, nonce)

	s := new(big.Int).Mul(c, kp.PrivateKey)
	s.Add(s, r)
	s.Mod(s, schnorrOrder)

	// Effacement best-effort du nonce local.
	r.SetInt64(0)

	return &SchnorrProof{TX: Tx, TY: Ty, S: s}, nil
}

// VerifyProof vérifie g^s == T + c·Y pour une SchnorrPublicKey enregistrée
// côté serveur (voir internal/handlers/schnorr_store.go).
func VerifyProof(pub *SchnorrPublicKey, proof *SchnorrProof, address string, timestamp int64, sequence uint64, nonce string) (bool, error) {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return false, fmt.Errorf("clé publique Schnorr invalide")
	}
	if proof == nil || proof.TX == nil || proof.TY == nil || proof.S == nil {
		return false, fmt.Errorf("preuve incomplète")
	}
	if !curve.IsOnCurve(pub.X, pub.Y) {
		return false, fmt.Errorf("clé publique hors courbe")
	}
	if !curve.IsOnCurve(proof.TX, proof.TY) {
		return false, fmt.Errorf("commitment T hors courbe")
	}
	if proof.S.Sign() <= 0 || proof.S.Cmp(schnorrOrder) >= 0 {
		return false, fmt.Errorf("réponse S hors intervalle valide")
	}

	c := computeDeviceChallenge(proof.TX, proof.TY, pub.X, pub.Y, address, timestamp, sequence, nonce)

	// left = s*G
	leftX, leftY := curve.ScalarBaseMult(proof.S.Bytes())

	// right = T + c*Y
	cYx, cYy := curve.ScalarMult(pub.X, pub.Y, c.Bytes())
	rightX, rightY := curve.Add(proof.TX, proof.TY, cYx, cYy)

	if leftX.Cmp(rightX) != 0 || leftY.Cmp(rightY) != 0 {
		return false, fmt.Errorf("équation Schnorr non vérifiée")
	}
	return true, nil
}

// computeDeviceChallenge — Fiat-Shamir sur T, Y et le contexte applicatif.
// Même construction que computeSchnorrChallenge (Schnorr.go), mais avec une
// adresse en string libre plutôt qu'une common.Address stricte.
func computeDeviceChallenge(Tx, Ty, Yx, Yy *big.Int, address string, timestamp int64, sequence uint64, nonce string) *big.Int {
	// [FIX-SCHNORR-CHALLENGE-CASE] Normalisation obligatoire — voir
	// commentaire au-dessus des imports. Sans ceci, un device envoyant une
	// adresse checksummée et un serveur la lowercasant avant vérification
	// ne calculeront JAMAIS le même challenge.
	normalizedAddress := strings.ToLower(address)

	data := make([]byte, 0, 4*32+len(normalizedAddress)+8+8+len(nonce))
	data = append(data, leftPad32(Tx)...)
	data = append(data, leftPad32(Ty)...)
	data = append(data, leftPad32(Yx)...)
	data = append(data, leftPad32(Yy)...)
	data = append(data, []byte(normalizedAddress)...)
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

var _ = crypto.S256 // garde l'import crypto utilisé même si Keccak256Hash suffit déjà
