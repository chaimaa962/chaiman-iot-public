#!/bin/bash
# reset-and-start.sh — Réinitialisation complète + démarrage blockchain PoA
# CHAIMAN IoT — Réseau 4 validators + 4 full nodes

set -euo pipefail

PROJECT_DIR="$HOME/iot-docker-project"
cd "$PROJECT_DIR"
GETH="ethereum/client-go:v1.10.26"

echo "═══════════════════════════════════════════════════════════════"
echo "🔄 RÉINITIALISATION COMPLÈTE + DÉMARRAGE"
echo "   4 Validators + 4 Full Nodes (PoA)"
echo "═══════════════════════════════════════════════════════════════"

# ============================================
# 1. ARRÊT TOTAL
# ============================================
echo ""
echo "🛑 Arrêt de tous les conteneurs..."
docker stop $(docker ps -aq --filter "name=^v" --filter "name=^fn") 2>/dev/null || true
docker rm $(docker ps -aq --filter "name=^v" --filter "name=^fn") 2>/dev/null || true

# ============================================
# 2. SUPPRESSION COMPLÈTE DES DONNÉES
# ============================================
echo ""
echo "🗑️  Suppression complète des données blockchain..."
sudo rm -rf ./nodes/validator*/geth
sudo rm -rf ./nodes/fullnode*/geth
sudo rm -f ./genesis.json
echo "   ✓ Données supprimées"

# ============================================
# 3. GÉNÉRATION DES CLÉS (si keystore manquant)
# ============================================
echo ""
echo "🔐 Vérification / génération des clés..."

generate_key_if_missing() {
    local node_dir="$1"
    if [ ! -f "$node_dir/keystore/UTC--"* ]; then
        echo "   📝 Génération clé pour $(basename $node_dir)..."
        mkdir -p "$node_dir"
        docker run --rm \
            -v "$PWD/$node_dir:/root/.ethereum" \
            $GETH account new --datadir=/root/.ethereum --password=/dev/null 2>/dev/null
        echo "   ✓ Clé générée"
    else
        echo "   ✓ Clé existante pour $(basename $node_dir)"
    fi
}

for i in 1 2 3 4; do
    generate_key_if_missing "nodes/validator${i}"
    generate_key_if_missing "nodes/fullnode${i}"
done

# ============================================
# 4. EXTRACTION DES ADRESSES
# ============================================
echo ""
echo "📋 Extraction des adresses..."

extract_address() {
    local node_dir="$1"
    local file=$(ls "$node_dir/keystore/UTC--"* 2>/dev/null | head -1)
    if [ -z "$file" ] || [ ! -s "$file" ]; then
        echo "ERREUR: Keystore vide ou manquant dans $node_dir" >&2
        exit 1
    fi
    python3 -c "import json,sys; print(json.load(open('$file'))['address'])" 2>/dev/null || \
    cat "$file" | python3 -c "import json,sys; print(json.load(sys.stdin)['address'])"
}

V1=$(extract_address nodes/validator1)
V2=$(extract_address nodes/validator2)
V3=$(extract_address nodes/validator3)
V4=$(extract_address nodes/validator4)

FN1=$(extract_address nodes/fullnode1)
FN2=$(extract_address nodes/fullnode2)
FN3=$(extract_address nodes/fullnode3)
FN4=$(extract_address nodes/fullnode4)

echo "   V1: 0x$V1"
echo "   V2: 0x$V2"
echo "   V3: 0x$V3"
echo "   V4: 0x$V4"
echo "   FN1: 0x$FN1"
echo "   FN2: 0x$FN2"
echo "   FN3: 0x$FN3"
echo "   FN4: 0x$FN4"

# ============================================
# 5. CRÉATION GENESIS
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
    "0x${FN1}": { "balance": "500000000000000000000" },
    "0x${FN2}": { "balance": "500000000000000000000" },
    "0x${FN3}": { "balance": "500000000000000000000" },
    "0x${FN4}": { "balance": "500000000000000000000" }
  }
}
EOF
echo "   ✓ genesis.json créé"

# ============================================
# 6. INITIALISATION GETH
# ============================================
echo ""
echo "🔧 Initialisation Geth..."

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

# ============================================
# 7. LANCEMENT DES NŒUDS
# ============================================
echo ""
echo "⛏️  Lancement des 4 validateurs..."

for i in 1 2 3 4; do
    PORT=$((8540+i)); P2P=$((30300+i)); IP="172.17.0.$((i+1))"
    
    docker run -d --name "v${i}" --restart unless-stopped \
        -p ${PORT}:8545 -p ${P2P}:30303 \
        -v "$PWD/nodes/validator${i}:/root/.ethereum" $GETH \
        --datadir=/root/.ethereum --networkid=1337 \
        --http --http.addr=0.0.0.0 --http.port=8545 \
        --http.api=eth,web3,net,admin,personal,miner,txpool,clique,debug \
        --http.corsdomain="*" --http.vhosts="*" \
        --port=30303 --nat=extip:${IP} --discovery.port=30303 \
        --mine --miner.threads=1 \
        --miner.gaslimit=8000000 \
        --miner.etherbase="0x${ADDR}" --unlock="0x${ADDR}" \
        --password=/dev/null --allow-insecure-unlock --verbosity=3 2>/dev/null
    echo "   ✓ v${i} : $IP (port $PORT) - 0x${ADDR}"
