#!/bin/bash
echo "🔍 ÉTAT DU SYSTÈME"
echo "═══════════════════════════════════════"

# Blockchain
curl -s -X POST http://localhost:8541 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  | grep -q "result" && echo "✅ Blockchain" || echo "❌ Blockchain"

# Backend
curl -s http://localhost:8080/api/health | grep -q "OK" && echo "✅ Backend ZKP" || echo "❌ Backend ZKP"

# IA
curl -s http://localhost:5001/health | grep -q "healthy" && echo "✅ Service IA" || echo "❌ Service IA"

# Simulateur
pgrep -f "geth_nodes_simulator" > /dev/null && echo "✅ Simulateur" || echo "❌ Simulateur"
