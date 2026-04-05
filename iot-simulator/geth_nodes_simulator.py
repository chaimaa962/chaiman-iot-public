import asyncio
import aiohttp
import json
import time
import hashlib
import secrets
import os
from datetime import datetime
from eth_account import Account
from eth_utils import keccak
from dotenv import load_dotenv

load_dotenv(override=True)

BN254_MODULUS = int("21888242871839275222246405745257275088548364400416034343698204186575808495617")

# LECTURE DE L'URL DU BACKEND
BACKEND_URL = os.getenv("BACKEND_URL", "http://localhost:8080")
print(f"🔗 Backend URL: {BACKEND_URL}")

# LES 4 NŒUDS GETH
GETH_NODES = [
    {
        "device_id": "GETH_NODE_1",
        "address": os.getenv("GETH_NODE_1_ADDRESS", "0xaf0c6bf76f11760b7ba90a852aaeadfe50ab9277"),
        "private_key": os.getenv("GETH_NODE_1_KEY"),
        "interval": 5,
        "message": "Heartbeat #{count} from GETH_1",
    },
    {
        "device_id": "GETH_NODE_2",
        "address": os.getenv("GETH_NODE_2_ADDRESS", "0xc4f26670f7539138a21e7f33f2b042dbd1da6f30"),
        "private_key": os.getenv("GETH_NODE_2_KEY"),
        "interval": 10,
        "message": "Heartbeat #{count} from GETH_2",
    },
    {
        "device_id": "GETH_NODE_3",
        "address": os.getenv("GETH_NODE_3_ADDRESS", "0x60e69259368a740e8fe91cc61c5306234e36e01d"),
        "private_key": os.getenv("GETH_NODE_3_KEY"),
        "interval": 15,
        "message": "Heartbeat #{count} from GETH_3",
    },
    {
        "device_id": "GETH_NODE_4",
        "address": os.getenv("GETH_NODE_4_ADDRESS", "0x29885af643612e8b72123ccc3d6f527cd9321319"),
        "private_key": os.getenv("GETH_NODE_4_KEY"),
        "interval": 25,
        "message": "Heartbeat #{count} from GETH_4",
    }
]

