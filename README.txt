# 🚀 IoT Blockchain + AI Anomaly Detection Platform

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](https://www.docker.com/)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)
[![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)](https://reactjs.org/)
[![TensorFlow](https://img.shields.io/badge/TensorFlow-2.16-FF6F00?logo=tensorflow)](https://www.tensorflow.org/)

---

## 📋 Guide de Démarrage - Copier-Coller dans 4 Terminaux

### 📥 Étape 0 - Cloner le projet

git clone https://github.com/chaimaa962/iot-project.git
cd iot-project
chmod +x *.sh

---------------------------------------------------------------------------------------------------------------------------------------------------------------------

## 🖥️ TERMINAL 1 - Blockchain

cd iot-project

# 1. Démarrer les conteneurs blockchain
docker-compose up -d validator1 validator2 validator3 validator4

# 2. Attendre 30 secondes que les nœuds démarrent
sleep 30

# 3. Connecter les validateurs entre eux
./demarer.sh
# Nettoyer les enodes (sans ?discport=0)
enode1="enode://9f07a02efe326bc16b46e356acf2960fec0a7e835b2f7fb061bad442b0d05084682fa0b3005e71c4d4baaecd65e6a33b726a88c5787dafbcea30bc29adc59481@172.20.0.11:30301"
enode2="enode://740eee8689016b4b3b515b223c7c5ea58da5a0f0c7bc72b5dca3fae2bc8e4ed3677e5fbb624c0438862a325ae1a385f9a843804991ab1878711640a9c6ba066d@172.20.0.12:30302"
enode3="enode://ab091bee69aea0ac1da14ed979e60c0ff396a6f28842ff40a83417f625c339e37876ecc1e4a5dee70d265e14c0ca9507100b680acfcb175acefb38b8fda003e6@172.20.0.13:30303"
enode4="enode://9e67d0919200a26220cd7d18f3d2f2fe6628cb4059ee7671a129f091e3681a6841df80c252ac64aaeffc6569ce13572e9c66e25711489e2070ed6f7d56727945@172.20.0.14:30304"

# Connecter
curl -X POST http://localhost:8541 -H 'Content-Type: application/json' -d "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"$enode2\"],\"id\":1}"
curl -X POST http://localhost:8541 -H 'Content-Type: application/json' -d "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"$enode3\"],\"id\":1}"
curl -X POST http://localhost:8541 -H 'Content-Type: application/json' -d "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"$enode4\"],\"id\":1}"

curl -X POST http://localhost:8542 -H 'Content-Type: application/json' -d "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"$enode1\"],\"id\":1}"
curl -X POST http://localhost:8542 -H 'Content-Type: application/json' -d "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"$enode3\"],\"id\":1}"
curl -X POST http://localhost:8542 -H 'Content-Type: application/json' -d "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"$enode4\"],\"id\":1}"

curl -X POST http://localhost:8543 -H 'Content-Type: application/json' -d "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"$enode1\"],\"id\":1}"
curl -X POST http://localhost:8543 -H 'Content-Type: application/json' -d "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"$enode2\"],\"id\":1}"
curl -X POST http://localhost:8543 -H 'Content-Type: application/json' -d "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"$enode4\"],\"id\":1}"

curl -X POST http://localhost:8544 -H 'Content-Type: application/json' -d "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"$enode1\"],\"id\":1}"
curl -X POST http://localhost:8544 -H 'Content-Type: application/json' -d "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"$enode2\"],\"id\":1}"
curl -X POST http://localhost:8544 -H 'Content-Type: application/json' -d "{\"jsonrpc\":\"2.0\",\"method\":\"admin_addPeer\",\"params\":[\"$enode3\"],\"id\":1}"

# Vérifier
sleep 3
for port in 8541 8542 8543 8544; do
    count=$(curl -s -X POST http://localhost:$port -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","method":"admin_peers","params":[],"id":1}' | jq '.result | length')
    echo "Port $port: $count peers"
done


# Résultat attendu : Port 8541: 3 peers / Port 8542: 3 peers / Port 8543: 3 peers / Port 8544: 3 peers

------------------------------------------------------------------------------------------------------------------------------------------------------------------

## 🖥️ TERMINAL 2 - Backend Go + ZKP

cd iot-project/iot-backend

# Lancer le backend sur le port 8080
PORT=8080 go run ./cmd/api/main_secure.go

# Résultat attendu : ✅ GETH_NODE_1 déjà enregistré ... 🚀 Serveur SÉCURISÉ démarré (ECDSA + ZKP) [port 8080]

---------------------------------------------------------------------------------------------------------------------------------------------------------------------------

## 🖥️ TERMINAL 3 - Simulateur IoT

cd iot-project

# Charger les clés privées
export $(cat .env | xargs 2>/dev/null)

# Démarrer le simulateur (4 nœuds)
python3 geth_nodes_simulator.py

# Résultat attendu : 🔐 [GETH_NODE_1] AUTHENTIFIÉ ✅ [GETH_NODE_1] #1 | 🔷 RÉELLE | Bloc #4583

-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

## 🖥️ TERMINAL 4 - Dashboard

cd iot-project/iot-dashboard

# Installer les dépendances (première fois uniquement)
npm install

# Démarrer le dashboard
npm start

# Le dashboard s'ouvre sur http://localhost:3000

-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

## 🛑 Arrêter le Système

cd iot-project
./stop-all.sh

# Résultat attendu : ✅ Système arrêté avec succès !

-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

## 🔄 Redémarrer le Système (après un arrêt)

cd iot-project
./start-all.sh

# Cette commande relance automatiquement :
# - La blockchain (Docker)
# - Le service IA
# - Le backend Go
# - Le simulateur
# - Le dashboard

-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

## 🔄 Synchroniser avec GitHub

# 1. Récupérer les derniers changements
git pull origin main --rebase

# 2. Ajouter vos modifications
git add .
git commit -m "Votre message"

# 3. Pousser vers GitHub
git push

# OU utiliser le script automatique
./sync.sh

-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

## 🌐 Points d'Accès

Dashboard       http://localhost:3000
Backend API     http://localhost:8080/api/health
Service IA      http://localhost:5001/health
Blockchain RPC  http://localhost:8541

-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

## 📊 Vérifier l'état du système

cd iot-project
./check-status.sh
