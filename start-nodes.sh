#!/bin/bash

# Arrêter les anciens conteneurs
docker stop geth-node1 geth-node2 2>/dev/null
docker rm geth-node1 geth-node2 2>/dev/null

# Node 1
docker run -d \
  --name geth-node1 \
  -v $(pwd)/node1_data:/data \
  -v $(pwd)/password1.txt:/data/password.txt \
  -p 8545:8545 \
  -p 30301:30301 \
  ethereum/client-go:latest \
  --datadir=/data \
  --networkid=1234 \
  --port=30301 \
  --http \
  --http.port=8545 \
  --http.addr=0.0.0.0 \
  --http.api=eth,web3,net,admin,personal,miner,clique,txpool,debug \
  --http.corsdomain=* \
  --mine \
  --miner.threads=1 \
  --unlock=0xc6a0a22e356ce0b09246dd7a1b0c3b223d39e0ff \
  --password=/data/password.txt \
  --allow-insecure-unlock \
  --verbosity=4

# Node 2
docker run -d \
  --name geth-node2 \
  -v $(pwd)/node2_data:/data \
  -v $(pwd)/password2.txt:/data/password.txt \
  -p 8546:8545 \
  -p 30302:30301 \
  ethereum/client-go:latest \
  --datadir=/data \
  --networkid=1234 \
  --port=30301 \
  --http \
  --http.port=8545 \
  --http.addr=0.0.0.0 \
  --http.api=eth,web3,net,admin,personal,miner,clique,txpool,debug \
  --http.corsdomain=* \
  --mine \
  --miner.threads=1 \
  --unlock=0x7d2b6759153a1625a757f36b63f7034dd8c095ed \
  --password=/data/password.txt \
  --allow-insecure-unlock \
  --verbosity=4

echo "✅ Nœuds Geth démarrés"
sleep 5

# Connecter les nœuds
echo "🔗 Connexion des nœuds..."
ENODE1=$(docker exec geth-node1 geth --exec "admin.nodeInfo.enode" attach /data/geth.ipc 2>/dev/null | tr -d '"')
docker exec geth-node2 geth --exec "admin.addPeer('$ENODE1')" attach /data/geth.ipc 2>/dev/null

echo "⛏️  Démarrage du minage..."
docker exec geth-node1 geth --exec "miner.start()" attach /data/geth.ipc 2>/dev/null
docker exec geth-node2 geth --exec "miner.start()" attach /data/geth.ipc 2>/dev/null

echo "✅ Nœuds connectés et minage démarré"
