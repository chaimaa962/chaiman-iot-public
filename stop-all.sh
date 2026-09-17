#!/bin/bash
echo "🛑 Arrêt du système..."

# Backend
pkill -f main_secure
lsof -ti:8080 | xargs kill -9 2>/dev/null

# Simulateur IoT
pkill -f iot_simulator

# IA V4
pkill -f detection.py
lsof -ti:5001 | xargs kill -9 2>/dev/null

# Conteneurs Docker
docker stop $(docker ps -aq --filter "name=^v") 2>/dev/null
docker stop $(docker ps -aq --filter "name=^fn") 2>/dev/null

echo "✅ Tout est arrêté"
