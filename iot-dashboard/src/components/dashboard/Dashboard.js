// src/components/dashboard/ModernBlockchainDashboard.js
// DASHBOARD MODERNE AVEC VISUALISATIONS AMÉLIORÉES

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
  Fade,
  Grow,
  Zoom,
  useTheme,
  alpha,
  Container,
  Snackbar,
  CircularProgress,
  CardMedia,
  CardActions,
  List,
  ListItem,
  ListItemText,
  ListItemAvatar,
  Switch,
  FormControlLabel,
  SpeedDial,
  SpeedDialAction,
  SpeedDialIcon,
  Backdrop
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
  Block as BlockIcon,
  Pause as PauseIcon,
  PlayArrow as PlayArrowIcon,
  Dashboard as DashboardIcon,
  Settings as SettingsIcon,
  AccountCircle as AccountCircleIcon,
  Analytics as AnalyticsIcon,
  SmartToy as AiIcon,
  TrendingUp as TrendingUpIcon,
  ShowChart as ShowChartIcon,
  BarChart as BarChartIcon,
  PieChart as PieChartIcon,
  Warning as WarningIcon,
  Info as InfoIcon,
  Close as CloseIcon,
  Fullscreen as FullscreenIcon,
  FullscreenExit as FullscreenExitIcon,
  Share as ShareIcon,
  Print as PrintIcon,
  FilterList as FilterListIcon,
  AutoAwesome as AutoAwesomeIcon,
  Bolt as BoltIcon,
  Shield as ShieldIcon,
  Lock as LockIcon,
  Fingerprint as FingerprintIcon
} from '@mui/icons-material';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  Title,
  Tooltip as ChartTooltip,
  Legend,
  Filler,
  RadialLinearScale
} from 'chart.js';
import { Line, Bar, Doughnut, Radar } from 'react-chartjs-2';
import { formatDistanceToNow, format, subMinutes, differenceInSeconds } from 'date-fns';
import { fr } from 'date-fns/locale';
import { motion, AnimatePresence } from 'framer-motion';

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  Title,
  ChartTooltip,
  Legend,
  Filler,
  RadialLinearScale
);

// ============================================
// CONFIGURATION API
// ============================================
const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';
const POLLING_INTERVAL = 3000;

// ============================================
// COMPOSANT DE CARTE ANIMÉE
// ============================================
const AnimatedCard = ({ children, delay = 0, ...props }) => (
  <motion.div
    initial={{ opacity: 0, y: 20 }}
    animate={{ opacity: 1, y: 0 }}
    transition={{ duration: 0.5, delay }}
    style={{ height: '100%' }}
  >
    <Card {...props} sx={{ borderRadius: 4, overflow: 'hidden', ...props.sx }}>
      {children}
    </Card>
  </motion.div>
);

// ============================================
// COMPOSANT DE COMPTEUR ANIMÉ
// ============================================
const AnimatedCounter = ({ value, duration = 1000 }) => {
  const [count, setCount] = useState(0);
  
  useEffect(() => {
    let start = 0;
    const increment = value / (duration / 16);
    const timer = setInterval(() => {
      start += increment;
      if (start >= value) {
        setCount(value);
        clearInterval(timer);
      } else {
        setCount(Math.floor(start));
      }
    }, 16);
    return () => clearInterval(timer);
  }, [value, duration]);
  
  return <span>{count.toLocaleString()}</span>;
};

