#!/usr/bin/env python3
# simulator.py
import asyncio
import aiohttp
import json
import time
import hashlib
import secrets
import random
from datetime import datetime
from ecdsa_keys import ECDSAManager

class IoTNode:
    def __init__(self, device_id, node_id, interval, message_template, private_key=None):
        self.device_id = device_id
        self.node_id = node_id
        self.interval = interval
        self.message_template = message_template
        self.sequence = 0
        self.stats = {
            "sent": 0,
            "success": 0,
            "failed": 0,
            "last_tx_time": 0
        }
        
        # Générer ou utiliser la clé existante
        if private_key:
            self.private_key = private_key
            # Reconstruire les clés à partir de la clé privée
            from ecdsa import SigningKey, SECP256k1
            sk = SigningKey.from_string(bytes.fromhex(private_key), curve=SECP256k1)
            vk = sk.get_verifying_key()
            self.public_key = vk.to_string().hex()
            # Adresse Ethereum simplifiée (à améliorer)
            self.address = hashlib.sha256(self.public_key.encode()).hexdigest()[:40]
            self.address = "0x" + self.address
        else:
            keys = ECDSAManager.generate_keys(device_id)
            self.private_key = keys["private_key"]
            self.public_key = keys["public_key"]
            self.address = keys["address"]
        
        print(f"🔑 [{device_id}] Adresse: {self.address[:15]}...")
    
    def generate_message(self):
        """Génère le message à envoyer"""
        # Message personnalisé selon le template
        if "{count}" in self.message_template:
            message = self.message_template.format(count=self.sequence)
        elif "{time}" in self.message_template:
            message = self.message_template.format(time=datetime.now().strftime("%H:%M:%S"))
        else:
            message = self.message_template
        
        return message
    
    def sign_message(self, message):
        """Signe le message avec ECDSA"""
        from ecdsa import SigningKey, SECP256k1
        sk = SigningKey.from_string(bytes.fromhex(self.private_key), curve=SECP256k1)
        signature = sk.sign(message.encode())
        return signature.hex()
    
    async def send_message(self, session, backend_url="http://localhost:8080"):
        """Envoie un message au backend"""
        self.sequence += 1
        
        # Générer le message
        message_content = self.generate_message()
        
        # Calculer le hash du message
        message_hash = hashlib.sha256(message_content.encode()).hexdigest()
        
        # Signer le message (ECDSA)
        signature = self.sign_message(message_content)
        
        # Générer nonce pour anti-rejeu
        nonce = secrets.token_hex(16)
        timestamp = int(time.time())
        
        # Simuler une preuve ZKP (pour cacher la signature)
        # Dans un vrai système, ce serait une vraie preuve
        proof_data = f"{signature}{message_hash}{self.sequence}".encode()
        proof = hashlib.sha256(proof_data).hexdigest()
        
        # Préparer la payload
        payload = {
            "device_id": self.device_id,
            "address": self.address,
            "message": message_content,
            "message_hash": message_hash,
            "signature": signature,
            "proof": proof,
            "timestamp": timestamp,
            "sequence": self.sequence,
            "nonce": nonce
        }
        
        start_time = time.time()
        
        try:
            async with session.post(
                f"{backend_url}/api/node/message",
                json=payload,
                timeout=5
            ) as resp:
                elapsed = (time.time() - start_time) * 1000
                result = await resp.json()
                
                self.stats["sent"] += 1
                
                if resp.status == 200:
                    self.stats["success"] += 1
                    self.stats["last_tx_time"] = time.time()
                    
                    print(f"✅ [{self.device_id}] #{self.sequence:3d} | "
                          f"⏱️ {elapsed:5.1f}ms | "
                          f"📨 '{message_content[:20]}...' | "
                          f"⛓️ Bloc #{result.get('block', 0)}")
                    return True
                else:
                    self.stats["failed"] += 1
                    print(f"❌ [{self.device_id}] Erreur {resp.status}: {result.get('message', '')}")
                    return False
                    
        except asyncio.TimeoutError:
            self.stats["failed"] += 1
            print(f"❌ [{self.device_id}] Timeout")
            return False
        except Exception as e:
            self.stats["failed"] += 1
            print(f"❌ [{self.device_id}] Exception: {e}")
            return False

