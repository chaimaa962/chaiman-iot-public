#!/usr/bin/env python3
"""
Simulateur 4 nœuds Geth avec Détection d'Anomalies IA
Version Finale - Intégration ML + ZKP
"""

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
import numpy as np
from collections import deque

load_dotenv(override=True)

BN254_MODULUS = int("21888242871839275222246405745257275088548364400416034343698204186575808495617")

# ============================================
# CONFIGURATION IA - DÉTECTION D'ANOMALIES
# ============================================
ML_ANALYZER_URL = "http://localhost:5000/analyze"
ENABLE_ML_DETECTION = True  # Activer/désactiver la détection IA
ANOMALY_ALERT_THRESHOLD = 0.7  # Seuil d'alerte (score normalisé)

# ============================================
# LES 4 NŒUDS GETH AVEC LEURS VRAIES ADRESSES ET CLÉS (VALIDATEURS POA)
# ============================================
GETH_NODES = [
    {
        "device_id": "GETH_NODE_1",
        "address": "0xaf0c6bf76f11760b7ba90a852aaeadfe50ab9277",
        "private_key": os.getenv("GETH_NODE_1_KEY"),
        "interval": 5,
        "message": "Bonjour je suis le nœud Geth 1 (autorité PoA) - Message #{count}",
        "node_type": "validator",
        "normal_latency_ms": 50,
        "normal_packet_size": 512,
    },
    {
        "device_id": "GETH_NODE_2",
        "address": "0xc4f26670f7539138a21e7f33f2b042dbd1da6f30",
        "private_key": os.getenv("GETH_NODE_2_KEY"),
        "interval": 10,
        "message": "Bonjour je suis le nœud Geth 2 (autorité PoA) - Message #{count}",
        "node_type": "validator",
        "normal_latency_ms": 55,
        "normal_packet_size": 480,
    },
    {
        "device_id": "GETH_NODE_3",
        "address": "0x60e69259368a740e8fe91cc61c5306234e36e01d",
        "private_key": os.getenv("GETH_NODE_3_KEY"),
        "interval": 15,
        "message": "Bonjour je suis le nœud Geth 3 (autorité PoA) - Message #{count}",
        "node_type": "validator",
        "normal_latency_ms": 45,
        "normal_packet_size": 500,
    },
    {
        "device_id": "GETH_NODE_4",
        "address": "0x29885af643612e8b72123ccc3d6f527cd9321319",
        "private_key": os.getenv("GETH_NODE_4_KEY"),
        "interval": 25,
        "message": "Bonjour je suis le nœud Geth 4 (autorité PoA) - Message #{count}",
        "node_type": "validator",
        "normal_latency_ms": 60,
        "normal_packet_size": 520,
    }
]

