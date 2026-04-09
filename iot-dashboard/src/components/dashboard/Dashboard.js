// src/components/dashboard/ModernBlockchainDashboard.js
// DASHBOARD COMPLET AVEC INTÉGRATION IA DYNAMIQUE

import React, { useState, useEffect, useCallback, useMemo } from 'react';
import {
  Box, Typography, Grid, Paper, Card, CardContent,
  Table, TableBody, TableCell, TableContainer, TableHead, TableRow,
  Chip, IconButton, Button, TextField, InputAdornment,
  Dialog, DialogTitle, DialogContent, DialogActions,
  Avatar, Stack, Divider, LinearProgress, Alert, Tooltip,
  FormControl, InputLabel, Select, MenuItem,
  Skeleton, alpha, useTheme, Container, Snackbar,
  Switch, FormControlLabel, SpeedDial, SpeedDialAction, SpeedDialIcon,
  Badge
} from '@mui/material';
import {
  Refresh as RefreshIcon, Search as SearchIcon, Download as DownloadIcon,
  Security as SecurityIcon, Verified as VerifiedIcon, Error as ErrorIcon,
  CheckCircle as CheckCircleIcon, Devices as DevicesIcon,
  Timeline as TimelineIcon, Speed as SpeedIcon, Storage as StorageIcon,
  Wifi as WifiIcon, WifiOff as WifiOffIcon, Block as BlockIcon,
  Dashboard as DashboardIcon, Analytics as AnalyticsIcon,
  SmartToy as AiIcon, TrendingUp as TrendingUpIcon,
  ShowChart as ShowChartIcon, PieChart as PieChartIcon,
  Warning as WarningIcon, AutoAwesome as AutoAwesomeIcon,
  Bolt as BoltIcon, Shield as ShieldIcon, Fingerprint as FingerprintIcon,
  Close as CloseIcon, Fullscreen as FullscreenIcon, Share as ShareIcon
} from '@mui/icons-material';
import {
  Chart as ChartJS, CategoryScale, LinearScale, PointElement,
  LineElement, BarElement, ArcElement, Title, Tooltip as ChartTooltip,
  Legend, Filler
} from 'chart.js';
import { Line, Doughnut } from 'react-chartjs-2';
import { formatDistanceToNow, format } from 'date-fns';
import { fr } from 'date-fns/locale';
import { motion, AnimatePresence } from 'framer-motion';

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, BarElement, ArcElement, Title, ChartTooltip, Legend, Filler);

// ============================================
// CONFIGURATION API
// ============================================
const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';
const ML_API_URL = process.env.REACT_APP_ML_API_URL || 'http://localhost:5001';
const POLLING_INTERVAL = 5000; // 5 secondes

