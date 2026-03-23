#!/usr/bin/env python3
# create_device.py
import asyncio
import aiohttp
import json
import time
import hashlib
import os
from ecdsa_keys import ECDSAManager

async def register_device(device_id, interval, message, backend_url="http://localhost:8080"):
    """Crée et enregistre un nouvel appareil"""
    
    print(f"\n📝 Création de {device_id}...")
    
    # Générer les clés ECDSA
    keys = ECDSAManager.generate_keys(device_id)
    
    # Créer un secret unique pour l'appareil (pour ZKP)
    secret_seed = f"{keys['address']}{device_id}{time.time()}".encode()
    secret = int(hashlib.sha256(secret_seed).hexdigest()[:8], 16)
    if secret == 0:
        secret = 42
    
    # Métadonnées enrichies
    metadata = {
        "device_id": device_id,
        "interval": interval,
        "message": message,
        "created_at": int(time.time()),
        "type": "iot_device",
        "zkp_secret": secret,  # Stocké localement, jamais envoyé !
        "capabilities": ["temperature", "humidity"]  # Exemple
    }
    
    # Préparer le payload pour l'enregistrement
    payload = {
        "address": keys["address"],
        "publicKey": keys["public_key"],
        "metadata": json.dumps(metadata)
    }
    
    try:
        async with aiohttp.ClientSession() as session:
            async with session.post(
                f"{backend_url}/api/device/register",
                json=payload,
                timeout=10
            ) as resp:
                result = await resp.json()
                
                if resp.status == 200:
                    print(f"✅ {device_id} enregistré avec succès!")
                    print(f"   Adresse: {keys['address']}")
                    print(f"   TX Hash: {result.get('txHash', 'N/A')}")
                    
                    # Sauvegarder TOUTES les informations dans un fichier
                    filename = f"{device_id}_keys.json"
                    device_data = {
                        "device_id": device_id,
                        "address": keys["address"],
                        "public_key": keys["public_key"],
                        "private_key": keys["private_key"],
                        "secret": secret,  # Important pour ZKP
                        "interval": interval,
                        "message": message,
                        "created_at": int(time.time())
                    }
                    
                    with open(filename, "w") as f:
                        json.dump(device_data, f, indent=2)
                    
                    print(f"💾 Données sauvegardées dans {filename}")
                    print(f"   🔑 Secret ZKP: {secret} (à garder précieusement!)")
                    
                    # Afficher les instructions pour utiliser l'appareil
                    print("\n📋 Pour simuler cet appareil plus tard:")
                    print(f"   python simulate_device.py --keys {filename}")
                    
                    return device_data
                else:
                    print(f"❌ Erreur d'enregistrement: {result}")
                    return None
    except asyncio.TimeoutError:
        print(f"❌ Timeout - Le backend n'a pas répondu")
        return None
    except Exception as e:
        print(f"❌ Exception: {e}")
        return None

async def register_multiple_devices():
    """Enregistre plusieurs appareils à la fois"""
    print("\n" + "="*60)
    print("📦 ENREGISTREMENT MULTIPLE D'APPAREILS")
    print("="*60)
    
    devices = []
    count = int(input("Nombre d'appareils à créer [défaut=1]: ") or "1")
    
    for i in range(count):
        print(f"\n--- Appareil #{i+1} ---")
        device_id = input(f"Nom (défaut=DEVICE_{100+i}): ").strip() or f"DEVICE_{100+i}"
        interval = int(input("Intervalle (s) [défaut=10]: ") or "10")
        message = input("Message [défaut='Bonjour...']: ").strip() or f"Bonjour je suis {device_id}"
        
        device_data = await register_device(device_id, interval, message)
        if device_data:
            devices.append(device_data)
        
        # Petit délai entre les enregistrements
        await asyncio.sleep(1)
    
    # Résumé
    print("\n" + "="*60)
    print(f"📊 RÉSUMÉ - {len(devices)} appareil(s) créé(s)")
    print("="*60)
    for d in devices:
        print(f"✅ {d['device_id']}: {d['address'][:20]}... (secret: {d['secret']})")
    
    return devices
async def generate_proof_from_backend(self, signature, message_hash, timestamp):
    """Utilise le backend pour générer une vraie preuve ZKP"""
    async with aiohttp.ClientSession() as session:
        payload = {
            "signature": signature,
            "message_hash": message_hash,
            "address": self.address,
            "sequence": self.sequence,
            "timestamp": timestamp
        }
        async with session.post("http://localhost:8080/api/zkp/generate", json=payload) as resp:
            if resp.status == 200:
                result = await resp.json()
                return result.get("proof")
    return None
async def main():
    print("="*60)
    print("➕ CRÉATEUR D'APPAREILS IoT (avec support ZKP)")
    print("="*60)
    print("\nChoisissez une option:")
    print("1. Créer un seul appareil")
    print("2. Créer plusieurs appareils")
    print("3. Quitter")
    
    choice = input("\nVotre choix [1-3]: ").strip() or "1"
    
    if choice == "1":
        device_id = input("\nNom de l'appareil (ex: CAPTEUR_001): ").strip()
        if not device_id:
            device_id = f"DEVICE_{int(time.time())}"
        
        interval = int(input("Intervalle d'envoi (secondes) [défaut=10]: ") or "10")
        
        message = input("Message à envoyer [défaut='Bonjour...']: ").strip()
        if not message:
            message = f"Bonjour je suis {device_id}"
        
        await register_device(device_id, interval, message)
    
    elif choice == "2":
        await register_multiple_devices()
    
    else:
        print("👋 Au revoir!")

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\n\n👋 Interruption par l'utilisateur")
