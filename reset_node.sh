#!/bin/bash
# reset-blockchain.sh — Réinitialisation complète de la blockchain CHAIMAN IoT
# Usage: ./reset-blockchain.sh

set -e

echo "═══════════════════════════════════════════════════════════════"
echo "🗑️  RÉINITIALISATION COMPLÈTE DE LA BLOCKCHAIN"
echo "   ⚠️  TOUTES LES DONNÉES SERONT PERDUES"
echo "═══════════════════════════════════════════════════════════════"

cd ~/iot-docker-project

# 1. Arrêter et supprimer TOUS les conteneurs
echo ""
echo "🛑 Arrêt des conteneurs..."
docker stop $(docker ps -aq --filter "name=^v" --filter "name=^fn") 2>/dev/null || true
docker rm -f $(docker ps -aq --filter "name=^v" --filter "name=^fn") 2>/dev/null || true
echo "   ✓ Conteneurs arrêtés et supprimés"

# 2. Supprimer TOUTES les données blockchain
echo ""
echo "🗑️  Suppression des données blockchain..."
sudo rm -rf ./nodes/validator*/geth
sudo rm -rf ./nodes/validator*/keystore
sudo rm -rf ./nodes/fullnode*/geth
sudo rm -rf ./nodes/fullnode*/keystore
sudo rm -f ./genesis.json
echo "   ✓ Données blockchain supprimées"

# 3. Recréer les structures de dossiers
echo ""
echo "📁 Recréation des dossiers..."
for i in 1 2 3 4; do
    mkdir -p ./nodes/validator${i}
    mkdir -p ./nodes/fullnode${i}
done
echo "   ✓ Dossiers recréés"

# 4. Générer de nouvelles clés
echo ""
echo "🔑 Génération des nouvelles clés Ethereum..."
GETH="ethereum/client-go:v1.10.26"

for i in 1 2 3 4; do
    # Validators
    docker run --rm \
      -v "$PWD/nodes/validator${i}:/root/.ethereum" \
      $GETH account new --datadir=/root/.ethereum --password=/dev/null 2>/dev/null
    echo "   ✓ validator${i} — clé générée"
    
    # Full nodes
    docker run --rm \
      -v "$PWD/nodes/fullnode${i}:/root/.ethereum" \
      $GETH account new --datadir=/root/.ethereum --password=/dev/null 2>/dev/null
    echo "   ✓ fullnode${i} — clé générée"
done

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "✅ RÉINITIALISATION TERMINÉE"
echo "   Blockchain remise à zéro — prêt pour start-node.sh"
echo "═══════════════════════════════════════════════════════════════"
