// src/components/dashboard/BlockchainDashboard.js
// DASHBOARD CORRIGÉ - Données 100% réelles en temps réel

import React, { useState, useEffect, useCallback, useRef, useMemo } from 'react';
import {
  Box,
  Typography,
  Grid,
  Paper,
  Card,
  CardContent,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TablePagination,
  Chip,
  IconButton,
  Button,
  TextField,
  InputAdornment,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Avatar,
  Badge,
  Stack,
  Divider,
  LinearProgress,
  Alert,
  Tooltip,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Tabs,
  Tab,
  Skeleton,
  Zoom,
  useTheme,
  alpha,
  Drawer,
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  AppBar,
  Toolbar,
  CssBaseline,
  Container,
  Snackbar
} from '@mui/material';
import {
  Refresh as RefreshIcon,
  Search as SearchIcon,
  Download as DownloadIcon,
  Security as SecurityIcon,
  Verified as VerifiedIcon,
  Error as ErrorIcon,
  CheckCircle as CheckCircleIcon,
  Devices as DevicesIcon,
  Timeline as TimelineIcon,
  Speed as SpeedIcon,
  Storage as StorageIcon,
  Wifi as WifiIcon,
  WifiOff as WifiOffIcon,
  Router as RouterIcon,
  DeveloperBoard as DevBoardIcon,
  Block as BlockIcon,
  Pause as PauseIcon,
  PlayArrow as PlayArrowIcon,
  Menu as MenuIcon,
  Dashboard as DashboardIcon,
  Settings as SettingsIcon,
  AccountCircle as AccountCircleIcon,
  Logout as LogoutIcon,
  Analytics as AnalyticsIcon,
  SmartToy as AiIcon,
  PrecisionManufacturing as PrecisionIcon,
  Close as CloseIcon
} from '@mui/icons-material';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip as ChartTooltip,
  Legend,
  Filler,
  TimeScale
} from 'chart.js';
import { Line, Bar } from 'react-chartjs-2';
import { formatDistanceToNow, format, subMinutes } from 'date-fns';
import { fr } from 'date-fns/locale';

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  ChartTooltip,
  Legend,
  Filler,
  TimeScale
);

// ============================================
// CONFIGURATION API
// ============================================
const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';
const POLLING_INTERVAL = 3000;
const BLOCKCHAIN_POLLING_INTERVAL = 5000;