done

echo ""
echo "📡 Lancement des 4 full nodes..."

for i in 1 2 3 4; do
    PORT=$((8552+i)); P2P=$((30312+i)); IP="172.17.0.$((13+i))"
    
    docker run -d --name "fn${i}" --restart unless-stopped \
        -p ${PORT}:8545 -p ${P2P}:30303 \
        -v "$PWD/nodes/fullnode${i}:/root/.ethereum" $GETH \
        --datadir=/root/.ethereum --networkid=1337 \
        --http --http.addr=0.0.0.0 --http.port=8545 \
        --http.api=eth,web3,net,admin,personal,txpool,clique,debug \
        --http.corsdomain="*" --http.vhosts="*" \
        --port=30303 --nat=extip:${IP} --discovery.port=30303 \
        --unlock="0x${ADDR}" --password=/dev/null --allow-insecure-unlock --verbosity=3 2>/dev/null
    echo "   ✓ fn${i} : $IP (port $PORT) - 0x${ADDR}"
done

# ============================================
# 8. CONNEXION PEERS
# ============================================
echo ""
echo "⏳ Attente démarrage (20s)..."
sleep 20

echo ""
echo "🔗 Récupération des enodes..."

declare -A ENODES
declare -A NODE_PORTS

for i in 1 2 3 4; do
    PORT=$((8540+i))
    NODE_PORTS["v${i}"]=$PORT
    ENODE=$(curl -s -X POST http://localhost:$PORT -H "Content-Type: application/json" \
        -d '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' | jq -r '.result.enode')
    IP="172.17.0.$((i+1))"
    ENODE_FIXED=$(echo "$ENODE" | sed "s/\\[::\\]/${IP}/g" | sed 's/?discport=[0-9]*//g')
    ENODES["v${i}"]="$ENODE_FIXED"
    echo "   v${i}: ${ENODE_FIXED:0:70}..."
done

for i in 1 2 3 4; do
    PORT=$((8552+i))
    NODE_PORTS["fn${i}"]=$PORT
    ENODE=$(curl -s -X POST http://localhost:$PORT -H "Content-Type: application/json" \
        -d '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' | jq -r '.result.enode')
    IP="172.17.0.$((13+i))"
    ENODE_FIXED=$(echo "$ENODE" | sed "s/\\[::\\]/${IP}/g" | sed 's/?discport=[0-9]*//g')
    ENODES["fn${i}"]="$ENODE_FIXED"
    echo "   fn${i}: ${ENODE_FIXED:0:70}..."
done

echo ""
echo "🔗 Connexion manuelle des peers..."

for name1 in v1 v2 v3 v4 fn1 fn2 fn3 fn4; do
    PORT1=${NODE_PORTS[$name1]}
    for name2 in v1 v2 v3 v4 fn1 fn2 fn3 fn4; do
        if [ "$name1" != "$name2" ]; then
            ENODE2=${ENODES[$name2]}
            curl -s -X POST http://localhost:$PORT1 -H "Content-Type: application/json" \
                -d "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"$ENODE2\"],\"id\":1}" > /dev/null 2>&1
        fi
    done
done

echo ""
echo "⏳ Attente synchronisation (15s)..."
sleep 15

# ============================================
# 9. VÉRIFICATION
# ============================================
echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "📊 VÉRIFICATION"
echo "═══════════════════════════════════════════════════════════════"

echo ""
echo "--- Peers connectés ---"
for name in v1 v2 v3 v4 fn1 fn2 fn3 fn4; do
    PORT=${NODE_PORTS[$name]}
    PEERS=$(curl -s -X POST http://localhost:$PORT -H "Content-Type: application/json" \
        -d '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' | jq -r '.result')
    PEERS_DEC=$(printf "%d" "$PEERS" 2>/dev/null || echo "$PEERS")
    echo "   $name (port $PORT) -> $PEERS_DEC peers"
done

echo ""
echo "--- Bloc actuel ---"
for name in v1 v2 v3 v4 fn1 fn2 fn3 fn4; do
    PORT=${NODE_PORTS[$name]}
    BLOCK=$(curl -s -X POST http://localhost:$PORT -H "Content-Type: application/json" \
        -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
    BLOCK_DEC=$(printf "%d" "$BLOCK" 2>/dev/null || echo "$BLOCK")
    echo "   $name (port $PORT) -> Bloc: $BLOCK_DEC"
done

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "✅ BLOCKCHAIN DÉMARRÉE — Bloc 0"
echo "   Adresses enregistrées dans : $PROJECT_DIR/nodes/"
echo "═══════════════════════════════════════════════════════════════"
