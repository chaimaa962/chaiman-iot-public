#!/usr/bin/env python3
"""API Flask pour le service de détection d'anomalies"""

from flask import Flask, request, jsonify
from detector import IoTAnomalyDetector

app = Flask(__name__)
detector = IoTAnomalyDetector(model_dir='/app/models')

@app.route('/analyze', methods=['POST'])
def analyze():
    """Analyse un message IoT"""
    try:
        data = request.get_json()
        device_id = data.get('device_id')
        
        if not device_id:
            return jsonify({'error': 'device_id manquant'}), 400
        
        # Ajouter au buffer
        has_enough = detector.add_message(device_id, data)
        
        if not has_enough:
            return jsonify({
                'status': 'buffering',
                'device_id': device_id,
                'message': 'Accumulation des données...'
            })
        
        # Prédire
        result = detector.predict(device_id)
        
        if result and result['is_anomaly']:
            print(f"🚨 ANOMALIE détectée: {result}")
        
        return jsonify(result or {'status': 'no_prediction'})
        
    except Exception as e:
        return jsonify({'error': str(e)}), 500

@app.route('/health', methods=['GET'])
def health():
    return jsonify({'status': 'healthy', 'model': 'loaded'})

@app.route('/stats', methods=['GET'])
def stats():
    """Statistiques du détecteur"""
    return jsonify({
        'devices': list(detector.buffers.keys()),
        'buffer_sizes': {k: len(v) for k, v in detector.buffers.items()},
        'threshold': detector.threshold
    })

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
