// src/services/api.js
import axios from 'axios';

// Configuration de base - pointe vers ton backend Go
const API_BASE_URL = 'http://localhost:8080/api';

// Création d'une instance axios
const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000, // Augmenté pour les requêtes ZKP
  headers: {
    'Content-Type': 'application/json',
  },
});

// Intercepteur pour les requêtes
api.interceptors.request.use(
  (config) => {
    console.log(`📤 Requête: ${config.method?.toUpperCase()} ${config.url}`);
    return config;
  },
  (error) => {
    console.error('❌ Erreur requête:', error);
    return Promise.reject(error);
  }
);

// Intercepteur pour les réponses
api.interceptors.response.use(
  (response) => {
    console.log(`📥 Réponse: ${response.status}`, response.data);
    return response;
  },
  (error) => {
    if (error.response) {
      console.error('❌ Erreur serveur:', error.response.status, error.response.data);
    } else if (error.request) {
      console.error('❌ Pas de réponse du serveur');
    } else {
      console.error('❌ Erreur:', error.message);
    }
    return Promise.reject(error);
  }
);

// ============================================
// FONCTIONS API - CORRESPONDANT AU BACKEND
// ============================================

// 1. Santé du serveur (GET /api/health)
export const healthCheck = async () => {
  try {
    const response = await api.get('/health');
    return response.data;
  } catch (error) {
    console.error('Health check failed:', error);
    throw error;
  }
};

// 2. Récupérer tous les appareils (GET /api/devices)
export const getAllDevices = async () => {
  try {
    const response = await api.get('/devices');
    return response.data;
  } catch (error) {
    console.error('❌ Erreur récupération appareils:', error);
    return [];
  }
};

// 3. Récupérer les infos d'un appareil (GET /api/device/{address})
export const getDeviceInfo = async (address) => {
  try {
    const response = await api.get(`/device/${address}`);
    return response.data;
  } catch (error) {
    console.error('Erreur getDeviceInfo:', error);
    throw error;
  }
};

// 4. Enregistrer un nouvel appareil (POST /api/device/register)
export const registerDevice = async (deviceData) => {
  try {
    const response = await api.post('/device/register', deviceData);
    return response.data;
  } catch (error) {
    console.error('Erreur registerDevice:', error);
    throw error;
  }
};

// 5. Récupérer les nœuds Geth (GET /api/geth-nodes)
export const getGethNodes = async () => {
  try {
    const response = await api.get('/geth-nodes');
    return response.data;
  } catch (error) {
    console.error('Erreur récupération nœuds Geth:', error);
    return [];
  }
};

// ============================================
// FONCTIONS ZKP - POUR LE SIMULATEUR
// ============================================

// 6. Générer une preuve ZKP (POST /api/zkp/generate-secure)
export const generateZKPProof = async (proofData) => {
  try {
    const response = await api.post('/zkp/generate-secure', {
      signature: proofData.signature,
      message_hash: proofData.messageHash,
      address: proofData.address,
      sequence: proofData.sequence,
      timestamp: proofData.timestamp,
      nonce: proofData.nonce
    });
    return response.data;
  } catch (error) {
    console.error('❌ Erreur génération ZKP:', error);
    throw error;
  }
};

// 7. Envoyer un message sécurisé (POST /api/node/message-secure)
export const sendSecureMessage = async (messageData) => {
  try {
    const response = await api.post('/node/message-secure', {
      device_id: messageData.device_id,
      address: messageData.address,
      message: messageData.message,
      message_hash: messageData.message_hash,
      signature: messageData.signature,
      proof: messageData.proof,
      expected_hash: messageData.expected_hash,
      timestamp: messageData.timestamp,
      sequence: messageData.sequence,
      nonce: messageData.nonce
    });
    return response.data;
  } catch (error) {
    console.error('❌ Erreur envoi message sécurisé:', error);
    throw error;
  }
};

// ============================================
// FONCTIONS POUR L'HISTORIQUE (À IMPLÉMENTER DANS LE BACKEND)
// ============================================

// NOTE: Ces routes n'existent pas encore dans votre backend
// Vous devrez les ajouter si nécessaire

// 8. Récupérer l'historique des authentifications (À AJOUTER DANS LE BACKEND)
export const getDeviceAuthHistory = async (address) => {
  try {
    // ⚠️ Cette route n'existe pas encore dans votre backend
    // Vous devez l'ajouter dans handlers/device.go
    const response = await api.get(`/device/${address}/history`);
    return response.data;
  } catch (error) {
    console.error('Erreur historique:', error);
    // Retourne des données simulées en attendant
    return generateMockHistory(address);
  }
};

// Fonction pour générer un historique simulé
const generateMockHistory = (address) => {
  const history = [];
  const now = Date.now();
  
  for (let i = 0; i < 10; i++) {
    history.push({
      timestamp: Math.floor((now - i * 3600000) / 1000),
      success: Math.random() > 0.3,
      proofHash: `0x${Math.random().toString(16).substring(2, 66)}`,
      blockNumber: 15000000 + Math.floor(Math.random() * 1000)
    });
  }
  
  return history;
};

