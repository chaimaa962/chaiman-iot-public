#!/bin/bash

echo "═══════════════════════════════════════════════════════════════"
echo "🚀 DÉMARRAGE : 4 Validators + 4 Full Nodes (PoA)"
echo "═══════════════════════════════════════════════════════════════"

cd ~/iot-docker-project
GETH="ethereum/client-go:v1.10.26"

# ============================================
# FONCTION : Extraire l'adresse du DERNIER keystore
# ============================================
get_last_address() {
    local dir=$1
    local latest
    latest=$(ls -t "$dir/keystore/UTC--"* 2>/dev/null | head -1)
    if [ -z "$latest" ]; then
        echo "ERREUR: Aucun keystore dans $dir" >&2
        exit 1
    fi
    python3 -c "import json; print(json.load(open('$latest'))['address'])"
}

# ============================================
# 1. NETTOYAGE (conteneurs uniquement)
# ============================================
echo "🧹 Arrêt des conteneurs existants..."

docker stop $(docker ps -aq --filter "name=^v" --filter "name=^fn") 2>/dev/null || true
docker rm   $(docker ps -aq --filter "name=^v" --filter "name=^fn") 2>/dev/null || true

# ============================================
# 2. VÉRIFIER SI INITIALISATION NÉCESSAIRE
# ============================================
NEED_INIT=false

for i in 1 2 3 4; do
    if [ ! -d "./nodes/validator${i}/geth/chaindata" ] || [ ! -f "./nodes/validator${i}/geth/LOCK" ]; then
        NEED_INIT=true
        echo "   ⚠️  validator${i} nécessite initialisation"
    fi
    if [ ! -d "./nodes/fullnode${i}/geth/chaindata" ] || [ ! -f "./nodes/fullnode${i}/geth/LOCK" ]; then
        NEED_INIT=true
        echo "   ⚠️  fullnode${i} nécessite initialisation"
    fi
done

# ============================================
# 3. INITIALISATION (si premier démarrage ou reset)
# ============================================
if [ "$NEED_INIT" = true ]; then
    echo ""
    echo "📝 Initialisation des datadirs..."
    
    V1=$(get_last_address nodes/validator1)
    V2=$(get_last_address nodes/validator2)
    V3=$(get_last_address nodes/validator3)
    V4=$(get_last_address nodes/validator4)
    F1=$(get_last_address nodes/fullnode1)
    F2=$(get_last_address nodes/fullnode2)
    F3=$(get_last_address nodes/fullnode3)
    F4=$(get_last_address nodes/fullnode4)

    echo "   V1=0x$V1"
    echo "   V2=0x$V2"
    echo "   V3=0x$V3"
    echo "   V4=0x$V4"
    echo "   F1=0x$F1"
    echo "   F2=0x$F2"
    echo "   F3=0x$F3"
    echo "   F4=0x$F4"

    cat > genesis.json << GENEOF
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
GENEOF

    echo "   ✓ genesis.json créé"

    # Initialiser les nœuds manquants
    for i in 1 2 3 4; do
        if [ ! -d "./nodes/validator${i}/geth" ]; then
            docker run --rm \
              -v "$PWD/nodes/validator${i}:/root/.ethereum" \
              -v "$PWD/genesis.json:/genesis.json" \
              $GETH init --datadir=/root/.ethereum /genesis.json 2>/dev/null
            echo "   ✓ validator${i} initialisé"
        fi
    done
    for i in 1 2 3 4; do
        if [ ! -d "./nodes/fullnode${i}/geth" ]; then
            docker run --rm \
              -v "$PWD/nodes/fullnode${i}:/root/.ethereum" \
              -v "$PWD/genesis.json:/genesis.json" \
              $GETH init --datadir=/root/.ethereum /genesis.json 2>/dev/null
            echo "   ✓ fullnode${i} initialisé"
        fi
    done