class SecureGethNode:
    def __init__(self, config):
        self.device_id = config["device_id"]
        self.address = config["address"].lower()
        self.interval = config["interval"]
        self.message_template = config["message"]
        self.sequence = 0
        self.stats = {"sent": 0, "success": 0, "failed": 0}
        self.authenticated = False
        self.pending_sequence = None

        private_key = config.get("private_key")
        if not private_key:
            raise ValueError(f"❌ Clé privée manquante pour {self.device_id}")

        private_key = private_key.strip()
        if private_key.startswith('0x'):
            private_key = private_key[2:]

        self.account = Account.from_key(private_key)
        self._private_key_bytes = bytes.fromhex(private_key.zfill(64))
        
        derived = self.account.address.lower()
        if derived != self.address:
            print(f"⚠️  {self.device_id}: ALERTE ADRESSE!")
            print(f"   Config: {self.address}")
            print(f"   Dérivée: {derived}")
            raise ValueError(f"Clé privée invalide pour {self.device_id}")
        else:
            print(f"✅ {self.device_id}: {self.address[:20]}... (clé OK)")

    def sign_message_ecdsa(self, message: str) -> tuple:
        msg_bytes = message.encode('utf-8')
        message_hash = keccak(msg_bytes)
        
        msg_hash_hex = message_hash.hex()
        if msg_hash_hex.startswith('0x'):
            msg_hash_hex = msg_hash_hex[2:]
        msg_hash_hex = msg_hash_hex.zfill(64)
        
        from eth_keys import keys
        private_key = keys.PrivateKey(self._private_key_bytes)
        signature = private_key.sign_msg_hash(message_hash)
        
        v = signature.v
        if v >= 27:
            v -= 27
        
        r_bytes = signature.r.to_bytes(32, 'big')
        s_bytes = signature.s.to_bytes(32, 'big')
        v_byte = bytes([v])
        
        signature_hex = (r_bytes + s_bytes + v_byte).hex()
        
        return signature_hex, msg_hash_hex

    def generate_zkp_secret(self, signature_hex: str) -> int:
        sig_bytes = bytes.fromhex(signature_hex)
        sig_hash = hashlib.sha256(sig_bytes).digest()
        secret = int.from_bytes(sig_hash, "big") % BN254_MODULUS
        return secret if secret != 0 else 1

    def compute_challenge(self, timestamp: int, sequence: int, nonce: str) -> int:
        data = f"{timestamp}:{sequence}:{nonce}"
        challenge_hash = hashlib.sha256(data.encode()).digest()
        return int.from_bytes(challenge_hash, "big") % BN254_MODULUS

    def compute_expected_hash(self, secret: int, challenge: int) -> str:
        addr_int = int(self.address.replace("0x", ""), 16) % BN254_MODULUS
        expected = (secret * addr_int + challenge) % BN254_MODULUS
        return format(expected, '064x')

    async def generate_zkp_from_backend(self, session, sig_hex, msg_hash_hex, ts, seq, nonce):
        payload = {
            "signature": sig_hex,
            "message_hash": msg_hash_hex,
            "address": self.address,
            "sequence": seq,
            "timestamp": ts,
            "nonce": nonce
        }
        try:
            url = f"{BACKEND_URL}/api/zkp/generate-secure"
            async with session.post(url, json=payload, timeout=aiohttp.ClientTimeout(total=30)) as resp:
                if resp.status == 200:
                    data = await resp.json()
                    proof = data.get("proof", "")
                    exp_hash = data.get("expected_hash", "")
                    if proof.startswith("0x"):
                        proof = proof[2:]
                    if exp_hash.startswith("0x"):
                        exp_hash = exp_hash[2:]
                    return proof, exp_hash, data.get("challenge")
                else:
                    text = await resp.text()
                    print(f"   ⚠️  Backend ZKP {resp.status}: {text[:100]}")
        except asyncio.TimeoutError:
            print(f"   ⚠️  Timeout génération ZKP pour {self.device_id}")
        except Exception as e:
            print(f"   ⚠️  Backend ZKP error: {e}")
        return None, None, None

    async def send_secure_message(self, session):
        try:
            if self.pending_sequence is None:
                self.sequence += 1
                current_seq = self.sequence
                self.pending_sequence = current_seq
            else:
                current_seq = self.pending_sequence
                print(f"   🔄 [{self.device_id}] Retry avec séquence #{current_seq}")
            
            message = self.message_template.format(count=current_seq)
            nonce = secrets.token_hex(16)
            timestamp = int(time.time())

            sig_hex, msg_hash = self.sign_message_ecdsa(message)
            secret = self.generate_zkp_secret(sig_hex)
            challenge = self.compute_challenge(timestamp, current_seq, nonce)

            proof, exp_hash, _ = await self.generate_zkp_from_backend(
                session, sig_hex, msg_hash, timestamp, current_seq, nonce
            )

            if not proof:
                exp_hash = self.compute_expected_hash(secret, challenge)
                proof = "00" * 192
                ptype = "🔶 SIMULÉE"
            else:
                ptype = "🔷 RÉELLE"

            payload = {
                "device_id": self.device_id,
                "address": self.address,
                "message": message,
                "message_hash": msg_hash,
                "signature": sig_hex,
                "proof": proof,
                "expected_hash": exp_hash,
                "timestamp": timestamp,
                "sequence": current_seq,
                "nonce": nonce,
            }

            start_time = time.time()
            url = f"{BACKEND_URL}/api/node/message-secure"
            async with session.post(url, json=payload, timeout=aiohttp.ClientTimeout(total=60)) as resp:
                elapsed = (time.time() - start_time) * 1000
                self.stats["sent"] += 1

                if resp.status == 200:
                    self.stats["success"] += 1
                    result = await resp.json()
                    self.pending_sequence = None

                    if not self.authenticated:
                        self.authenticated = True
                        print(f"🔐 [{self.device_id}] AUTHENTIFIÉ")

                    print(f"✅ [{self.device_id}] #{current_seq:3d} | {elapsed:5.0f}ms | {ptype} | Bloc #{result.get('block', 0)}")
                    return True
                elif resp.status == 400:
                    text = await resp.text()
                    if "Séquence invalide" in text:
                        print(f"⚠️  [{self.device_id}] #{current_seq:3d} | Séquence déjà enregistrée")
                        self.pending_sequence = None
                        return True
                    else:
                        self.stats["failed"] += 1
                        print(f"❌ [{self.device_id}] #{current_seq:3d} | Erreur 400: {text[:80]}")
                        return False
                else:
                    self.stats["failed"] += 1
                    text = await resp.text()
                    print(f"❌ [{self.device_id}] #{current_seq:3d} | Erreur {resp.status}: {text[:80]}")
                    return False

        except asyncio.TimeoutError:
            self.stats["failed"] += 1
            print(f"⏱️  [{self.device_id}] Timeout - réessai en cours...")
            return False
        except Exception as e:
            self.stats["failed"] += 1
            print(f"❌ [{self.device_id}] Exception: {str(e)[:80]}")
            return False

