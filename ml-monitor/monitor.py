#!/usr/bin/env python3
"""Moniteur ML pour nœuds Geth - Version simplifiée"""

import asyncio
import aiohttp
import os
import json
import time
from datetime import datetime
from web3 import Web3

ML_ANALYZER_URL = os.getenv('ML_ANALYZER_URL', 'http://ml-analyzer:5000')
GETH_NODES = os.getenv('GETH_NODES', 'validator1:8541,validator2:8542,validator3:8543,validator4:8544').split(',')
POLL_INTERVAL = int(os.getenv('POLL_INTERVAL', 10))

class GethNodeMonitor:
    def __init__(self, node_url: str):
        self.node_name = node_url.split(':')[0]
        self.w3 = Web3(Web3.HTTPProvider(f"http://{node_url}"))
        self.stats = {'analyzed': 0, 'anomalies': 0}
        print(f"  ✅ Moniteur initialisé pour {self.node_name}")
    
    async def collect_metrics(self) -> dict:
        metrics = {
            'device_id': f"geth_{self.node_name}",
            'timestamp': datetime.now().isoformat(),
            'packet_size': 512,
            'publish_rate': 1.0 / POLL_INTERVAL,
            'inter_arrival_time': POLL_INTERVAL,
            'qos_level': 2,
            'sampling_rate': 1.0 / POLL_INTERVAL,
        }
        
        try:
            if self.w3.is_connected():
                metrics['block_number'] = self.w3.eth.block_number
                metrics['peer_count'] = self.w3.net.peer_count
                metrics['gas_price'] = float(self.w3.eth.gas_price) / 1e9
                metrics['syncing'] = bool(self.w3.eth.syncing)
                
                # Métriques réseau estimées
                metrics['bytes_sent'] = 1024 * (metrics.get('peer_count', 0) + 1)
                metrics['bytes_received'] = 2048 * (metrics.get('peer_count', 0) + 1)
        except Exception as e:
            metrics['error'] = str(e)[:100]
        
        return metrics
    
    async def analyze_with_ml(self, session: aiohttp.ClientSession):
        metrics = await self.collect_metrics()
        self.stats['analyzed'] += 1
        
        try:
            async with session.post(
                f"{ML_ANALYZER_URL}/analyze",
                json=metrics,
                timeout=aiohttp.ClientTimeout(total=5)
            ) as resp:
                if resp.status == 200:
                    result = await resp.json()
                    if result.get('is_anomaly'):
                        self.stats['anomalies'] += 1
                        severity = result.get('severity', 'UNKNOWN')
                        score = result.get('anomaly_score', 0)
                        print(f"🚨 [{self.node_name}] ANOMALIE {severity} (score: {score:.2f})")
                        print(f"   Block: {metrics.get('block_number', 'N/A')}, Peers: {metrics.get('peer_count', 'N/A')}")
                    return result
        except Exception as e:
            if self.stats['analyzed'] % 10 == 0:
                print(f"⚠️  [{self.node_name}] ML error: {str(e)[:50]}")
        
        return None

async def monitor_loop():
    print("🔄 Démarrage du monitoring ML pour nœuds Geth...")
    print(f"   ML Analyzer: {ML_ANALYZER_URL}")
    print(f"   Intervalle: {POLL_INTERVAL}s")
    print()
    
    # Attendre que le service ML soit prêt
    await asyncio.sleep(5)
    
    monitors = []
    for node in GETH_NODES:
        if node.strip():
            try:
                monitors.append(GethNodeMonitor(node.strip()))
            except Exception as e:
                print(f"⚠️  Erreur initialisation {node}: {e}")
    
    print(f"\n📊 {len(monitors)} nœuds en surveillance\n")
    
    async with aiohttp.ClientSession() as session:
        cycle = 0
        while True:
            cycle += 1
            tasks = [m.analyze_with_ml(session) for m in monitors]
            await asyncio.gather(*tasks, return_exceptions=True)
            
            if cycle % 6 == 0:  # Toutes les minutes
                total_analyzed = sum(m.stats['analyzed'] for m in monitors)
                total_anomalies = sum(m.stats['anomalies'] for m in monitors)
                print(f"📊 [Cycle {cycle}] {total_analyzed} analyses, {total_anomalies} anomalies")
            
            await asyncio.sleep(POLL_INTERVAL)

if __name__ == "__main__":
    try:
        asyncio.run(monitor_loop())
    except KeyboardInterrupt:
        print("\n👋 Monitoring arrêté")