// 9. Récupérer les statistiques globales (À AJOUTER DANS LE BACKEND)
export const getGlobalStats = async () => {
  try {
    // ⚠️ Cette route n'existe pas encore
    // Vous pouvez utiliser /api/devices et calculer
    const devices = await getAllDevices();
    const gethNodes = await getGethNodes();
    
    const totalDevices = devices.length;
    const activeDevices = devices.filter(d => d.isActive).length;
    const totalAuths = devices.reduce((sum, d) => sum + (d.authCount || 0), 0);
    const zkpVerified = devices.filter(d => d.zkpValid).length;
    
    return {
      totalDevices,
      activeDevices,
      totalAuths,
      zkpVerified,
      successRate: totalDevices > 0 ? (zkpVerified / totalDevices * 100).toFixed(1) : 0,
      gethNodes: gethNodes.length,
      timestamp: Date.now()
    };
  } catch (error) {
    console.error('Erreur stats globales:', error);
    return {
      totalDevices: 0,
      activeDevices: 0,
      totalAuths: 0,
      zkpVerified: 0,
      successRate: 0,
      gethNodes: 4,
      timestamp: Date.now()
    };
  }
};

// 10. Récupérer les authentifications récentes (À AJOUTER DANS LE BACKEND)
export const getRecentAuths = async (limit = 20) => {
  try {
    // ⚠️ Cette route n'existe pas encore
    // Simulé pour l'instant
    const devices = await getAllDevices();
    const auths = [];
    
    for (const device of devices.slice(0, 5)) {
      for (let i = 0; i < 4; i++) {
        auths.push({
          id: `auth-${Date.now()}-${i}`,
          deviceAddress: device.address,
          deviceName: device.name || 'ESP32',
          success: Math.random() > 0.2,
          timestamp: Date.now() - i * 300000,
          latency: Math.floor(Math.random() * 200) + 50,
          blockNumber: 15000000 + Math.floor(Math.random() * 1000),
          proofHash: `0x${Math.random().toString(16).substring(2, 34)}`
        });
      }
    }
    
    return auths.sort((a, b) => b.timestamp - a.timestamp).slice(0, limit);
  } catch (error) {
    console.error('Erreur récupérations authentifications:', error);
    return [];
  }
};

// ============================================
// WEBSOCKET POUR TEMPS RÉEL
// ============================================

// 11. WebSocket pour les mises à jour en temps réel
export const subscribeToMessages = (callback) => {
  try {
    const ws = new WebSocket('ws://localhost:8080/ws');
    
    ws.onopen = () => {
      console.log('✅ WebSocket connecté');
    };
    
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        callback(data);
      } catch (e) {
        console.error('Erreur parsing WebSocket:', e);
      }
    };
    
    ws.onerror = (error) => {
      console.error('❌ WebSocket error:', error);
    };
    
    ws.onclose = () => {
      console.log('WebSocket déconnecté');
    };
    
    return ws;
  } catch (error) {
    console.error('❌ Erreur création WebSocket:', error);
    return null;
  }
};

// ============================================
// AUTHENTIFICATION ZKP (Route existante)
// ============================================

// 12. Authentifier un appareil avec ZKP
export const authenticateDevice = async (deviceAddress) => {
  try {
    console.log('🔐 Authentification ZKP pour:', deviceAddress);
    
    // Pour le simulateur, on utilise la route /api/node/message-secure
    const nonce = Math.random().toString(36).substring(2, 18);
    const timestamp = Math.floor(Date.now() / 1000);
    const sequence = 1;
    
    const payload = {
      device_id: `device-${deviceAddress.substring(0, 8)}`,
      address: deviceAddress,
      message: `Authentication request for ${deviceAddress.substring(0, 10)}`,
      message_hash: `0x${Math.random().toString(16).substring(2, 66)}`,
      signature: `0x${Math.random().toString(16).substring(2, 130)}`,
      proof: `0x${Math.random().toString(16).substring(2, 386)}`,
      expected_hash: `0x${Math.random().toString(16).substring(2, 66)}`,
      timestamp: timestamp,
      sequence: sequence,
      nonce: nonce
    };
    
    const response = await api.post('/node/message-secure', payload);
    
    return {
      success: response.status === 200,
      message: 'Authentification ZKP réussie',
      txHash: response.data.txHash,
      block: response.data.block
    };
  } catch (error) {
    console.error('❌ Erreur authentification ZKP:', error);
    return {
      success: false,
      message: error.response?.data?.message || 'Erreur authentification'
    };
  }
};

// ============================================
// VÉRIFICATION DE CONNEXION
// ============================================

// 13. Vérifier toutes les connexions
export const checkAllConnections = async () => {
  const results = {
    backend: false,
    blockchain: false,
    gethNodes: []
  };
  
  try {
    // Vérifier backend
    const health = await healthCheck();
    results.backend = health.status === 'OK';
    
    // Vérifier appareils
    const devices = await getAllDevices();
    results.totalDevices = devices.length;
    
    // Vérifier nœuds Geth
    const gethNodes = await getGethNodes();
    results.gethNodes = gethNodes;
    
  } catch (error) {
    console.error('❌ Erreur vérification connexions:', error);
  }
  
  return results;
};

// Export par défaut
export default api;