else
    echo "   ✓ Tous les datadirs existent — pas d'initialisation nécessaire"
fi

# ============================================
# 4. LANCER LES 4 VALIDATEURS
# ============================================
echo ""
echo "⛏️  Lancement des 4 validateurs..."

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
      --mine --miner.threads=1 \
      --miner.gaslimit=8000000 \
      --miner.etherbase="0x${ADDR}" --unlock="0x${ADDR}" \
      --password=/dev/null --allow-insecure-unlock --verbosity=3 2>/dev/null
    echo "   ✓ v${i} : $IP (port $PORT) - 0x${ADDR}"
done

# ============================================
# 5. LANCER LES 4 FULL NODES
# ============================================
echo "📡 Lancement des 4 full nodes..."

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
    echo "   ✓ fn${i} : $IP (port $PORT) - 0x${ADDR}"
done

echo ""
echo "⏳ Attente démarrage (20s)..."
sleep 20

# ============================================
# 6. RÉCUPÉRER LES ENODES
# ============================================
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
    ENODE_FIXED=$(echo "$ENODE" | sed "s/\[::\]/${IP}/g" | sed 's/?discport=[0-9]*//g')
    ENODES["v${i}"]="$ENODE_FIXED"
    echo "   v${i}: ${ENODE_FIXED:0:70}..."
done

for i in 1 2 3 4; do
    PORT=$((8552+i))
    NODE_PORTS["fn${i}"]=$PORT
    ENODE=$(curl -s -X POST http://localhost:$PORT -H "Content-Type: application/json" \
      -d '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' | jq -r '.result.enode')
    
    IP="172.17.0.$((13+i))"
    ENODE_FIXED=$(echo "$ENODE" | sed "s/\[::\]/${IP}/g" | sed 's/?discport=[0-9]*//g')
    ENODES["fn${i}"]="$ENODE_FIXED"
    echo "   fn${i}: ${ENODE_FIXED:0:70}..."
done

# ============================================
# 7. CONNEXION MANUELLE DES PEERS
# ============================================
echo ""
echo "🔗 Connexion manuelle des peers..."

for name1 in v1 v2 v3 v4 fn1 fn2 fn3 fn4; do
    PORT1=${NODE_PORTS[$name1]}
    echo "   Connecting $name1 (port $PORT1)..."
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
# 8. VÉRIFICATION
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
    PEERS_DEC=$(python3 -c "print(int('$PEERS', 16))" 2>/dev/null || echo "0")
    echo "   $name (port $PORT) -> $PEERS_DEC peers"
done

echo ""
echo "--- Bloc actuel ---"
for name in v1 v2 v3 v4 fn1 fn2 fn3 fn4; do
    PORT=${NODE_PORTS[$name]}
    BLOCK=$(curl -s -X POST http://localhost:$PORT -H "Content-Type: application/json" \
      -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
    BLOCK_DEC=$(python3 -c "print(int('$BLOCK', 16))" 2>/dev/null || echo "0")
    echo "   $name (port $PORT) -> Bloc: $BLOCK_DEC"
done

echo ""
echo "--- Minage (validateurs) ---"
for i in 1 2 3 4; do
    PORT=$((8540+i))
    MINING=$(curl -s -X POST http://localhost:$PORT -H "Content-Type: application/json" \
      -d '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' | jq -r '.result')
    HASHRATE=$(curl -s -X POST http://localhost:$PORT -H "Content-Type: application/json" \
      -d '{"jsonrpc":"2.0","method":"eth_hashrate","params":[],"id":1}' | jq -r '.result')
    echo "   v${i} -> Mining: $MINING, Hashrate: $HASHRATE"
done

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "✅ TERMINÉ !"
if [ "$NEED_INIT" = true ]; then
    echo "   🆕 Nouvelle blockchain initialisée"
else
    echo "   🔄 Blockchain existante préservée"
fi
echo "═══════════════════════════════════════════════════════════════"
