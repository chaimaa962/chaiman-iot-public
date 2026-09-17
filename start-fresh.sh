#!/bin/bash
set -e

cd ~/iot-docker-project
GETH="ethereum/client-go:v1.10.26"

# ============================================
# FONCTION : Extraire l'adresse du DERNIER keystore
# ============================================
get_last_address() {
    local dir=$1
    local latest=$(ls -t "$dir/keystore/UTC--"* 2>/dev/null | head -1)
    if [ -z "$latest" ]; then
        echo "ERREUR: Aucun keystore dans $dir" >&2
        exit 1
    fi
    python3 -c "import json,sys; print(json.load(open('$latest'))['address'])"
}

# ============================================
# 1. EXTRAIRE LES ADRESSES
# ============================================
echo "🔐 Extraction des adresses (dernier keystore)..."

V1=$(get_last_address nodes/validator1)
V2=$(get_last_address nodes/validator2)
V3=$(get_last_address nodes/validator3)
V4=$(get_last_address nodes/validator4)
F1=$(get_last_address nodes/fullnode1)
F2=$(get_last_address nodes/fullnode2)
F3=$(get_last_address nodes/fullnode3)
F4=$(get_last_address nodes/fullnode4)

echo "  V1=0x$V1"
echo "  V2=0x$V2"
echo "  V3=0x$V3"
echo "  V4=0x$V4"
echo "  F1=0x$F1"
echo "  F2=0x$F2"
echo "  F3=0x$F3"
echo "  F4=0x$F4"

# ============================================
# 2. CRÉER GENESIS.JSON
# ============================================
echo ""
echo "📝 Création genesis.json..."

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
  "extradata": "0x0000000000000000000000000000000000000000000000000000000000000000${V1}${V2}${V3}${V4}0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
  "alloc": {
    "0x${V1}": { "balance": "1000000000000000000000" },
    "0x${V2}": { "balance": "1000000000000000000000" },
    "0x${V3}": { "balance": "1000000000000000000000" },
    "0x${V4}": { "balance": "1000000000000000000000" },
    "0x${F1}": { "balance": "500000000000000000000" },
    "0x${F2}": { "balance": "500000000000000000000" },
    "0x${F3}": { "balance": "500000000000000000000" },
    "0x${F4}": { "balance": "500000000000000000000" }
  }
}
EOF

echo "  ✓ genesis.json créé"

# ============================================
# 3. INITIALISER LES NŒUDS
# ============================================
echo ""
echo "🗄️  Initialisation des datadirs..."

for i in 1 2 3 4; do
    docker run --rm \
      -v "$PWD/nodes/validator${i}:/root/.ethereum" \
      -v "$PWD/genesis.json:/genesis.json" \
      $GETH init --datadir=/root/.ethereum /genesis.json 2>/dev/null
    echo "  ✓ validator${i} initialisé"
done

for i in 1 2 3 4; do
    docker run --rm \
      -v "$PWD/nodes/fullnode${i}:/root/.ethereum" \
      -v "$PWD/genesis.json:/genesis.json" \
      $GETH init --datadir=/root/.ethereum /genesis.json 2>/dev/null
    echo "  ✓ fullnode${i} initialisé"
done

# ============================================
# 4. LANCER LES VALIDATEURS
# ============================================
echo ""
echo "⛏️  Lancement des validateurs..."

for i in 1 2 3 4; do
    PORT=$((8540+i)); P2P=$((30300+i)); IP="172.17.0.$((i+1))"
    ADDR=$(get_last_address nodes/validator${i})
    
    docker run -d --name "v${i}" --restart unless-stopped \
      -p ${PORT}:8545 -p ${P2P}:30303 \
      -v "$PWD/nodes/validator${i}:/root/.ethereum" $GETH \
      --datadir=/root/.ethereum --networkid=1337 \
      --http --http.addr=0.0.0.0 --http.port=8545 \
      --http.api=eth,web3,net,admin,personal,miner,txpool,clique,debug \
      --http.corsdomain="*" --http.vhosts="*" \
      --port=30303 --nat=extip:${IP} --discovery.port=30303 \
      --mine --miner.threads=1 --miner.gaslimit=8000000 \
      --miner.etherbase="0x${ADDR}" --unlock="0x${ADDR}" \
      --password=/dev/null --allow-insecure-unlock --verbosity=3 2>/dev/null
    
    echo "  ✓ v${i} -> port $PORT | 0x${ADDR}"
done

# ============================================
# 5. LANCER LES FULL NODES
# ============================================
echo ""
echo "📡 Lancement des full nodes..."