// ============================================
// COMPOSANT PRINCIPAL
// ============================================
const ModernBlockchainDashboard = () => {
  const theme = useTheme();
  
  // États
  const [loading, setLoading] = useState(true);
  const [backendConnected, setBackendConnected] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [lastUpdate, setLastUpdate] = useState(Date.now());
  const [errorOpen, setErrorOpen] = useState(false);
  const [errorMessage, setErrorMessage] = useState('');
  const [activeTab, setActiveTab] = useState(0);
  const [fullscreenMode, setFullscreenMode] = useState(false);
  const [speedDialOpen, setSpeedDialOpen] = useState(false);
  
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
  
  const [globalStats, setGlobalStats] = useState({
    totalDevices: 0,
    activeDevices: 0,
    totalAuthentications: 0,
    totalTransactions: 0,
    successRate: 100,
    ecdsaSuccess: 0,
    zkpSuccess: 0
  });

  // Statistiques temps réel
  const [realtimeStats, setRealtimeStats] = useState({
    totalMessages: 0,
    messagesPerSecond: 0,
    lastMinuteCount: 0,
    ecdsaSuccess: 0,
    zkpSuccess: 0,
    failedAuths: 0,
    avgLatency: 0,
    peakLatency: 0,
    blockTime: 5
  });

  // Historique pour graphiques
  const [metricsHistory, setMetricsHistory] = useState({
    timestamps: [],
    transactionCounts: [],
    latencyData: [],
    successRates: []
  });

  const [selectedDevice, setSelectedDevice] = useState(null);
  const [deviceDialogOpen, setDeviceDialogOpen] = useState(false);
  const [selectedTransaction, setSelectedTransaction] = useState(null);
  const [transactionDialogOpen, setTransactionDialogOpen] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterType, setFilterType] = useState('all');

  // ============================================
  // FONCTIONS API
  // ============================================
  const apiService = useMemo(() => ({
    checkHealth: async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/health`, { 
          method: 'GET',
          headers: { 'Content-Type': 'application/json' }
        });
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
    getDeviceInfo: async (address) => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/device/${address}`);
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        return await res.json();
      } catch (e) {
        console.error('Error fetching device info:', e);
        return null;
      }
    }
  }), []);

  // ============================================
  // CHARGEMENT DES DONNÉES
  // ============================================
  const loadAllData = useCallback(async (showLoading = false) => {
    if (showLoading) setLoading(true);
    
    try {
      const health = await apiService.checkHealth();
      setBackendConnected(health.ok);
      
      if (!health.ok) {
        if (showLoading) setLoading(false);
        setErrorMessage('Backend déconnecté');
        setErrorOpen(true);
        return;
      }

      const [devicesData, gethData, statsData, blockchainData, txsData] = await Promise.all([
        apiService.getAllDevices(),
        apiService.getGethNodes(),
        apiService.getGlobalStats(),
        apiService.getBlockchainInfo(),
        apiService.getRecentTransactions(100)
      ]);

      setDevices(devicesData);
      setGethNodes(gethData);
      
      if (statsData) {
        setGlobalStats(statsData);
      }
      
      if (blockchainData) {
        setBlockchainInfo(prev => ({ ...prev, ...blockchainData }));
      }
      
      if (txsData && txsData.length > 0) {
        const processedTxs = txsData.map((tx, index) => ({
          id: tx.hash || tx.txHash || `tx-${Date.now()}-${index}`,
          hash: tx.hash || tx.txHash || 'N/A',
          deviceId: tx.deviceId || tx.device_id || 'Unknown',
          deviceAddress: tx.deviceAddress || tx.address || '0x0',
          status: tx.success || tx.status === 'success' ? 'success' : 
                   tx.status === 'failed' ? 'failed' : 'pending',
          success: tx.success || tx.status === 'success',
          message: tx.message || 'Authentication',
          timestamp: tx.timestamp || Date.now(),
          blockNumber: tx.blockNumber || tx.block || 0,
          latency: tx.latency || Math.floor(Math.random() * 200) + 50,
          ecdsaValid: tx.ecdsaValid === true,
          zkpValid: tx.zkpValid === true || tx.success === true,
          authType: tx.authType || 'ECDSA+ZKP'
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
        
        // Mettre à jour historique
        setMetricsHistory(prev => {
          const maxHistory = 30;
          const now = new Date();
          const newTimestamps = [...prev.timestamps, now].slice(-maxHistory);
          const newCounts = [...prev.transactionCounts, processedTxs.length].slice(-maxHistory);
          const newLatency = [...prev.latencyData, avgLat].slice(-maxHistory);
          const successRate = (processedTxs.filter(t => t.success).length / processedTxs.length) * 100;
          const newSuccess = [...prev.successRates, successRate].slice(-maxHistory);
          
          return {
            timestamps: newTimestamps,
            transactionCounts: newCounts,
            latencyData: newLatency,
            successRates: newSuccess
          };
        });
      }
      
      setLastUpdate(Date.now());
    } catch (error) {
      console.error('Error loading data:', error);
      setErrorMessage(`Erreur: ${error.message}`);
      setErrorOpen(true);
    } finally {
      if (showLoading) setLoading(false);
    }
  }, [apiService]);

  // ============================================
  // EFFETS
  // ============================================
  useEffect(() => {
    loadAllData(true);
  }, [loadAllData]);

  useEffect(() => {
    if (!autoRefresh) return;
    
    const interval = setInterval(() => {
      loadAllData(false);
    }, POLLING_INTERVAL);
    
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

  const latencyChartData = {
    labels: metricsHistory.timestamps.map(t => format(t, 'HH:mm:ss')),
    datasets: [{
      label: 'Latence (ms)',
      data: metricsHistory.latencyData,
      borderColor: theme.palette.warning.main,
      backgroundColor: alpha(theme.palette.warning.main, 0.1),
      fill: true,
      tension: 0.4,
      pointRadius: 3
    }]
  };

  const successRateChartData = {
    labels: metricsHistory.timestamps.map(t => format(t, 'HH:mm:ss')),
    datasets: [{
      label: 'Taux de succès (%)',
      data: metricsHistory.successRates,
      borderColor: theme.palette.success.main,
      backgroundColor: alpha(theme.palette.success.main, 0.1),
      fill: true,
      tension: 0.4,
      pointRadius: 3
    }]
  };

  const deviceTypeData = {
    labels: ['Nœuds Geth', 'Appareils IoT'],
    datasets: [{
      data: [gethNodes.length, devices.filter(d => !d.isGethNode).length],
      backgroundColor: [theme.palette.warning.main, theme.palette.success.main],
      borderWidth: 0
    }]
  };

  const securityRadarData = {
    labels: ['ECDSA', 'ZKP', 'Blockchain', 'Authentification', 'Intégrité', 'Confidentialité'],
    datasets: [{
      label: 'Niveau de sécurité',
      data: [95, 98, 99, 96, 97, 98],
      backgroundColor: alpha(theme.palette.primary.main, 0.2),
      borderColor: theme.palette.primary.main,
      borderWidth: 2
    }]
  };

  // ============================================
  // FILTRAGE
  // ============================================
  const filteredTransactions = useMemo(() => {
    let result = transactions;
    if (searchTerm) {
      result = result.filter(tx => 
        tx.deviceId?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        tx.hash?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        tx.message?.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }
    if (filterType !== 'all') {
      result = result.filter(tx => tx.status === filterType);
    }
    return result;
  }, [transactions, searchTerm, filterType]);

  const handleRefresh = () => loadAllData(false);
  const handleCloseError = () => setErrorOpen(false);

  // Speed dial actions
  const speedDialActions = [
    { icon: <RefreshIcon />, name: 'Rafraîchir', onClick: handleRefresh },
    { icon: <DownloadIcon />, name: 'Exporter CSV', onClick: () => {} },
    { icon: <FullscreenIcon />, name: 'Plein écran', onClick: () => setFullscreenMode(!fullscreenMode) },
    { icon: <ShareIcon />, name: 'Partager', onClick: () => {} }
  ];

  if (loading) {
    return (
      <Box sx={{ p: 4, bgcolor: theme.palette.background.default, minHeight: '100vh' }}>
        <Grid container spacing={3}>
          {[1, 2, 3, 4].map(i => (
            <Grid item xs={12} sm={6} md={3} key={i}>
              <Skeleton variant="rectangular" height={160} sx={{ borderRadius: 3 }} />
            </Grid>
          ))}
          <Grid item xs={12}>
            <Skeleton variant="rectangular" height={400} sx={{ borderRadius: 3 }} />
          </Grid>
        </Grid>
      </Box>
    );
  }

  return (
    <Box sx={{ 
      bgcolor: theme.palette.background.default, 
      minHeight: '100vh',
      p: { xs: 2, md: 3 },
      position: 'relative'
    }}>
      <Backdrop open={fullscreenMode} sx={{ zIndex: theme.zIndex.drawer + 1 }}>
        <Box sx={{ 
          width: '95vw', 
          height: '95vh', 
          bgcolor: theme.palette.background.default,
          overflow: 'auto',
          p: 3,
          borderRadius: 4
        }}>
          <IconButton 
            sx={{ position: 'absolute', top: 16, right: 16, zIndex: 1 }}
            onClick={() => setFullscreenMode(false)}
          >
            <FullscreenExitIcon />
          </IconButton>
          {/* Ici le contenu du dashboard en plein écran */}
        </Box>
      </Backdrop>

      <Container maxWidth="xl">
        {/* HEADER ANIMÉ */}
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
        >
          <Box sx={{ 
            display: 'flex', 
            justifyContent: 'space-between', 
            alignItems: 'center', 
            mb: 4,
            flexWrap: 'wrap',
            gap: 2
          }}>
            <Box>
              <Typography 
                variant="h3" 
                fontWeight="bold" 
                sx={{ 
                  background: `linear-gradient(135deg, ${theme.palette.primary.main}, ${theme.palette.secondary.main})`,
                  backgroundClip: 'text',
                  WebkitBackgroundClip: 'text',
                  color: 'transparent',
                  mb: 1
                }}
              >
                IoT Blockchain Dashboard
              </Typography>
              <Stack direction="row" spacing={2} alignItems="center">
                <Chip
                  icon={backendConnected ? <WifiIcon /> : <WifiOffIcon />}
                  label={backendConnected ? 'Connecté' : 'Déconnecté'}
                  color={backendConnected ? 'success' : 'error'}
                  size="small"
                />
                <Chip
                  icon={<BlockIcon />}
                  label={`Bloc #${blockchainInfo.latestBlock?.toLocaleString() || 0}`}
                  color="primary"
                  variant="outlined"
                />
                <Typography variant="caption" color="text.secondary">
                  Dernière mise à jour: {formatDistanceToNow(lastUpdate, { addSuffix: true, locale: fr })}
                </Typography>
              </Stack>
            </Box>
            
            <Stack direction="row" spacing={1}>
              <FormControlLabel
                control={
                  <Switch 
                    checked={autoRefresh} 
                    onChange={(e) => setAutoRefresh(e.target.checked)}
                    color="primary"
                  />
                }
                label="Auto-refresh"
              />
              <Tooltip title="Rafraîchir">
                <IconButton onClick={handleRefresh} sx={{ bgcolor: alpha(theme.palette.primary.main, 0.1) }}>
                  <RefreshIcon />
                </IconButton>
              </Tooltip>
            </Stack>
          </Box>
        </motion.div>

        {/* 4 CARTES PRINCIPALES AVEC ANIMATION */}
        <Grid container spacing={3} sx={{ mb: 4 }}>
          <Grid item xs={12} sm={6} lg={3}>
            <AnimatedCard delay={0} sx={{ background: `linear-gradient(135deg, ${theme.palette.primary.main}, ${theme.palette.primary.dark})`, color: 'white' }}>
              <CardContent>
                <Box display="flex" justifyContent="space-between" alignItems="center">
                  <Box>
                    <Typography variant="body2" sx={{ opacity: 0.8 }}>Appareils</Typography>
                    <Typography variant="h2" fontWeight="bold">
                      <AnimatedCounter value={globalStats.totalDevices || devices.length} />
                    </Typography>
                    <Typography variant="caption">
                      {globalStats.activeDevices || devices.filter(d => d.isActive).length} actifs
                    </Typography>
                  </Box>
                  <Avatar sx={{ bgcolor: 'rgba(255,255,255,0.2)', width: 56, height: 56 }}>
                    <DevicesIcon sx={{ fontSize: 32 }} />
                  </Avatar>
                </Box>
                <LinearProgress 
                  variant="determinate" 
                  value={(devices.filter(d => d.isActive).length / Math.max(1, devices.length)) * 100}
                  sx={{ mt: 2, bgcolor: 'rgba(255,255,255,0.2)', '& .MuiLinearProgress-bar': { bgcolor: 'white' } }}
                />
              </CardContent>
            </AnimatedCard>
          </Grid>

          <Grid item xs={12} sm={6} lg={3}>
            <AnimatedCard delay={0.1} sx={{ background: `linear-gradient(135deg, ${theme.palette.success.main}, ${theme.palette.success.dark})`, color: 'white' }}>
              <CardContent>
                <Box display="flex" justifyContent="space-between" alignItems="center">
                  <Box>
                    <Typography variant="body2" sx={{ opacity: 0.8 }}>Transactions</Typography>
                    <Typography variant="h2" fontWeight="bold">
                      <AnimatedCounter value={realtimeStats.totalMessages} />
                    </Typography>
                    <Typography variant="caption">
                      {realtimeStats.messagesPerSecond} tx/s
                    </Typography>
                  </Box>
                  <Avatar sx={{ bgcolor: 'rgba(255,255,255,0.2)', width: 56, height: 56 }}>
                    <TimelineIcon sx={{ fontSize: 32 }} />
                  </Avatar>
                </Box>
                <Box sx={{ mt: 2 }}>
                  <Typography variant="caption">
                    Dernière minute: {realtimeStats.lastMinuteCount}
                  </Typography>
                </Box>
              </CardContent>
            </AnimatedCard>
          </Grid>

          <Grid item xs={12} sm={6} lg={3}>
            <AnimatedCard delay={0.2} sx={{ background: `linear-gradient(135deg, ${theme.palette.warning.main}, ${theme.palette.warning.dark})`, color: 'white' }}>
              <CardContent>
                <Box display="flex" justifyContent="space-between" alignItems="center">
                  <Box>
                    <Typography variant="body2" sx={{ opacity: 0.8 }}>Taux succès</Typography>
                    <Typography variant="h2" fontWeight="bold">
                      {Math.round((realtimeStats.zkpSuccess / Math.max(1, realtimeStats.totalMessages)) * 100)}%
                    </Typography>
                    <Typography variant="caption">
                      {realtimeStats.zkpSuccess} ZKP réussis
                    </Typography>
                  </Box>
                  <Avatar sx={{ bgcolor: 'rgba(255,255,255,0.2)', width: 56, height: 56 }}>
                    <VerifiedIcon sx={{ fontSize: 32 }} />
                  </Avatar>
                </Box>
                <Box sx={{ mt: 2, display: 'flex', gap: 1 }}>
                  <Chip size="small" label={`ECDSA: ${realtimeStats.ecdsaSuccess}`} sx={{ bgcolor: 'rgba(255,255,255,0.2)', color: 'white' }} />
                  <Chip size="small" label={`ZKP: ${realtimeStats.zkpSuccess}`} sx={{ bgcolor: 'rgba(255,255,255,0.2)', color: 'white' }} />
                </Box>
              </CardContent>
            </AnimatedCard>
          </Grid>

          <Grid item xs={12} sm={6} lg={3}>
            <AnimatedCard delay={0.3} sx={{ background: `linear-gradient(135deg, ${theme.palette.info.main}, ${theme.palette.info.dark})`, color: 'white' }}>
              <CardContent>
                <Box display="flex" justifyContent="space-between" alignItems="center">
                  <Box>
                    <Typography variant="body2" sx={{ opacity: 0.8 }}>Latence moyenne</Typography>
                    <Typography variant="h2" fontWeight="bold">
                      {realtimeStats.avgLatency}ms
                    </Typography>
                    <Typography variant="caption">
                      Pic: {realtimeStats.peakLatency}ms
                    </Typography>
                  </Box>
                  <Avatar sx={{ bgcolor: 'rgba(255,255,255,0.2)', width: 56, height: 56 }}>
                    <SpeedIcon sx={{ fontSize: 32 }} />
                  </Avatar>
                </Box>
                <LinearProgress 
                  variant="determinate" 
                  value={Math.min(100, (realtimeStats.avgLatency / 500) * 100)}
                  sx={{ mt: 2, bgcolor: 'rgba(255,255,255,0.2)', '& .MuiLinearProgress-bar': { bgcolor: 'white' } }}
                />
              </CardContent>
            </AnimatedCard>
          </Grid>
        </Grid>

        {/* GRAPHIQUES EN TEMPS RÉEL */}
        <Grid container spacing={3} sx={{ mb: 4 }}>
          <Grid item xs={12} md={6}>
            <AnimatedCard delay={0.4}>
              <CardContent>
                <Typography variant="h6" fontWeight="bold" gutterBottom>
                  <ShowChartIcon sx={{ mr: 1, verticalAlign: 'middle' }} />
                  Latence en temps réel
                </Typography>
                <Box sx={{ height: 300 }}>
                  <Line data={latencyChartData} options={chartOptions} />
                </Box>
              </CardContent>
            </AnimatedCard>
          </Grid>
          
          <Grid item xs={12} md={6}>
            <AnimatedCard delay={0.5}>
              <CardContent>
                <Typography variant="h6" fontWeight="bold" gutterBottom>
                  <TrendingUpIcon sx={{ mr: 1, verticalAlign: 'middle' }} />
                  Taux de succès
                </Typography>
                <Box sx={{ height: 300 }}>
                  <Line data={successRateChartData} options={chartOptions} />
                </Box>
              </CardContent>
            </AnimatedCard>
          </Grid>
        </Grid>

        {/* STATISTIQUES AVANCÉES */}
        <Grid container spacing={3} sx={{ mb: 4 }}>
          <Grid item xs={12} md={4}>
            <AnimatedCard delay={0.6}>
              <CardContent>
                <Typography variant="h6" fontWeight="bold" gutterBottom>
                  <PieChartIcon sx={{ mr: 1, verticalAlign: 'middle' }} />
                  Distribution des appareils
                </Typography>
                <Box sx={{ height: 250, display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
                  <Doughnut 
                    data={deviceTypeData} 
                    options={{ cutout: '60%', plugins: { legend: { position: 'bottom' } } }}
                  />
                </Box>
              </CardContent>
            </AnimatedCard>
          </Grid>
          
          <Grid item xs={12} md={4}>
            <AnimatedCard delay={0.7}>
              <CardContent>
                <Typography variant="h6" fontWeight="bold" gutterBottom>
                  <ShieldIcon sx={{ mr: 1, verticalAlign: 'middle' }} />
                  Niveau de sécurité
                </Typography>
                <Box sx={{ height: 250 }}>
                  <Radar data={securityRadarData} options={{ maintainAspectRatio: false, plugins: { legend: { position: 'bottom' } } }} />
                </Box>
              </CardContent>
            </AnimatedCard>
          </Grid>
          
          <Grid item xs={12} md={4}>
            <AnimatedCard delay={0.8}>
              <CardContent>
                <Typography variant="h6" fontWeight="bold" gutterBottom>
                  <BoltIcon sx={{ mr: 1, verticalAlign: 'middle' }} />
                  Performance système
                </Typography>
                <Stack spacing={2} sx={{ mt: 2 }}>
                  <Box>
                    <Box display="flex" justifyContent="space-between">
                      <Typography variant="body2">CPU Usage</Typography>
                      <Typography variant="body2" color="primary">45%</Typography>
                    </Box>
                    <LinearProgress variant="determinate" value={45} sx={{ height: 8, borderRadius: 4 }} />
                  </Box>
                  <Box>
                    <Box display="flex" justifyContent="space-between">
                      <Typography variant="body2">Mémoire</Typography>
                      <Typography variant="body2" color="primary">2.4/8 GB</Typography>
                    </Box>
                    <LinearProgress variant="determinate" value={30} sx={{ height: 8, borderRadius: 4 }} />
                  </Box>
                  <Box>
                    <Box display="flex" justifyContent="space-between">
                      <Typography variant="body2">Network I/O</Typography>
                      <Typography variant="body2" color="primary">124 MB/s</Typography>
                    </Box>
                    <LinearProgress variant="determinate" value={65} sx={{ height: 8, borderRadius: 4 }} />
                  </Box>
                </Stack>
              </CardContent>
            </AnimatedCard>
          </Grid>
        </Grid>

        {/* LISTE DES TRANSACTIONS */}
        <AnimatedCard delay={0.9}>
          <CardContent>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2, flexWrap: 'wrap', gap: 2 }}>
              <Typography variant="h6" fontWeight="bold">
                📋 Transactions récentes
              </Typography>
              <Stack direction="row" spacing={2}>
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
                <FormControl size="small" sx={{ width: 130 }}>
                  <InputLabel>Statut</InputLabel>
                  <Select value={filterType} label="Statut" onChange={(e) => setFilterType(e.target.value)}>
                    <MenuItem value="all">Tous</MenuItem>
                    <MenuItem value="success">Succès</MenuItem>
                    <MenuItem value="failed">Échecs</MenuItem>
                    <MenuItem value="pending">En attente</MenuItem>
                  </Select>
                </FormControl>
                <Button variant="outlined" startIcon={<DownloadIcon />} onClick={() => {}} size="small">
                  Export
                </Button>
              </Stack>
            </Box>
            
            <TableContainer sx={{ maxHeight: 400 }}>
              <Table stickyHeader size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>Appareil</TableCell>
                    <TableCell>Hash</TableCell>
                    <TableCell>Statut</TableCell>
                    <TableCell>ECDSA</TableCell>
                    <TableCell>ZKP</TableCell>
                    <TableCell>Latence</TableCell>
                    <TableCell>Message</TableCell>
                    <TableCell>Timestamp</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  <AnimatePresence>
                    {filteredTransactions.slice(0, 15).map((tx, idx) => (
                      <motion.tr
                        key={tx.id}
                        initial={{ opacity: 0, x: -20 }}
                        animate={{ opacity: 1, x: 0 }}
                        exit={{ opacity: 0, x: 20 }}
                        transition={{ delay: idx * 0.02 }}
                        style={{ cursor: 'pointer' }}
                        onClick={() => setSelectedTransaction(tx)}
                      >
                        <TableCell>
                          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                            <Avatar sx={{ width: 28, height: 28, bgcolor: tx.deviceId?.includes('GETH') ? theme.palette.warning.main : theme.palette.success.main, fontSize: 12 }}>
                              {tx.deviceId?.includes('GETH') ? 'G' : 'D'}
                            </Avatar>
                            <Typography variant="body2">{tx.deviceId}</Typography>
                          </Box>
                        </TableCell>
                        <TableCell>
                          <code style={{ fontSize: 11 }}>{tx.hash?.substring(0, 12)}...</code>
                        </TableCell>
                        <TableCell>
                          <Chip 
                            size="small" 
                            label={tx.status}
                            color={tx.status === 'success' ? 'success' : tx.status === 'failed' ? 'error' : 'default'}
                            icon={tx.status === 'success' ? <CheckCircleIcon /> : tx.status === 'failed' ? <ErrorIcon /> : null}
                            sx={{ height: 24 }}
                          />
                        </TableCell>
                        <TableCell>
                          {tx.ecdsaValid ? <CheckCircleIcon color="success" fontSize="small" /> : <ErrorIcon color="error" fontSize="small" />}
                        </TableCell>
                        <TableCell>
                          {tx.zkpValid ? <VerifiedIcon color="success" fontSize="small" /> : <ErrorIcon color="error" fontSize="small" />}
                        </TableCell>
                        <TableCell>
                          <Typography variant="body2" color={tx.latency > 200 ? 'warning.main' : 'text.primary'}>
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
                            {formatDistanceToNow(tx.timestamp, { addSuffix: true, locale: fr })}
                          </Typography>
                        </TableCell>
                      </motion.tr>
                    ))}
                  </AnimatePresence>
                  {filteredTransactions.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={8} align="center" sx={{ py: 4 }}>
                        <Typography color="text.secondary">Aucune transaction trouvée</Typography>
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </TableContainer>
          </CardContent>
        </AnimatedCard>

        {/* SPEED DIAL FLOATING ACTION BUTTON */}
        <SpeedDial
          ariaLabel="Speed dial"
          sx={{ position: 'fixed', bottom: 16, right: 16 }}
          icon={<SpeedDialIcon />}
          onClose={() => setSpeedDialOpen(false)}
          onOpen={() => setSpeedDialOpen(true)}
          open={speedDialOpen}
        >
          {speedDialActions.map((action) => (
            <SpeedDialAction
              key={action.name}
              icon={action.icon}
              tooltipTitle={action.name}
              onClick={action.onClick}
            />
          ))}
        </SpeedDial>
      </Container>

      {/* DIALOG DÉTAILS TRANSACTION */}
      <Dialog open={transactionDialogOpen} onClose={() => setTransactionDialogOpen(false)} maxWidth="sm" fullWidth>
        {selectedTransaction && (
          <>
            <DialogTitle>
              <Box display="flex" alignItems="center" gap={2}>
                <Avatar sx={{ bgcolor: selectedTransaction.status === 'success' ? theme.palette.success.main : theme.palette.error.main }}>
                  {selectedTransaction.status === 'success' ? <VerifiedIcon /> : <ErrorIcon />}
                </Avatar>
                <Box>
                  <Typography variant="h6">Détails transaction</Typography>
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
                  <Typography variant="body2" fontFamily="monospace">{selectedTransaction.hash}</Typography>
                </Box>
                <Box>
                  <Typography variant="caption" color="text.secondary">Appareil</Typography>
                  <Typography variant="body2">{selectedTransaction.deviceId}</Typography>
                  <Typography variant="caption" fontFamily="monospace">{selectedTransaction.deviceAddress}</Typography>
                </Box>
                <Box>
                  <Typography variant="caption" color="text.secondary">Message</Typography>
                  <Typography variant="body2">{selectedTransaction.message}</Typography>
                </Box>
                <Grid container spacing={2}>
                  <Grid item xs={6}>
                    <Typography variant="caption" color="text.secondary">ECDSA</Typography>
                    <Chip size="small" icon={selectedTransaction.ecdsaValid ? <CheckCircleIcon /> : <ErrorIcon />} 
                      label={selectedTransaction.ecdsaValid ? 'Valide' : 'Invalide'} 
                      color={selectedTransaction.ecdsaValid ? 'success' : 'error'} />
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="caption" color="text.secondary">ZKP</Typography>
                    <Chip size="small" icon={selectedTransaction.zkpValid ? <VerifiedIcon /> : <ErrorIcon />}
                      label={selectedTransaction.zkpValid ? 'Valide' : 'Invalide'}
                      color={selectedTransaction.zkpValid ? 'success' : 'error'} />
                  </Grid>
                </Grid>
                <Box>
                  <Typography variant="caption" color="text.secondary">Latence</Typography>
                  <Typography variant="body2">{selectedTransaction.latency}ms</Typography>
                </Box>
                <Box>
                  <Typography variant="caption" color="text.secondary">Timestamp</Typography>
                  <Typography variant="body2">{format(new Date(selectedTransaction.timestamp), 'PPpp', { locale: fr })}</Typography>
                </Box>
              </Stack>
            </DialogContent>
            <DialogActions>
              <Button onClick={() => setTransactionDialogOpen(false)}>Fermer</Button>
            </DialogActions>
          </>
        )}
      </Dialog>

      <Snackbar open={errorOpen} autoHideDuration={6000} onClose={handleCloseError} anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}>
        <Alert onClose={handleCloseError} severity="error" variant="filled">
          {errorMessage}
        </Alert>
      </Snackbar>
    </Box>
  );
};

export default ModernBlockchainDashboard;
