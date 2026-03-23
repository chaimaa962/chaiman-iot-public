# add_new_device.py
#!/usr/bin/env python3
import asyncio
import aiohttp
import json
import time
import hashlib
from ecdsa_keys import ECDSAManager

async def register_new_device():
    print("\n" + "="*60)
    print("➕ AJOUT D'UN NOUVEL APPAREIL IoT")
    print("="*60)
    
    # Saisie des informations
    name = input("Nom de l'appareil (ex: CAPTEUR_001): ").strip()
    if not name:
        name = f"DEVICE_{int(time.time())}"
    
    interval = int(input("Intervalle d'envoi (secondes) [10]: ") or "10")
    
    message = input("Message à envoyer [Bonjour...]: ").strip()
    if not message:
        message = f"Bonjour je suis {name}"
    
    # Générer les clés
    print(f"\n🔑 Génération des clés ECDSA pour {name}...")
    keys = ECDSAManager.generate_keys(name)
    
    print(f"   Adresse: {keys['address'][:20]}...")
    print(f"   Clé publique: {keys['public_key'][:30]}...")
    
    # Enregistrer dans la blockchain
    print(f"\n📝 Enregistrement dans la blockchain...")
    
    payload = {
        "address": keys["address"],
        "publicKey": keys["public_key"],
        "metadata": json.dumps({
            "device_id": name,
            "interval": interval,
            "message": message,
            "type": "sensor",
            "created_at": int(time.time())
        })
    }
    
    async with aiohttp.ClientSession() as session:
        async with session.post(
            "http://localhost:8080/api/device/register",
            json=payload,
            timeout=10
        ) as resp:
            result = await resp.json()
            
            if resp.status == 200:
                print(f"✅ Appareil enregistré avec succès!")
                print(f"   TX Hash: {result.get('txHash', 'N/A')[:30]}...")
                
                # Sauvegarder les clés
                filename = f"{name}_keys.json"
                with open(filename, "w") as f:
                    json.dump({
                        "device_id": name,
                        "address": keys["address"],
                        "public_key": keys["public_key"],
                        "private_key": keys["private_key"],
                        "interval": interval,
                        "message": message
                    }, f, indent=2)
                
                print(f"💾 Clés sauvegardées dans {filename}")
                
                # Afficher les instructions
                print("\n📋 Pour utiliser cet appareil:")
                print(f"   python send_message.py {filename}")
                print(f"   python simulator.py --device {filename}")
                
            else:
                print(f"❌ Erreur: {result}")

if __name__ == "__main__":
    asyncio.run(register_new_device())