# ============================================
# CLASSE DE DÉTECTION D'ANOMALIES IA
# ============================================
class MLAnomalyDetector:
    """Client pour le service de détection d'anomalies ML"""
    
    def __init__(self, analyzer_url: str = ML_ANALYZER_URL):
        self.analyzer_url = analyzer_url
        self.enabled = ENABLE_ML_DETECTION
        self.anomaly_history = deque(maxlen=100)  # Historique des anomalies
        self.stats = {
            'analyzed': 0,
            'anomalies': 0,
            'alerts_sent': 0
        }
        
    async def analyze_message(self, session: aiohttp.ClientSession, node_data: dict, 
                             metrics: dict) -> dict:
        """
        Analyse un message avec le modèle ML
        
        Args:
            session: Session aiohttp
            node_data: Données du nœud
            metrics: Métriques collectées (latence, taille, etc.)
        
        Returns:
            Résultat de l'analyse ou None si désactivé/erreur
        """
        if not self.enabled:
            return None
            
        try:
            # Préparer les features pour le modèle ML
            features = self._prepare_features(node_data, metrics)
            
            payload = {
                "device_id": node_data['device_id'],
                **features
            }
            
            async with session.post(
                self.analyzer_url,
                json=payload,
                timeout=aiohttp.ClientTimeout(total=5)
            ) as resp:
                
                if resp.status == 200:
                    result = await resp.json()
                    self.stats['analyzed'] += 1
                    
                    if result.get('is_anomaly'):
                        self.stats['anomalies'] += 1
                        self.anomaly_history.append({
                            'timestamp': datetime.now().isoformat(),
                            'device_id': node_data['device_id'],
                            'result': result
                        })
                        
                        # Alerte si score élevé
                        if result.get('anomaly_score', 0) > ANOMALY_ALERT_THRESHOLD:
                            self.stats['alerts_sent'] += 1
                            await self._send_alert(session, node_data, result)
                    
                    return result
                    
        except asyncio.TimeoutError:
            pass  # Ignorer silencieusement les timeouts ML
        except Exception as e:
            print(f"   ⚠️  ML Analyzer error: {str(e)[:50]}")
            
        return None
    
    def _prepare_features(self, node_data: dict, metrics: dict) -> dict:
        """Prépare les features pour le modèle ML"""
        
        # Features attendues par le modèle (18 features)
        features = {
            # Métriques réseau
            'packet_size': metrics.get('message_size', 500),
            'inter_arrival_time': node_data.get('interval', 10),
            'publish_rate': 1.0 / node_data.get('interval', 10),
            'bytes_sent': metrics.get('message_size', 500),
            'bytes_received': metrics.get('response_size', 200),
            
            # Métriques de performance
            'connection_duration': metrics.get('latency_ms', 50),
            'topic_count': 1,  # Un seul topic par message
            'qos_level': 2,    # QoS élevé pour blockchain
            
            # Métriques calculées
            'sampling_rate': metrics.get('message_rate', 0.1),
            'value_change': metrics.get('latency_variance', 0),
            
            # Valeurs par défaut pour les autres features
            'sensor_value': metrics.get('latency_ms', 50),  # Utiliser la latence comme valeur capteur
            'packet_size_ma': metrics.get('avg_packet_size', 500),
            'publish_rate_std': metrics.get('rate_std', 0.01),
        }
        
        return features
    
    async def _send_alert(self, session: aiohttp.ClientSession, node_data: dict, result: dict):
        """Envoie une alerte au backend"""
        try:
            alert = {
                'type': 'ML_ANOMALY_DETECTED',
                'device_id': node_data['device_id'],
                'address': node_data['address'],
                'severity': result.get('severity', 'UNKNOWN'),
                'anomaly_score': result.get('anomaly_score', 0),
                'timestamp': datetime.now().isoformat(),
                'details': result
            }
            
            async with session.post(
                "http://localhost:8080/api/alerts/ml-anomaly",
                json=alert,
                timeout=aiohttp.ClientTimeout(total=5)
            ) as resp:
                if resp.status == 200:
                    print(f"   🚨 ALERTE ML ENVOYÉE: {node_data['device_id']} - {result['severity']}")
                    
        except Exception as e:
            print(f"   ⚠️  Erreur envoi alerte: {e}")
    
    def get_stats(self) -> dict:
        """Retourne les statistiques du détecteur ML"""
        return {
            **self.stats,
            'detection_rate': self.stats['anomalies'] / max(1, self.stats['analyzed']) * 100,
            'history_size': len(self.anomaly_history),
            'enabled': self.enabled
        }

