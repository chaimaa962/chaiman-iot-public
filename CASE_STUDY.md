# CHAIMAN IoT — Security Case Study

This document is a technical write-up of the real security and correctness issues found and fixed during the development of CHAIMAN IoT, in the order they were discovered. Each section covers: the problem, the root cause, the fix, and how it was verified.

---

## 1. Groth16 master-secret architecture — single point of compromise

### The problem

The original ZKP authentication layer used a Groth16 zk-SNARK circuit. To avoid asking every IoT device to run a heavy trusted-setup ceremony, the backend derived each device's secret witness from a single **master secret** held server-side (`GenerateSecureSecretFromMaster`, HKDF-based).

### Why this was a real vulnerability, not a style issue

A "zero-knowledge proof" is only meaningful if the verifier does not already know the witness. Here, the server — the verifier — could recompute the secret of *any* device on demand. The "zero-knowledge" property was fiction: the server already held everything the proof was supposed to hide.

The practical consequence: leaking `masterSecret` (a single value, in one place) would compromise **every device in the fleet simultaneously**. This is the textbook definition of a single point of compromise, and it defeated the entire purpose of using a ZKP in the first place.

### The fix

Migrated to a custom **Schnorr identification protocol** over secp256k1 (`pkg/zkp/Schnorr.go`, `schnorr_device.go`):

- Each device generates its own key pair `(x, Y = x·G)` **locally**, once, at first boot.
- `x` never leaves the device — not at registration, not at authentication, not ever serialized to the network.
- The server stores only `Y` (public, meaningless without `x`) and verifies proofs using nothing but public data: `s·G == T + c·Y`.
- The challenge `c` is derived deterministically via Fiat-Shamir (`Keccak256(T, Y, address, timestamp, sequence, nonce)`), so the whole exchange fits in a single HTTP round-trip — no separate `/api/zkp/generate-secure` call, no challenge pre-registration store.

Removed entirely: `GenerateSecureSecretFromMaster`, `SecureZKPManager`, `GetSecureZKPManager`, the Groth16 circuit (`auth_circuit_secure.go`), and every route that depended on them (`POST /api/auth/zkp`, `POST /api/zkp/generate-secure`).

### Why removal, not just deprecation

Leaving the old Groth16/master-secret code present but "unused" in the codebase is a risk in itself: a future contributor could re-enable it by mistake, or an overlooked code path could still invoke it. A security migration is only complete when the dangerous predecessor is physically removed, not merely disconnected.

---

## 2. Identity-squatting at registration

### The problem

`RegisterDevice` accepted an ECDSA signature *if provided*, but did not require one:

```go
if req.Signature != "" && req.MessageHash != "" {
    // verify signature...
}
```

Any client that omitted `signature`/`message_hash` skipped verification entirely. Nothing tied `deviceAddress` in the request to actual possession of the corresponding private key — it was taken from `req.Address` at face value.

### Attack scenario

Ethereum addresses are derivable/guessable ahead of a device's first real registration (e.g. addresses generated offline before deployment, or simply observed once a device starts announcing itself). An attacker who learns a future device's address could register first, submitting a fabricated Schnorr public key for that address.

Because `RegisterSchnorrPublicKey` refuses to overwrite an existing Schnorr key for an address (a correct anti-takeover check on its own), this didn't let the attacker impersonate the device at the authentication layer — but it permanently blocked the *real* device from ever registering its real key. Net effect: identity squatting / denial of service against a specific device, triggerable by anyone who knew (or predicted) its address.

### The fix

Proof-of-possession is now mandatory, and the message being signed is computed server-side — never trusted from the client:

```go
if req.Signature == "" {
    sendError(w, "Signature ECDSA requise pour prouver la possession de l'adresse", http.StatusBadRequest)
    return
}

registrationMsg := buildRegistrationProofMessage(req.Address, req.SchnorrPublicKey)
// "REGISTER:<address>:<schnorrPublicKey>"
expectedHash := crypto.Keccak256Hash([]byte(registrationMsg))
validECDSA, err := zkp.VerifyECDSASignature(deviceAddr, expectedHashHex, req.Signature)
```

Two things matter here:

1. **The signature is no longer optional.** A missing signature is now a hard `400`, not a skipped check.
2. **The signed message is server-derived, not client-supplied.** Previously trusting a client-provided `message_hash` would have reopened the same class of bug (sign an arbitrary message, claim it's the expected one). Binding the message to the *specific* Schnorr key submitted in *that* request also prevents replaying a valid signature from one registration with a different Schnorr key.

### Verification

Added a dedicated end-to-end test device (`IOT_0004` in `iot_simulator.go`) that is never pre-registered, specifically to exercise the mandatory proof-of-possession path in real conditions — the first three devices would short-circuit to `409 Already registered` before ever reaching signature verification, and wouldn't have caught a regression here.

---

## 3. Case-sensitivity bug — 100% registration failure

### Symptom

After deploying the fix above, **every** registration attempt from **every** device failed with:

```
401 Signature ECDSA invalide — preuve de possession de l'adresse échouée
```

Not intermittent — deterministic, on every attempt, for every device.

### Root cause

`crypto.Address.Hex()` in go-ethereum returns an address in **EIP-55 checksummed** form (mixed case, e.g. `0xBCD3351e42171A8b2211C61f423C3E17A455161a`). The device signed the registration message built from this mixed-case string.

Server-side, `RegisterDevice` normalizes the incoming address to lowercase (`req.Address = strings.ToLower(req.Address)`) as its very first step — a reasonable-looking normalization for storage/lookup consistency — and then rebuilt the *same* canonical message from this now-lowercased address before verifying the signature.

Keccak256 is fully case-sensitive: two strings differing only in letter case hash to completely unrelated digests. The device's signature, computed over the checksummed-address hash, could therefore never match the server's recomputed hash over the lowercased address. The two sides were — silently, deterministically — signing and verifying two different messages that happened to look almost identical to a human reading the logs.

### The fix

Rather than trusting every call site (device code, server code, tests) to independently remember to lowercase the address before hashing, normalization was moved **inside the single shared challenge-construction function** that both the device and the server call:

```go
func computeDeviceChallenge(Tx, Ty, Yx, Yy *big.Int, address string, ...) *big.Int {
    normalizedAddress := strings.ToLower(address)
    // ... hash normalizedAddress, not the raw input
}
```

Same principle applied to the registration proof message on the device side (`lowerAddress := strings.ToLower(d.Address)` in `iot_simulator.go`, matching the server's own `strings.ToLower(req.Address)`).

### Verification

Confirmed by an independent client-side/server-side hash comparison: before the fix, the two computed digests differed on 100% of runs; after the fix, they matched on 100% of runs, across all four simulated devices.

### The general lesson

Whenever a security check depends on **byte-for-byte string equality between two independently-implemented parties**, any formatting difference — case, whitespace, ordering, a stray `0x` prefix — breaks the check silently. It doesn't surface as an obvious bug; it surfaces as an opaque, "correct-looking" authentication failure. The durable fix is not "remember to normalize everywhere" but centralizing the normalization in the one function every caller is forced to go through.

---

## 4. Blockchain nonce management under concurrency

### The problem

Ethereum requires each transaction from an account to carry a strictly sequential nonce. Under any concurrency — device registration, activation/deactivation, and Geth node auto-registration all happening around the same time — naive nonce handling (e.g. "read `PendingNonceAt`, use it, increment locally") races: two goroutines can read the same nonce, one transaction gets rejected, and recovering incorrectly compounds the problem into cascading nonce gaps.

### The fix

- **Single-owner serializer** (`NonceSerializer`, `nonce_manager.go`): one goroutine owns the in-memory nonce counter exclusively; every request for a nonce goes through a channel to this goroutine. No two callers can ever be handed the same nonce.
- **`TxSemaphore`** (capacity 1, `device_ops.go`): serializes all *write* transactions across the whole backend (registration, activation, deactivation), so failures from one operation can't corrupt the nonce state of a concurrently in-flight one. Deliberately **not** applied to routine authentication recording, which is too frequent to serialize globally and has its own retry logic.
- **Two distinct recovery strategies**, chosen based on *why* a transaction failed:
  - `ReleaseNonceSync` (synchronous, ~500–700ms): for errors returned by go-ethereum *before* the transaction reaches the mempool (`nonce too low`, `already known`). Geth never saw the transaction, so there is nothing to purge — a fast resync is correct and safe.
  - `ReleaseNonce` (asynchronous, ~2s): for `replacement transaction underpriced`, where the *previous* transaction may still be sitting in the mempool. Geth needs time to purge it before the same nonce can be reused; resyncing too early would just race into the same conflict again.
- **Gas price bump on retry** (+15% per attempt) to avoid repeatedly hitting `replacement underpriced` on the same stuck nonce.
- **Circuit breakers** around blockchain and ML calls to stop cascading failures from a degraded dependency (open → half-open → closed), instead of every request queuing up against a service that is already failing.

### Why this distinction matters

Treating both failure classes the same way — either always resyncing fast or always waiting — either reintroduces the race (resyncing too early while Geth still holds the stale transaction) or wastes time unnecessarily (waiting 2 seconds for a transaction that was never submitted). Diagnosing *which* failure occurred, from the error string returned by go-ethereum, and routing to the matching recovery path, was what actually stopped the intermittent nonce corruption seen under load.

---

## Status of each fix

| Issue | Status | Verified by |
|---|---|---|
| Groth16 master-secret single point of compromise | Fixed — architecture replaced | Manual review; old code paths physically removed |
| Identity-squatting at registration | Fixed | Dedicated fresh-device end-to-end test (`IOT_0004`) |
| Case-sensitivity signature mismatch | Fixed | Independent client/server hash comparison, 0/N → N/N match rate |
| Nonce management under concurrency | Fixed | Load testing with concurrent registration/activation/deactivation |

None of these fixes required a third-party security audit to identify — they surfaced through direct testing (a device registering for the first time, load-testing concurrent writes) and through architectural review of what the server could and couldn't derive. That said, a formal external review is still recommended before any production deployment, as noted in the main README.