async def node_loop(node, session):
    await asyncio.sleep(node.interval * 0.1)
    consecutive_failures = 0
    while True:
        success = await node.send_secure_message(session)
        if success:
            consecutive_failures = 0
            await asyncio.sleep(node.interval)
        else:
            consecutive_failures += 1
            wait_time = min(2 ** consecutive_failures, 30)
            print(f"   ⏳ [{node.device_id}] Attente {wait_time}s...")
            await asyncio.sleep(wait_time)

async def stats_reporter(nodes):
    await asyncio.sleep(10)
    while True:
        await asyncio.sleep(30)
        print("\n" + "="*70)
        print(f"📊 STATISTIQUES - {datetime.now().strftime('%H:%M:%S')}")
        print("="*70)
        total_sent = total_success = 0
        for node in nodes:
            rate = node.stats["success"] / max(1, node.stats["sent"]) * 100 if node.stats["sent"] > 0 else 0
            status = "✅" if node.authenticated else "⏳"
            print(f"{status} {node.device_id:12}: {node.stats['sent']:4d} envois, "
                  f"{node.stats['success']:4d} succès, {node.stats['failed']:4d} échecs, "
                  f"{rate:5.1f}%")
            total_sent += node.stats['sent']
            total_success += node.stats['success']
        print("-"*70)
        global_rate = total_success / max(1, total_sent) * 100 if total_sent > 0 else 0
        print(f"   {'TOTAL':12}: {total_sent:4d} envois, {total_success:4d} succès, "
              f"{global_rate:5.1f}%")
        print("="*70)

async def check_backend():
    try:
        async with aiohttp.ClientSession() as session:
            url = f"{BACKEND_URL}/api/health"
            async with session.get(url, timeout=5) as resp:
                return True, "✅ Backend OK" if resp.status == 200 else f"❌ Backend HTTP {resp.status}"
    except Exception as e:
        return False, f"❌ Erreur: {e}"

async def main():
    print("\n" + "="*70)
    print("🚀 SIMULATEUR 4 NŒUDS GETH - VERSION CORRIGÉE")
    print(f"🔗 Backend URL: {BACKEND_URL}")
    print("="*70)

    backend_ok, backend_msg = await check_backend()
    print(backend_msg)
    if not backend_ok:
        return

    print("\n📋 Chargement des nœuds...")
    nodes = []
    for cfg in GETH_NODES:
        try:
            node = SecureGethNode(cfg)
            nodes.append(node)
        except Exception as e:
            print(f"❌ {cfg['device_id']}: {e}")

    if not nodes:
        print("\n❌ Aucun nœud valide!")
        return

    print(f"\n📋 RÉSUMÉ DES NŒUDS")
    print("="*70)
    for node in nodes:
        print(f"✅ {node.device_id}: {node.address[:25]}... (interval: {node.interval}s)")

    print(f"\n🔄 Démarrage de {len(nodes)} nœuds...")
    print("   Appuyez sur Ctrl+C pour arrêter\n")

    timeout = aiohttp.ClientTimeout(total=120, connect=10)
    async with aiohttp.ClientSession(timeout=timeout) as session:
        tasks = [asyncio.create_task(node_loop(node, session)) for node in nodes]
        tasks.append(asyncio.create_task(stats_reporter(nodes)))
        try:
            await asyncio.gather(*tasks)
        except KeyboardInterrupt:
            print("\n\n👋 Arrêt demandé!")

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\n\n👋 Au revoir!")
