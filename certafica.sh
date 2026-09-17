#!/bin/bash
# ============================================================
# CHAIMAN IoT — Script d'initialisation securisee
# Genere: genesis.json, certificats TLS, JWT secret, bootnode key
# ============================================================

set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_DIR"

echo "🔐 Initialisation du reseau blockchain CHAIMAN IoT"
echo "============================================================"

# --- 1. Repertoires ---
mkdir -p nodes/{validator1,validator2,validator3,validator4,fullnode1,fullnode2,fullnode3,fullnode4}
mkdir -p nginx/ssl
mkdir -p prometheus
mkdir -p certs

# --- 2. JWT Secret pour Auth RPC ---
echo "📄 Generation du JWT secret..."
openssl rand -hex 32 > jwt-secret
chmod 600 jwt-secret

# --- 3. Bootnode Key ---
echo "📄 Generation de la cle bootnode..."
if ! command -v bootnode &> /dev/null; then
    echo "⚠️  bootnode non trouve. Utilisation d'une cle aleatoire."
    openssl rand -hex 64 > bootnode.key
else
    bootnode -genkey bootnode.key
fi
chmod 600 bootnode.key

# --- 4. Certificats TLS (auto-signes pour dev, remplacer par CA en prod) ---
echo "📄 Generation des certificats TLS..."

# CA interne
openssl req -x509 -newkey rsa:4096 -keyout nginx/ssl/ca.key -out nginx/ssl/ca.crt \
    -days 3650 -nodes -subj "/CN=CHAIMAN-IoT-CA/O=CHAIMAN/C=MA" \
    -addext "basicConstraints=critical,CA:TRUE" \
    -addext "keyUsage=critical,keyCertSign,cRLSign"

# Server cert (1 an)
openssl req -newkey rsa:4096 -keyout nginx/ssl/server.key -out nginx/ssl/server.csr \
    -nodes -subj "/CN=iot-backend.chaiman.local/O=CHAIMAN/C=MA" \
    -addext "subjectAltName=DNS:localhost,DNS:iot-backend,DNS:iot-backend.chaiman.local,IP:127.0.0.1"

openssl x509 -req -in nginx/ssl/server.csr -CA nginx/ssl/ca.crt -CAkey nginx/ssl/ca.key \
    -CAcreateserial -out nginx/ssl/server.crt -days 365 \
    -extfile <(printf "subjectAltName=DNS:localhost,DNS:iot-backend,DNS:iot-backend.chaiman.local,IP:127.0.0.1")

chmod 600 nginx/ssl/*.key
rm -f nginx/ssl/server.csr nginx/ssl/ca.srl

echo "✅ Certificat serveur valide 365 jours (1 an)"
echo "   Renouvellement: $(date -d '+365 days' '+%Y-%m-%d')"

# --- 5. Fichiers password pour les keystores Geth ---
echo "📄 Generation des mots de passe keystore..."
for i in {1..4}; do
    openssl rand -base64 32 > "nodes/validator${i}/password.txt"
    chmod 600 "nodes/validator${i}/password.txt"
done

# --- 6. Verification ---
echo ""
echo "============================================================"
echo "✅ Initialisation terminee !"
echo "============================================================"
echo ""
echo "Prochaines etapes:"
echo "  1. Copier votre genesis.json dans ce repertoire"
echo "  2. Initialiser les datadirs:"
echo "     for i in {1..4}; do docker run --rm -v \"$(pwd)/nodes/validator\$i:/data\" -v \"$(pwd)/genesis.json:/genesis.json\" ethereum/client-go:v1.14.0 init /genesis.json --datadir=/data; done"
echo "  3. Configurer .env avec vos valeurs: cp .env.example .env"
echo "  4. Lancer: docker compose -f docker-compose-hardened.yml up -d"
echo ""
echo "🔒 SECURITE: Les comptes validateurs ne sont PAS deverrouilles automatiquement."
echo "   Utilisez Clef ou deverrouillez manuellement via IPC securise."
echo ""