# ============================================
# CLASSE SECUREGETHNODE AMÉLIORÉE AVEC ML
# ============================================
class SecureGethNode:
    def __init__(self, config, ml_detector: MLAnomalyDetector = None):
        self.device_id = config["device_id"]
        self.address = config["address"].lower()
        self.interval = config["interval"]
        self.message_template = config["message"]
        self.node_type = config.get("node_type", "validator")
        self.normal_latency_ms = config.get("normal_latency_ms", 50)
        self.normal_packet_size = config.get("normal_packet_size", 500)
        
        self.sequence = 0
        self.stats = {
            "sent": 0, 
            "success": 0, 
            "failed": 0,
            "ml_anomalies": 0,
            "avg_latency": 0
        }
        self.authenticated = False
        self.pending_sequence = None
        self.ml_detector = ml_detector
        
        # Historique pour ML
        self.latency_history = deque(maxlen=30)
        self.packet_size_history = deque(maxlen=30)
        
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
            print(f"✅ {self.device_id}: {self.address[:20]}... (clé OK, ML activé)")

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
            async with session.post("http://localhost:8080/api/zkp/generate-secure", 
                                   json=payload, timeout=aiohttp.ClientTimeout(total=30)) as resp:
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

    def _calculate_metrics(self, message: str, latency_ms: float, response_size: int = 200) -> dict:
        """Calcule les métriques pour l'analyse ML"""
        
        # Mettre à jour l'historique
        self.latency_history.append(latency_ms)
        self.packet_size_history.append(len(message))
        
        # Calculer les statistiques
        avg_latency = np.mean(self.latency_history) if self.latency_history else latency_ms
        latency_variance = np.var(self.latency_history) if len(self.latency_history) > 1 else 0
        avg_packet_size = np.mean(self.packet_size_history) if self.packet_size_history else len(message)
        
        # Détecter les anomalies simples (complément à l'IA)
        is_latency_anomaly = latency_ms > self.normal_latency_ms * 3
        is_size_anomaly = len(message) > self.normal_packet_size * 5
        
        return {
            'message_size': len(message),
            'response_size': response_size,
            'latency_ms': latency_ms,
            'avg_latency': avg_latency,
            'latency_variance': latency_variance,
            'avg_packet_size': avg_packet_size,
            'message_rate': 1.0 / self.interval,
            'rate_std': 0.01,
            'is_latency_anomaly': is_latency_anomaly,
            'is_size_anomaly': is_size_anomaly
        }

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

            # Mesurer la latence
            start_time = time.time()
            
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

            async with session.post("http://localhost:8080/api/node/message-secure", 
                                   json=payload, 
                                   timeout=aiohttp.ClientTimeout(total=60)) as resp:
                
                latency_ms = (time.time() - start_time) * 1000
                self.stats["sent"] += 1
                
                # Mettre à jour la latence moyenne
                alpha = 0.3
                self.stats['avg_latency'] = (alpha * latency_ms + 
                                            (1 - alpha) * self.stats.get('avg_latency', latency_ms))

                # ============================================
                # ANALYSE ML - DÉTECTION D'ANOMALIES
                # ============================================
                if self.ml_detector:
                    metrics = self._calculate_metrics(message, latency_ms)
                    ml_result = await self.ml_detector.analyze_message(
                        session,
                        {'device_id': self.device_id, 'address': self.address, 'interval': self.interval},
                        metrics
                    )
                    
                    if ml_result and ml_result.get('is_anomaly'):
                        self.stats['ml_anomalies'] += 1
                        anomaly_emoji = "🚨" if ml_result['severity'] in ['HIGH', 'CRITICAL'] else "⚠️"
                    else:
                        anomaly_emoji = ""
                else:
                    anomaly_emoji = ""

                if resp.status == 200:
                    self.stats["success"] += 1
                    result = await resp.json()
                    self.pending_sequence = None

                    if not self.authenticated:
                        self.authenticated = True
                        print(f"🔐 [{self.device_id}] AUTHENTIFIÉ")

                    status_line = (f"✅ [{self.device_id}] #{current_seq:3d} | "
                                 f"{latency_ms:5.0f}ms | {ptype} | Bloc #{result.get('block', 0)}")
                    
                    if anomaly_emoji:
                        status_line += f" | {anomaly_emoji} ML:{ml_result['severity']}"
                    
                    print(status_line)
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

# ============================================
# FONCTIONS PRINCIPALES
# ============================================
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

