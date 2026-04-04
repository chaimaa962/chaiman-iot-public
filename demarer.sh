#!/bin/bash

# ============================================
# Script de Redémarrage du Réseau POA
# Conserve toutes les données et reconnecte les peers
# ============================================

set -e

echo "=========================================="
echo "🔄 Redémarrage du réseau POA"
echo "=========================================="

cd ~/iot-docker-project

# 1. Arrêter les nœuds proprement
echo ""
echo "🛑 Arrêt des nœuds..."
docker-compose down

sleep 3

# 2. Vérifier que le réseau Docker existe
if ! docker network inspect poa-network >/dev/null 2>&1; then
    echo "🌐 Création du réseau Docker..."
    docker network create --subnet=172.20.0.0/16 poa-network
fi

# 3. Redémarrer les nœuds
echo ""
echo "🚀 Démarrage des nœuds..."
docker-compose up -d

echo ""
echo "⏳ Attente de 25 secondes pour le démarrage complet..."
sleep 25

# 4. Récupérer les ENODEs avec IP Docker
echo ""
echo "🔗 Récupération des ENODEs..."

enode1=$(curl -s -X POST http://localhost:8541 -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' 2>/dev/null | jq -r '.result.enode' | sed "s/127.0.0.1/172.20.0.11/" | sed "s/\[::\]/172.20.0.11/" || echo "")
enode2=$(curl -s -X POST http://localhost:8542 -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' 2>/dev/null | jq -r '.result.enode' | sed "s/127.0.0.1/172.20.0.12/" | sed "s/\[::\]/172.20.0.12/" || echo "")
enode3=$(curl -s -X POST http://localhost:8543 -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' 2>/dev/null | jq -r '.result.enode' | sed "s/127.0.0.1/172.20.0.13/" | sed "s/\[::\]/172.20.0.13/" || echo "")
enode4=$(curl -s -X POST http://localhost:8544 -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' 2>/dev/null | jq -r '.result.enode' | sed "s/127.0.0.1/172.20.0.14/" | sed "s/\[::\]/172.20.0.14/" || echo "")

# 5. Connecter les peers
echo ""
echo "🔗 Connexion des peers..."

# Val1 -> Val2,3,4
if [ -n "$enode2" ]; then curl -s -X POST http://localhost:8541 -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode2\\"],\\"id\\":1}" > /dev/null; fi
if [ -n "$enode3" ]; then curl -s -X POST http://localhost:8541 -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode3\\"],\\"id\\":1}" > /dev/null; fi
if [ -n "$enode4" ]; then curl -s -X POST http://localhost:8541 -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode4\\"],\\"id\\":1}" > /dev/null; fi

# Val2 -> Val1,3,4
if [ -n "$enode1" ]; then curl -s -X POST http://localhost:8542 -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode1\\"],\\"id\\":1}" > /dev/null; fi
if [ -n "$enode3" ]; then curl -s -X POST http://localhost:8542 -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode3\\"],\\"id\\":1}" > /dev/null; fi
if [ -n "$enode4" ]; then curl -s -X POST http://localhost:8542 -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode4\\"],\\"id\\":1}" > /dev/null; fi

# Val3 -> Val1,2,4
if [ -n "$enode1" ]; then curl -s -X POST http://localhost:8543 -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode1\\"],\\"id\\":1}" > /dev/null; fi
if [ -n "$enode2" ]; then curl -s -X POST http://localhost:8543 -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode2\\"],\\"id\\":1}" > /dev/null; fi
if [ -n "$enode4" ]; then curl -s -X POST http://localhost:8543 -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode4\\"],\\"id\\":1}" > /dev/null; fi

# Val4 -> Val1,2,3
if [ -n "$enode1" ]; then curl -s -X POST http://localhost:8544 -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode1\\"],\\"id\\":1}" > /dev/null; fi
if [ -n "$enode2" ]; then curl -s -X POST http://localhost:8544 -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode2\\"],\\"id\\":1}" > /dev/null; fi
if [ -n "$enode3" ]; then curl -s -X POST http://localhost:8544 -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode3\\"],\\"id\\":1}" > /dev/null; fi

# Full nodes -> tous les validateurs
for port in 8545 8546 8547 8548; do
    for enode in "$enode1" "$enode2" "$enode3" "$enode4"; do
        if [ -n "$enode" ]; then
            curl -s -X POST http://localhost:$port -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode\\"],\\"id\\":1}" > /dev/null 2>&1 || true
        fi
    done
done

echo "✅ Peers connectés"

# 6. Attendre et vérifier
echo ""
echo "⏳ Attente de 10 secondes pour la stabilisation..."
sleep 10

# 7. Vérification finale
echo ""
echo "=========================================="
echo "📊 Vérification du réseau"
echo "=========================================="

for port in 8541 8542 8543 8544; do
    peers=$(curl -s -X POST http://localhost:$port -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","method":"admin_peers","params":[],"id":1}' 2>/dev/null | jq '.result | length' 2>/dev/null || echo "0")
    block=$(curl -s -X POST http://localhost:$port -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' 2>/dev/null | jq -r '.result' 2>/dev/null | xargs printf "%d\\n" 2>/dev/null || echo "0")
    mining=$(curl -s -X POST http://localhost:$port -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' 2>/dev/null | jq -r '.result' 2>/dev/null || echo "false")
    echo "Validator (port $port): $peers peers | Bloc #$block | Mining: $mining"
done

echo ""
echo "=========================================="
echo "✅ Redémarrage terminé avec succès !"
echo "=========================================="
echo ""
echo "📌 Commandes utiles :"
echo "  - Voir les logs : docker logs -f geth-validator1"
echo "  - Arrêter : docker-compose down"
echo "  - Statut : docker-compose ps"
'''

# Sauvegarder le script
with open('/mnt/kimi/output/restart-poa-network.sh', 'w') as f:
    f.write(restart_script)

print("✅ Script de redémarrage créé : restart-poa-network.sh")
print("\n" + "="*60)
print(restart_script)
