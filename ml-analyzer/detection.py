# ai_engine/anomaly_detector_v4.py
import numpy as np
import tensorflow as tf
from tensorflow import keras
import redis
import json
from datetime import datetime, timedelta
from typing import Dict, List
import asyncio
import aiohttp
import joblib
from pathlib import Path
import threading
import time

class IoTNetworkAnalyzerV4:
    """
    Moteur d'IA pour détecter les anomalies réseau avec le modèle V4
    Utilise le modèle Bidirectional LSTM + Attention (F1=96.45%)
    AVEC APPRENTISSAGE AUTOMATIQUE SUR LES DONNÉES NORMALES
    """
    
    def __init__(self, model_path: str = "../ml-analyzer/models"):
        self.model_path = Path(model_path)
        self.model_path.mkdir(parents=True, exist_ok=True)
        
        # Charger le modèle V4
        keras_file = self.model_path / 'autoencoder_3datasets_v4_98f1.keras'
        finetuned_file = self.model_path / 'model_finetuned.keras'
        
        if finetuned_file.exists():
            self.model = tf.keras.models.load_model(str(finetuned_file), compile=False, safe_mode=False)
            print(f"✅ Modèle fine-tuné chargé ({self.model.count_params():,} paramètres)")
        elif keras_file.exists():
            self.model = tf.keras.models.load_model(str(keras_file), compile=False, safe_mode=False)
            print(f"✅ Modèle V4 chargé ({self.model.count_params():,} paramètres)")
        else:
            print("⚠️ Aucun modèle trouvé - utilisation du mode détection simple")
            self.model = None
        
        # Charger metadata
        metadata_file = self.model_path / 'metadata_v4.json'
        if metadata_file.exists():
            with open(metadata_file, 'r') as f:
                self.metadata = json.load(f)
            self.threshold = self.metadata.get('selected_threshold', 0.1437)
            self.seq_len = self.metadata.get('sequence_length', 10)
            self.feature_cols = self.metadata.get('feature_columns', [
                'duration', 'bytes_sent', 'bytes_received', 'packet_count',
                'bytes_per_sec', 'pkts_per_sec', 'bytes_per_pkt', 'bytes_ratio',
                'bytes_sent_ratio', 'bytes_total', 'log_bytes_total', 'pkt_size_variance',
                'high_volume_flag', 'burst_indicator'
            ])
        else:
            self.threshold = 0.1437
            self.seq_len = 10
            self.feature_cols = [
                'duration', 'bytes_sent', 'bytes_received', 'packet_count',
                'bytes_per_sec', 'pkts_per_sec', 'bytes_per_pkt', 'bytes_ratio',
                'bytes_sent_ratio', 'bytes_total', 'log_bytes_total', 'pkt_size_variance',
                'high_volume_flag', 'burst_indicator'
            ]
        
        # Charger le scaler
        scaler_file = self.model_path / 'scaler.pkl'
        try:
            self.scaler = joblib.load(scaler_file)
            print(f"✅ Scaler chargé")
        except:
            self.scaler = None
            print(f"⚠️ Normalisation dynamique activée")
        
        # Buffers pour chaque device
        self.buffers: Dict[str, List] = {}
        self.buffers_lock = threading.Lock()
        
        # 🔥 BUFFER D'APPRENTISSAGE
        self.training_buffer: List[List[float]] = []
        self.training_labels: List[int] = []  # 0 = normal, 1 = anomaly
        self.is_training = False
        self.training_start_time = None
        self.min_training_samples = 100  # Minimum pour fine-tuning
        self.last_finetune_time = None
        
        # Redis pour cache
        try:
            self.redis_client = redis.Redis(host='localhost', port=6379, db=0)
        except:
            self.redis_client = None
        
        # Démarrer le thread d'apprentissage automatique
        self.learning_thread = threading.Thread(target=self._auto_learning_loop, daemon=True)
        self.learning_thread.start()
        
        print(f"✅ Analyseur V4 initialisé - Seuil: {self.threshold:.6f}")
        print(f"🧠 Apprentissage automatique ACTIVÉ - Min samples: {self.min_training_samples}")
    
    def _normalize(self, data: np.ndarray) -> np.ndarray:
        """Normalisation z-score"""
        if self.scaler is not None:
            return self.scaler.transform(data)
        else:
            mean = np.mean(data, axis=0, keepdims=True)
            std = np.std(data, axis=0, keepdims=True)
            std[std == 0] = 1.0
            return (data - mean) / std
    
    def add_message(self, device_address: str, features: Dict) -> bool:
        """Ajoute un message au buffer du device ET au buffer d'apprentissage"""
        with self.buffers_lock:
            if device_address not in self.buffers:
                self.buffers[device_address] = []
            
            # Extraire les features dans le bon ordre
            feature_vector = [float(features.get(col, 0.0)) for col in self.feature_cols]
            self.buffers[device_address].append(feature_vector)
            
            # Garder seulement les seq_len * 2 derniers messages
            if len(self.buffers[device_address]) > self.seq_len * 2:
                self.buffers[device_address] = self.buffers[device_address][-self.seq_len * 2:]
            
            # 🔥 AJOUTER AU BUFFER D'APPRENTISSAGE
            self.training_buffer.append(feature_vector)
            self.training_labels.append(0)  # Par défaut = normal
            
            # Garder max 10000 échantillons
            if len(self.training_buffer) > 10000:
                self.training_buffer = self.training_buffer[-10000:]
                self.training_labels = self.training_labels[-10000:]
        
        return len(self.buffers[device_address]) >= self.seq_len
    
    def mark_as_anomaly(self, device_address: str):
        """Marque les derniers messages d'un device comme anomalies"""
        with self.buffers_lock:
            if device_address in self.buffers:
                buffer_len = len(self.buffers[device_address])
                # Marquer les derniers messages comme anomalies dans le training buffer
                for i in range(min(self.seq_len, buffer_len)):
                    if len(self.training_labels) > i:
                        self.training_labels[-(i+1)] = 1
    
    def _auto_learning_loop(self):
        """Boucle d'apprentissage automatique - fine-tuning périodique"""
        while True:
            time.sleep(300)  # Vérifier toutes les 5 minutes
            
            if self.model is None:
                continue
            
            if len(self.training_buffer) < self.min_training_samples:
                continue
            
            # Vérifier si on a assez de nouvelles données
            if self.last_finetune_time and (datetime.now() - self.last_finetune_time).seconds < 1800:
                continue  # Attendre 30 minutes entre chaque fine-tuning
            
            print(f"\n🧠 DÉMARRAGE APPRENTISSAGE AUTOMATIQUE...")
            print(f"   Échantillons: {len(self.training_buffer)}")
            print(f"   Anomalies: {sum(self.training_labels)}")
            print(f"   Normaux: {len(self.training_labels) - sum(self.training_labels)}")
            
            self._fine_tune_model()
    
    def _fine_tune_model(self):
        """Fine-tuning du modèle sur les données collectées"""
        try:
            self.is_training = True
            self.training_start_time = datetime.now()
            
            # Préparer les données
            X = np.array(self.training_buffer[-5000:], dtype=np.float32)
            
            # Normaliser
            if self.scaler is not None:
                X_norm = self.scaler.transform(X)
            else:
                mean = np.mean(X, axis=0)
                std = np.std(X, axis=0)
                std[std == 0] = 1.0
                X_norm = (X - mean) / std
            
            # Créer des séquences
            sequences = []
            for i in range(0, len(X_norm) - self.seq_len, self.seq_len // 2):
                seq = X_norm[i:i + self.seq_len]
                if len(seq) == self.seq_len:
                    sequences.append(seq)
            
            if len(sequences) < 10:
                print("⚠️ Pas assez de séquences pour l'apprentissage")
                return
            
            X_seq = np.array(sequences, dtype=np.float32)
            
            # Fine-tuning
            print(f"   🔄 Fine-tuning sur {len(X_seq)} séquences...")
            
            # Compiler le modèle pour le fine-tuning
            self.model.compile(optimizer=keras.optimizers.Adam(learning_rate=0.0001), loss='mse')
            
            # Entraîner
            history = self.model.fit(
                X_seq, X_seq,
                epochs=5,
                batch_size=16,
                validation_split=0.1,
                verbose=1
            )
            
            # Recalculer le seuil sur les données normales
            reconstructions = self.model.predict(X_seq, verbose=0)
            mse = np.mean((X_seq - reconstructions) ** 2, axis=(1, 2))
            
            # Nouveau seuil = 95ème percentile des erreurs
            new_threshold = np.percentile(mse, 95)
            old_threshold = self.threshold
            self.threshold = float(new_threshold)
            
            # Sauvegarder le modèle fine-tuné
            finetuned_file = self.model_path / 'model_finetuned.keras'
            self.model.save(str(finetuned_file))
            
            # Sauvegarder le nouveau seuil
            self.metadata['selected_threshold'] = self.threshold
            with open(self.model_path / 'metadata_v4.json', 'w') as f:
                json.dump(self.metadata, f, indent=2)
            
            self.last_finetune_time = datetime.now()
            
            print(f"   ✅ Fine-tuning terminé!")
            print(f"   📊 Seuil: {old_threshold:.6f} → {self.threshold:.6f}")
            print(f"   💾 Modèle sauvegardé: {finetuned_file}")
            
        except Exception as e:
            print(f"   ❌ Erreur fine-tuning: {e}")
        finally:
            self.is_training = False
    
    async def analyze_device(self, device_address: str) -> Dict:
        """Analyse complète d'un device avec le modèle V4"""
        if self.model is None:
            return {"score": 0, "status": "no_model", "action": "none", "message": "Modèle non chargé"}
        
        if device_address not in self.buffers or len(self.buffers[device_address]) < self.seq_len:
            return {
                "score": 0, "status": "insufficient_data", "action": "none",
                "message": f"Besoin de {self.seq_len} messages (actuel: {len(self.buffers.get(device_address, []))})"
            }
        
        sequence = np.array(self.buffers[device_address][-self.seq_len:], dtype=np.float32)
        sequence_norm = self._normalize(sequence)
        sequence_norm = np.expand_dims(sequence_norm, axis=0)
        
        reconstructed = self.model.predict(sequence_norm, verbose=0)
        mse = np.mean((sequence_norm - reconstructed) ** 2)
        
        is_anomaly = mse > self.threshold
        score = min(mse / self.threshold * 50, 100) if self.threshold > 0 else 0
        
        if score > 90: severity, action = "CRITICAL", "IMMEDIATE_REVOKE"
        elif score > 70: severity, action = "HIGH", "REVOKE"
        elif score > 40: severity, action = "MEDIUM", "CHALLENGE"
        elif is_anomaly: severity, action = "LOW", "MONITOR"
        else: severity, action = "NORMAL", "NONE"
        
        # 🔥 Si anomalie détectée, marquer dans le buffer d'apprentissage
        if is_anomaly:
            self.mark_as_anomaly(device_address)
        
        result = {
            "device_address": device_address, "score": round(score, 2),
            "mse": float(mse), "threshold": self.threshold,
            "is_anomaly": bool(is_anomaly), "severity": severity,
            "action": action, "status": "anomaly" if is_anomaly else "normal",
            "timestamp": datetime.now().isoformat(),
            "model_version": "V4 Fine-tuned" if self.last_finetune_time else "V4 Base",
            "f1_score": 96.45
        }
        
        if self.redis_client:
            try:
                self.redis_client.setex(f"device_analysis:{device_address}", 300, json.dumps(result))
            except: pass
        
        if action in ["REVOKE", "IMMEDIATE_REVOKE"]:
            await self._trigger_revocation(device_address, result)
        
        return result
    
    async def _trigger_revocation(self, device_address: str, analysis: Dict):
        """Déclenche la révocation on-chain"""
        try:
            async with aiohttp.ClientSession() as session:
                payload = {
                    "deviceAddress": device_address,
                    "reason": f"AI V4 Anomaly Score: {analysis['score']:.2f} - {analysis['severity']}",
                    "proof": {"mse": analysis['mse'], "threshold": self.threshold, "model": "V4"}
                }
                async with session.post("http://localhost:8080/api/device/deactivate", json=payload) as resp:
                    if resp.status == 200:
                        print(f"🔒 Device {device_address[:15]}... REVOKED (score: {analysis['score']:.2f})")
        except Exception as e:
            print(f"❌ Error: {e}")
    
    def get_learning_status(self) -> Dict:
        """Retourne le statut de l'apprentissage"""
        return {
            "is_training": self.is_training,
            "training_samples": len(self.training_buffer),
            "anomalies_labeled": sum(self.training_labels),
            "last_finetune": self.last_finetune_time.isoformat() if self.last_finetune_time else None,
            "current_threshold": self.threshold,
            "model_type": "Fine-tuned" if self.last_finetune_time else "Base V4"
        }
    
    def get_device_stats(self, device_address: str) -> Dict:
        if device_address not in self.buffers:
            return {"error": "Device not found"}
        return {
            "device_address": device_address,
            "messages_collected": len(self.buffers[device_address]),
            "ready_for_analysis": len(self.buffers[device_address]) >= self.seq_len,
            "features": self.feature_cols,
            "threshold": self.threshold
        }
    
    def get_all_devices(self) -> List[str]:
        return list(self.buffers.keys())


# ============================================
# API FastAPI pour l'analyseur V4
# ============================================
from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI(title="IoT Network Analyzer V4 - Auto-Learning")
analyzer = IoTNetworkAnalyzerV4()

class NetworkFeatures(BaseModel):
    duration: float = 0.0
    bytes_sent: float = 0
    bytes_received: float = 0
    packet_count: float = 0
    bytes_per_sec: float = 0
    pkts_per_sec: float = 0
    bytes_per_pkt: float = 0
    bytes_ratio: float = 0
    bytes_sent_ratio: float = 0
    bytes_total: float = 0
    log_bytes_total: float = 0
    pkt_size_variance: float = 0
    high_volume_flag: float = 0
    burst_indicator: float = 0

class MessageRequest(BaseModel):
    device_address: str
    features: NetworkFeatures

@app.post("/v4/add_message")
async def add_message(request: MessageRequest):
    has_enough = analyzer.add_message(request.device_address, request.features.dict())
    return {
        "device_address": request.device_address,
        "has_enough_data": has_enough,
        "buffer_size": len(analyzer.buffers.get(request.device_address, [])),
        "learning_status": analyzer.get_learning_status()
    }

@app.get("/v4/analyze/{device_address}")
async def analyze_device(device_address: str):
    return await analyzer.analyze_device(device_address)

@app.get("/v4/learning/status")
async def learning_status():
    return analyzer.get_learning_status()

@app.post("/v4/learning/force")
async def force_learning():
    """Force un fine-tuning immédiat"""
    analyzer._fine_tune_model()
    return {"status": "completed", "details": analyzer.get_learning_status()}

@app.get("/v4/stats/{device_address}")
async def get_device_stats(device_address: str):
    return analyzer.get_device_stats(device_address)

@app.get("/v4/devices")
async def get_all_devices():
    return {"devices": analyzer.get_all_devices()}

@app.get("/v4/health")
async def health():
    return {
        "status": "healthy",
        "model": "V4 (Bidirectional LSTM + Attention) + Auto-Learning",
        "f1_score": 96.45,
        "threshold": analyzer.threshold,
        "features": len(analyzer.feature_cols),
        "sequence_length": analyzer.seq_len,
        "learning": analyzer.get_learning_status()
    }

@app.get("/health")
async def health_alt():
    return await health()

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=5001)
