# CHAIMAN IoT — Secure IoT Device Authentication on Blockchain

A backend system for registering, authenticating, and monitoring IoT devices using a private Ethereum blockchain, a Schnorr zero-knowledge proof protocol, and an ML-based anomaly detection engine.

![status](https://img.shields.io/badge/status-prototype-orange)
![go](https://img.shields.io/badge/backend-Go-00ADD8)
![ml](https://img.shields.io/badge/anomaly--detection-Python%2FFastAPI-3776AB)
![blockchain](https://img.shields.io/badge/blockchain-Ethereum%20(PoA)-3C3C3D)

---

## What this project demonstrates

This is not a toy demo — it's a working system that was audited and hardened through several real security fixes, documented as they happened:

- **Custom Schnorr ZKP protocol** (secp256k1, Fiat-Shamir) implemented from scratch in Go, migrating away from a centralized Groth16 setup where the server could derive any device's secret from a single master key — a single point of compromise.
- **A critical case-sensitivity bug** in the registration flow (mismatched Keccak256 hash between device and server) that caused 100% registration failures — root-caused and fixed at the shared protocol level.
- **An identity-squatting vulnerability**: proof-of-possession was optional at registration, allowing an attacker to register a fabricated Schnorr key for an address before its legitimate owner. Fixed by making the ECDSA proof-of-possession mandatory and server-derived.
- **Blockchain nonce management** under concurrency: a single-owner nonce serializer, distinct retry strategies for "nonce too low" vs "replacement underpriced," and circuit breakers to prevent cascading failures when the blockchain or ML service is degraded.

See [`CASE_STUDY.md`](CASE_STUDY.md) for the full write-up of these fixes.

---

## Architecture

```
┌─────────────────┐      ┌──────────────────────┐      ┌─────────────────────┐
│   IoT Device     │      │   Go Backend          │      │  Ethereum (PoA)      │
│  (simulator.go)  │─────▶│  - ECDSA signature     │─────▶│  IoTDeviceManager    │
│  - ECDSA sign    │      │  - Schnorr ZKP verify  │      │  smart contract      │
│  - Schnorr proof │      │  - Rate limiting       │      └─────────────────────┘
│    (local only)  │      │  - Circuit breakers    │
└─────────────────┘      └──────────┬───────────┘
                                     │
                                     ▼
                          ┌──────────────────────┐
                          │  ML Anomaly Engine     │
                          │  (FastAPI + Keras)     │
                          │  Autoencoder on        │
                          │  network features      │
                          └──────────┬───────────┘
                                     │
                                     ▼
                          ┌──────────────────────┐
                          │  React Dashboard       │
                          │  Real-time monitoring  │
                          └──────────────────────┘
```

---

## Key components

| Component | Path | Responsibility |
|---|---|---|
| Schnorr ZKP protocol | `pkg/zkp/Schnorr.go`, `schnorr_device.go` | Proof generation/verification, key persistence |
| Device registration & auth | `internal/handlers/device.go`, `messages_secure.go` | Proof-of-possession, ECDSA + Schnorr double auth |
| Blockchain client | `internal/blockchain/client.go`, `nonce_manager.go` | Encrypted key in memory, nonce serialization |
| ML anomaly detection | `detection_dev.py` | Autoencoder inference, adaptive threshold logic |
| Device simulator | `iot_simulator.go` | End-to-end test client mimicking real IoT devices |
| Dashboard | `Dashboard.js` | Live metrics, transaction feed, anomaly alerts |

---

## Getting started (local demo)

> This is a development/demo setup. **Do not use the default secrets or test keys in any real deployment.**

```bash
# 1. Start the local PoA blockchain + backend services
docker-compose up -d

# 2. Run the ML anomaly detection service
cd iot-backend/ai
python detection_dev.py

# 3. Run the device simulator (registers 4 devices, starts authenticating)
go run iot-backend/simulators/iot_simulator.go -backend=http://localhost:8080

# 4. Open the dashboard
cd iot-dashboard && npm start
```

Required environment variables are documented in `.env.example` (see `internal/config`).

---

## Security notes

- Secrets (blockchain private key, admin API key, AI backend signature secret) are loaded from environment variables or HashiCorp Vault (`internal/config/vault.go`) — never hardcoded in production paths.
- Device private keys (Schnorr) never leave the device; only public keys are transmitted and stored server-side.
- All authentication requests require **both** a valid ECDSA signature (identity) and a valid Schnorr proof (possession of the registered secret) — see `messages_secure.go`.

---

## Status

Prototype / demo-stage project, used as a technical showcase. Some hardening steps documented in `CASE_STUDY.md` are applied; production deployment would additionally require TLS certificate automation, secret rotation, and a formal third-party security audit.
Related links :
AI anomaly detection (Kaggle Notebook): https://www.kaggle.com/code/benadl/iot-anomaly-detection

---

## Contact

Available for freelance work on: blockchain security audits, Go backend development, IoT authentication systems, and cryptographic protocol design/migration.

📧 **chaimaa.benadla@gmail.com**