// ============================================
// SERVICE API - AMÉLIORÉ AVEC GESTION D'ERREURS
// ============================================
const apiService = {
  checkHealth: async () => {
    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 5000);
      
      const res = await fetch(`${API_BASE_URL}/api/health`, { 
        method: 'GET',
        headers: { 'Content-Type': 'application/json' },
        signal: controller.signal
      });
      
      clearTimeout(timeoutId);
      return { ok: res.ok, status: res.status };
    } catch (e) {
      return { ok: false, error: e.message };
    }
  },

  getAllDevices: async () => {
    try {
      const res = await fetch(`${API_BASE_URL}/api/devices`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return await res.json();
    } catch (e) {
      console.error('Error fetching devices:', e);
      return [];
    }
  },

  getGethNodes: async () => {
    try {
      const res = await fetch(`${API_BASE_URL}/api/geth-nodes`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return await res.json();
    } catch (e) {
      console.error('Error fetching Geth nodes:', e);
      return [];
    }
  },

  getGlobalStats: async () => {
    try {
      const res = await fetch(`${API_BASE_URL}/api/stats/global`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return await res.json();
    } catch (e) {
      console.error('Error fetching global stats:', e);
      return null;
    }
  },

  getBlockchainInfo: async () => {
    try {
      const res = await fetch(`${API_BASE_URL}/api/blockchain/info`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return await res.json();
    } catch (e) {
      console.error('Error fetching blockchain info:', e);
      return null;
    }
  },

  getRecentTransactions: async (limit = 100) => {
    try {
      const res = await fetch(`${API_BASE_URL}/api/transactions/recent?limit=${limit}`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return await res.json();
    } catch (e) {
      console.error('Error fetching transactions:', e);
      return [];
    }
  },

  getDeviceHistory: async (address, limit = 50) => {
    try {
      const res = await fetch(`${API_BASE_URL}/api/device/${address}/history?limit=${limit}`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return await res.json();
    } catch (e) {
      console.error(`Error fetching history for ${address}:`, e);
      return [];
    }
  },

  getTransactionByHash: async (hash) => {
    try {
      const res = await fetch(`${API_BASE_URL}/api/transaction/${hash}`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return await res.json();
    } catch (e) {
      console.error('Error fetching transaction:', e);
      return null;
    }
  },

  getDeviceInfo: async (address) => {
    try {
      const res = await fetch(`${API_BASE_URL}/api/device/${address}`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return await res.json();
    } catch (e) {
      console.error('Error fetching device info:', e);
      return null;
    }
  },

  getBlockMetrics: async () => {
    try {
      const res = await fetch(`${API_BASE_URL}/api/blockchain/metrics`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return await res.json();
    } catch (e) {
      console.error('Error fetching block metrics:', e);
      return null;
    }
  },

  getAIMetrics: async () => {
    try {
      const res = await fetch(`${API_BASE_URL}/api/ai/metrics`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return await res.json();
    } catch (e) {
      console.error('Error fetching AI metrics:', e);
      return null;
    }
  }
};

// ============================================
// COMPOSANT SNACKBAR D'ERREUR
// ============================================
const ErrorSnackbar = ({ open, message, onClose }) => (
  <Snackbar
    open={open}
    autoHideDuration={6000}
    onClose={onClose}
    anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
  >
    <Alert onClose={onClose} severity="error" variant="filled" sx={{ width: '100%' }}>
      {message}
    </Alert>
  </Snackbar>
);

// ============================================
// COMPOSANT MENU DE NAVIGATION
// ============================================
const NavigationMenu = ({ open, onClose, onNavigate, currentView }) => {
  const theme = useTheme();
  
  const menuItems = [
    { id: 'dashboard', label: 'Dashboard', icon: <DashboardIcon />, color: 'primary' },
    { id: 'devices', label: 'Appareils', icon: <DevicesIcon />, color: 'success' },
    { id: 'transactions', label: 'Transactions', icon: <TimelineIcon />, color: 'info' },
    { id: 'geth-nodes', label: 'Nœuds Geth', icon: <RouterIcon />, color: 'warning' },
    { id: 'analytics', label: 'Analytics', icon: <AnalyticsIcon />, color: 'secondary' },
    { id: 'ai-detection', label: 'Détection IA', icon: <AiIcon />, color: 'error' },
    { id: 'settings', label: 'Paramètres', icon: <SettingsIcon />, color: 'default' },
  ];

  const getColorValue = (color) => {
    const colorMap = {
      primary: theme.palette.primary.main,
      success: theme.palette.success.main,
      info: theme.palette.info.main,
      warning: theme.palette.warning.main,
      secondary: theme.palette.secondary.main,
      error: theme.palette.error.main,
      default: theme.palette.text.primary
    };
    return colorMap[color] || theme.palette.primary.main;
  };

  return (
    <Drawer
      anchor="left"
      open={open}
      onClose={onClose}
      sx={{
        '& .MuiDrawer-paper': {
          width: 280,
          boxSizing: 'border-box',
          background: (theme) => `linear-gradient(180deg, ${theme.palette.background.paper} 0%, ${alpha(theme.palette.primary.main, 0.05)} 100%)`,
        },
      }}
    >
      <Box sx={{ p: 3, display: 'flex', alignItems: 'center', gap: 2 }}>
        <Avatar sx={{ bgcolor: 'primary.main', width: 48, height: 48 }}>
          <SecurityIcon />
        </Avatar>
        <Box>
          <Typography variant="h6" fontWeight="bold">
            IoT ZKP
          </Typography>
          <Typography variant="caption" color="text.secondary">
            Blockchain Dashboard
          </Typography>
        </Box>
      </Box>
      
      <Divider sx={{ mb: 2 }} />
      
      <List>
        {menuItems.map((item) => (
          <ListItem key={item.id} disablePadding>
            <ListItemButton
              selected={currentView === item.id}
              onClick={() => {
                onNavigate(item.id);
                onClose();
              }}
              sx={{
                mx: 1,
                borderRadius: 2,
                mb: 0.5,
                '&.Mui-selected': {
                  bgcolor: alpha(getColorValue(item.color), 0.1),
                  '&:hover': {
                    bgcolor: alpha(getColorValue(item.color), 0.2),
                  },
                },
              }}
            >
              <ListItemIcon sx={{ color: getColorValue(item.color) }}>
                {item.icon}
              </ListItemIcon>
              <ListItemText 
                primary={item.label}
                primaryTypographyProps={{ fontWeight: currentView === item.id ? 'bold' : 'normal' }}
              />
            </ListItemButton>
          </ListItem>
        ))}
      </List>
      
      <Box sx={{ mt: 'auto', p: 2 }}>
        <Divider sx={{ mb: 2 }} />
        <ListItem disablePadding>
          <ListItemButton sx={{ mx: 1, borderRadius: 2 }}>
            <ListItemIcon>
              <AccountCircleIcon />
            </ListItemIcon>
            <ListItemText primary="Mon Compte" />
          </ListItemButton>
        </ListItem>
        <ListItem disablePadding>
          <ListItemButton sx={{ mx: 1, borderRadius: 2, color: 'error.main' }}>
            <ListItemIcon sx={{ color: 'error.main' }}>
              <LogoutIcon />
            </ListItemIcon>
            <ListItemText primary="Déconnexion" />
          </ListItemButton>
        </ListItem>
      </Box>
    </Drawer>
  );
};

// ============================================
// COMPOSANT PRINCIPAL
// ============================================
const BlockchainDashboard = () => {
  const theme = useTheme();
  
  // ============================================
  // ÉTATS
  // ============================================
  const [menuOpen, setMenuOpen] = useState(false);
  const [currentView, setCurrentView] = useState('dashboard');
  
  const [loading, setLoading] = useState(true);
  const [backendConnected, setBackendConnected] = useState(false);
  const [lastUpdate, setLastUpdate] = useState(Date.now());
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [refreshInterval] = useState(POLLING_INTERVAL);
  
  const [errorOpen, setErrorOpen] = useState(false);
  const [errorMessage, setErrorMessage] = useState('');
  
  // Données
  const [devices, setDevices] = useState([]);
  const [gethNodes, setGethNodes] = useState([]);
  const [transactions, setTransactions] = useState([]);
  const [blockchainInfo, setBlockchainInfo] = useState({
    latestBlock: 0,
    gasPrice: '0',
    networkId: '1234',
    peers: 4,
    syncing: false
  });
  
  // Statistiques globales
  const [globalStats, setGlobalStats] = useState({
    totalDevices: 0,
    activeDevices: 0,
    totalAuthentications: 0,
    totalTransactions: 0,
    successRate: 100,
    ecdsaSuccess: 0,
    zkpSuccess: 0,
    gethNodesCount: 4,
    gethNodesActive: 4
  });

  // Métriques
  const [blockMetrics, setBlockMetrics] = useState({
    blockTimes: [],
    validationTimes: [],
    lastBlockTime: 0,
    averageBlockTime: 5
  });

  const [aiMetrics, setAiMetrics] = useState({
    detections: [],
    accuracy: 0,
    totalDetections: 0,
    falsePositives: 0,
    analysisTime: 0
  });

  // UI
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterType, setFilterType] = useState('all');
  const [sortOrder] = useState('desc');
  const [activeTab, setActiveTab] = useState(0);

  const [selectedDevice, setSelectedDevice] = useState(null);
  const [deviceDialogOpen, setDeviceDialogOpen] = useState(false);
  const [selectedTransaction, setSelectedTransaction] = useState(null);
  const [transactionDialogOpen, setTransactionDialogOpen] = useState(false);

  // Statistiques en temps réel
  const [realtimeStats, setRealtimeStats] = useState({
    totalMessages: 0,
    messagesPerSecond: 0,
    lastMinuteCount: 0,
    ecdsaSuccess: 0,
    zkpSuccess: 0,
    failedAuths: 0,
    avgLatency: 0,
    peakLatency: 0,
    blockTime: 5,
    aiDetections: 0,
    aiAccuracy: 0,
    blockValidationTime: 2.5
  });

  // Historique pour les graphiques
  const [metricsHistory, setMetricsHistory] = useState({
    timestamps: [],
    transactionCounts: [],
    latencyData: [],
    successRates: [],
    blockTimes: [],
    validationTimes: [],
    aiDetections: [],
    aiAccuracy: []
  });

  // Détection des nouvelles transactions
  const [prevTxCount, setPrevTxCount] = useState(0);
  const [newTxIds, setNewTxIds] = useState([]);

  const pollingTimerRef = useRef(null);
  const blockchainTimerRef = useRef(null);
  const messageCountRef = useRef({ lastMinute: 0, total: 0, lastMinuteReset: Date.now() });

  // ============================================
  // FONCTIONS UTILITAIRES
  // ============================================
  const showError = (message) => {
    setErrorMessage(message);
    setErrorOpen(true);
  };

  // ============================================
  // CALCUL DES STATISTIQUES RÉELLES
  // ============================================
  const calculateRealtimeStats = useCallback((txs, devicesList, blockInfo, aiData) => {
    const ecdsaValid = txs.filter(t => t.ecdsaValid !== false).length;
    const zkpValid = txs.filter(t => t.zkpValid || t.success).length;
    const failed = txs.filter(t => t.status === 'failed' || t.success === false).length;
    
    const avgLat = txs.length > 0 
      ? txs.reduce((acc, t) => acc + (t.latency || 0), 0) / txs.length 
      : 0;
    
    const peakLat = txs.length > 0
      ? Math.max(...txs.map(t => t.latency || 0))
      : 0;
    
    const now = Date.now();
    const oneMinuteAgo = now - 60000;
    const recentTxs = txs.filter(t => t.timestamp > oneMinuteAgo);
    const messagesPerSecond = recentTxs.length / 60;
    
    messageCountRef.current.total = txs.length;
    
    const blockTime = blockMetrics.averageBlockTime || 5;
    const aiDetections = aiData?.totalDetections || 0;
    const aiAccuracy = aiData?.accuracy || 0;
    const blockValidationTime = blockMetrics.validationTimes.length > 0
      ? blockMetrics.validationTimes[blockMetrics.validationTimes.length - 1]
      : 2.5;

    return {
      totalMessages: txs.length,
      messagesPerSecond: parseFloat(messagesPerSecond.toFixed(2)),
      lastMinuteCount: recentTxs.length,
      ecdsaSuccess: ecdsaValid,
      zkpSuccess: zkpValid,
      failedAuths: failed,
      avgLatency: Math.round(avgLat),
      peakLatency: Math.round(peakLat),
      blockTime: blockTime,
      aiDetections: aiDetections,
      aiAccuracy: aiAccuracy,
      blockValidationTime: blockValidationTime
    };
  }, [blockMetrics]);

  // ============================================
  // MISE À JOUR DE L'HISTORIQUE
  // ============================================
  const updateMetricsHistory = useCallback((txs, stats, blockData, aiData) => {
    const now = new Date();
    
    setMetricsHistory(prev => {
      const maxHistory = 60;
      
      const newTimestamps = [...prev.timestamps, now].slice(-maxHistory);
      const newCounts = [...prev.transactionCounts, txs.length].slice(-maxHistory);
      
      const avgLat = txs.length > 0 
        ? txs.reduce((acc, t) => acc + (t.latency || 0), 0) / txs.length 
        : 0;
      const newLatency = [...prev.latencyData, avgLat].slice(-maxHistory);
      
      const successCount = txs.filter(t => t.success || t.status === 'success').length;
      const successRate = txs.length > 0 ? (successCount / txs.length) * 100 : 100;
      const newSuccess = [...prev.successRates, successRate].slice(-maxHistory);
      
      const newBlockTimes = blockData?.blockTimes 
        ? [...prev.blockTimes, ...blockData.blockTimes].slice(-maxHistory)
        : [...prev.blockTimes, prev.blockTimes[prev.blockTimes.length - 1] || 5].slice(-maxHistory);
      
      const newValidationTimes = blockData?.validationTimes
        ? [...prev.validationTimes, ...blockData.validationTimes].slice(-maxHistory)
        : [...prev.validationTimes, prev.validationTimes[prev.blockTimes.length - 1] || 2.5].slice(-maxHistory);
      
      const newAiDetections = aiData?.detections
        ? [...prev.aiDetections, ...aiData.detections].slice(-maxHistory)
        : prev.aiDetections;
      
      const newAiAccuracy = aiData?.accuracyHistory
        ? [...prev.aiAccuracy, ...aiData.accuracyHistory].slice(-maxHistory)
        : prev.aiAccuracy;
      
      return {
        timestamps: newTimestamps,
        transactionCounts: newCounts,
        latencyData: newLatency,
        successRates: newSuccess,
        blockTimes: newBlockTimes,
        validationTimes: newValidationTimes,
        aiDetections: newAiDetections,
        aiAccuracy: newAiAccuracy
      };
    });
  }, []);

  // ============================================
  // CHARGEMENT DES DONNÉES
  // ============================================
  const loadAllData = useCallback(async (showLoading = false) => {
    if (showLoading) setLoading(true);
    
    try {
      console.log('🔄 Chargement des données réelles...');
      
      const health = await apiService.checkHealth();
      setBackendConnected(health.ok);
      
      if (!health.ok) {
        if (showLoading) setLoading(false);
        showError('Backend déconnecté - Vérifiez que le serveur Go est démarré');
        return;
      }

      const results = await Promise.allSettled([
        apiService.getAllDevices(),
        apiService.getGethNodes(),
        apiService.getGlobalStats(),
        apiService.getBlockchainInfo(),
        apiService.getRecentTransactions(100),
        apiService.getBlockMetrics().catch(() => null),
        apiService.getAIMetrics().catch(() => null)
      ]);

      const [devicesResult, gethResult, statsResult, blockchainResult, txsResult, blockMetricsResult, aiMetricsResult] = results;

      if (devicesResult.status === 'fulfilled') {
        setDevices(devicesResult.value || []);
        console.log('📱 Appareils chargés:', devicesResult.value?.length || 0);
      }

      if (gethResult.status === 'fulfilled') {
        setGethNodes(gethResult.value || []);
        console.log('⛓️ Nœuds Geth chargés:', gethResult.value?.length || 0);
      }

      if (statsResult.status === 'fulfilled' && statsResult.value) {
        const statsData = statsResult.value;
        console.log('📊 Stats globales:', statsData);
        setGlobalStats(prev => ({
          ...prev,
          totalDevices: statsData.totalDevices || devicesResult.value?.length || 0,
          activeDevices: statsData.activeDevices || devicesResult.value?.filter(d => d.isActive).length || 0,
          totalAuthentications: statsData.totalAuthentications || 0,
          totalTransactions: statsData.totalAuthentications || 0,
          successRate: statsData.successRate || 100,
          gethNodesCount: statsData.gethNodesCount || 4,
          gethNodesActive: statsData.gethNodesActive || 4
        }));
      }

      if (blockchainResult.status === 'fulfilled' && blockchainResult.value) {
        const blockchainData = blockchainResult.value;
        console.log('🔗 Infos blockchain:', blockchainData);
        setBlockchainInfo(prev => ({
          ...prev,
          ...blockchainData,
          latestBlock: blockchainData.latestBlock || blockchainData.blockNumber || prev.latestBlock
        }));
      }

      if (blockMetricsResult?.status === 'fulfilled' && blockMetricsResult.value) {
        setBlockMetrics(blockMetricsResult.value);
        console.log('⏱️ Métriques de blocs:', blockMetricsResult.value);
      }

      if (aiMetricsResult?.status === 'fulfilled' && aiMetricsResult.value) {
        setAiMetrics(aiMetricsResult.value);
        console.log('🤖 Métriques IA:', aiMetricsResult.value);
      }

      if (txsResult.status === 'fulfilled' && txsResult.value) {
        const recentTxs = txsResult.value;
        console.log('📥 Transactions brutes du backend:', recentTxs);
        
        // FORMATAGE CORRIGÉ : timestamp déjà en millisecondes
        const processedTxs = recentTxs.map((tx, index) => ({
          id: tx.hash || tx.txHash || `tx-${tx.timestamp}-${index}`,
          hash: tx.hash || tx.txHash || 'N/A',
          deviceId: tx.deviceId || tx.device_id || 'Unknown',
          deviceAddress: tx.deviceAddress || tx.address || '0x0',
          type: tx.type || 'authentication',
          status: tx.success || tx.status === 'success' ? 'success' : 
                 tx.status === 'failed' ? 'failed' : 'pending',
          success: tx.success || tx.status === 'success',
          message: tx.message || 'Authentication',
          timestamp: tx.timestamp, // DÉJÀ EN MILLISECONDES
          blockNumber: tx.blockNumber || tx.block || 0,
          latency: tx.latency || 0,
          ecdsaValid: tx.ecdsaValid === true,
          zkpValid: tx.zkpValid === true || tx.success === true,
          authType: tx.authType || 'ECDSA+ZKP',
          sequence: tx.sequence || 0,
          proofHash: tx.proofHash || null
        }));

        console.log('📤 Transactions formatées:', processedTxs);
        
        // Détection des nouvelles transactions
        if (prevTxCount > 0 && processedTxs.length > prevTxCount) {
          const newIds = processedTxs.slice(0, processedTxs.length - prevTxCount).map(tx => tx.id);
          setNewTxIds(newIds);
          setTimeout(() => setNewTxIds([]), 3000);
        }
        setPrevTxCount(processedTxs.length);
        
        setTransactions(processedTxs);

        const newRealtimeStats = calculateRealtimeStats(
          processedTxs, 
          devicesResult.value || [],
          blockchainResult.value,
          aiMetricsResult?.value
        );
        setRealtimeStats(newRealtimeStats);

        updateMetricsHistory(
          processedTxs,
          statsResult.value,
          blockMetricsResult?.value,
          aiMetricsResult?.value
        );
      }

      setLastUpdate(Date.now());
      
    } catch (error) {
      console.error('❌ Erreur chargement dashboard:', error);
      showError(`Erreur de chargement: ${error.message}`);
    } finally {
      if (showLoading) setLoading(false);
    }
  }, [calculateRealtimeStats, updateMetricsHistory, prevTxCount]);

  // ============================================
  // EFFETS
  // ============================================
  useEffect(() => {
    loadAllData(true);
  }, [loadAllData]);

  useEffect(() => {
    if (!autoRefresh) return;

    pollingTimerRef.current = setInterval(() => {
      loadAllData(false);
    }, POLLING_INTERVAL);

    return () => {
      if (pollingTimerRef.current) {
        clearInterval(pollingTimerRef.current);
      }
    };
  }, [autoRefresh, loadAllData]);

  useEffect(() => {
    if (!autoRefresh) return;

    blockchainTimerRef.current = setInterval(async () => {
      const info = await apiService.getBlockchainInfo();
      if (info) {
        setBlockchainInfo(prev => ({
          ...prev,
          ...info,
          latestBlock: info.latestBlock || info.blockNumber || prev.latestBlock
        }));
      }
    }, BLOCKCHAIN_POLLING_INTERVAL);

    return () => {
      if (blockchainTimerRef.current) {
        clearInterval(blockchainTimerRef.current);
      }
    };
  }, [autoRefresh]);

  // ============================================
  // HANDLERS
  // ============================================
  const handleRefresh = () => {
    loadAllData(false);
  };

  const handleExportCSV = () => {
    try {
      const csvContent = [
        ['Timestamp', 'Transaction Hash', 'Device ID', 'Device Address', 'Type', 'Status', 'ECDSA', 'ZKP', 'Block', 'Latency (ms)', 'Message'],
        ...transactions.map(tx => [
          new Date(tx.timestamp).toISOString(),
          tx.hash,
          tx.deviceId,
          tx.deviceAddress,
          tx.authType,
          tx.status,
          tx.ecdsaValid ? 'Yes' : 'No',
          tx.zkpValid ? 'Yes' : 'No',
          tx.blockNumber,
          tx.latency,
          `"${(tx.message || '').replace(/"/g, '""')}"`
        ])
      ].map(row => row.join(',')).join('\n');

      const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
      const link = document.createElement('a');
      link.href = URL.createObjectURL(blob);
      link.download = `blockchain_transactions_${format(new Date(), 'yyyy-MM-dd_HH-mm')}.csv`;
      link.click();
    } catch (error) {
      showError(`Erreur export CSV: ${error.message}`);
    }
  };

  const handleViewDevice = async (device) => {
    try {
      const details = await apiService.getDeviceInfo(device.address);
      setSelectedDevice({ ...device, ...details });
      setDeviceDialogOpen(true);
    } catch (error) {
      showError(`Erreur chargement détails appareil: ${error.message}`);
    }
  };

  const handleViewTransaction = (tx) => {
    setSelectedTransaction(tx);
    setTransactionDialogOpen(true);
  };

  const handleNavigate = (viewId) => {
    setCurrentView(viewId);
    const viewToTab = {
      'dashboard': 0,
      'transactions': 0,
      'devices': 1,
      'geth-nodes': 2,
      'analytics': 0,
      'ai-detection': 3,
      'settings': 0
    };
    setActiveTab(viewToTab[viewId] || 0);
  };

  const handleCloseError = () => {
    setErrorOpen(false);
  };

  // ============================================
  // FILTRAGE DES DONNÉES
  // ============================================
  const filteredDevices = useMemo(() => {
    let result = devices.filter(device => {
      const searchLower = searchTerm.toLowerCase();
      const matchesSearch = 
        (device.device_id || '').toLowerCase().includes(searchLower) ||
        (device.address || '').toLowerCase().includes(searchLower) ||
        (device.metadata || '').toLowerCase().includes(searchLower);
      
      if (filterType === 'all') return matchesSearch;
      if (filterType === 'active') return matchesSearch && device.isActive;
      if (filterType === 'inactive') return matchesSearch && !device.isActive;
      if (filterType === 'geth') return matchesSearch && (device.isGethNode || (device.device_id || '').includes('GETH'));
      if (filterType === 'verified') return matchesSearch && (device.authCount > 0);
      
      return matchesSearch;
    });

    result.sort((a, b) => {
      const aVal = sortOrder === 'desc' ? b.lastSeen || 0 : a.lastSeen || 0;
      const bVal = sortOrder === 'desc' ? a.lastSeen || 0 : b.lastSeen || 0;
      return aVal - bVal;
    });

    return result;
  }, [devices, searchTerm, filterType, sortOrder]);

  const paginatedDevices = filteredDevices.slice(
    page * rowsPerPage,
    page * rowsPerPage + rowsPerPage
  );

  // ============================================
  // CONFIGURATION DES GRAPHIQUES
  // ============================================
  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: 'top',
        labels: { color: theme.palette.text.primary, font: { size: 11 } }
      },
      tooltip: {
        mode: 'index',
        intersect: false,
        backgroundColor: alpha(theme.palette.background.paper, 0.9),
        titleColor: theme.palette.text.primary,
        bodyColor: theme.palette.text.secondary,
        borderColor: theme.palette.divider,
        borderWidth: 1
      }
    },
    scales: {
      y: {
        beginAtZero: true,
        grid: { color: alpha(theme.palette.divider, 0.1) },
        ticks: { color: theme.palette.text.secondary, font: { size: 10 } }
      },
      x: {
        grid: { display: false },
        ticks: { 
          color: theme.palette.text.secondary, 
          font: { size: 10 },
          maxRotation: 0,
          autoSkip: true,
          maxTicksLimit: 8
        }
      }
    },
    animation: { duration: 750, easing: 'easeOutQuart' }
  };

  const latencyChartData = {
    labels: metricsHistory.timestamps.map(t => format(t, 'HH:mm:ss')),
    datasets: [{
      label: 'Latence moyenne (ms)',
      data: metricsHistory.latencyData.map(l => Math.round(l)),
      borderColor: theme.palette.warning.main,
      backgroundColor: alpha(theme.palette.warning.main, 0.1),
      fill: true,
      tension: 0.4,
      pointRadius: 3,
      pointHoverRadius: 6
    }]
  };

  const blockValidationChartData = {
    labels: metricsHistory.timestamps.map(t => format(t, 'HH:mm:ss')),
    datasets: [{
      label: 'Temps validation block (s)',
      data: metricsHistory.validationTimes.map(t => t || 0),
      borderColor: theme.palette.info.main,
      backgroundColor: alpha(theme.palette.info.main, 0.1),
      fill: true,
      tension: 0.4,
      borderDash: [5, 5]
    }]
  };

  const aiDetectionsChartData = {
    labels: metricsHistory.timestamps.map(t => format(t, 'HH:mm:ss')),
    datasets: [{
      label: 'Détections IA',
      data: metricsHistory.aiDetections.length > 0 
        ? metricsHistory.aiDetections 
        : Array(metricsHistory.timestamps.length).fill(null),
      borderColor: theme.palette.error.main,
      backgroundColor: alpha(theme.palette.error.main, 0.1),
      fill: true,
      tension: 0.4,
      pointRadius: 4,
      spanGaps: true
    }]
  };

  const aiAccuracyChartData = {
    labels: metricsHistory.timestamps.map(t => format(t, 'HH:mm:ss')),
    datasets: [{
      label: 'Précision IA (%)',
      data: metricsHistory.aiAccuracy.length > 0
        ? metricsHistory.aiAccuracy
        : Array(metricsHistory.timestamps.length).fill(null),
      borderColor: theme.palette.success.main,
      backgroundColor: alpha(theme.palette.success.main, 0.1),
      fill: true,
      tension: 0.4,
      pointRadius: 4,
      spanGaps: true,
      yAxisID: 'y',
    }]
  };

  // ============================================
  // RENDU
  // ============================================
  if (loading) {
    return (
      <Box sx={{ p: 3, bgcolor: theme.palette.background.default, minHeight: '100vh' }}>
        <Skeleton variant="rectangular" height={80} sx={{ mb: 3, borderRadius: 2 }} />
        <Grid container spacing={3} sx={{ mb: 3 }}>
          {[1, 2, 3, 4].map(i => (
            <Grid item xs={12} sm={6} md={3} key={i}>
              <Skeleton variant="rectangular" height={120} sx={{ borderRadius: 2 }} />
            </Grid>
          ))}
        </Grid>
        <Skeleton variant="rectangular" height={400} sx={{ borderRadius: 2 }} />
      </Box>
    );
  }

  return (
    <Box sx={{ display: 'flex', bgcolor: theme.palette.background.default, minHeight: '100vh' }}>
      <CssBaseline />
      
      <NavigationMenu 
        open={menuOpen} 
        onClose={() => setMenuOpen(false)}
        onNavigate={handleNavigate}
        currentView={currentView}
      />

      <AppBar 
        position="fixed" 
        sx={{ 
          width: { sm: `calc(100%)` },
          ml: { sm: 0 },
          bgcolor: 'background.paper',
          color: 'text.primary',
          boxShadow: theme.shadows[2]
        }}
      >
        <Toolbar>
          <IconButton
            color="inherit"
            edge="start"
            onClick={() => setMenuOpen(true)}
            sx={{ mr: 2 }}
          >
            <MenuIcon />
          </IconButton>
          
          <Typography variant="h6" noWrap component="div" sx={{ flexGrow: 1, fontWeight: 'bold' }}>
            🔐 IoT Blockchain Dashboard
          </Typography>

          <Stack direction="row" spacing={1} alignItems="center">
            <Chip
              icon={backendConnected ? <WifiIcon /> : <WifiOffIcon />}
              label={backendConnected ? 'Connecté' : 'Déconnecté'}
              color={backendConnected ? 'success' : 'error'}
              size="small"
            />
            <Chip
              icon={<BlockIcon />}
              label={`Bloc #${blockchainInfo.latestBlock.toLocaleString()}`}
              color="primary"
              variant="outlined"
              size="small"
            />
            <Tooltip title={autoRefresh ? "Pause" : "Reprendre"}>
              <IconButton 
                onClick={() => setAutoRefresh(!autoRefresh)} 
                color={autoRefresh ? 'primary' : 'default'}
                size="small"
              >
                {autoRefresh ? <PauseIcon /> : <PlayArrowIcon />}
              </IconButton>
            </Tooltip>
            <Tooltip title="Rafraîchir">
              <IconButton onClick={handleRefresh} size="small">
                <RefreshIcon />
              </IconButton>
            </Tooltip>
          </Stack>
        </Toolbar>
        
        {!backendConnected && (
          <Alert severity="error" sx={{ mx: 2, mb: 1 }}>
            Backend déconnecté - Vérifiez que le serveur Go est démarré sur le port 8080
          </Alert>
        )}
        
        {autoRefresh && (
          <LinearProgress sx={{ height: 2 }} color="primary" />
        )}
      </AppBar>

      <Box
        component="main"
        sx={{
          flexGrow: 1,
          bgcolor: theme.palette.background.default,
          p: 3,
          width: '100%',
          mt: 8
        }}
      >
        <Container maxWidth="xl">
          {/* HEADER */}
          <Paper sx={{ 
            p: 3, 
            mb: 3, 
            borderRadius: 3,
            boxShadow: theme.shadows[4],
            background: `linear-gradient(135deg, ${theme.palette.background.paper} 0%, ${alpha(theme.palette.primary.main, 0.05)} 100%)`
          }}>
            <Box sx={{ 
              display: 'flex', 
              justifyContent: 'space-between', 
              alignItems: 'center', 
              flexWrap: 'wrap', 
              gap: 2 
            }}>
              <Box>
                <Typography variant="h4" fontWeight="bold" sx={{ mb: 0.5 }}>
                  Tableau de Bord Blockchain
                </Typography>
                <Stack direction="row" spacing={2} alignItems="center" flexWrap="wrap">
                  <Typography variant="caption" color="text.secondary">
                    Dernière mise à jour: {formatDistanceToNow(lastUpdate, { addSuffix: true, locale: fr })}
                  </Typography>
                  <Badge 
                    badgeContent={realtimeStats.messagesPerSecond} 
                    color="secondary"
                    max={999}
                  >
                    <SpeedIcon color="action" fontSize="small" />
                  </Badge>
                  <Typography variant="caption" color="text.secondary">
                    {realtimeStats.messagesPerSecond} tx/s
                  </Typography>
                </Stack>
              </Box>
              
              <Stack direction="row" spacing={1}>
                <Tooltip title="Exporter CSV">
                  <Button 
                    variant="outlined" 
                    startIcon={<DownloadIcon />}
                    onClick={handleExportCSV}
                    size="small"
                  >
                    Export
                  </Button>
                </Tooltip>
              </Stack>
            </Box>
          </Paper>

          {/* STATS CARDS */}
          <Grid container spacing={3} sx={{ mb: 4 }}>
            <Grid item xs={12} sm={6} lg={3}>
              <Zoom in={true} style={{ transitionDelay: '0ms' }}>
                <Card sx={{ 
                  borderRadius: 3, 
                  boxShadow: theme.shadows[4],
                  background: `linear-gradient(135deg, ${theme.palette.primary.main} 0%, ${alpha(theme.palette.primary.dark, 0.8)} 100%)`,
                  color: 'white',
                  height: '100%'
                }}>
                  <CardContent>
                    <Box display="flex" alignItems="center" justifyContent="space-between">
                      <Box>
                        <Typography variant="body2" sx={{ opacity: 0.9, mb: 0.5 }}>
                          Total Appareils
                        </Typography>
                        <Typography variant="h3" fontWeight="bold">
                          {globalStats.totalDevices}
                        </Typography>
                        <Typography variant="caption" sx={{ opacity: 0.8 }}>
                          {globalStats.activeDevices} actifs
                        </Typography>
                      </Box>
                      <Avatar sx={{ 
                        bgcolor: 'rgba(255,255,255,0.2)', 
                        width: 64, 
                        height: 64 
                      }}>
                        <DevicesIcon sx={{ fontSize: 32 }} />
                      </Avatar>
                    </Box>
                  </CardContent>
                </Card>
              </Zoom>
            </Grid>

            <Grid item xs={12} sm={6} lg={3}>
              <Zoom in={true} style={{ transitionDelay: '100ms' }}>
                <Card sx={{ 
                  borderRadius: 3, 
                  boxShadow: theme.shadows[4],
                  background: `linear-gradient(135deg, ${theme.palette.info.main} 0%, ${alpha(theme.palette.info.dark, 0.8)} 100%)`,
                  color: 'white',
                  height: '100%'
                }}>
                  <CardContent>
                    <Box display="flex" alignItems="center" justifyContent="space-between">
                      <Box>
                        <Typography variant="body2" sx={{ opacity: 0.9, mb: 0.5 }}>
                          ECDSA Réussis
                        </Typography>
                        <Typography variant="h3" fontWeight="bold">
                          {realtimeStats.ecdsaSuccess}
                        </Typography>
                        <Typography variant="caption" sx={{ opacity: 0.8 }}>
                          {globalStats.totalAuthentications > 0 
                            ? Math.round((realtimeStats.ecdsaSuccess / globalStats.totalAuthentications) * 100) 
                            : 0}% taux succès
                        </Typography>
                      </Box>
                      <Avatar sx={{ 
                        bgcolor: 'rgba(255,255,255,0.2)', 
                        width: 64, 
                        height: 64 
                      }}>
                        <SecurityIcon sx={{ fontSize: 32 }} />
                      </Avatar>
                    </Box>
                  </CardContent>
                </Card>
              </Zoom>
            </Grid>

            <Grid item xs={12} sm={6} lg={3}>
              <Zoom in={true} style={{ transitionDelay: '200ms' }}>
                <Card sx={{ 
                  borderRadius: 3, 
                  boxShadow: theme.shadows[4],
                  background: `linear-gradient(135deg, ${theme.palette.success.main} 0%, ${alpha(theme.palette.success.dark, 0.8)} 100%)`,
                  color: 'white',
                  height: '100%'
                }}>
                  <CardContent>
                    <Box display="flex" alignItems="center" justifyContent="space-between">
                      <Box>
                        <Typography variant="body2" sx={{ opacity: 0.9, mb: 0.5 }}>
                          ZKP Réussis
                        </Typography>
                        <Typography variant="h3" fontWeight="bold">
                          {realtimeStats.zkpSuccess}
                        </Typography>
                        <Typography variant="caption" sx={{ opacity: 0.8 }}>
                          {globalStats.totalAuthentications > 0 
                            ? Math.round((realtimeStats.zkpSuccess / globalStats.totalAuthentications) * 100) 
                            : 0}% taux succès
                        </Typography>
                      </Box>
                      <Avatar sx={{ 
                        bgcolor: 'rgba(255,255,255,0.2)', 
                        width: 64, 
                        height: 64 
                      }}>
                        <VerifiedIcon sx={{ fontSize: 32 }} />
                      </Avatar>
                    </Box>
                  </CardContent>
                </Card>
              </Zoom>
            </Grid>

            <Grid item xs={12} sm={6} lg={3}>
              <Zoom in={true} style={{ transitionDelay: '300ms' }}>
                <Card sx={{ 
                  borderRadius: 3, 
                  boxShadow: theme.shadows[4],
                  background: `linear-gradient(135deg, ${theme.palette.warning.main} 0%, ${alpha(theme.palette.warning.dark, 0.8)} 100%)`,
                  color: 'white',
                  height: '100%'
                }}>
                  <CardContent>
                    <Box display="flex" alignItems="center" justifyContent="space-between">
                      <Box>
                        <Typography variant="body2" sx={{ opacity: 0.9, mb: 0.5 }}>
                          Total Transactions
                        </Typography>
                        <Typography variant="h3" fontWeight="bold">
                          {globalStats.totalTransactions || transactions.length}
                        </Typography>
                        <Typography variant="caption" sx={{ opacity: 0.8 }}>
                          {realtimeStats.failedAuths} échecs
                        </Typography>
                      </Box>
                      <Avatar sx={{ 
                        bgcolor: 'rgba(255,255,255,0.2)', 
                        width: 64, 
                        height: 64 
                      }}>
                        <TimelineIcon sx={{ fontSize: 32 }} />
                      </Avatar>
                    </Box>
                  </CardContent>
                </Card>
              </Zoom>
            </Grid>
          </Grid>

          {/* 4 DIAGRAMMES */}
          <Grid container spacing={3} sx={{ mb: 4 }}>
            <Grid item xs={12} md={6} lg={3}>
              <Paper sx={{ 
                p: 2, 
                borderRadius: 3, 
                height: 280,
                boxShadow: theme.shadows[2]
              }}>
                <Typography variant="subtitle2" fontWeight="bold" gutterBottom>
                  ⏱️ Latence Authentification
                </Typography>
                <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 1 }}>
                  Moyenne: {realtimeStats.avgLatency}ms | Pic: {realtimeStats.peakLatency}ms
                </Typography>
                <Box sx={{ height: 200 }}>
                  <Line data={latencyChartData} options={{
                    ...chartOptions,
                    maintainAspectRatio: false,
                    plugins: { legend: { display: false } }
                  }} />
                </Box>
              </Paper>
            </Grid>

            <Grid item xs={12} md={6} lg={3}>
              <Paper sx={{ 
                p: 2, 
                borderRadius: 3, 
                height: 280,
                boxShadow: theme.shadows[2]
              }}>
                <Typography variant="subtitle2" fontWeight="bold" gutterBottom>
                  ⛓️ Temps Validation Blocks
                </Typography>
                {metricsHistory.validationTimes.some(t => t > 0) ? (
                  <>
                    <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 1 }}>
                      Moyen: {realtimeStats.blockValidationTime.toFixed(1)}s | Dernier: {realtimeStats.blockTime}s
                    </Typography>
                    <Box sx={{ height: 200 }}>
                      <Line data={blockValidationChartData} options={{
                        ...chartOptions,
                        maintainAspectRatio: false,
                        plugins: { legend: { display: false } }
                      }} />
                    </Box>
                  </>
                ) : (
                  <Box sx={{ 
                    height: 200, 
                    display: 'flex', 
                    alignItems: 'center', 
                    justifyContent: 'center',
                    flexDirection: 'column',
                    gap: 1
                  }}>
                    <StorageIcon sx={{ fontSize: 40, color: 'text.disabled', opacity: 0.3 }} />
                    <Typography variant="caption" color="text.disabled" align="center">
                      Données de validation<br />non disponibles
                    </Typography>
                  </Box>
                )}
              </Paper>
            </Grid>

            <Grid item xs={12} md={6} lg={3}>
              <Paper sx={{ 
                p: 2, 
                borderRadius: 3, 
                height: 280,
                boxShadow: theme.shadows[2]
              }}>
                <Typography variant="subtitle2" fontWeight="bold" gutterBottom>
                  🤖 IA - Détections Temps Réel
                </Typography>
                {metricsHistory.aiDetections.length > 0 ? (
                  <>
                    <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 1 }}>
                      Total: {realtimeStats.aiDetections} anomalies | Actuel: {metricsHistory.aiDetections[metricsHistory.aiDetections.length-1] || 0}
                    </Typography>
                    <Box sx={{ height: 200 }}>
                      <Bar data={aiDetectionsChartData} options={{
                        ...chartOptions,
                        maintainAspectRatio: false,
                        plugins: { legend: { display: false } }
                      }} />
                    </Box>
                  </>
                ) : (
                  <Box sx={{ 
                    height: 200, 
                    display: 'flex', 
                    alignItems: 'center', 
                    justifyContent: 'center',
                    flexDirection: 'column',
                    gap: 1
                  }}>
                    <AiIcon sx={{ fontSize: 40, color: 'text.disabled', opacity: 0.3 }} />
                    <Typography variant="caption" color="text.disabled" align="center">
                      Module IA non<br />configuré
                    </Typography>
                  </Box>
                )}
              </Paper>
            </Grid>

            <Grid item xs={12} md={6} lg={3}>
              <Paper sx={{ 
                p: 2, 
                borderRadius: 3, 
                height: 280,
                boxShadow: theme.shadows[2]
              }}>
                <Typography variant="subtitle2" fontWeight="bold" gutterBottom>
                  🎯 IA - Précision
                </Typography>
                {metricsHistory.aiAccuracy.length > 0 ? (
                  <>
                    <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 1 }}>
                      Actuelle: {realtimeStats.aiAccuracy.toFixed(1)}% | Cible: 99.9%
                    </Typography>
                    <Box sx={{ height: 200 }}>
                      <Line data={aiAccuracyChartData} options={{
                        ...chartOptions,
                        maintainAspectRatio: false,
                        plugins: { legend: { display: false } },
                        scales: {
                          y: { min: 98, max: 100 }
                        }
                      }} />
                    </Box>
                  </>
                ) : (
                  <Box sx={{ 
                    height: 200, 
                    display: 'flex', 
                    alignItems: 'center', 
                    justifyContent: 'center',
                    flexDirection: 'column',
                    gap: 1
                  }}>
                    <PrecisionIcon sx={{ fontSize: 40, color: 'text.disabled', opacity: 0.3 }} />
                    <Typography variant="caption" color="text.disabled" align="center">
                      Données de précision<br />non disponibles
                    </Typography>
                  </Box>
                )}
              </Paper>
            </Grid>
          </Grid>

          {/* TABS */}
          <Paper sx={{ borderRadius: 3, boxShadow: theme.shadows[2], mb: 3 }}>
            <Tabs 
              value={activeTab} 
              onChange={(e, v) => setActiveTab(v)}
              sx={{ borderBottom: 1, borderColor: 'divider' }}
            >
              <Tab 
                icon={<TimelineIcon />} 
                label={`Transactions (${transactions.length})`} 
                iconPosition="start"
              />
              <Tab 
                icon={<DevicesIcon />} 
                label={`Appareils (${filteredDevices.length})`} 
                iconPosition="start"
              />
              <Tab 
                icon={<RouterIcon />} 
                label={`Nœuds Geth (${gethNodes.length})`} 
                iconPosition="start"
              />
              <Tab 
                icon={<AiIcon />} 
                label={`IA Détection`} 
                iconPosition="start"
              />
            </Tabs>

            {/* TAB 1: TRANSACTIONS */}
            {activeTab === 0 && (
              <Box sx={{ p: 3 }}>
                <Box sx={{ 
                  display: 'flex', 
                  justifyContent: 'space-between', 
                  alignItems: 'center', 
                  mb: 2,
                  flexWrap: 'wrap',
                  gap: 2
                }}>
                  <Typography variant="h6">
                    🔄 Transactions Blockchain en Temps Réel
                  </Typography>
                  <Stack direction="row" spacing={1}>
                    <TextField
                      size="small"
                      placeholder="Rechercher..."
                      value={searchTerm}
                      onChange={(e) => setSearchTerm(e.target.value)}
                      InputProps={{
                        startAdornment: (
                          <InputAdornment position="start">
                            <SearchIcon fontSize="small" />
                          </InputAdornment>
                        ),
                      }}
                      sx={{ width: 200 }}
                    />
                    <Button
                      variant="outlined"
                      size="small"
                      startIcon={<DownloadIcon />}
                      onClick={handleExportCSV}
                    >
                      Export CSV
                    </Button>
                  </Stack>
                </Box>

                <TableContainer sx={{ maxHeight: 500, borderRadius: 2 }}>
                  <Table stickyHeader size="small">
                    <TableHead>
                      <TableRow>
                        <TableCell>Appareil</TableCell>
                        <TableCell>Hash Transaction</TableCell>
                        <TableCell>Type</TableCell>
                        <TableCell>Statut</TableCell>
                        <TableCell>ECDSA</TableCell>
                        <TableCell>ZKP</TableCell>
                        <TableCell>Bloc</TableCell>
                        <TableCell>Latence</TableCell>
                        <TableCell>Message</TableCell>
                        <TableCell>Timestamp</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {transactions.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage).map((tx, index) => (
                        <TableRow 
                          key={tx.id || index}
                          hover
                          onClick={() => handleViewTransaction(tx)}
                          sx={{ 
                            cursor: 'pointer',
                            backgroundColor: newTxIds.includes(tx.id) ? alpha(theme.palette.success.main, 0.1) : 'inherit',
                            transition: 'background-color 0.5s ease'
                          }}
                        >
                          <TableCell>
                            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                              <Avatar sx={{ 
                                width: 24, 
                                height: 24, 
                                bgcolor: tx.deviceId?.includes('GETH') 
                                  ? theme.palette.warning.main 
                                  : theme.palette.success.main,
                                fontSize: 10
                              }}>
                                {tx.deviceId?.includes('GETH') ? 'G' : 'D'}
                              </Avatar>
                              <Typography variant="body2" fontWeight="medium">
                                {tx.deviceId}
                              </Typography>
                            </Box>
                          </TableCell>
                          <TableCell>
                            <code style={{ fontSize: 11 }}>
                              {tx.hash?.substring(0, 16)}...
                            </code>
                          </TableCell>
                          <TableCell>
                            <Chip 
                              size="small" 
                              label={tx.authType} 
                              variant="outlined"
                              sx={{ height: 20, fontSize: 10 }}
                            />
                          </TableCell>
                          <TableCell>
                            <Chip 
                              size="small" 
                              label={tx.status}
                              color={tx.status === 'success' ? 'success' : 
                                     tx.status === 'failed' ? 'error' : 'default'}
                              icon={tx.status === 'success' ? <CheckCircleIcon /> : 
                                    tx.status === 'failed' ? <ErrorIcon /> : null}
                              sx={{ height: 24 }}
                            />
                          </TableCell>
                          <TableCell>
                            {tx.ecdsaValid ? (
                              <CheckCircleIcon color="success" fontSize="small" />
                            ) : (
                              <ErrorIcon color="error" fontSize="small" />
                            )}
                          </TableCell>
                          <TableCell>
                            {tx.zkpValid ? (
                              <VerifiedIcon color="success" fontSize="small" />
                            ) : (
                              <ErrorIcon color="error" fontSize="small" />
                            )}
                          </TableCell>
                          <TableCell>
                            <Typography variant="body2" color="primary" fontWeight="medium">
                              #{tx.blockNumber || '?'}
                            </Typography>
                          </TableCell>
                          <TableCell>
                            <Typography variant="body2">
                              {tx.latency}ms
                            </Typography>
                          </TableCell>
                          <TableCell>
                            <Typography variant="body2" noWrap sx={{ maxWidth: 150 }}>
                              {tx.message}
                            </Typography>
                          </TableCell>
                          <TableCell>
                            <Typography variant="caption" color="text.secondary">
                              {tx.timestamp ? formatDistanceToNow(tx.timestamp, { addSuffix: true, locale: fr }) : 'Inconnu'}
                            </Typography>
                          </TableCell>
                        </TableRow>
                      ))}
                      {transactions.length === 0 && (
                        <TableRow>
                          <TableCell colSpan={10} align="center" sx={{ py: 4 }}>
                            <Typography color="text.secondary">
                              Aucune transaction trouvée
                            </Typography>
                          </TableCell>
                        </TableRow>
                      )}
                    </TableBody>
                  </Table>
                </TableContainer>

                <TablePagination
                  rowsPerPageOptions={[10, 25, 50, 100]}
                  component="div"
                  count={transactions.length}
                  rowsPerPage={rowsPerPage}
                  page={page}
                  onPageChange={(e, newPage) => setPage(newPage)}
                  onRowsPerPageChange={(e) => {
                    setRowsPerPage(parseInt(e.target.value, 10));
                    setPage(0);
                  }}
                  labelRowsPerPage="Lignes par page"
                  labelDisplayedRows={({ from, to, count }) => `${from}-${to} sur ${count}`}
                />
              </Box>
            )}

            {/* TAB 2: APPAREILS */}
            {activeTab === 1 && (
              <Box sx={{ p: 3 }}>
                <Box sx={{ 
                  display: 'flex', 
                  justifyContent: 'space-between', 
                  alignItems: 'center', 
                  mb: 2,
                  flexWrap: 'wrap',
                  gap: 2
                }}>
                  <Typography variant="h6">
                    📱 Appareils Connectés - Infos Détaillées
                  </Typography>
                  <Stack direction="row" spacing={1}>
                    <TextField
                      size="small"
                      placeholder="Rechercher..."
                      value={searchTerm}
                      onChange={(e) => setSearchTerm(e.target.value)}
                      InputProps={{
                        startAdornment: (
                          <InputAdornment position="start">
                            <SearchIcon fontSize="small" />
                          </InputAdornment>
                        ),
                      }}
                      sx={{ width: 200 }}
                    />
                    <FormControl size="small" sx={{ width: 150 }}>
                      <InputLabel>Filtrer</InputLabel>
                      <Select 
                        value={filterType} 
                        label="Filtrer" 
                        onChange={(e) => setFilterType(e.target.value)}
                      >
                        <MenuItem value="all">Tous</MenuItem>
                        <MenuItem value="active">Actifs</MenuItem>
                        <MenuItem value="inactive">Inactifs</MenuItem>
                        <MenuItem value="geth">Nœuds Geth</MenuItem>
                        <MenuItem value="verified">Vérifiés</MenuItem>
                      </Select>
                    </FormControl>
                  </Stack>
                </Box>

                <Grid container spacing={3}>
                  {paginatedDevices.map((device) => (
                    <Grid item xs={12} key={device.address}>
                      <Card 
                        sx={{ 
                          borderRadius: 2,
                          boxShadow: theme.shadows[1],
                          '&:hover': { boxShadow: theme.shadows[4] }
                        }}
                      >
                        <CardContent>
                          <Grid container spacing={2} alignItems="center">
                            <Grid item xs={12} md={3}>
                              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                                <Avatar sx={{ 
                                  width: 56, 
                                  height: 56,
                                  bgcolor: device.isGethNode 
                                    ? theme.palette.warning.main 
                                    : theme.palette.success.main
                                }}>
                                  {device.isGethNode ? <RouterIcon /> : <DevBoardIcon />}
                                </Avatar>
                                <Box>
                                  <Typography variant="h6" fontWeight="bold">
                                    {device.device_id || 'Appareil IoT'}
                                  </Typography>
                                  <Chip 
                                    size="small" 
                                    label={device.isGethNode ? 'Nœud Geth PoA' : 'Device IoT'}
                                    color={device.isGethNode ? 'warning' : 'success'}
                                    sx={{ mt: 0.5 }}
                                  />
                                </Box>
                              </Box>
                            </Grid>

                            <Grid item xs={12} md={3}>
                              <Typography variant="caption" color="text.secondary">Adresse</Typography>
                              <Typography variant="body2" fontFamily="monospace" fontSize={12}>
                                {device.address}
                              </Typography>
                              <Box sx={{ mt: 1 }}>
                                <Typography variant="caption" color="text.secondary">Clé Publique</Typography>
                                <Typography variant="caption" fontFamily="monospace" display="block">
                                  {device.publicKey?.substring(0, 30)}...
                                </Typography>
                              </Box>
                            </Grid>

                            <Grid item xs={6} md={2}>
                              <Typography variant="caption" color="text.secondary">Statut</Typography>
                              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mt: 0.5 }}>
                                <Chip 
                                  size="small" 
                                  label={device.isActive ? 'Actif' : 'Inactif'}
                                  color={device.isActive ? 'success' : 'default'}
                                />
                              </Box>
                              <Box sx={{ mt: 1 }}>
                                <Typography variant="caption" color="text.secondary">Authentifications</Typography>
                                <Typography variant="h6" color="primary">
                                  {device.authCount || 0}
                                </Typography>
                              </Box>
                            </Grid>

                            <Grid item xs={6} md={2}>
                              <Typography variant="caption" color="text.secondary">Dernière activité</Typography>
                              <Typography variant="body2">
                                {device.lastSeen 
                                  ? formatDistanceToNow(device.lastSeen * 1000, { addSuffix: true, locale: fr })
                                  : 'Jamais'}
                              </Typography>
                              {device.isGethNode && (
                                <>
                                  <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
                                    Intervalle
                                  </Typography>
                                  <Typography variant="body2">
                                    {device.interval || 5}s
                                  </Typography>
                                </>
                              )}
                            </Grid>

                            <Grid item xs={12} md={2}>
                              <Stack direction="row" spacing={1} justifyContent="flex-end">
                                <Button
                                  size="small"
                                  variant="contained"
                                  onClick={() => handleViewDevice(device)}
                                >
                                  Détails complets
                                </Button>
                              </Stack>
                            </Grid>
                          </Grid>

                          {device.metadata && device.metadata !== '{}' && (
                            <Box sx={{ mt: 2, p: 1.5, bgcolor: alpha(theme.palette.primary.main, 0.05), borderRadius: 2 }}>
                              <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 0.5 }}>
                                Métadonnées:
                              </Typography>
                              <Typography variant="caption" fontFamily="monospace" sx={{ whiteSpace: 'pre-wrap' }}>
                                {typeof device.metadata === 'string' 
                                  ? device.metadata.substring(0, 200) + (device.metadata.length > 200 ? '...' : '')
                                  : JSON.stringify(device.metadata, null, 2).substring(0, 200)}
                              </Typography>
                            </Box>
                          )}
                        </CardContent>
                      </Card>
                    </Grid>
                  ))}
                </Grid>

                <TablePagination
                  rowsPerPageOptions={[5, 10, 25]}
                  component="div"
                  count={filteredDevices.length}
                  rowsPerPage={rowsPerPage}
                  page={page}
                  onPageChange={(e, newPage) => setPage(newPage)}
                  onRowsPerPageChange={(e) => {
                    setRowsPerPage(parseInt(e.target.value, 10));
                    setPage(0);
                  }}
                  labelRowsPerPage="Lignes par page"
                  sx={{ mt: 2 }}
                />
              </Box>
            )}

            {/* TAB 3: NŒUDS GETH */}
            {activeTab === 2 && (
              <Box sx={{ p: 3 }}>
                <Typography variant="h6" gutterBottom>
                  ⛏️ Nœuds Geth (Proof of Authority)
                </Typography>
                
                <Grid container spacing={3}>
                  {gethNodes.map((node, index) => (
                    <Grid item xs={12} md={6} key={node.address || index}>
                      <Card sx={{ 
                        borderRadius: 3,
                        boxShadow: theme.shadows[2],
                        border: `2px solid ${alpha(theme.palette.warning.main, 0.3)}`
                      }}>
                        <CardContent>
                          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
                            <Avatar sx={{ 
                              bgcolor: theme.palette.warning.main,
                              width: 56,
                              height: 56
                            }}>
                              <RouterIcon sx={{ fontSize: 28 }} />
                            </Avatar>
                            <Box>
                              <Typography variant="h6" fontWeight="bold">
                                {node.device_id || `GETH_NODE_${index + 1}`}
                              </Typography>
                              <Typography variant="caption" color="text.secondary">
                                PoA Validator Node
                              </Typography>
                            </Box>
                            <Chip 
                              label={node.isActive ? 'Mining' : 'Idle'}
                              color={node.isActive ? 'success' : 'default'}
                              sx={{ ml: 'auto' }}
                            />
                          </Box>

                          <Divider sx={{ my: 2 }} />

                          <Grid container spacing={2}>
                            <Grid item xs={6}>
                              <Typography variant="caption" color="text.secondary">
                                Adresse
                              </Typography>
                              <Typography variant="body2" fontFamily="monospace" fontSize={11}>
                                {node.address?.substring(0, 25)}...
                              </Typography>
                            </Grid>
                            <Grid item xs={6}>
                              <Typography variant="caption" color="text.secondary">
                                Intervalle
                              </Typography>
                              <Typography variant="body2" fontWeight="medium">
                                {node.interval || 5}s
                              </Typography>
                            </Grid>
                            <Grid item xs={6}>
                              <Typography variant="caption" color="text.secondary">
                                Authentifications
                              </Typography>
                              <Typography variant="body2" fontWeight="bold" color="primary">
                                {node.authCount || 0}
                              </Typography>
                            </Grid>
                            <Grid item xs={6}>
                              <Typography variant="caption" color="text.secondary">
                                Dernière activité
                              </Typography>
                              <Typography variant="body2">
                                {node.lastSeen 
                                  ? formatDistanceToNow(node.lastSeen * 1000, { addSuffix: true, locale: fr })
                                  : 'Jamais'}
                              </Typography>
                            </Grid>
                          </Grid>
                          
                          <Box sx={{ mt: 2 }}>
                            <Typography variant="caption" color="text.secondary">
                              Message par défaut
                            </Typography>
                            <Typography variant="body2" fontStyle="italic">
                              "{node.message || 'Heartbeat from Geth node'}"
                            </Typography>
                          </Box>
                        </CardContent>
                      </Card>
                    </Grid>
                  ))}
                </Grid>
              </Box>
            )}

            {/* TAB 4: IA Détection */}
            {activeTab === 3 && (
              <Box sx={{ p: 3 }}>
                <Typography variant="h6" gutterBottom>
                  🤖 Intelligence Artificielle - Détection d'Anomalies
                </Typography>
                
                <Grid container spacing={3}>
                  {metricsHistory.aiDetections.length > 0 ? (
                    <>
                      <Grid item xs={12} md={6}>
                        <Card sx={{ p: 2 }}>
                          <Typography variant="subtitle1" fontWeight="bold" gutterBottom>
                            Détections en Temps Réel
                          </Typography>
                          <Box sx={{ height: 300 }}>
                            <Bar data={aiDetectionsChartData} options={{
                              ...chartOptions,
                              maintainAspectRatio: false,
                            }} />
                          </Box>
                        </Card>
                      </Grid>
                      
                      <Grid item xs={12} md={6}>
                        <Card sx={{ p: 2 }}>
                          <Typography variant="subtitle1" fontWeight="bold" gutterBottom>
                            Précision du Modèle
                          </Typography>
                          <Box sx={{ height: 300 }}>
                            <Line data={aiAccuracyChartData} options={{
                              ...chartOptions,
                              maintainAspectRatio: false,
                              scales: {
                                y: { min: 98, max: 100 }
                              }
                            }} />
                          </Box>
                        </Card>
                      </Grid>
                      
                      <Grid item xs={12}>
                        <Card sx={{ p: 3, bgcolor: alpha(theme.palette.info.main, 0.05) }}>
                          <Typography variant="h6" gutterBottom>
                            Statistiques IA
                          </Typography>
                          <Grid container spacing={3}>
                            <Grid item xs={3}>
                              <Typography variant="caption" color="text.secondary">Détections totales</Typography>
                              <Typography variant="h4">{realtimeStats.aiDetections}</Typography>
                            </Grid>
                            <Grid item xs={3}>
                              <Typography variant="caption" color="text.secondary">Précision actuelle</Typography>
                              <Typography variant="h4">{realtimeStats.aiAccuracy.toFixed(1)}%</Typography>
                            </Grid>
                            <Grid item xs={3}>
                              <Typography variant="caption" color="text.secondary">Faux positifs</Typography>
                              <Typography variant="h4">{aiMetrics.falsePositives || 0}</Typography>
                            </Grid>
                            <Grid item xs={3}>
                              <Typography variant="caption" color="text.secondary">Temps d'analyse</Typography>
                              <Typography variant="h4">{aiMetrics.analysisTime || 0}ms</Typography>
                            </Grid>
                          </Grid>
                        </Card>
                      </Grid>
                    </>
                  ) : (
                    <Grid item xs={12}>
                      <Card sx={{ p: 4, textAlign: 'center' }}>
                        <AiIcon sx={{ fontSize: 80, color: 'text.disabled', mb: 2, opacity: 0.5 }} />
                        <Typography variant="h6" color="text.secondary" gutterBottom>
                          Module IA non configuré
                        </Typography>
                        <Typography variant="body2" color="text.disabled">
                          Les métriques de détection d'anomalies ne sont pas disponibles.
                          <br />
                          Configurez le module IA pour activer cette fonctionnalité.
                        </Typography>
                      </Card>
                    </Grid>
                  )}
                </Grid>
              </Box>
            )}
          </Paper>
        </Container>

        {/* Dialog détails appareil */}
        <Dialog 
          open={deviceDialogOpen} 
          onClose={() => setDeviceDialogOpen(false)}
          maxWidth="md"
          fullWidth
        >
          {selectedDevice && (
            <>
              <DialogTitle>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                  <Avatar sx={{ 
                    bgcolor: selectedDevice.isGethNode 
                      ? theme.palette.warning.main 
                      : theme.palette.success.main,
                    width: 50,
                    height: 50
                  }}>
                    {selectedDevice.isGethNode ? <RouterIcon /> : <DevBoardIcon />}
                  </Avatar>
                  <Box>
                    <Typography variant="h6">
                      {selectedDevice.device_id || 'Appareil'}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                      {selectedDevice.address}
                    </Typography>
                  </Box>
                </Box>
              </DialogTitle>
              <DialogContent dividers>
                <Grid container spacing={3}>
                  <Grid item xs={6}>
                    <Typography variant="subtitle2" gutterBottom>Informations Générales</Typography>
                    <Stack spacing={1}>
                      <Box>
                        <Typography variant="caption" color="text.secondary">ID</Typography>
                        <Typography variant="body2">{selectedDevice.device_id || 'N/A'}</Typography>
                      </Box>
                      <Box>
                        <Typography variant="caption" color="text.secondary">Adresse</Typography>
                        <Typography variant="body2" fontFamily="monospace">{selectedDevice.address}</Typography>
                      </Box>
                      <Box>
                        <Typography variant="caption" color="text.secondary">Clé Publique</Typography>
                        <Typography variant="body2" fontFamily="monospace" sx={{ wordBreak: 'break-all' }}>
                          {selectedDevice.publicKey || 'N/A'}
                        </Typography>
                      </Box>
                    </Stack>
                  </Grid>
                  
                  <Grid item xs={6}>
                    <Typography variant="subtitle2" gutterBottom>Statistiques</Typography>
                    <Stack spacing={1}>
                      <Box>
                        <Typography variant="caption" color="text.secondary">Statut</Typography>
                        <Chip 
                          size="small" 
                          label={selectedDevice.isActive ? 'Actif' : 'Inactif'}
                          color={selectedDevice.isActive ? 'success' : 'default'}
                          sx={{ ml: 1 }}
                        />
                      </Box>
                      <Box>
                        <Typography variant="caption" color="text.secondary">Authentifications</Typography>
                        <Typography variant="h6" color="primary">{selectedDevice.authCount || 0}</Typography>
                      </Box>
                      <Box>
                        <Typography variant="caption" color="text.secondary">Dernière activité</Typography>
                        <Typography variant="body2">
                          {selectedDevice.lastSeen 
                            ? format(new Date(selectedDevice.lastSeen * 1000), 'PPpp', { locale: fr })
                            : 'Jamais'}
                        </Typography>
                      </Box>
                    </Stack>
                  </Grid>
                  
                  {selectedDevice.metadata && selectedDevice.metadata !== '{}' && (
                    <Grid item xs={12}>
                      <Typography variant="subtitle2" gutterBottom>Métadonnées</Typography>
                      <Paper sx={{ p: 2, bgcolor: alpha(theme.palette.primary.main, 0.05) }}>
                        <pre style={{ margin: 0, fontSize: 12, overflow: 'auto' }}>
                          {typeof selectedDevice.metadata === 'string' 
                            ? JSON.stringify(JSON.parse(selectedDevice.metadata), null, 2)
                            : JSON.stringify(selectedDevice.metadata, null, 2)}
                        </pre>
                      </Paper>
                    </Grid>
                  )}
                </Grid>
              </DialogContent>
              <DialogActions>
                <Button onClick={() => setDeviceDialogOpen(false)}>
                  Fermer
                </Button>
              </DialogActions>
            </>
          )}
        </Dialog>

        {/* Dialog détails transaction */}
        <Dialog 
          open={transactionDialogOpen} 
          onClose={() => setTransactionDialogOpen(false)}
          maxWidth="sm"
          fullWidth
        >
          {selectedTransaction && (
            <>
              <DialogTitle>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                  <Avatar sx={{ 
                    bgcolor: selectedTransaction.status === 'success' 
                      ? theme.palette.success.main 
                      : theme.palette.error.main,
                    width: 50,
                    height: 50
                  }}>
                    {selectedTransaction.status === 'success' ? <VerifiedIcon /> : <ErrorIcon />}
                  </Avatar>
                  <Box>
                    <Typography variant="h6">
                      Transaction {selectedTransaction.status === 'success' ? 'Réussie' : 'Échouée'}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                      Bloc #{selectedTransaction.blockNumber}
                    </Typography>
                  </Box>
                </Box>
              </DialogTitle>
              <DialogContent dividers>
                <Stack spacing={2}>
                  <Box>
                    <Typography variant="caption" color="text.secondary">Hash</Typography>
                    <Typography variant="body2" fontFamily="monospace" sx={{ wordBreak: 'break-all' }}>
                      {selectedTransaction.hash}
                    </Typography>
                  </Box>
                  <Box>
                    <Typography variant="caption" color="text.secondary">Appareil</Typography>
                    <Typography variant="body2" fontWeight="medium">
                      {selectedTransaction.deviceId}
                    </Typography>
                    <Typography variant="caption" fontFamily="monospace">
                      {selectedTransaction.deviceAddress}
                    </Typography>
                  </Box>
                  <Box>
                    <Typography variant="caption" color="text.secondary">Message</Typography>
                    <Typography variant="body2">
                      {selectedTransaction.message}
                    </Typography>
                  </Box>
                  <Grid container spacing={2}>
                    <Grid item xs={6}>
                      <Typography variant="caption" color="text.secondary">ECDSA</Typography>
                      <Box>
                        {selectedTransaction.ecdsaValid ? (
                          <Chip size="small" icon={<CheckCircleIcon />} label="Validé" color="success" />
                        ) : (
                          <Chip size="small" icon={<ErrorIcon />} label="Invalide" color="error" />
                        )}
                      </Box>
                    </Grid>
                    <Grid item xs={6}>
                      <Typography variant="caption" color="text.secondary">ZKP</Typography>
                      <Box>
                        {selectedTransaction.zkpValid ? (
                          <Chip size="small" icon={<VerifiedIcon />} label="Validé" color="success" />
                        ) : (
                          <Chip size="small" icon={<ErrorIcon />} label="Invalide" color="error" />
                        )}
                      </Box>
                    </Grid>
                  </Grid>
                  <Box>
                    <Typography variant="caption" color="text.secondary">Timestamp</Typography>
                    <Typography variant="body2">
                      {selectedTransaction.timestamp 
                        ? format(new Date(selectedTransaction.timestamp), 'PPpp', { locale: fr })
                        : 'Inconnu'}
                    </Typography>
                  </Box>
                  <Box>
                    <Typography variant="caption" color="text.secondary">Latence</Typography>
                    <Typography variant="body2">
                      {selectedTransaction.latency}ms
                    </Typography>
                  </Box>
                  {selectedTransaction.proofHash && (
                    <Box>
                      <Typography variant="caption" color="text.secondary">Preuve ZKP</Typography>
                      <Typography variant="caption" fontFamily="monospace" sx={{ wordBreak: 'break-all' }}>
                        {selectedTransaction.proofHash}
                      </Typography>
                    </Box>
                  )}
                </Stack>
              </DialogContent>
              <DialogActions>
                <Button onClick={() => setTransactionDialogOpen(false)}>
                  Fermer
                </Button>
              </DialogActions>
            </>
          )}
        </Dialog>
      </Box>

      <ErrorSnackbar 
        open={errorOpen}
        message={errorMessage}
        onClose={handleCloseError}
      />
    </Box>
  );
};

export default BlockchainDashboard;
