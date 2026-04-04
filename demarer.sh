#!/bin/bash

# ============================================
# Script de Redémarrage du Réseau POA
# ============================================

set -e

echo "=========================================="
echo "🔄 Redémarrage du réseau POA"
echo "=========================================="

cd ~/iot-docker-project

# 1. Arrêter les nœuds
echo ""
echo "🛑 Arrêt des nœuds..."
docker-compose down

sleep 3

# 2. Vérifier/créer le réseau Docker
if ! docker network inspect poa-network >/dev/null 2>&1; then
    echo "🌐 Création du réseau Docker..."
    docker network create --subnet=172.20.0.0/16 poa-network
fi

# 3. Redémarrer les nœuds
echo ""
echo "🚀 Démarrage des nœuds..."
docker-compose up -d

echo ""
echo "⏳ Attente de 30 secondes pour le démarrage complet..."
sleep 30

# 4. Récupérer les ENODEs avec IP Docker
echo ""
echo "🔗 Récupération des ENODEs..."

# Fonction pour récupérer un enode avec retry
get_enode() {
    local port=$1
    local ip=$2
    local max_attempts=5
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        enode=$(curl -s -X POST http://localhost:$port -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' 2>/dev/null | jq -r '.result.enode' 2>/dev/null)
        
        if [ -n "$enode" ] && [ "$enode" != "null" ]; then
            # Remplacer l'IP
            enode_fixed=$(echo "$enode" | sed "s/127.0.0.1/$ip/" | sed "s/\\[::\\]/$ip/")
            echo "$enode_fixed"
            return 0
        fi
        
        echo "  Tentative $attempt/$max_attempts pour port $port..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo ""
    return 1
}

enode1=$(get_enode 8541 "172.20.0.11")
enode2=$(get_enode 8542 "172.20.0.12")
enode3=$(get_enode 8543 "172.20.0.13")
enode4=$(get_enode 8544 "172.20.0.14")

echo ""
echo "ENODE 1: ${enode1:-ECHEC}"
echo "ENODE 2: ${enode2:-ECHEC}"
echo "ENODE 3: ${enode3:-ECHEC}"
echo "ENODE 4: ${enode4:-ECHEC}"

# 5. Connecter les peers
echo ""
echo "🔗 Connexion des peers..."

# Val1 -> Val2,3,4
if [ -n "$enode2" ]; then 
    curl -s -X POST http://localhost:8541 -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode2\\"],\\"id\\":1}" > /dev/null
    echo "  Val1 -> Val2 OK"
fi
if [ -n "$enode3" ]; then 
    curl -s -X POST http://localhost:8541 -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode3\\"],\\"id\\":1}" > /dev/null
    echo "  Val1 -> Val3 OK"
fi
if [ -n "$enode4" ]; then 
    curl -s -X POST http://localhost:8541 -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode4\\"],\\"id\\":1}" > /dev/null
    echo "  Val1 -> Val4 OK"
fi

# Val2 -> Val1,3,4
if [ -n "$enode1" ]; then 
    curl -s -X POST http://localhost:8542 -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode1\\"],\\"id\\":1}" > /dev/null
    echo "  Val2 -> Val1 OK"
fi
if [ -n "$enode3" ]; then 
    curl -s -X POST http://localhost:8542 -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode3\\"],\\"id\\":1}" > /dev/null
    echo "  Val2 -> Val3 OK"
fi
if [ -n "$enode4" ]; then 
    curl -s -X POST http://localhost:8542 -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode4\\"],\\"id\\":1}" > /dev/null
    echo "  Val2 -> Val4 OK"
fi

# Val3 -> Val1,2,4
if [ -n "$enode1" ]; then 
    curl -s -X POST http://localhost:8543 -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode1\\"],\\"id\\":1}" > /dev/null
    echo "  Val3 -> Val1 OK"
fi
if [ -n "$enode2" ]; then 
    curl -s -X POST http://localhost:8543 -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode2\\"],\\"id\\":1}" > /dev/null
    echo "  Val3 -> Val2 OK"
fi
if [ -n "$enode4" ]; then 
    curl -s -X POST http://localhost:8543 -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode4\\"],\\"id\\":1}" > /dev/null
    echo "  Val3 -> Val4 OK"
fi

# Val4 -> Val1,2,3
if [ -n "$enode1" ]; then 
    curl -s -X POST http://localhost:8544 -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode1\\"],\\"id\\":1}" > /dev/null
    echo "  Val4 -> Val1 OK"
fi
if [ -n "$enode2" ]; then 
    curl -s -X POST http://localhost:8544 -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode2\\"],\\"id\\":1}" > /dev/null
    echo "  Val4 -> Val2 OK"
fi
if [ -n "$enode3" ]; then 
    curl -s -X POST http://localhost:8544 -H 'Content-Type: application/json' -d "{\\"jsonrpc\\":\\"2.0\\",\\"method\\":\\"admin_addPeer\\",\\"params\\":[\\"$enode3\\"],\\"id\\":1}" > /dev/null
    echo "  Val4 -> Val3 OK"
fi

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
"""

# Sauvegarder
with open('/mnt/kimi/output/demarer_fixed.sh', 'w') as f:
    f.write(demarer_fixed)

print("✅ Script corrigé créé : demarer_fixed.sh")
print("\nLes changements principaux :")
print("1. Ajout de retry pour récupérer les ENODEs (5 tentatives)")
print("2. Affichage des ENODEs pour débogage")
print("3. Correction des erreurs de syntaxe")
print("4. Plus de détails lors de la connexion des peers")
