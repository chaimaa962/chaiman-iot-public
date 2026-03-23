# ai_engine/anomaly_detector.py
import numpy as np
import pandas as pd
from sklearn.ensemble import IsolationForest
from sklearn.preprocessing import StandardScaler
import tensorflow as tf
from tensorflow import keras
import redis
import json
from datetime import datetime, timedelta
from typing import Dict, List, Tuple
import asyncio

class IoTBehaviorAnalyzer:
    """
    Moteur d'IA pour détecter les comportements anormaux dans l'authentification IoT
    Combine ML classique et Deep Learning pour détection en temps réel
    """
    
    def __init__(self):
        self.scaler = StandardScaler()
        self.isolation_forest = IsolationForest(contamination=0.1, random_state=42)
        self.lstm_model = self._build_lstm_model()
        self.redis_client = redis.Redis(host='localhost', port=6379, db=0)
        self.device_profiles = {}
        
    def _build_lstm_model(self) -> keras.Model:
        """LSTM pour détection de séquences temporelles anormales"""
        model = keras.Sequential([
            keras.layers.LSTM(64, return_sequences=True, input_shape=(10, 5)),
            keras.layers.Dropout(0.2),
            keras.layers.LSTM(32),
            keras.layers.Dropout(0.2),
            keras.layers.Dense(16, activation='relu'),
            keras.layers.Dense(1, activation='sigmoid')  # Anomalie ou non
        ])
        model.compile(optimizer='adam', loss='binary_crossentropy', metrics=['accuracy'])
        return model
    
    def extract_features(self, auth_events: List[Dict]) -> np.ndarray:
        """
        Extrait les features comportementales d'une séquence d'authentifications
        
        Features:
        - Fréquence d'authentification
        - Taux de succès/échec
        - Variabilité des horaires (entropie temporelle)
        - Temps de réponse moyen
        - Patterns géographiques (si disponible)
        """
        if len(auth_events) < 5:
            return np.zeros(5)
            
        df = pd.DataFrame(auth_events)
        df['timestamp'] = pd.to_datetime(df['timestamp'], unit='s')
        
        features = []
        
        # 1. Fréquence (auths par heure)
        time_diffs = df['timestamp'].diff().dt.total_seconds()
        avg_interval = time_diffs.mean()
        frequency = 3600 / avg_interval if avg_interval > 0 else 0
        features.append(min(frequency, 100))  # Cap à 100/h
        
        # 2. Taux de succès
        success_rate = df['success'].mean()
        features.append(success_rate)
        
        # 3. Entropie temporelle (régularité suspecte = faible entropie)
        hours = df['timestamp'].dt.hour
        hour_dist = hours.value_counts(normalize=True)
        entropy = -sum(p * np.log2(p) for p in hour_dist if p > 0)
        features.append(entropy)
        
        # 4. Temps de réponse
        avg_response = df.get('response_time', pd.Series([0]*len(df))).mean()
        features.append(avg_response)
        
        # 5. Variabilité des signatures ZKP (si nullifiers uniques réutilisés)
        unique_nullifiers = df.get('nullifier', pd.Series()).nunique()
        reuse_ratio = 1 - (unique_nullifiers / len(df))
        features.append(reuse_ratio)
        
        return np.array(features)
    
    async def analyze_device(self, device_address: str) -> Dict:
        """
        Analyse complète d'un device et retourne un score d'anomalie
        """
        # Récupérer l'historique depuis Redis + Blockchain
        history = await self._get_device_history(device_address)
        
        if len(history) < 10:
            return {"score": 0, "status": "insufficient_data", "action": "none"}
        
        # Features actuelles
        current_features = self.extract_features(history[-20:])
        
        # 1. Détection Isolation Forest (comportement hors distribution)
        features_scaled = self.scaler.fit_transform([current_features])
        isolation_score = self.isolation_forest.fit_predict(features_scaled)[0]
        
        # 2. Détection LSTM (patterns temporels)
        sequence = self._create_sequence(history[-50:])
        lstm_pred = self.lstm_model.predict(sequence.reshape(1, 10, 5))[0][0]
        
        # 3. Règles métier
        rule_score = self._apply_business_rules(history)
        
        # Score composite 0-100
        composite_score = (
            (50 if isolation_score == -1 else 0) * 0.3 +  # IF
            lstm_pred * 100 * 0.4 +  # LSTM
            rule_score * 0.3  # Règles
        )
        
        # Décision
        action = self._decide_action(composite_score, history)
        
        result = {
            "score": round(composite_score, 2),
            "isolation_forest": isolation_score == -1,
            "lstm_anomaly": float(lstm_pred),
            "rule_violations": rule_score,
            "status": "anomaly" if composite_score > 70 else "warning" if composite_score > 40 else "normal",
            "action": action,
            "timestamp": datetime.now().isoformat()
        }
        
        # Mettre à jour le profil
        await self._update_device_profile(device_address, result)
        
        return result
    
    def _apply_business_rules(self, history: List[Dict]) -> float:
        """Règles métier spécifiques à l'IoT"""
        violations = 0
        
        # Règle 1: Burst d'authentification (>10 en 1 minute)
        recent = [h for h in history if h['timestamp'] > time.time() - 60]
        if len(recent) > 10:
            violations += 30
        
        # Règle 2: Échecs répétés (>5 consécutifs)
        recent_success = [h['success'] for h in history[-10:]]
        if recent_success.count(False) > 5:
            violations += 25
            
        # Règle 3: Horaires suspects (3h-5h du matin)
        night_auths = sum(1 for h in history[-20:] 
                         if 3 <= datetime.fromtimestamp(h['timestamp']).hour <= 5)
        if night_auths > 3:
            violations += 20
            
        return min(violations, 100)
    
    def _decide_action(self, score: float, history: List[Dict]) -> str:
        """
        Décide de l'action à prendre basée sur le score et l'historique
        """
        if score > 90:
            return "IMMEDIATE_REVOKE"  # Révocation immédiate
        elif score > 70:
            # Vérifier si pattern récurrent
            recent_anomalies = sum(1 for h in history[-5:] if h.get('score', 0) > 70)
            if recent_anomalies >= 3:
                return "REVOKE"  # Révocation après confirmation
            return "CHALLENGE"  # Demande authentification supplémentaire
        elif score > 40:
            return "MONITOR"  # Surveillance renforcée
        return "NONE"
    
    async def _get_device_history(self, device_address: str) -> List[Dict]:
        """Récupère l'historique mixte Redis + Blockchain"""
        # Cache Redis (données récentes)
        cached = self.redis_client.get(f"auth_history:{device_address}")
        if cached:
            return json.loads(cached)
        
        # Sinon, récupérer depuis l'API blockchain
        # (implémentation dépend de ton backend)
        return []
    
    def _create_sequence(self, history: List[Dict]) -> np.ndarray:
        """Crée une séquence temporelle pour le LSTM"""
        sequences = []
        for i in range(len(history) - 9):
            window = history[i:i+10]
            seq_features = [self.extract_features([w]) for w in window]
            sequences.append(seq_features)
        return np.array(sequences[-1]) if sequences else np.zeros((10, 5))
    
    async def _update_device_profile(self, device_address: str, analysis: Dict):
        """Met à jour le profil et déclenche les actions"""
        self.redis_client.setex(
            f"device_profile:{device_address}",
            3600,  # TTL 1h
            json.dumps(analysis)
        )
        
        # Si révocation nécessaire, appeler le smart contract
        if analysis["action"] in ["REVOKE", "IMMEDIATE_REVOKE"]:
            await self._trigger_revocation(device_address, analysis)
    
    async def _trigger_revocation(self, device_address: str, analysis: Dict):
        """Déclenche la révocation on-chain via l'API backend"""
        import aiohttp
        
        async with aiohttp.ClientSession() as session:
            payload = {
                "deviceAddress": device_address,
                "reason": f"AI Anomaly Score: {analysis['score']}",
                "proof": analysis
            }
            async with session.post(
                "http://localhost:8080/api/device/revoke",
                json=payload
            ) as resp:
                if resp.status == 200:
                    print(f"🔒 Device {device_address} revoked by AI decision")
                else:
                    print(f"❌ Failed to revoke {device_address}")

# Service d'API pour l'IA
from fastapi import FastAPI, BackgroundTasks

app = FastAPI()
analyzer = IoTBehaviorAnalyzer()

@app.post("/analyze/{device_address}")
async def analyze_device(device_address: str, background_tasks: BackgroundTasks):
    """Endpoint pour analyse à la volée"""
    result = await analyzer.analyze_device(device_address)
    
    # Si anomalie, vérifier périodiquement
    if result["score"] > 40:
        background_tasks.add_task(periodic_recheck, device_address)
    
    return result

async def periodic_recheck(device_address: str):
    """Recheck périodique pour les devices suspects"""
    for _ in range(3):  # 3 vérifications supplémentaires
        await asyncio.sleep(60)  # Toutes les minutes
        result = await analyzer.analyze_device(device_address)
        if result["action"] == "REVOKE":
            break