for i in 1 2 3 4; do
    PORT=$((8552+i)); P2P=$((30312+i)); IP="172.17.0.$((13+i))"
    ADDR=$(get_last_address nodes/fullnode${i})
    
    docker run -d --name "fn${i}" --restart unless-stopped \
      -p ${PORT}:8545 -p ${P2P}:30303 \
      -v "$PWD/nodes/fullnode${i}:/root/.ethereum" $GETH \
      --datadir=/root/.ethereum --networkid=1337 \
      --http --http.addr=0.0.0.0 --http.port=8545 \
      --http.api=eth,web3,net,admin,personal,txpool,clique,debug \
      --http.corsdomain="*" --http.vhosts="*" \
      --port=30303 --nat=extip:${IP} --discovery.port=30303 \
      --unlock="0x${ADDR}" --password=/dev/null --allow-insecure-unlock --verbosity=3 2>/dev/null
    
    echo "  ✓ fn${i} -> port $PORT | 0x${ADDR}"
done

# ============================================
# 6. ATTENDRE ET CONNECTER LES PEERS
# ============================================
echo ""
echo "⏳ Attente démarrage (20s)..."
sleep 20

echo ""
echo "🔗 Récupération des enodes..."

# Stocker les enodes dans un fichier temporaire
> /tmp/enodes.txt

for i in 1 2 3 4; do
    PORT=$((8540+i))
    ENODE=$(curl -s -X POST http://localhost:$PORT -H "Content-Type: application/json" \
      -d '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' | jq -r '.result.enode')
    IP="172.17.0.$((i+1))"
    ENODE_FIXED=$(echo "$ENODE" | sed "s/\\[::\\]/${IP}/g" | sed 's/?discport=[0-9]*//g')
    echo "$ENODE_FIXED" >> /tmp/enodes.txt
    echo "  v${i}: ${ENODE_FIXED:0:65}..."
done

for i in 1 2 3 4; do
    PORT=$((8552+i))
    ENODE=$(curl -s -X POST http://localhost:$PORT -H "Content-Type: application/json" \
      -d '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' | jq -r '.result.enode')
    IP="172.17.0.$((13+i))"
    ENODE_FIXED=$(echo "$ENODE" | sed "s/\\[::\\]/${IP}/g" | sed 's/?discport=[0-9]*//g')
    echo "$ENODE_FIXED" >> /tmp/enodes.txt
    echo "  fn${i}: ${ENODE_FIXED:0:65}..."
done

echo ""
echo "🔗 Connexion des peers..."

for PORT in 8541 8542 8543 8544 8553 8554 8555 8556; do
    while IFS= read -r ENODE; do
        curl -s -X POST http://localhost:$PORT -H "Content-Type: application/json" \
          -d "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"$ENODE\"],\"id\":1}" > /dev/null
    done < /tmp/enodes.txt
    echo "  ✓ Port $PORT connecté"
done

# ============================================
# 7. VÉRIFICATION
# ============================================
echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "📊 VÉRIFICATION"
echo "═══════════════════════════════════════════════════════════════"

sleep 5

echo ""
echo "--- Peers ---"
for name in v1 v2 v3 v4 fn1 fn2 fn3 fn4; do
    if [[ $name == v* ]]; then
        PORT=$((8540 + ${name:1}))
    else
        PORT=$((8552 + ${name:2}))
    fi
    PEERS=$(curl -s -X POST http://localhost:$PORT -H "Content-Type: application/json" \
      -d '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' | jq -r '.result')
    PEERS_DEC=$(printf "%d" "$PEERS" 2>/dev/null || echo "0")
    echo "  $name (port $PORT) -> $PEERS_DEC peers"
done

echo ""
echo "--- Blocs ---"
for name in v1 v2 v3 v4 fn1 fn2 fn3 fn4; do
    if [[ $name == v* ]]; then
        PORT=$((8540 + ${name:1}))
    else
        PORT=$((8552 + ${name:2}))
    fi
    BLOCK=$(curl -s -X POST http://localhost:$PORT -H "Content-Type: application/json" \
      -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
    # Convertir hex en decimal avec python3
    BLOCK_DEC=$(python3 -c "print(int('$BLOCK', 16))")
    echo "  $name (port $PORT) -> Bloc: $BLOCK_DEC"
done

echo ""
echo "--- Minage ---"
for i in 1 2 3 4; do
    PORT=$((8540+i))
    MINING=$(curl -s -X POST http://localhost:$PORT -H "Content-Type: application/json" \
      -d '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' | jq -r '.result')
    echo "  v${i} -> Mining: $MINING"
done

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "✅ RÉSEAU POA DÉMARRÉ"
echo "═══════════════════════════════════════════════════════════════"