class NetworkSimulator:
    def __init__(self, backend_url="http://localhost:8080"):
        self.backend_url = backend_url
        self.nodes = []
        self.stats = {
            "start_time": time.time(),
            "total_messages": 0
        }
    
    def add_node(self, device_id, node_id, interval, message_template, private_key=None):
        node = IoTNode(device_id, node_id, interval, message_template, private_key)
        self.nodes.append(node)
        return node
    
    async def node_loop(self, node, session):
        """Boucle d'envoi pour un nœud"""
        print(f"🚀 [{node.device_id}] Démarrage (intervalle: {node.interval}s)")
        
        while True:
            await node.send_message(session, self.backend_url)
            self.stats["total_messages"] += 1
            await asyncio.sleep(node.interval)
    
    async def stats_reporter(self):
        """Rapport périodique"""
        while True:
            await asyncio.sleep(15)
            
            runtime = time.time() - self.stats["start_time"]
            
            print("\n" + "=" * 80)
            print(f"📊 RAPPORT STATISTIQUES - {runtime:.0f}s")
            print("=" * 80)
            print(f"{'Appareil':12} {'Envoyés':8} {'✅':6} {'❌':6} {'Taux':8} {'Dernier':12}")
            print("-" * 80)
            
            for node in self.nodes:
                taux = node.stats["success"] / max(1, node.stats["sent"]) * 100
                last = datetime.fromtimestamp(node.stats["last_tx_time"]).strftime("%H:%M:%S") if node.stats["last_tx_time"] > 0 else "jamais"
                print(f"{node.device_id:12} {node.stats['sent']:8d} "
                      f"{node.stats['success']:6d} {node.stats['failed']:6d} "
                      f"{taux:6.1f}% {last:12}")
            
            print("=" * 80)
    
    async def register_all_devices(self):
        """Enregistre tous les nœuds dans la blockchain"""
        print("\n📝 ENREGISTREMENT DES APPAREILS DANS LA BLOCKCHAIN")
        print("=" * 50)
        
        async with aiohttp.ClientSession() as session:
            for node in self.nodes:
                # Créer le payload d'enregistrement
                payload = {
                    "address": node.address,
                    "publicKey": node.public_key,
                    "metadata": json.dumps({
                        "device_id": node.device_id,
                        "node_id": node.node_id,
                        "interval": node.interval,
                        "type": "iot_node",
                        "registered_at": int(time.time())
                    })
                }
                
                try:
                    async with session.post(
                        f"{self.backend_url}/api/device/register",
                        json=payload,
                        timeout=10
                    ) as resp:
                        result = await resp.json()
                        
                        if resp.status == 200:
                            print(f"✅ [{node.device_id}] Enregistré - TX: {result.get('txHash', 'N/A')[:20]}...")
                        else:
                            print(f"❌ [{node.device_id}] Erreur: {result}")
                except Exception as e:
                    print(f"❌ [{node.device_id}] Exception: {e}")
                
                await asyncio.sleep(1)
        
        print("=" * 50)
    
    async def run(self, register_first=True):
        """Lance le simulateur"""
        
        print("\n" + "=" * 60)
        print("🚀 SIMULATEUR IoT - MESSAGES ENTRE NŒUDS")
        print("=" * 60)
        
        if register_first:
            await self.register_all_devices()
        
        print(f"\n✅ {len(self.nodes)} appareils chargés")
        print("Démarrage des envois de messages...\n")
        
        async with aiohttp.ClientSession() as session:
            tasks = []
            
            # Tâches pour chaque nœud
            for node in self.nodes:
                task = asyncio.create_task(self.node_loop(node, session))
                tasks.append(task)
            
            # Tâche pour les statistiques
            stats_task = asyncio.create_task(self.stats_reporter())
            tasks.append(stats_task)
            
            await asyncio.gather(*tasks)

async def create_new_device(simulator):
    """Crée un nouvel appareil et l'enregistre"""
    print("\n" + "=" * 50)
    print("➕ CRÉATION D'UN NOUVEL APPAREIL")
    print("=" * 50)
    
    device_id = input("Nom du nouvel appareil (ex: ESP32_005): ").strip()
    if not device_id:
        device_id = f"ESP32_{random.randint(100, 999)}"
    
    interval = int(input("Intervalle d'envoi (secondes) [défaut=10]: ") or "10")
    
    message = input("Message à envoyer [défaut='Bonjour je suis nouveau']: ").strip()
    if not message:
        message = f"Bonjour je suis {device_id}"
    
    # Créer le nœud
    node = simulator.add_node(device_id, len(simulator.nodes)+1, interval, message)
    
    print(f"\n✅ Appareil créé:")
    print(f"   ID: {node.device_id}")
    print(f"   Adresse: {node.address}")
    print(f"   Intervalle: {node.interval}s")
    print(f"   Message: '{node.message_template}'")
    
    return node

async def main():
    print("\n📋 SIMULATEUR IoT - MESSAGES ENTRE NŒUDS")
    print("1. Lancer les 4 nœuds par défaut")
    print("2. Lancer avec création de nouveaux appareils")
    
    choice = input("\nChoix (1-2) [défaut=1]: ").strip() or "1"
    
    # Vérifier le backend
    try:
        async with aiohttp.ClientSession() as session:
            async with session.get("http://localhost:8080/api/health", timeout=5) as resp:
                if resp.status == 200:
                    print("✅ Backend accessible")
                else:
                    print("❌ Backend inaccessible")
                    return
    except:
        print("❌ Backend non joignable")
        return
    
    # Créer le simulateur
    sim = NetworkSimulator()
    
    if choice == "1":
        # Ajouter les 4 nœuds par défaut avec leurs messages personnalisés
        sim.add_node("ESP32_001", 1, 5, "Bonjour je suis node 1 (toutes les 5s)")
        sim.add_node("ESP32_002", 2, 10, "Bonjour je suis node 2 (toutes les 10s)")
        sim.add_node("ESP32_003", 3, 15, "Bonjour je suis node 3 (toutes les 15s)")
        sim.add_node("ESP32_004", 4, 25, "Bonjour je suis node 4 (toutes les 25s)")
        
        await sim.run(register_first=True)
    
    else:
        # Mode création interactive
        print("\nCombien d'appareils voulez-vous créer ?")
        count = int(input("Nombre [défaut=4]: ") or "4")
        
        for i in range(count):
            print(f"\n--- Appareil {i+1} ---")
            device_id = input(f"Nom (défaut=ESP32_{100+i}): ").strip() or f"ESP32_{100+i}"
            interval = int(input("Intervalle (s) [défaut=10]: ") or "10")
            message = input("Message [défaut='Bonjour...']: ").strip() or f"Bonjour je suis {device_id}"
            
            sim.add_node(device_id, i+1, interval, message)
        
        # Demander si enregistrement automatique
        register = input("\nEnregistrer automatiquement dans la blockchain? (o/n) [défaut=o]: ").lower() or "o"
        register_first = (register == "o")
        
        await sim.run(register_first=register_first)

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\n\n👋 Arrêt du simulateur")
