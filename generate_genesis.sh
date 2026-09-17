#!/bin/bash

# Script pour générer genesis.json avec les adresses des nœuds existants

echo "🔍 Recherche des adresses des nœuds existants..."

# Récupérer les adresses des nœuds existants
ADDRESSES=""
declare -a ADDR_LIST

for i in $(seq 1 20); do
    if [ -d "node${i}_data/keystore" ] && [ -f "node${i}_data/keystore/"* ]; then
        ADDR=$(sudo cat node${i}_data/keystore/* 2>/dev/null | grep -o '"address":"[^"]*"' | cut -d'"' -f4)
        if [ -n "$ADDR" ]; then
            ADDR_LIST+=("$ADDR")
            echo "  ✅ Node$i: $ADDR"
        fi
    fi
done

NODE_COUNT=${#ADDR_LIST[@]}

if [ $NODE_COUNT -eq 0 ]; then
    echo "❌ Aucun nœud trouvé. Veuillez d'abord créer des nœuds."
    exit 1
fi

echo ""
echo "📊 $NODE_COUNT nœuds trouvés"

# Construire l'extradata
EXTRA_PREFIX="0x0000000000000000000000000000000000000000000000000000000000000000"
EXTRA_SUFFIX="00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"

EXTRA_DATA="$EXTRA_PREFIX"
for addr in "${ADDR_LIST[@]}"; do
    EXTRA_DATA="${EXTRA_DATA}${addr#0x}"
done
EXTRA_DATA="${EXTRA_DATA}${EXTRA_SUFFIX}"

# Générer le genesis.json
cat > genesis.json << GENEOF
{
  "config": {
    "chainId": 1234,
    "homesteadBlock": 0,
    "eip150Block": 0,
    "eip155Block": 0,
    "eip158Block": 0,
    "byzantiumBlock": 0,
    "constantinopleBlock": 0,
    "petersburgBlock": 0,
    "istanbulBlock": 0,
    "berlinBlock": 0,
    "londonBlock": 0,
    "clique": {
      "period": 5,
      "epoch": 30000
    }
  },
  "difficulty": "1",
  "gasLimit": "8000000",
  "extradata": "${EXTRA_DATA}",
  "alloc": {
GENEOF

# Ajouter les allocations
for i in "${!ADDR_LIST[@]}"; do
    addr="${ADDR_LIST[$i]}"
    if [ $i -eq $((${#ADDR_LIST[@]} - 1)) ]; then
        echo "    \"${addr}\": {\"balance\": \"1000000000000000000000000\"}" >> genesis.json
    else
        echo "    \"${addr}\": {\"balance\": \"1000000000000000000000000\"}," >> genesis.json
    fi
done

cat >> genesis.json << GENEOF
  }
}
GENEOF

echo ""
echo "✅ Genesis.json généré avec succès !"
echo "   📝 Fichier: genesis.json"
echo "   👥 Nombre de signeurs: $NODE_COUNT"
