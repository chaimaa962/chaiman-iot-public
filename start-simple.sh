#!/bin/bash

cd ~/iot-docker-project

echo "🚀 Démarrage PoA simple..."

# Arrêter et supprimer les anciens conteneurs
docker rm -f v1 v2 v3 v4 fn1 fn2 fn3 fn4 2>/dev/null

# Démarrer les validateurs
for i in 1 2 3 4; do
    PORT=$((8540+i))
    P2P=$((30300+i))
    ADDR=$(cat nodes/validator${i}/address.txt | tr -d '\n' | sed 's/0x//')
    
    echo "Démarrage v${i} (port $PORT) avec adresse 0x${ADDR}"
    
    docker run -d --name "v${i}" --restart unless-stopped \
      -p ${PORT}:8545 -p ${P2P}:30303 \
      -v "$PWD/nodes/validator${i}:/root/.ethereum" \
      -v "$PWD/genesis.json:/root/genesis.json" \
      ethereum/client-go:v1.10.26 \
      /bin/sh -c "geth init /root/genesis.json && geth --datadir=/root/.ethereum --networkid=2026 --http --http.addr=0.0.0.0 --http.port=8545 --http.api=eth,web3,net,admin,personal,miner,txpool,clique --http.corsdomain='*' --port=30303 --nat=extip:172.17.0.$((i+1)) --mine --miner.threads=1 --miner.etherbase=0x${ADDR} --unlock=0x${ADDR} --password=/dev/null --allow-insecure-unlock --verbosity=3"
done

echo "⏳ Attente 15 secondes..."
sleep 15

# Démarrer le minage
for port in 8541 8542 8543 8544; do
    curl -s -X POST http://localhost:$port \
      -H "Content-Type: application/json" \
      --data '{"jsonrpc":"2.0","method":"miner_start","params":[1],"id":1}'
    echo "Minage démarré sur port $port"
done

sleep 5
BLOCK=$(curl -s -X POST http://localhost:8541 -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
echo "📊 Bloc actuel: $BLOCK"

echo "✅ PoA démarré!"
