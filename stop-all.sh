#!/bin/bash

echo "════════════════════════════════════════════════════════════════"
echo "🛑 ARRÊT DU SYSTÈME"
echo "════════════════════════════════════════════════════════════════"

# Arrêter le simulateur
echo "📡 Arrêt du simulateur..."
pkill -f "geth_nodes_simulator" 2>/dev/null && echo "   ✅ Simulateur arrêté"

# Arrêter le backend Go
echo "🔐 Arrêt du backend Go..."
pkill -f "main_secure" 2>/dev/null && echo "   ✅ Backend Go arrêté"

# Arrêter les conteneurs Docker
echo "🐳 Arrêt des conteneurs Docker..."
docker-compose down 2>/dev/null && echo "   ✅ Conteneurs arrêtés"

# Nettoyer les logs
rm -f logs/*.pid 2>/dev/null

echo ""
echo "✅ Système arrêté avec succès !"
