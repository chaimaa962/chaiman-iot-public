#!/usr/bin/env python3
import os, json, numpy as np
from pathlib import Path
from typing import Dict, List, Optional, Any
from datetime import datetime
from collections import deque
os.environ['TF_CPP_MIN_LOG_LEVEL'] = '3'
import tensorflow as tf
tf.get_logger().setLevel('ERROR')

class IoTAnomalyDetector:
    def __init__(self, model_dir: str = '/app/models'):
        self.model_dir = Path(model_dir)
        with open(self.model_dir / 'metadata_v3_tflite.json', 'r') as f:
            self.metadata = json.load(f)
        self.threshold = self.metadata.get('selected_threshold', 0.2654)
        self.seq_len = self.metadata.get('sequence_length', 30)
        self.feature_cols = self.metadata.get('feature_columns', [])
        
        # CHARGER LE MODÈLE KERAS (pas TFLite)
        keras_files = list(self.model_dir.glob('*.keras'))
        print(f"📦 Chargement du modèle Keras: {keras_files[0].name}")
        self.model = tf.keras.models.load_model(str(keras_files[0]), compile=False)
        self.buffers: Dict[str, deque] = {}
        print(f"✅ Modèle Keras chargé - Seuil: {self.threshold:.4f}")
    
    def _normalize(self, data):
        return (data - np.mean(data, axis=0)) / (np.std(data, axis=0) + 1e-6)
    
    def preprocess_message(self, msg):
        return [float(msg.get(c, 0)) for c in self.feature_cols]
    
    def add_message(self, dev, msg):
        f = self.preprocess_message(msg)
        if dev not in self.buffers:
            self.buffers[dev] = deque(maxlen=self.seq_len*2)
        self.buffers[dev].append(f)
        return len(self.buffers[dev]) >= self.seq_len
    
    def predict(self, dev):
        if dev not in self.buffers or len(self.buffers[dev]) < self.seq_len:
            return None
        seq = np.array(list(self.buffers[dev])[-self.seq_len:], dtype=np.float32)
        seq_norm = np.expand_dims(self._normalize(seq), 0)
        rec = self.model.predict(seq_norm, verbose=0)
        mse = np.mean((seq_norm - rec)**2)
        is_anom = mse > self.threshold
        score = mse / self.threshold
        sev = "CRITICAL" if score>5 else ("HIGH" if score>3 else ("MEDIUM" if is_anom else "NORMAL"))
        return {'device_id':dev, 'is_anomaly':bool(is_anom), 'anomaly_score':float(score), 'severity':sev}
