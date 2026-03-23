# authentication/register_device.py
import asyncio
import aiohttp
import json
import time
from ecdsa_keys import ECDSAManager

class DeviceAuthenticator:
    def __init__(self, backend_url="http://localhost:8080"):
        self.backend_url = backend_url
        
    async def register_device(self, device_id, node_id, interval):
        print(f"\n📝 [{device_id}] Enregistrement...")
        
        keys = ECDSAManager.generate_keys(device_id)
        
        # ✅ Structure CORRECTE qui correspond au backend Go
        payload = {
            "address": keys["address"],           # ← Clé "address" attendue
            "publicKey": keys["public_key"],       # ← Clé "publicKey" attendue
            "metadata": json.dumps({               # ← "metadata" doit être une string JSON
                "device_id": device_id,
                "node_id": node_id,
                "interval": interval,
                "type": "sensor",
                "firmware": "v1.0",
                "capabilities": ["temperature", "humidity", "battery"]
            })
        }
        
        try:
            async with aiohttp.ClientSession() as session:
                async with session.post(
                    f"{self.backend_url}/api/device/register",
                    json=payload,
                    timeout=10
                ) as resp:
                    result = await resp.json()
                    
                    if resp.status == 200:
                        print(f"✅ [{device_id}] Enregistré!")
                        # La réponse contient le txHash
                        print(f"   TX Hash: {result.get('txHash', 'N/A')[:20]}...")
                        return keys
                    else:
                        print(f"❌ [{device_id}] Erreur {resp.status}: {result}")
                        return None
        except Exception as e:
            print(f"❌ [{device_id}] Exception: {e}")
            return None

async def register_all_devices():
    auth = DeviceAuthenticator()
    # Configuration des 4 appareils
    devices_config = [
        ("ESP32_001", 1, 2),   # device_id, node_id, interval
        ("ESP32_002", 2, 4),
        ("ESP32_003", 3, 5),
        ("ESP32_004", 4, 6),
    ]
    
    registered_devices = []
    for device_id, node_id, interval in devices_config:
        keys = await auth.register_device(device_id, node_id, interval)
        if keys:
            registered_devices.append({
                "device_id": device_id,
                "node_id": node_id,
                "interval": interval,
                "keys": keys
            })
        await asyncio.sleep(1)  # Espacement entre enregistrements
    
    return registered_devices
