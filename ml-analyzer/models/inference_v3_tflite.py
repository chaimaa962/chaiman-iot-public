#!/usr/bin/env python3
import os
os.environ['TF_CPP_MIN_LOG_LEVEL'] = '2'

import numpy as np
import tensorflow as tf
import pickle, json
from pathlib import Path

class IoTAnomalyDetectorTFLite:
    def __init__(self, model_dir='./'):
        self.model_dir = Path(model_dir)
        
        with open(self.model_dir / 'metadata_v3_tflite.json') as f:
            self.meta = json.load(f)
        with open(self.model_dir / 'scaler_v3_tflite.pkl', 'rb') as f:
            self.scaler = pickle.load(f)
        
        self.threshold = self.meta['selected_threshold']
        self.seq_len = self.meta['sequence_length']
        self.feature_cols = self.meta['feature_columns']
        self.buffers = {}
        
        # Charger TFLite
        tflite_file = self.meta.get('tflite_path')
        if tflite_file:
            tflite_path = self.model_dir / Path(tflite_file).name
        else:
            tflite_path = self.model_dir / f"{self.meta['model_name']}.tflite"
        
        self.interpreter = tf.lite.Interpreter(model_path=str(tflite_path))
        self.interpreter.allocate_tensors()
        self.input_details = self.interpreter.get_input_details()
        self.output_details = self.interpreter.get_output_details()
        
        print(f"✅ Détecteur TFLite chargé (seuil={self.threshold:.4f})")
    
    def preprocess(self, data):
        return [data.get(col, 0.0) for col in self.feature_cols]
    
    def add_to_buffer(self, device_id, features):
        if device_id not in self.buffers:
            self.buffers[device_id] = []
        self.buffers[device_id].append(features)
        if len(self.buffers[device_id]) > self.seq_len:
            self.buffers[device_id].pop(0)
        return len(self.buffers[device_id]) >= self.seq_len
    
    def predict(self, device_id):
        if device_id not in self.buffers or len(self.buffers[device_id]) < self.seq_len:
            return None
        
        seq = np.array(self.buffers[device_id])
        seq_scaled = self.scaler.transform(seq)
        seq_scaled = np.expand_dims(seq_scaled, axis=0).astype(np.float32)
        
        self.interpreter.set_tensor(self.input_details[0]['index'], seq_scaled)
        self.interpreter.invoke()
        reconstructed = self.interpreter.get_tensor(self.output_details[0]['index'])
        
        mse = np.mean(np.power(seq_scaled - reconstructed, 2))
        is_anomaly = mse > self.threshold
        
        return {
            'device_id': device_id,
            'mse': float(mse),
            'is_anomaly': bool(is_anomaly),
            'score': float(mse / self.threshold),
            'severity': 'HIGH' if mse / self.threshold > 3 else 'MEDIUM' if is_anomaly else 'NONE'
        }

if __name__ == '__main__':
    import argparse
    parser = argparse.ArgumentParser()
    parser.add_argument('--model-dir', default='./export_v3_tflite')
    parser.add_argument('--test', action='store_true')
    args = parser.parse_args()
    
    detector = IoTAnomalyDetectorTFLite(args.model_dir)
    
    if args.test:
        for i in range(50):
            msg = {'sensor_value': 22, 'packet_size': 64, 'inter_arrival_time': 5,
                   'topic_count': 2, 'publish_rate': 0.2, 'connection_duration': 300,
                   'bytes_sent': 200, 'bytes_received': 150, 'qos_level': 0,
                   'bytes_ratio': 1.3, 'traffic_intensity': 50, 'packet_efficiency': 3,
                   'value_change': 0.1, 'hour_sin': 0, 'hour_cos': 1, 'sampling_rate': 5,
                   'packet_size_ma': 64, 'publish_rate_std': 0.01}
            
            if i == 25:
                msg['packet_size'] = 5000
                msg['publish_rate'] = 10
            
            if detector.add_to_buffer(f'device_{i%4}', detector.preprocess(msg)):
                result = detector.predict(f'device_{i%4}')
                status = "🚨 ANOMALIE" if result['is_anomaly'] else "✓ Normal"
                print(f"t={i:2d}: {status} (score={result['score']:.2f})")
