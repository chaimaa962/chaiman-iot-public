#!/bin/bash
# reset-blockchain.sh — Réinitialisation TOTALE blockchain IoT
# Usage: ./reset-blockchain.sh

set -euo pipefail

cd ~/iot-docker-project
GETH="ethereum/client-go:v1.10.26"

echo "═══════════════════════════════════════════════════════════════"
echo "🗑️  RÉINITIALISATION COMPLÈTE — Blockchain remise à zéro"
echo "═══════════════════════════════════════════════════════════════"

# 1. ARRÊT & SUPPRESSION CONTENEURS
echo ""
echo "🛑 Arrêt des nœuds..."
docker stop $(docker ps -aq --filter "name=^v" --filter "name=^fn") 2>/dev/null || true
docker rm $(docker ps -aq --filter "name=^v" --filter "name=^fn") 2>/dev/null || true

# 2. SUPPRESSION DATADIRS
echo "🗑️  Suppression des données blockchain..."
sudo rm -rf ./nodes/validator*/geth ./nodes/validator*/keystore
sudo rm -rf ./nodes/fullnode*/geth ./nodes/fullnode*/keystore
sudo rm -f ./genesis.json

# 3. CRÉATION RÉPERTOIRES
echo "📁 Création des répertoires..."
for i in 1 2 3 4; do
    mkdir -p nodes/validator${i} nodes/fullnode${i}
done

# 4. GÉNÉRATION CLÉS (via Docker geth account new)
echo ""
echo "🔑 Génération des clés..."

generate_key() {
    local datadir=$1
    local name=$2
    # geth account new avec mot de passe vide (dev only)
    docker run --rm -v "$PWD/$datadir:/root/.ethereum" $GETH \
        account new --password=/dev/null --datadir=/root/.ethereum 2>/dev/null | \
        grep "Public address" | awk '{print $5}'
}

V1=$(generate_key nodes/validator1 "validator1")
echo "   ✓ validator1 — $V1"
V2=$(generate_key nodes/validator2 "validator2")
echo "   ✓ validator2 — $V2"
V3=$(generate_key nodes/validator3 "validator3")
echo "   ✓ validator3 — $V3"
V4=$(generate_key nodes/validator4 "validator4")
echo "   ✓ validator4 — $V4"

FN1=$(generate_key nodes/fullnode1 "fullnode1")
echo "   ✓ fullnode1  — $FN1"
FN2=$(generate_key nodes/fullnode2 "fullnode2")
echo "   ✓ fullnode2  — $FN2"
FN3=$(generate_key nodes/fullnode3 "fullnode3")
echo "   ✓ fullnode3  — $FN3"
FN4=$(generate_key nodes/fullnode4 "fullnode4")
echo "   ✓ fullnode4  — $FN4"

# 5. GÉNÉRATION GENESIS
echo ""
echo "📝 Génération genesis.json..."

# Extraire les adresses sans préfixe 0x pour extradata
V1_HEX=${V1#0x}
V2_HEX=${V2#0x}
V3_HEX=${V3#0x}
V4_HEX=${V4#0x}

cat > genesis.json << EOF
{
  "config": {
    "chainId": 1337,
    "homesteadBlock": 0,
    "eip150Block": 0,
    "eip155Block": 0,
    "eip158Block": 0,
    "byzantiumBlock": 0,
    "constantinopleBlock": 0,
    "petersburgBlock": 0,
    "istanbulBlock": 0,
    "clique": {
      "period": 5,
      "epoch": 30000
    }
  },
  "difficulty": "1",
  "gasLimit": "8000000",
  "extradata": "0x0000000000000000000000000000000000000000000000000000000000000000${V1_HEX}${V2_HEX}${V3_HEX}${V4_HEX}0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
  "alloc": {
    "${V1}": { "balance": "1000000000000000000000" },
    "${V2}": { "balance": "1000000000000000000000" },
    "${V3}": { "balance": "1000000000000000000000" },
    "${V4}": { "balance": "1000000000000000000000" },
    "${FN1}": { "balance": "500000000000000000000" },
    "${FN2}": { "balance": "500000000000000000000" },
    "${FN3}": { "balance": "500000000000000000000" },
    "${FN4}": { "balance": "500000000000000000000" }
  }
}
EOF

echo "   ✓ genesis.json créé"

# 6. INIT GETH POUR TOUS LES NŒUDS
echo ""
echo "⚙️  Initialisation geth..."

for i in 1 2 3 4; do
    docker run --rm \
      -v "$PWD/nodes/validator${i}:/root/.ethereum" \
      -v "$PWD/genesis.json:/genesis.json" \
      $GETH init --datadir=/root/.ethereum /genesis.json 2>/dev/null
    echo "   ✓ validator${i} initialisé"
done

for i in 1 2 3 4; do
    docker run --rm \
      -v "$PWD/nodes/fullnode${i}:/root/.ethereum" \
      -v "$PWD/genesis.json:/genesis.json" \
      $GETH init --datadir=/root/.ethereum /genesis.json 2>/dev/null
    echo "   ✓ fullnode${i} initialisé"
done

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "✅ RÉINITIALISATION TERMINÉE"
echo "   Blockchain remise à zéro — prêt pour start-node.sh"
echo "═══════════════════════════════════════════════════════════════"