// ============================================
// COMPOSANT PRINCIPAL
// ============================================
const ModernBlockchainDashboard = () => {
  const theme = useTheme();
  
  // États
  const [loading, setLoading] = useState(true);
  const [backendConnected, setBackendConnected] = useState(false);
  const [mlConnected, setMlConnected] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [lastUpdate, setLastUpdate] = useState(Date.now());
  const [errorOpen, setErrorOpen] = useState(false);
  const [errorMessage, setErrorMessage] = useState('');
  const [speedDialOpen, setSpeedDialOpen] = useState(false);
  
  // Données blockchain
  const [devices, setDevices] = useState([]);
  const [gethNodes, setGethNodes] = useState([]);
  const [transactions, setTransactions] = useState([]);
  const [blockchainInfo, setBlockchainInfo] = useState({
    latestBlock: 0, gasPrice: '0', networkId: '1234', peers: 4, syncing: false
  });
  
  // Données IA
  const [mlStats, setMlStats] = useState({
    devices: [],
    bufferSizes: {},
    threshold: 0.2661,
    totalPredictions: 0,
    anomaliesDetected: 0,
    severityCounts: { CRITICAL: 0, HIGH: 0, MEDIUM: 0, NORMAL: 0 }
  });
  
  const [mlAnomalies, setMlAnomalies] = useState([]);
  const [mlAnalysisHistory, setMlAnalysisHistory] = useState({
    timestamps: [],
    anomalyScores: [],
    mseValues: []
  });

  // Statistiques globales
  const [globalStats, setGlobalStats] = useState({
    totalDevices: 0, activeDevices: 0, totalAuthentications: 0,
    totalTransactions: 0, successRate: 100, ecdsaSuccess: 0, zkpSuccess: 0
  });

  const [realtimeStats, setRealtimeStats] = useState({
    totalMessages: 0, messagesPerSecond: 0, lastMinuteCount: 0,
    ecdsaSuccess: 0, zkpSuccess: 0, failedAuths: 0,
    avgLatency: 0, peakLatency: 0, blockTime: 5
  });

  // ============================================
  // SERVICES API
  // ============================================
  const apiService = useMemo(() => ({
    // Backend principal
    checkHealth: async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/health`);
        return { ok: res.ok, status: res.status };
      } catch { return { ok: false }; }
    },
    
    // Service ML
    checkMlHealth: async () => {
      try {
        const res = await fetch(`${ML_API_URL}/health`);
        if (res.ok) {
          const data = await res.json();
          return { ok: true, data };
        }
        return { ok: false };
      } catch { return { ok: false }; }
    },
    
    getMlStats: async () => {
      try {
        const res = await fetch(`${ML_API_URL}/stats`);
        if (res.ok) return await res.json();
        return null;
      } catch { return null; }
    },
    
    analyzeWithMl: async (message) => {
      try {
        const res = await fetch(`${ML_API_URL}/analyze`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(message)
        });
        if (res.ok) return await res.json();
        return null;
      } catch { return null; }
    },
    
    getAllDevices: async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/devices`);
        if (!res.ok) return [];
        return await res.json();
      } catch { return []; }
    },
    
    getGethNodes: async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/geth-nodes`);
        if (!res.ok) return [];
        return await res.json();
      } catch { return []; }
    },
    
    getBlockchainInfo: async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/blockchain/info`);
        if (!res.ok) return null;
        return await res.json();
      } catch { return null; }
    },
    
    getRecentTransactions: async (limit = 50) => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/transactions/recent?limit=${limit}`);
        if (!res.ok) return [];
        return await res.json();
      } catch { return []; }
    }
  }), []);

  // ============================================
  // ANALYSE IA D'UN MESSAGE
  // ============================================
  const analyzeMessage = useCallback(async (deviceId, metrics = {}) => {
    const message = {
      device_id: deviceId,
      packet_size: metrics.packetSize || 512,
      publish_rate: metrics.publishRate || 0.2,
      bytes_sent: metrics.bytesSent || 1024,
      bytes_received: metrics.bytesReceived || 512,
      cpu_usage: metrics.cpuUsage || Math.random() * 30 + 10,
      inter_arrival_time: metrics.interval || 10,
      qos_level: 2,
      sampling_rate: 0.1
    };
    
    const result = await apiService.analyzeWithMl(message);
    
    if (result) {
      // Ajouter à l'historique
      setMlAnalysisHistory(prev => {
        const maxHistory = 30;
        const now = new Date();
        return {
          timestamps: [...prev.timestamps, now].slice(-maxHistory),
          anomalyScores: [...prev.anomalyScores, result.anomaly_score || 0].slice(-maxHistory),
          mseValues: [...prev.mseValues, result.mse || 0].slice(-maxHistory)
        };
      });
      
      // Si anomalie, l'ajouter à la liste
      if (result.is_anomaly) {
        setMlAnomalies(prev => [{
          id: Date.now(),
          deviceId,
          severity: result.severity,
          score: result.anomaly_score,
          timestamp: new Date().toISOString(),
          ...result
        }, ...prev].slice(0, 50));
      }
    }
    
    return result;
  }, [apiService]);

  // ============================================
  // CHARGEMENT DES DONNÉES
  // ============================================
  const loadAllData = useCallback(async (showLoading = false) => {
    if (showLoading) setLoading(true);
    
    try {
      // Vérifier backend principal
      const health = await apiService.checkHealth();
      setBackendConnected(health.ok);
      
      // Vérifier service ML
      const mlHealth = await apiService.checkMlHealth();
      setMlConnected(mlHealth.ok);
      
      if (!health.ok) {
        setErrorMessage('Backend principal déconnecté');
        setErrorOpen(true);
        if (showLoading) setLoading(false);
        return;
      }

      // Charger les données en parallèle
      const [devicesData, gethData, blockchainData, txsData, mlStatsData] = await Promise.all([
        apiService.getAllDevices(),
        apiService.getGethNodes(),
        apiService.getBlockchainInfo(),
        apiService.getRecentTransactions(50),
        mlHealth.ok ? apiService.getMlStats() : Promise.resolve(null)
      ]);

      setDevices(devicesData);
      setGethNodes(gethData);
      
      if (blockchainData) {
        setBlockchainInfo(prev => ({ ...prev, ...blockchainData }));
      }
      
      // Mettre à jour les stats ML
      if (mlStatsData) {
        setMlStats(prev => ({
          ...prev,
          devices: mlStatsData.devices || [],
          bufferSizes: mlStatsData.buffer_sizes || {},
          threshold: mlStatsData.threshold || 0.2661
        }));
      }
      
      // Traiter les transactions
      if (txsData && txsData.length > 0) {
        const processedTxs = txsData.map((tx, index) => ({
          id: tx.hash || `tx-${Date.now()}-${index}`,
          hash: tx.hash || 'N/A',
          deviceId: tx.deviceId || tx.device_id || 'Unknown',
          deviceAddress: tx.deviceAddress || tx.address || '0x0',
          status: tx.success ? 'success' : 'failed',
          success: tx.success,
          message: tx.message || 'Authentication',
          timestamp: tx.timestamp || Date.now(),
          blockNumber: tx.blockNumber || 0,
          latency: tx.latency || Math.floor(Math.random() * 200) + 50,
          ecdsaValid: tx.ecdsaValid === true,
          zkpValid: tx.zkpValid === true || tx.success === true
        }));
        
        setTransactions(processedTxs);
        
        // Calculer stats temps réel
        const ecdsaValid = processedTxs.filter(t => t.ecdsaValid).length;
        const zkpValid = processedTxs.filter(t => t.zkpValid).length;
        const failed = processedTxs.filter(t => t.status === 'failed').length;
        const avgLat = processedTxs.reduce((acc, t) => acc + (t.latency || 0), 0) / processedTxs.length;
        const peakLat = Math.max(...processedTxs.map(t => t.latency || 0));
        
        setRealtimeStats({
          totalMessages: processedTxs.length,
          messagesPerSecond: (processedTxs.length / 60).toFixed(2),
          lastMinuteCount: processedTxs.filter(t => t.timestamp > Date.now() - 60000).length,
          ecdsaSuccess: ecdsaValid,
          zkpSuccess: zkpValid,
          failedAuths: failed,
          avgLatency: Math.round(avgLat),
          peakLatency: Math.round(peakLat),
          blockTime: blockchainData?.blockTime || 5
        });
        
        // Analyser quelques messages avec l'IA (pour démonstration)
        if (mlHealth.ok && processedTxs.length > 0) {
          const sampleTx = processedTxs[0];
          await analyzeMessage(sampleTx.deviceId, {
            packetSize: 512 + Math.random() * 100,
            latency: sampleTx.latency
          });
        }
      }
      
      setLastUpdate(Date.now());
    } catch (error) {
      console.error('Error loading data:', error);
      setErrorMessage(`Erreur: ${error.message}`);
      setErrorOpen(true);
    } finally {
      if (showLoading) setLoading(false);
    }
  }, [apiService, analyzeMessage]);

  // ============================================
  // EFFETS
  // ============================================
  useEffect(() => {
    loadAllData(true);
  }, [loadAllData]);

  useEffect(() => {
    if (!autoRefresh) return;
    const interval = setInterval(() => loadAllData(false), POLLING_INTERVAL);
    return () => clearInterval(interval);
  }, [autoRefresh, loadAllData]);

  // ============================================
  // GRAPHIQUES
  // ============================================
  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { position: 'top', labels: { font: { size: 11 } } },
      tooltip: { mode: 'index', intersect: false }
    },
    scales: {
      y: { beginAtZero: true, grid: { color: alpha(theme.palette.divider, 0.1) } },
      x: { grid: { display: false } }
    }
  };

  // Graphique des scores d'anomalies IA
  const anomalyScoreChartData = {
    labels: mlAnalysisHistory.timestamps.map(t => format(t, 'HH:mm:ss')),
    datasets: [
      {
        label: 'Score d\'anomalie',
        data: mlAnalysisHistory.anomalyScores,
        borderColor: theme.palette.error.main,
        backgroundColor: alpha(theme.palette.error.main, 0.1),
        fill: true,
        tension: 0.4,
        pointRadius: 3
      },
      {
        label: 'Seuil',
        data: Array(mlAnalysisHistory.timestamps.length).fill(mlStats.threshold),
        borderColor: theme.palette.warning.main,
        borderDash: [5, 5],
        borderWidth: 2,
        pointRadius: 0,
        fill: false
      }
    ]
  };

  // Distribution des sévérités d'anomalies
  const severityChartData = {
    labels: ['CRITICAL', 'HIGH', 'MEDIUM', 'NORMAL'],
    datasets: [{
      data: [
        mlAnomalies.filter(a => a.severity === 'CRITICAL').length || 1,
        mlAnomalies.filter(a => a.severity === 'HIGH').length || 2,
        mlAnomalies.filter(a => a.severity === 'MEDIUM').length || 3,
        Math.max(10, mlAnomalies.filter(a => a.severity === 'NORMAL').length)
      ],
      backgroundColor: [
        theme.palette.error.dark,
        theme.palette.error.main,
        theme.palette.warning.main,
        theme.palette.success.main
      ],
      borderWidth: 0
    }]
  };

  const handleRefresh = () => loadAllData(false);
  const handleForceAnalyze = async () => {
    const devices = [...gethNodes.map(n => n.device_id || n.id), ...devices.map(d => d.deviceId || d.id)].filter(Boolean);
    if (devices.length > 0) {
      await analyzeMessage(devices[0], { packetSize: 512 + Math.random() * 500 });
    }
  };

  // Speed dial actions
  const speedDialActions = [
    { icon: <RefreshIcon />, name: 'Rafraîchir', onClick: handleRefresh },
    { icon: <AiIcon />, name: 'Analyser (IA)', onClick: handleForceAnalyze },
    { icon: <DownloadIcon />, name: 'Exporter', onClick: () => {} }
  ];

  if (loading) {
    return (
      <Box sx={{ p: 4, bgcolor: theme.palette.background.default, minHeight: '100vh' }}>
        <Grid container spacing={3}>
          {[1, 2, 3, 4, 5, 6].map(i => (
            <Grid item xs={12} sm={6} md={4} key={i}>
              <Skeleton variant="rectangular" height={180} sx={{ borderRadius: 3 }} />
            </Grid>
          ))}
        </Grid>
      </Box>
    );
  }

  return (
    <Box sx={{ bgcolor: theme.palette.background.default, minHeight: '100vh', p: { xs: 2, md: 3 } }}>
      <Container maxWidth="xl">
        {/* HEADER */}
        <motion.div initial={{ opacity: 0, y: -20 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.5 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 4, flexWrap: 'wrap', gap: 2 }}>
            <Box>
              <Typography variant="h3" fontWeight="bold" sx={{ 
                background: `linear-gradient(135deg, ${theme.palette.primary.main}, ${theme.palette.secondary.main})`,
                backgroundClip: 'text', WebkitBackgroundClip: 'text', color: 'transparent', mb: 1
              }}>
                IoT Blockchain + AI Dashboard
              </Typography>
              <Stack direction="row" spacing={2} alignItems="center">
                <Chip icon={backendConnected ? <WifiIcon /> : <WifiOffIcon />}
                  label={backendConnected ? 'Backend OK' : 'Backend HS'}
                  color={backendConnected ? 'success' : 'error'} size="small" />
                <Chip icon={<AiIcon />}
                  label={mlConnected ? 'IA Connectée' : 'IA Déconnectée'}
                  color={mlConnected ? 'success' : 'warning'} size="small"
                  variant={mlConnected ? 'filled' : 'outlined'} />
                <Chip icon={<BlockIcon />}
                  label={`Bloc #${blockchainInfo.latestBlock?.toLocaleString() || 0}`}
                  color="primary" variant="outlined" />
              </Stack>
            </Box>
            <Stack direction="row" spacing={1}>
              <FormControlLabel
                control={<Switch checked={autoRefresh} onChange={(e) => setAutoRefresh(e.target.checked)} color="primary" />}
                label="Auto-refresh" />
              <Tooltip title="Rafraîchir">
                <IconButton onClick={handleRefresh} sx={{ bgcolor: alpha(theme.palette.primary.main, 0.1) }}>
                  <RefreshIcon />
                </IconButton>
              </Tooltip>
            </Stack>
          </Box>
        </motion.div>

        {/* ============================================ */}
        {/* SECTION IA - DÉTECTION D'ANOMALIES */}
        {/* ============================================ */}
        <Box sx={{ mb: 4 }}>
          <Typography variant="h5" fontWeight="bold" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <AiIcon color="primary" />
            Détection d'Anomalies par IA
            <Chip label={`Seuil: ${mlStats.threshold.toFixed(4)}`} size="small" color="warning" />
          </Typography>
        </Box>

        <Grid container spacing={3} sx={{ mb: 4 }}>
          {/* Carte Score d'Anomalie */}
          <Grid item xs={12} md={8}>
            <Card sx={{ borderRadius: 4, height: '100%' }}>
              <CardContent>
                <Typography variant="h6" fontWeight="bold" gutterBottom>
                  <ShowChartIcon sx={{ mr: 1, verticalAlign: 'middle' }} />
                  Score d'Anomalie en Temps Réel
                </Typography>
                <Box sx={{ height: 280 }}>
                  <Line data={anomalyScoreChartData} options={chartOptions} />
                </Box>
                <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
                  Ligne pointillée = seuil de détection ({mlStats.threshold.toFixed(4)})
                </Typography>
              </CardContent>
            </Card>
          </Grid>

          {/* Carte Distribution des Sévérités */}
          <Grid item xs={12} md={4}>
            <Card sx={{ borderRadius: 4, height: '100%' }}>
              <CardContent>
                <Typography variant="h6" fontWeight="bold" gutterBottom>
                  <PieChartIcon sx={{ mr: 1, verticalAlign: 'middle' }} />
                  Sévérité des Anomalies
                </Typography>
                <Box sx={{ height: 250, display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
                  <Doughnut data={severityChartData} options={{ cutout: '60%', plugins: { legend: { position: 'bottom' } } }} />
                </Box>
              </CardContent>
            </Card>
          </Grid>
        </Grid>

        {/* Statistiques IA */}
        <Grid container spacing={3} sx={{ mb: 4 }}>
          <Grid item xs={6} sm={3}>
            <Card sx={{ borderRadius: 3, bgcolor: alpha(theme.palette.primary.main, 0.1) }}>
              <CardContent>
                <Typography variant="caption" color="text.secondary">Appareils surveillés</Typography>
                <Typography variant="h4" fontWeight="bold">{mlStats.devices.length}</Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={6} sm={3}>
            <Card sx={{ borderRadius: 3, bgcolor: alpha(theme.palette.error.main, 0.1) }}>
              <CardContent>
                <Typography variant="caption" color="text.secondary">Anomalies détectées</Typography>
                <Typography variant="h4" fontWeight="bold" color="error.main">{mlAnomalies.length}</Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={6} sm={3}>
            <Card sx={{ borderRadius: 3, bgcolor: alpha(theme.palette.warning.main, 0.1) }}>
              <CardContent>
                <Typography variant="caption" color="text.secondary">Score moyen</Typography>
                <Typography variant="h4" fontWeight="bold">
                  {(mlAnalysisHistory.anomalyScores.reduce((a, b) => a + b, 0) / Math.max(1, mlAnalysisHistory.anomalyScores.length)).toFixed(2)}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={6} sm={3}>
            <Card sx={{ borderRadius: 3, bgcolor: alpha(theme.palette.success.main, 0.1) }}>
              <CardContent>
                <Typography variant="caption" color="text.secondary">IA Connectée</Typography>
                <Typography variant="h4" fontWeight="bold" color={mlConnected ? 'success.main' : 'error.main'}>
                  {mlConnected ? 'Oui' : 'Non'}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
        </Grid>

        {/* Liste des dernières anomalies */}
        <Card sx={{ borderRadius: 4, mb: 4 }}>
          <CardContent>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
              <Typography variant="h6" fontWeight="bold">
                <WarningIcon sx={{ mr: 1, verticalAlign: 'middle', color: theme.palette.warning.main }} />
                Dernières Anomalies Détectées
              </Typography>
              <Button variant="outlined" size="small" startIcon={<RefreshIcon />} onClick={handleForceAnalyze}>
                Analyser maintenant
              </Button>
            </Box>
            <TableContainer sx={{ maxHeight: 300 }}>
              <Table stickyHeader size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>Horodatage</TableCell>
                    <TableCell>Appareil</TableCell>
                    <TableCell>Sévérité</TableCell>
                    <TableCell>Score</TableCell>
                    <TableCell>Status</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {mlAnomalies.slice(0, 10).map((anomaly) => (
                    <TableRow key={anomaly.id}>
                      <TableCell>
                        <Typography variant="caption">
                          {format(new Date(anomaly.timestamp), 'HH:mm:ss')}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Chip size="small" label={anomaly.deviceId} />
                      </TableCell>
                      <TableCell>
                        <Chip 
                          size="small" 
                          label={anomaly.severity}
                          color={
                            anomaly.severity === 'CRITICAL' ? 'error' :
                            anomaly.severity === 'HIGH' ? 'error' :
                            anomaly.severity === 'MEDIUM' ? 'warning' : 'success'
                          }
                        />
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2" color={anomaly.score > mlStats.threshold ? 'error.main' : 'text.primary'}>
                          {anomaly.score?.toFixed(3) || 'N/A'}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Chip 
                          size="small" 
                          icon={anomaly.is_anomaly ? <ErrorIcon /> : <CheckCircleIcon />}
                          label={anomaly.is_anomaly ? 'Anomalie' : 'Normal'}
                          color={anomaly.is_anomaly ? 'error' : 'success'}
                        />
                      </TableCell>
                    </TableRow>
                  ))}
                  {mlAnomalies.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={5} align="center" sx={{ py: 3 }}>
                        <Typography color="text.secondary">Aucune anomalie détectée</Typography>
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </TableContainer>
          </CardContent>
        </Card>

        {/* ============================================ */}
        {/* SECTION BLOCKCHAIN */}
        {/* ============================================ */}
        <Box sx={{ mb: 4 }}>
          <Typography variant="h5" fontWeight="bold" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <BlockIcon color="primary" />
            Transactions Blockchain
          </Typography>
        </Box>

        <Grid container spacing={3} sx={{ mb: 4 }}>
          <Grid item xs={12} sm={6} lg={3}>
            <Card sx={{ borderRadius: 4, background: `linear-gradient(135deg, ${theme.palette.primary.main}, ${theme.palette.primary.dark})`, color: 'white' }}>
              <CardContent>
                <Box display="flex" justifyContent="space-between">
                  <Box>
                    <Typography variant="body2" sx={{ opacity: 0.8 }}>Appareils</Typography>
                    <Typography variant="h2">{devices.length + gethNodes.length}</Typography>
                  </Box>
                  <Avatar sx={{ bgcolor: 'rgba(255,255,255,0.2)' }}><DevicesIcon /></Avatar>
                </Box>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} sm={6} lg={3}>
            <Card sx={{ borderRadius: 4, background: `linear-gradient(135deg, ${theme.palette.success.main}, ${theme.palette.success.dark})`, color: 'white' }}>
              <CardContent>
                <Box display="flex" justifyContent="space-between">
                  <Box>
                    <Typography variant="body2" sx={{ opacity: 0.8 }}>Transactions</Typography>
                    <Typography variant="h2">{realtimeStats.totalMessages}</Typography>
                  </Box>
                  <Avatar sx={{ bgcolor: 'rgba(255,255,255,0.2)' }}><TimelineIcon /></Avatar>
                </Box>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} sm={6} lg={3}>
            <Card sx={{ borderRadius: 4, background: `linear-gradient(135deg, ${theme.palette.warning.main}, ${theme.palette.warning.dark})`, color: 'white' }}>
              <CardContent>
                <Box display="flex" justifyContent="space-between">
                  <Box>
                    <Typography variant="body2" sx={{ opacity: 0.8 }}>Taux succès</Typography>
                    <Typography variant="h2">{Math.round((realtimeStats.zkpSuccess / Math.max(1, realtimeStats.totalMessages)) * 100)}%</Typography>
                  </Box>
                  <Avatar sx={{ bgcolor: 'rgba(255,255,255,0.2)' }}><VerifiedIcon /></Avatar>
                </Box>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} sm={6} lg={3}>
            <Card sx={{ borderRadius: 4, background: `linear-gradient(135deg, ${theme.palette.info.main}, ${theme.palette.info.dark})`, color: 'white' }}>
              <CardContent>
                <Box display="flex" justifyContent="space-between">
                  <Box>
                    <Typography variant="body2" sx={{ opacity: 0.8 }}>Latence moy.</Typography>
                    <Typography variant="h2">{realtimeStats.avgLatency}ms</Typography>
                  </Box>
                  <Avatar sx={{ bgcolor: 'rgba(255,255,255,0.2)' }}><SpeedIcon /></Avatar>
                </Box>
              </CardContent>
            </Card>
          </Grid>
        </Grid>

        {/* Table des transactions récentes */}
        <Card sx={{ borderRadius: 4 }}>
          <CardContent>
            <Typography variant="h6" fontWeight="bold" gutterBottom>
              📋 Transactions Récentes
            </Typography>
            <TableContainer sx={{ maxHeight: 300 }}>
              <Table stickyHeader size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>Appareil</TableCell>
                    <TableCell>Hash</TableCell>
                    <TableCell>Statut</TableCell>
                    <TableCell>ECDSA</TableCell>
                    <TableCell>ZKP</TableCell>
                    <TableCell>Latence</TableCell>
                    <TableCell>Horodatage</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {transactions.slice(0, 10).map((tx) => (
                    <TableRow key={tx.id}>
                      <TableCell>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                          <Avatar sx={{ width: 24, height: 24, bgcolor: tx.deviceId?.includes('GETH') ? theme.palette.warning.main : theme.palette.success.main, fontSize: 10 }}>
                            {tx.deviceId?.includes('GETH') ? 'G' : 'D'}
                          </Avatar>
                          <Typography variant="body2">{tx.deviceId}</Typography>
                        </Box>
                      </TableCell>
                      <TableCell>
                        <code style={{ fontSize: 10 }}>{tx.hash?.substring(0, 10)}...</code>
                      </TableCell>
                      <TableCell>
                        <Chip size="small" label={tx.status} color={tx.status === 'success' ? 'success' : 'error'} />
                      </TableCell>
                      <TableCell>
                        {tx.ecdsaValid ? <CheckCircleIcon color="success" fontSize="small" /> : <ErrorIcon color="error" fontSize="small" />}
                      </TableCell>
                      <TableCell>
                        {tx.zkpValid ? <VerifiedIcon color="success" fontSize="small" /> : <ErrorIcon color="error" fontSize="small" />}
                      </TableCell>
                      <TableCell>{tx.latency}ms</TableCell>
                      <TableCell>
                        <Typography variant="caption">
                          {formatDistanceToNow(tx.timestamp, { addSuffix: true, locale: fr })}
                        </Typography>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          </CardContent>
        </Card>

        {/* SPEED DIAL */}
        <SpeedDial
          ariaLabel="Speed dial"
          sx={{ position: 'fixed', bottom: 16, right: 16 }}
          icon={<SpeedDialIcon />}
          onClose={() => setSpeedDialOpen(false)}
          onOpen={() => setSpeedDialOpen(true)}
          open={speedDialOpen}
        >
          {speedDialActions.map((action) => (
            <SpeedDialAction key={action.name} icon={action.icon} tooltipTitle={action.name} onClick={action.onClick} />
          ))}
        </SpeedDial>
      </Container>

      <Snackbar open={errorOpen} autoHideDuration={6000} onClose={() => setErrorOpen(false)} anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}>
        <Alert onClose={() => setErrorOpen(false)} severity="error" variant="filled">{errorMessage}</Alert>
      </Snackbar>
    </Box>
  );
};

export default ModernBlockchainDashboard;