async def stats_reporter(nodes, ml_detector=None):
    await asyncio.sleep(10)
    while True:
        await asyncio.sleep(30)
        print("\n" + "="*80)
        print(f"📊 STATISTIQUES - {datetime.now().strftime('%H:%M:%S')}")
        print("="*80)
        
        total_sent = total_success = total_ml_anomalies = 0
        for node in nodes:
            rate = node.stats["success"] / max(1, node.stats["sent"]) * 100 if node.stats["sent"] > 0 else 0
            status = "✅" if node.authenticated else "⏳"
            
            print(f"{status} {node.device_id:12}: {node.stats['sent']:4d} envois, "
                  f"{node.stats['success']:4d} succès, {node.stats['failed']:4d} échecs, "
                  f"{rate:5.1f}%, {node.stats['avg_latency']:5.0f}ms moy, "
                  f"ML:{node.stats['ml_anomalies']:3d}")
            
            total_sent += node.stats['sent']
            total_success += node.stats['success']
            total_ml_anomalies += node.stats['ml_anomalies']
        
        print("-"*80)
        global_rate = total_success / max(1, total_sent) * 100 if total_sent > 0 else 0
        print(f"   {'TOTAL':12}: {total_sent:4d} envois, {total_success:4d} succès, "
              f"{global_rate:5.1f}%, ML anomalies: {total_ml_anomalies}")
        
        # Statistiques ML
        if ml_detector and ml_detector.enabled:
            ml_stats = ml_detector.get_stats()
            print(f"   {'ML Detector':12}: {ml_stats['analyzed']} analysés, "
                  f"{ml_stats['anomalies']} anomalies ({ml_stats['detection_rate']:.1f}%), "
                  f"{ml_stats['alerts_sent']} alertes")
        
        print("="*80)

async def check_backend():
    try:
        async with aiohttp.ClientSession() as session:
            async with session.get("http://localhost:8080/api/health", timeout=5) as resp:
                return True, "✅ Backend OK" if resp.status == 200 else f"❌ Backend HTTP {resp.status}"
    except Exception as e:
        return False, f"❌ Erreur: {e}"

async def check_ml_service():
    """Vérifie si le service ML est disponible"""
    try:
        async with aiohttp.ClientSession() as session:
            async with session.get("http://localhost:5000/health", timeout=3) as resp:
                if resp.status == 200:
                    data = await resp.json()
                    return True, f"✅ ML Service OK (model: {data.get('model', 'loaded')})"
                return False, f"⚠️  ML Service HTTP {resp.status}"
    except Exception as e:
        return False, f"⚠️  ML Service non disponible: {str(e)[:50]}"

async def main():
    print("\n" + "="*80)
    print("🚀 SIMULATEUR 4 NŒUDS GETH + DÉTECTION ANOMALIES IA")
    print("="*80)

    # Vérifier le backend
    backend_ok, backend_msg = await check_backend()
    print(backend_msg)
    if not backend_ok:
        return

    # Vérifier le service ML
    ml_ok, ml_msg = await check_ml_service()
    print(ml_msg)
    
    # Initialiser le détecteur ML
    ml_detector = MLAnomalyDetector() if ml_ok else None
    if ml_detector:
        print("✅ Détection d'anomalies IA ACTIVÉE")
    else:
        print("⚠️  Détection d'anomalies IA DÉSACTIVÉE (service non disponible)")

    print("\n📋 Chargement des nœuds...")
    nodes = []
    for cfg in GETH_NODES:
        try:
            node = SecureGethNode(cfg, ml_detector)
            nodes.append(node)
        except Exception as e:
            print(f"❌ {cfg['device_id']}: {e}")

    if not nodes:
        print("\n❌ Aucun nœud valide!")
        return

    print(f"\n📋 RÉSUMÉ DES NŒUDS")
    print("="*80)
    for node in nodes:
        print(f"✅ {node.device_id}: {node.address[:25]}... "
              f"(interval: {node.interval}s, type: {node.node_type})")

    print(f"\n🔄 Démarrage de {len(nodes)} nœuds avec surveillance IA...")
    print("   Appuyez sur Ctrl+C pour arrêter\n")

    timeout = aiohttp.ClientTimeout(total=120, connect=10)
    async with aiohttp.ClientSession(timeout=timeout) as session:
        tasks = [asyncio.create_task(node_loop(node, session)) for node in nodes]
        tasks.append(asyncio.create_task(stats_reporter(nodes, ml_detector)))
        try:
            await asyncio.gather(*tasks)
        except KeyboardInterrupt:
            print("\n\n👋 Arrêt demandé!")
            
            # Afficher résumé final
            if ml_detector:
                ml_stats = ml_detector.get_stats()
                print(f"\n📊 RÉSUMÉ FINAL ML:")
                print(f"   Messages analysés: {ml_stats['analyzed']}")
                print(f"   Anomalies détectées: {ml_stats['anomalies']}")
                print(f"   Alertes envoyées: {ml_stats['alerts_sent']}")

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\n\n👋 Au revoir!")
