#!/bin/bash

echo "════════════════════════════════════════════════════════════════"
echo "🚀 DÉMARRAGE SYSTÈME IoT + PoA + IA + ZKP (VERSION FIXED)"
echo "════════════════════════════════════════════════════════════════"

cd ~/iot-docker-project
mkdir -p logs
GETH="ethereum/client-go:v1.10.26"

echo "📦 1/3 - Démarrage PoA (8 nœuds)..."

for i in 1 2 3 4; do
    PORT=$((8540+i))
    P2P=$((30300+i))
    IP="172.17.0.$((i+1))"
    ADDR=$(cat nodes/validator${i}/address.txt 2>/dev/null | tr -d '\n')
    if [ -z "$ADDR" ]; then
        ADDR=$(ls nodes/validator${i}/keystore/UTC--* 2>/dev/null | head -1 | xargs cat 2>/dev/null | python3 -c "import json,sys; print(json.load(sys.stdin)['address'])" 2>/dev/null)
    fi
    echo "   v${i} : adresse 0x${ADDR}"
    
    docker run -d --name "v${i}" --restart unless-stopped \
      -p ${PORT}:8545 -p ${P2P}:30303 \
      -v "$PWD/nodes/validator${i}:/root/.ethereum" \
      -v "$PWD/genesis.json:/root/genesis.json" $GETH \
      init /root/genesis.json 2>/dev/null
    
    docker run -d --name "v${i}" --restart unless-stopped \
      -p ${PORT}:8545 -p ${P2P}:30303 \
      -v "$PWD/nodes/validator${i}:/root/.ethereum" $GETH \
      --datadir=/root/.ethereum --networkid=2026 \
      --http --http.addr=0.0.0.0 --http.port=8545 \
      --http.api=eth,web3,net,admin,personal,miner,txpool,clique,debug \
      --http.corsdomain="*" \
      --port=30303 --nat=extip:${IP} --discovery.port=30303 \
      --mine --miner.threads=1 --miner.gaslimit=8000000 \
      --miner.etherbase="0x${ADDR}" --unlock="0x${ADDR}" \
      --password=/dev/null --allow-insecure-unlock --verbosity=4
done

echo "   ⏳ Attente démarrage..."
sleep 20

echo "⛏️  Démarrage du minage..."
for port in 8541 8542 8543 8544; do
    curl -X POST http://localhost:$port \
      -H "Content-Type: application/json" \
      --data '{"jsonrpc":"2.0","method":"miner_start","params":[1],"id":1}' 2>/dev/null
    echo "   Minage démarré sur port $port"
done

sleep 5
BLOCK=$(curl -s -X POST http://localhost:8541 -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
echo "   📊 Bloc actuel: $BLOCK"

echo ""
echo "✅ SYSTÈME DÉMARRÉ !"
echo "   PoA: http://localhost:8541"
