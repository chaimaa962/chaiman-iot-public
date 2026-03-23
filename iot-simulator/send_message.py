#!/usr/bin/env python3
# send_message.py
import asyncio
import aiohttp
import json
import time
import hashlib
import secrets
from ecdsa_keys import ECDSAManager

async def send_message(keys_file=None, custom_message=None):
    """Envoie un message avec un appareil existant"""
    
    if keys_file:
        # Charger les clés depuis un fichier
        with open(keys_file, 'r') as f:
            keys = json.load(f)
        device_id = keys["device_id"]
        private_key = keys["private_key"]
        address = keys["address"]
    else:
        # Créer un appareil temporaire
        device_id = f"TEMP_{int(time.time())}"
        keys = ECDSAManager.generate_keys(device_id)
        private_key = keys["private_key"]
        address = keys["address"]
        print(f"⚠️ Appareil temporaire créé: {address[:15]}...")
    
    # Préparer le message
    if custom_message:
        message = custom_message
    else:
        message = input("Message à envoyer: ").strip()
    
    # Calculer le hash
    message_hash = hashlib.sha256(message.encode()).hexdigest()
    
    # Signer
    from ecdsa import SigningKey, SECP256k1
    sk = SigningKey.from_string(bytes.fromhex(private_key), curve=SECP256k1)
    signature = sk.sign(message.encode()).hex()
    
    # Préparer la payload
    payload = {
        "device_id": device_id,
        "address": address,
        "message": message,
        "message_hash": message_hash,
        "signature": signature,
        "proof": hashlib.sha256(f"{signature}{message_hash}".encode()).hexdigest(),
        "timestamp": int(time.time()),
        "sequence": 1,
        "nonce": secrets.token_hex(16)
    }
    
    print(f"\n📤 Envoi du message...")
    print(f"   De: {device_id}")
    print(f"   Message: '{message}'")
    
    try:
        async with aiohttp.ClientSession() as session:
            async with session.post(
                "http://localhost:8080/api/node/message",
                json=payload,
                timeout=5
            ) as resp:
                result = await resp.json()
                
                if resp.status == 200:
                    print(f"✅ Message envoyé!")
                    print(f"   TX Hash: {result.get('txHash', 'N/A')[:30]}...")
                    print(f"   Bloc: #{result.get('block', 0)}")
                else:
                    print(f"❌ Erreur: {result}")
    except Exception as e:
        print(f"❌ Exception: {e}")

async def main():
    print("=" * 50)
    print("📨 ENVOI DE MESSAGE IoT")
    print("=" * 50)
    
    keys_file = input("Fichier de clés (laisser vide pour temporaire): ").strip() or None
    
    message = input("Message à envoyer: ").strip()
    
    await send_message(keys_file, message)

if __name__ == "__main__":
    asyncio.run(main())
