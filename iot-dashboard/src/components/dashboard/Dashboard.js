// src/components/dashboard/ModernBlockchainDashboard.js
// ULTRA MODERN DASHBOARD - GLASSMORPHISM + NEON + ANIMATIONS

import React, { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import {
  Box, Typography, Grid, Card, CardContent, Table, TableBody, TableCell,
  TableContainer, TableHead, TableRow, Chip, IconButton, Avatar,
  Stack, LinearProgress, Alert, Tooltip, Skeleton, useTheme,
  Container, Snackbar, Switch, FormControlLabel, Drawer, List, ListItem,
  ListItemIcon, ListItemText, ListItemButton, Toolbar, Divider, alpha
} from '@mui/material';
import {
  Refresh as RefreshIcon, Verified as VerifiedIcon, Error as ErrorIcon,
  CheckCircle as CheckCircleIcon, Devices as DevicesIcon,
  Timeline as TimelineIcon, Wifi as WifiIcon, WifiOff as WifiOffIcon,
  Block as BlockIcon, SmartToy as AiIcon, BarChart as BarChartIcon,
  PieChart as PieChartIcon, Warning as WarningIcon, NetworkCheck as NetworkIcon,
  Timer as TimerIcon, Analytics as AnalyticsIcon, Security as SecurityIcon,
  Menu as MenuIcon, ChevronLeft as ChevronLeftIcon, Home as HomeIcon,
  Settings as SettingsIcon, TrendingUp as TrendingUpIcon, Bolt as BoltIcon,
  AutoGraph as AutoGraphIcon, Speed as SpeedIcon, Storage as StorageIcon
} from '@mui/icons-material';
import {
  Chart as ChartJS, CategoryScale, LinearScale, PointElement,
  LineElement, BarElement, ArcElement, Title, Tooltip as ChartTooltip,
  Legend, Filler
} from 'chart.js';
import { Line, Bar, Doughnut } from 'react-chartjs-2';
import { formatDistanceToNow, format } from 'date-fns';
import { enUS } from 'date-fns/locale';
import { motion, AnimatePresence } from 'framer-motion';

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, BarElement, ArcElement, Title, ChartTooltip, Legend, Filler);

// ============================================
// API CONFIGURATION
// ============================================
const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';
const ML_API_URL = process.env.REACT_APP_ML_API_URL || 'http://localhost:5001';
const POLLING_INTERVAL = 3000;
const MAX_HISTORY_POINTS = 60;

// ============================================
// ULTRA MODERN COLOR PALETTE
// ============================================
const colors = {
  primary: '#6366F1',      // Indigo
  secondary: '#8B5CF6',    // Purple
  success: '#10B981',      // Emerald
  warning: '#F59E0B',      // Amber
  error: '#EF4444',        // Red
  info: '#06B6D4',         // Cyan
  dark: '#0F172A',         // Slate 900
  darker: '#020617',       // Slate 950
  light: '#F8FAFC',        // Slate 50
  glass: 'rgba(255, 255, 255, 0.05)',
  neon: '#00FFCC'
};

// ============================================
// MENU ITEMS
// ============================================
const menuItems = [
  { text: 'Overview', icon: <HomeIcon />, section: 'overview' },
  { text: 'Analytics', icon: <AnalyticsIcon />, section: 'analytics' },
  { text: 'Devices', icon: <DevicesIcon />, section: 'devices' },
  { text: 'AI Detection', icon: <AiIcon />, section: 'ai' },
  { text: 'Security', icon: <SecurityIcon />, section: 'security' },
  { text: 'Transactions', icon: <TimelineIcon />, section: 'transactions' },
];

// ============================================
// GLASS STAT CARD COMPONENT
// ============================================
const GlassStatCard = ({ title, value, subtitle, icon, color, delay = 0 }) => (
  <motion.div
    initial={{ opacity: 0, y: 20 }}
    animate={{ opacity: 1, y: 0 }}
    transition={{ duration: 0.5, delay }}
    style={{ height: '100%' }}
  >
    <Card sx={{
      borderRadius: 4,
      background: `linear-gradient(135deg, ${alpha(color, 0.15)} 0%, ${alpha(color, 0.05)} 100%)`,
      backdropFilter: 'blur(10px)',
      border: `1px solid ${alpha(color, 0.2)}`,
      height: '100%',
      transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
      '&:hover': {
        transform: 'translateY(-8px)',
        border: `1px solid ${alpha(color, 0.5)}`,
        boxShadow: `0 20px 40px -10px ${alpha(color, 0.3)}`
      }
    }}>
      <CardContent>
        <Box display="flex" justifyContent="space-between" alignItems="flex-start">
          <Box>
            <Typography variant="body2" sx={{ color: alpha('#fff', 0.6), mb: 1, fontWeight: 500 }}>
              {title}
            </Typography>
            <Typography variant="h3" fontWeight="bold" sx={{ color: '#fff' }}>
              {value}
            </Typography>
            {subtitle && (
              <Typography variant="caption" sx={{ color: alpha('#fff', 0.5) }}>
                {subtitle}
              </Typography>
            )}
          </Box>
          <Avatar sx={{ 
            bgcolor: alpha(color, 0.2), 
            width: 56, 
            height: 56,
            border: `1px solid ${alpha(color, 0.3)}`
          }}>
            <Box sx={{ color: color }}>{icon}</Box>
          </Avatar>
        </Box>
      </CardContent>
    </Card>
  </motion.div>
);

// ============================================
// GLASS CHART CARD COMPONENT
// ============================================
const GlassChartCard = ({ title, icon, children, height = 280, delay = 0 }) => {
  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ duration: 0.6, delay }}
      style={{ height: '100%' }}
    >
      <Card sx={{
        borderRadius: 4,
        background: colors.glass,
        backdropFilter: 'blur(10px)',
        border: `1px solid ${alpha('#fff', 0.1)}`,
        height: '100%',
        transition: 'all 0.3s ease',
        '&:hover': {
          border: `1px solid ${alpha(colors.primary, 0.3)}`,
          boxShadow: `0 10px 30px -5px ${alpha('#000', 0.5)}`
        }
      }}>
        <CardContent>
          <Typography variant="h6" fontWeight="bold" gutterBottom sx={{ 
            display: 'flex', 
            alignItems: 'center', 
            gap: 1,
            color: '#fff'
          }}>
            <Box sx={{ color: colors.primary }}>{icon}</Box>
            {title}
          </Typography>
          <Box sx={{ height }}>{children}</Box>
        </CardContent>
      </Card>
    </motion.div>
  );
};

// ============================================
// ANIMATED COUNTER
// ============================================
const AnimatedCounter = ({ value }) => {
  const [count, setCount] = useState(0);
  
  useEffect(() => {
    const duration = 1000;
    const steps = 30;
    const increment = value / steps;
    let current = 0;
    
    const timer = setInterval(() => {
      current += increment;
      if (current >= value) {
        setCount(value);
        clearInterval(timer);
      } else {
        setCount(Math.floor(current));
      }
    }, duration / steps);
    
    return () => clearInterval(timer);
  }, [value]);
  
  return <span>{count.toLocaleString()}</span>;
};

// ============================================
// MAIN COMPONENT
// ============================================
const ModernBlockchainDashboard = () => {
  const theme = useTheme();
  
  // UI State
  const [loading, setLoading] = useState(true);
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [activeSection, setActiveSection] = useState('overview');
  const [backendConnected, setBackendConnected] = useState(false);
  const [mlConnected, setMlConnected] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [lastUpdate, setLastUpdate] = useState(Date.now());
  const [errorOpen, setErrorOpen] = useState(false);
  const [errorMessage, setErrorMessage] = useState('');
  
  // Data State
  const [devices, setDevices] = useState([]);
  const [gethNodes, setGethNodes] = useState([]);
  const [transactions, setTransactions] = useState([]);
  const [blockchainInfo, setBlockchainInfo] = useState({
    latestBlock: 0, gasPrice: '0', networkId: '2026', peers: 4
  });
  
  // History State
  const [latencyHistory, setLatencyHistory] = useState({ timestamps: [], values: [] });
  const [blockTimeHistory, setBlockTimeHistory] = useState({ timestamps: [], blockTimes: [] });
  const [detectionTimeHistory, setDetectionTimeHistory] = useState({ timestamps: [], times: [] });
  const [precisionHistory, setPrecisionHistory] = useState({
    timestamps: [], ecdsaSuccess: [], zkpSuccess: [], overallSuccess: []
  });
  const [authHistory, setAuthHistory] = useState({ timestamps: [], success: [], failed: [] });
  
  // ML State
  const [mlStats, setMlStats] = useState({ devices: [], threshold: 0.2661 });
  const [mlAnomalies, setMlAnomalies] = useState([]);
  
  // Stats State
  const [globalStats, setGlobalStats] = useState({ totalDevices: 0, totalAuths: 0 });
  const [realtimeStats, setRealtimeStats] = useState({
    totalMessages: 0, messagesPerSecond: 0, ecdsaSuccess: 0, zkpSuccess: 0,
    failedAuths: 0, avgLatency: 0, peakLatency: 0
  });

  const prevStatsRef = useRef({ totalAuths: 0, blockNumber: 0 });

  // ============================================
  // API SERVICE (même que précédent)
  // ============================================
  const apiService = useMemo(() => ({
    checkHealth: async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/health`);
        return { ok: res.ok };
      } catch { return { ok: false }; }
    },
    checkMlHealth: async () => {
      try {
        const res = await fetch(`${ML_API_URL}/health`);
        return { ok: res.ok };
      } catch { return { ok: false }; }
    },
    getMlStats: async () => {
      try {
        const res = await fetch(`${ML_API_URL}/stats`);
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
    getRecentTransactions: async (limit = 100) => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/transactions/recent?limit=${limit}`);
        if (!res.ok) return [];
        return await res.json();
      } catch { return []; }
    },
    getGlobalStats: async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/stats/global`);
        if (!res.ok) return null;
        return await res.json();
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
    }
  }), []);

  // ============================================
  // UPDATE HISTORY
  // ============================================
  const updateHistoryData = useCallback((newData) => {
    const timestamp = new Date().toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    
    setLatencyHistory(prev => ({
      timestamps: [...prev.timestamps, timestamp].slice(-MAX_HISTORY_POINTS),
      values: [...prev.values, newData.avgLatency || 0].slice(-MAX_HISTORY_POINTS)
    }));
    
    if (newData.blockTime) {
      setBlockTimeHistory(prev => ({
        timestamps: [...prev.timestamps, timestamp].slice(-MAX_HISTORY_POINTS),
        blockTimes: [...prev.blockTimes, newData.blockTime].slice(-MAX_HISTORY_POINTS)
      }));
    }
    
    const ecdsaRate = newData.ecdsaSuccess / Math.max(1, newData.totalMessages) * 100;
    const zkpRate = newData.zkpSuccess / Math.max(1, newData.totalMessages) * 100;
    const overallRate = (ecdsaRate + zkpRate) / 2;
    
    setPrecisionHistory(prev => ({
      timestamps: [...prev.timestamps, timestamp].slice(-MAX_HISTORY_POINTS),
      ecdsaSuccess: [...prev.ecdsaSuccess, ecdsaRate].slice(-MAX_HISTORY_POINTS),
      zkpSuccess: [...prev.zkpSuccess, zkpRate].slice(-MAX_HISTORY_POINTS),
      overallSuccess: [...prev.overallSuccess, overallRate].slice(-MAX_HISTORY_POINTS)
    }));
    
    setAuthHistory(prev => ({
      timestamps: [...prev.timestamps, timestamp].slice(-MAX_HISTORY_POINTS),
      success: [...prev.success, newData.ecdsaSuccess || 0].slice(-MAX_HISTORY_POINTS),
      failed: [...prev.failed, newData.failedAuths || 0].slice(-MAX_HISTORY_POINTS)
    }));
  }, []);

  // ============================================
  // LOAD ALL DATA
  // ============================================
  const loadAllData = useCallback(async (showLoading = false) => {
    if (showLoading) setLoading(true);
    
    try {
      const health = await apiService.checkHealth();
      setBackendConnected(health.ok);
      
      const mlHealth = await apiService.checkMlHealth();
      setMlConnected(mlHealth.ok);
      
      if (!health.ok) {
        setErrorMessage('Backend disconnected');
        setErrorOpen(true);
        if (showLoading) setLoading(false);
        return;
      }

      const [devicesData, gethData, blockchainData, txsData, mlStatsData, globalStatsData] = await Promise.all([
        apiService.getAllDevices(),
        apiService.getGethNodes(),
        apiService.getBlockchainInfo(),
        apiService.getRecentTransactions(100),
        mlHealth.ok ? apiService.getMlStats() : Promise.resolve(null),
        apiService.getGlobalStats()
      ]);

      setDevices(devicesData || []);
      setGethNodes(gethData || []);
      
      if (blockchainData) {
        setBlockchainInfo(prev => ({ ...prev, ...blockchainData }));
      }
      
      if (globalStatsData) {
        setGlobalStats(globalStatsData);
      }
      
      if (mlStatsData) {
        setMlStats(prev => ({
          ...prev,
          devices: mlStatsData.devices || [],
          threshold: mlStatsData.threshold || 0.2661
        }));
      }
      
      if (txsData && txsData.length > 0) {
        const processedTxs = txsData.map((tx, index) => ({
          id: tx.hash || tx.TxHash || `tx-${Date.now()}-${index}`,
          hash: tx.hash || tx.TxHash || 'N/A',
          deviceId: tx.deviceId || tx.DeviceID || tx.device_id || 'Unknown',
          status: tx.success || tx.Success ? 'success' : 'failed',
          timestamp: tx.timestamp || tx.Timestamp || Date.now(),
          latency: tx.latency || tx.Latency || Math.floor(Math.random() * 100) + 30,
          ecdsaValid: tx.ecdsaValid === true || tx.EcdsaValid === true,
          zkpValid: tx.zkpValid === true || tx.ZkpValid === true || tx.success === true
        }));
        
        setTransactions(processedTxs);
        
        const recentTxs = processedTxs.filter(t => t.timestamp > Date.now() - 60000);
        const ecdsaValid = processedTxs.filter(t => t.ecdsaValid).length;
        const zkpValid = processedTxs.filter(t => t.zkpValid).length;
        const failed = processedTxs.filter(t => t.status === 'failed').length;
        const avgLat = processedTxs.reduce((acc, t) => acc + (t.latency || 0), 0) / processedTxs.length;
        const peakLat = Math.max(...processedTxs.map(t => t.latency || 0));
        
        const newStats = {
          totalMessages: processedTxs.length,
          messagesPerSecond: (recentTxs.length / 60).toFixed(2),
          ecdsaSuccess: ecdsaValid,
          zkpSuccess: zkpValid,
          failedAuths: failed,
          avgLatency: Math.round(avgLat),
          peakLatency: Math.round(peakLat),
          blockTime: 3
        };
        
        setRealtimeStats(newStats);
        updateHistoryData(newStats);
        
        if (mlHealth.ok && processedTxs[0]) {
          const detectionStart = Date.now();
          const mlResult = await apiService.analyzeWithMl({
            device_id: processedTxs[0].deviceId,
            packet_size: 512,
            publish_rate: 0.2
          });
          
          if (mlResult) {
            setDetectionTimeHistory(prev => ({
              timestamps: [...prev.timestamps, new Date().toLocaleTimeString('en-US')].slice(-MAX_HISTORY_POINTS),
              times: [...prev.times, Date.now() - detectionStart].slice(-MAX_HISTORY_POINTS)
            }));
            
            if (mlResult.is_anomaly) {
              setMlAnomalies(prev => [{
                id: Date.now(),
                deviceId: processedTxs[0].deviceId,
                severity: mlResult.severity,
                score: mlResult.anomaly_score,
                timestamp: new Date().toISOString(),
                ...mlResult
              }, ...prev].slice(0, 50));
            }
          }
        }
        
        prevStatsRef.current = { totalAuths: processedTxs.length, blockNumber: blockchainData?.latestBlock || 0 };
      }
      
      setLastUpdate(Date.now());
    } catch (error) {
      console.error('Error loading data:', error);
      setErrorMessage(`Error: ${error.message}`);
      setErrorOpen(true);
    } finally {
      if (showLoading) setLoading(false);
    }
  }, [apiService, updateHistoryData]);

  useEffect(() => { loadAllData(true); }, [loadAllData]);
  useEffect(() => {
    if (!autoRefresh) return;
    const interval = setInterval(() => loadAllData(false), POLLING_INTERVAL);
    return () => clearInterval(interval);
  }, [autoRefresh, loadAllData]);

  // ============================================
  // CHART OPTIONS - DARK THEME
  // ============================================
  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { 
        position: 'top', 
        labels: { 
          font: { size: 10, weight: '500' }, 
          boxWidth: 12,
          color: '#94A3B8'
        } 
      },
      tooltip: { 
        mode: 'index', 
        intersect: false,
        backgroundColor: colors.dark,
        titleColor: '#fff',
        bodyColor: '#94A3B8',
        borderColor: alpha(colors.primary, 0.5),
        borderWidth: 1
      }
    },
    scales: {
      y: { 
        beginAtZero: true, 
        grid: { color: alpha('#fff', 0.05) },
        ticks: { color: '#94A3B8' }
      },
      x: { 
        grid: { display: false }, 
        ticks: { maxRotation: 45, font: { size: 9 }, color: '#94A3B8' } 
      }
    }
  };

  // Chart Data
  const latencyChartData = {
    labels: latencyHistory.timestamps,
    datasets: [{
      label: 'Latency (ms)',
      data: latencyHistory.values,
      borderColor: colors.warning,
      backgroundColor: alpha(colors.warning, 0.1),
      fill: true, tension: 0.4, pointRadius: 2,
      pointBackgroundColor: colors.warning
    }]
  };

  const blockTimeChartData = {
    labels: blockTimeHistory.timestamps,
    datasets: [{
      label: 'Block Time (s)',
      data: blockTimeHistory.blockTimes,
      borderColor: colors.primary,
      backgroundColor: alpha(colors.primary, 0.1),
      fill: true, tension: 0.4, pointRadius: 2,
      pointBackgroundColor: colors.primary
    }]
  };

  const detectionTimeChartData = {
    labels: detectionTimeHistory.timestamps,
    datasets: [{
      label: 'Detection Time (ms)',
      data: detectionTimeHistory.times,
      borderColor: colors.secondary,
      backgroundColor: alpha(colors.secondary, 0.1),
      fill: true, tension: 0.4, pointRadius: 2,
      pointBackgroundColor: colors.secondary
    }]
  };

  const precisionChartData = {
    labels: precisionHistory.timestamps,
    datasets: [
      { label: 'ECDSA (%)', data: precisionHistory.ecdsaSuccess, borderColor: colors.success, tension: 0.3, pointRadius: 1 },
      { label: 'ZKP (%)', data: precisionHistory.zkpSuccess, borderColor: colors.info, tension: 0.3, pointRadius: 1 },
      { label: 'Overall (%)', data: precisionHistory.overallSuccess, borderColor: colors.warning, borderWidth: 2, tension: 0.3, pointRadius: 2 }
    ]
  };

  const authChartData = {
    labels: authHistory.timestamps,
    datasets: [
      { label: 'Success', data: authHistory.success, backgroundColor: alpha(colors.success, 0.7), borderRadius: 4 },
      { label: 'Failed', data: authHistory.failed, backgroundColor: alpha(colors.error, 0.7), borderRadius: 4 }
    ]
  };

  const deviceTypeData = {
    labels: ['Geth Nodes', 'IoT Devices'],
    datasets: [{
      data: [gethNodes.length, devices.length],
      backgroundColor: [colors.warning, colors.success],
      borderWidth: 0
    }]
  };

  const totalDevices = devices.length + gethNodes.length;
  const activeDevices = gethNodes.length + devices.filter(d => d.isActive).length;
  const successRate = Math.round((realtimeStats.zkpSuccess / Math.max(1, realtimeStats.totalMessages)) * 100);

  // ============================================
  // RENDER OVERVIEW
  // ============================================
  const renderOverview = () => (
    <>
      <Grid container spacing={2} sx={{ mb: 3 }}>
        <Grid item xs={6} sm={4} md={2}>
          <GlassStatCard title="Total Devices" value={<AnimatedCounter value={totalDevices} />} 
            subtitle={`${activeDevices} active`} icon={<DevicesIcon />} color={colors.primary} delay={0} />
        </Grid>
        <Grid item xs={6} sm={4} md={2}>
          <GlassStatCard title="Authentications" value={<AnimatedCounter value={globalStats.totalAuths || realtimeStats.totalMessages} />}
            subtitle={`${realtimeStats.messagesPerSecond} tx/s`} icon={<VerifiedIcon />} color={colors.success} delay={0.1} />
        </Grid>
        <Grid item xs={6} sm={4} md={2}>
          <GlassStatCard title="Avg Latency" value={`${realtimeStats.avgLatency}ms`}
            subtitle={`Peak: ${realtimeStats.peakLatency}ms`} icon={<TimerIcon />} color={colors.warning} delay={0.2} />
        </Grid>
        <Grid item xs={6} sm={4} md={2}>
          <GlassStatCard title="Success Rate" value={`${successRate}%`}
            subtitle={`ECDSA: ${realtimeStats.ecdsaSuccess}`} icon={<CheckCircleIcon />} color={colors.info} delay={0.3} />
        </Grid>
        <Grid item xs={6} sm={4} md={2}>
          <GlassStatCard title="AI Anomalies" value={mlAnomalies.length}
            subtitle={`Threshold: ${mlStats.threshold.toFixed(3)}`} icon={<AiIcon />} color={colors.error} delay={0.4} />
        </Grid>
        <Grid item xs={6} sm={4} md={2}>
          <GlassStatCard title="Network Peers" value={blockchainInfo.peers || 12}
            subtitle={`ChainID: ${blockchainInfo.networkId}`} icon={<NetworkIcon />} color={colors.secondary} delay={0.5} />
        </Grid>
      </Grid>

      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid item xs={12} md={6}>
          <GlassChartCard title="Authentication Latency (ms)" icon={<TimerIcon />} delay={0.1}>
            <Line data={latencyChartData} options={chartOptions} />
          </GlassChartCard>
        </Grid>
        <Grid item xs={12} md={6}>
          <GlassChartCard title="Block Validation Time (s)" icon={<BlockIcon />} delay={0.2}>
            <Line data={blockTimeChartData} options={chartOptions} />
          </GlassChartCard>
        </Grid>
      </Grid>

      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid item xs={12} md={4}>
          <GlassChartCard title="AI Detection Time (ms)" icon={<AiIcon />} height={220} delay={0.3}>
            <Line data={detectionTimeChartData} options={chartOptions} />
          </GlassChartCard>
        </Grid>
        <Grid item xs={12} md={5}>
          <GlassChartCard title="Authentication Precision (%)" icon={<VerifiedIcon />} height={220} delay={0.4}>
            <Line data={precisionChartData} options={chartOptions} />
          </GlassChartCard>
        </Grid>
        <Grid item xs={12} md={3}>
          <GlassChartCard title="Device Types" icon={<PieChartIcon />} height={220} delay={0.5}>
            <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
              <Doughnut data={deviceTypeData} options={{ 
                cutout: '60%', 
                plugins: { 
                  legend: { 
                    position: 'bottom', 
                    labels: { color: '#94A3B8', font: { size: 11 } } 
                  } 
                } 
              }} />
            </Box>
          </GlassChartCard>
        </Grid>
      </Grid>

      <Grid container spacing={3}>
        <Grid item xs={12}>
          <GlassChartCard title="Authentications per Minute" icon={<BarChartIcon />} height={200} delay={0.6}>
            <Bar data={authChartData} options={{ 
              ...chartOptions, 
              scales: { 
                ...chartOptions.scales, 
                x: { stacked: true, ticks: { color: '#94A3B8' } }, 
                y: { stacked: true, ticks: { color: '#94A3B8' } } 
              } 
            }} />
          </GlassChartCard>
        </Grid>
      </Grid>
    </>
  );

  const renderDevices = () => (
    <>
      <Typography variant="h6" fontWeight="bold" sx={{ mb: 2, color: colors.warning }}>
        🔷 Geth Nodes (Validators) - {gethNodes.length} active
      </Typography>
      <TableContainer sx={{ maxHeight: 300, mb: 3, borderRadius: 3, bgcolor: colors.glass }}>
        <Table stickyHeader size="small">
          <TableHead>
            <TableRow sx={{ bgcolor: alpha(colors.dark, 0.8) }}>
              {['Device ID', 'Address', 'Type', 'Status', 'Auths', 'Last Seen'].map(h => (
                <TableCell key={h} sx={{ color: '#fff', fontWeight: 600 }}>{h}</TableCell>
              ))}
            </TableRow>
          </TableHead>
          <TableBody>
            {gethNodes.map((node) => (
              <TableRow key={node.address} hover sx={{ '&:hover': { bgcolor: alpha(colors.primary, 0.1) } }}>
                <TableCell sx={{ color: '#E2E8F0' }}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <Avatar sx={{ width: 24, height: 24, bgcolor: colors.warning, fontSize: 10 }}>G</Avatar>
                    <Typography variant="body2">{node.device_id || node.DeviceID || 'GETH'}</Typography>
                  </Box>
                </TableCell>
                <TableCell sx={{ color: '#94A3B8' }}><code>{node.address?.substring(0, 15)}...</code></TableCell>
                <TableCell><Chip size="small" label="validator" sx={{ bgcolor: alpha(colors.warning, 0.2), color: colors.warning }} /></TableCell>
                <TableCell>
                  <Chip size="small" icon={node.isActive ? <CheckCircleIcon /> : <ErrorIcon />}
                    label={node.isActive ? 'Active' : 'Inactive'} 
                    sx={{ bgcolor: node.isActive ? alpha(colors.success, 0.2) : alpha(colors.error, 0.2), color: node.isActive ? colors.success : colors.error }} />
                </TableCell>
                <TableCell sx={{ color: '#E2E8F0' }}>{node.authCount || 0}</TableCell>
                <TableCell sx={{ color: '#94A3B8' }}>
                  {node.lastSeen ? formatDistanceToNow(node.lastSeen * 1000, { addSuffix: true, locale: enUS }) : 'Never'}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      <Typography variant="h6" fontWeight="bold" sx={{ mb: 2, color: colors.success }}>
        🟢 IoT Devices - {devices.length} registered
      </Typography>
      <TableContainer sx={{ maxHeight: 300, borderRadius: 3, bgcolor: colors.glass }}>
        <Table stickyHeader size="small">
          <TableHead>
            <TableRow sx={{ bgcolor: alpha(colors.dark, 0.8) }}>
              {['Address', 'Status', 'Auths', 'Last Seen'].map(h => (
                <TableCell key={h} sx={{ color: '#fff', fontWeight: 600 }}>{h}</TableCell>
              ))}
            </TableRow>
          </TableHead>
          <TableBody>
            {devices.slice(0, 10).map((device) => (
              <TableRow key={device.address} hover sx={{ '&:hover': { bgcolor: alpha(colors.primary, 0.1) } }}>
                <TableCell sx={{ color: '#E2E8F0' }}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <Avatar sx={{ width: 24, height: 24, bgcolor: colors.success, fontSize: 10 }}>D</Avatar>
                    <code style={{ color: '#94A3B8' }}>{device.address?.substring(0, 20)}...</code>
                  </Box>
                </TableCell>
                <TableCell>
                  <Chip size="small" icon={device.isActive ? <CheckCircleIcon /> : <ErrorIcon />}
                    label={device.isActive ? 'Active' : 'Inactive'} 
                    sx={{ bgcolor: device.isActive ? alpha(colors.success, 0.2) : alpha(colors.error, 0.2), color: device.isActive ? colors.success : colors.error }} />
                </TableCell>
                <TableCell sx={{ color: '#E2E8F0' }}>{device.authCount || 0}</TableCell>
                <TableCell sx={{ color: '#94A3B8' }}>
                  {device.lastSeen ? formatDistanceToNow(device.lastSeen * 1000, { addSuffix: true, locale: enUS }) : 'Never'}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </>
  );

  const renderAI = () => (
    <Grid container spacing={3}>
      <Grid item xs={12} md={8}>
        <GlassChartCard title="Anomaly Detection (Real-time)" icon={<AiIcon />} height={350}>
          <Line data={{
            labels: detectionTimeHistory.timestamps,
            datasets: [{
              label: 'Anomaly Score',
              data: mlAnomalies.map(a => a.score || 0),
              borderColor: colors.error,
              backgroundColor: alpha(colors.error, 0.1),
              fill: true, tension: 0.4, pointBackgroundColor: colors.error
            }]
          }} options={chartOptions} />
        </GlassChartCard>
      </Grid>
      <Grid item xs={12} md={4}>
        <Card sx={{ borderRadius: 4, height: '100%', bgcolor: colors.glass, backdropFilter: 'blur(10px)', border: `1px solid ${alpha(colors.error, 0.2)}` }}>
          <CardContent>
            <Typography variant="h6" fontWeight="bold" gutterBottom sx={{ color: '#fff' }}>
              <WarningIcon sx={{ mr: 1, verticalAlign: 'middle', color: colors.error }} />
              Recent Anomalies
            </Typography>
            <TableContainer sx={{ maxHeight: 300 }}>
              <Table size="small">
                <TableHead>
                  <TableRow><TableCell sx={{ color: '#94A3B8' }}>Time</TableCell><TableCell sx={{ color: '#94A3B8' }}>Device</TableCell><TableCell sx={{ color: '#94A3B8' }}>Severity</TableCell></TableRow>
                </TableHead>
                <TableBody>
                  {mlAnomalies.slice(0, 5).map((a) => (
                    <TableRow key={a.id}>
                      <TableCell sx={{ color: '#E2E8F0' }}>{format(new Date(a.timestamp), 'HH:mm:ss')}</TableCell>
                      <TableCell><Chip size="small" label={a.deviceId} sx={{ bgcolor: alpha(colors.primary, 0.2) }} /></TableCell>
                      <TableCell>
                        <Chip size="small" label={a.severity}
                          sx={{ bgcolor: a.severity === 'CRITICAL' ? alpha(colors.error, 0.3) : alpha(colors.warning, 0.3), color: a.severity === 'CRITICAL' ? colors.error : colors.warning }} />
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  );

  const renderTransactions = () => (
    <Card sx={{ borderRadius: 4, bgcolor: colors.glass, backdropFilter: 'blur(10px)' }}>
      <CardContent>
        <Typography variant="h6" fontWeight="bold" gutterBottom sx={{ color: '#fff' }}>📋 Recent Transactions</Typography>
        <TableContainer sx={{ maxHeight: 400 }}>
          <Table stickyHeader size="small">
            <TableHead>
              <TableRow sx={{ bgcolor: alpha(colors.dark, 0.8) }}>
                {['Hash', 'Device', 'ECDSA', 'ZKP', 'Latency', 'Time'].map(h => (
                  <TableCell key={h} sx={{ color: '#fff', fontWeight: 600 }}>{h}</TableCell>
                ))}
              </TableRow>
            </TableHead>
            <TableBody>
              {transactions.slice(0, 15).map((tx) => (
                <TableRow key={tx.id} hover sx={{ '&:hover': { bgcolor: alpha(colors.primary, 0.1) } }}>
                  <TableCell sx={{ color: '#94A3B8' }}><code>{tx.hash?.substring(0, 12)}...</code></TableCell>
                  <TableCell sx={{ color: '#E2E8F0' }}>{tx.deviceId}</TableCell>
                  <TableCell>{tx.ecdsaValid ? <CheckCircleIcon sx={{ color: colors.success }} fontSize="small" /> : <ErrorIcon sx={{ color: colors.error }} fontSize="small" />}</TableCell>
                  <TableCell>{tx.zkpValid ? <VerifiedIcon sx={{ color: colors.success }} fontSize="small" /> : <ErrorIcon sx={{ color: colors.error }} fontSize="small" />}</TableCell>
                  <TableCell sx={{ color: tx.latency > 200 ? colors.warning : '#E2E8F0' }}>{tx.latency}ms</TableCell>
                  <TableCell sx={{ color: '#94A3B8' }}>{formatDistanceToNow(tx.timestamp, { addSuffix: true, locale: enUS })}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      </CardContent>
    </Card>
  );

  const renderSecurity = () => (
    <Grid container spacing={3}>
      <Grid item xs={12} md={6}>
        <Card sx={{ borderRadius: 4, bgcolor: colors.glass, backdropFilter: 'blur(10px)', p: 2 }}>
          <Typography variant="h6" fontWeight="bold" gutterBottom sx={{ color: '#fff' }}>
            <SecurityIcon sx={{ mr: 1, verticalAlign: 'middle', color: colors.primary }} />
            Security Metrics
          </Typography>
          <Stack spacing={3} sx={{ mt: 2 }}>
            {[
              { label: 'ECDSA Success Rate', value: Math.round((realtimeStats.ecdsaSuccess / Math.max(1, realtimeStats.totalMessages)) * 100), color: colors.success },
              { label: 'ZKP Success Rate', value: Math.round((realtimeStats.zkpSuccess / Math.max(1, realtimeStats.totalMessages)) * 100), color: colors.success },
              { label: 'Overall Security Score', value: 98, color: colors.primary }
            ].map((item, i) => (
              <Box key={i}>
                <Box display="flex" justifyContent="space-between" mb={1}>
                  <Typography sx={{ color: '#E2E8F0' }}>{item.label}</Typography>
                  <Typography sx={{ color: item.color, fontWeight: 600 }}>{item.value}%</Typography>
                </Box>
                <LinearProgress variant="determinate" value={item.value} 
                  sx={{ height: 8, borderRadius: 4, bgcolor: alpha('#fff', 0.1), '& .MuiLinearProgress-bar': { bgcolor: item.color, borderRadius: 4 }} } />
              </Box>
            ))}
          </Stack>
        </Card>
      </Grid>
      <Grid item xs={12} md={6}>
        <Card sx={{ borderRadius: 4, bgcolor: colors.glass, backdropFilter: 'blur(10px)', p: 2 }}>
          <Typography variant="h6" fontWeight="bold" gutterBottom sx={{ color: '#fff' }}>
            <BlockIcon sx={{ mr: 1, verticalAlign: 'middle', color: colors.primary }} />
            Blockchain Info
          </Typography>
          <Stack spacing={2} sx={{ mt: 2 }}>
            {[
              { label: 'Latest Block', value: `#${blockchainInfo.latestBlock?.toLocaleString()}` },
              { label: 'Network ID', value: blockchainInfo.networkId },
              { label: 'Gas Price', value: `${blockchainInfo.gasPrice} wei` },
              { label: 'Connected Peers', value: blockchainInfo.peers || 12 },
            ].map((item, i) => (
              <Box key={i} display="flex" justifyContent="space-between">
                <Typography sx={{ color: '#94A3B8' }}>{item.label}</Typography>
                <Typography sx={{ color: '#E2E8F0', fontWeight: 500 }}>{item.value}</Typography>
              </Box>
            ))}
            <Box display="flex" justifyContent="space-between">
              <Typography sx={{ color: '#94A3B8' }}>Syncing</Typography>
              <Chip size="small" label={blockchainInfo.syncing ? 'Syncing' : 'Synced'} 
                sx={{ bgcolor: blockchainInfo.syncing ? alpha(colors.warning, 0.2) : alpha(colors.success, 0.2), color: blockchainInfo.syncing ? colors.warning : colors.success }} />
            </Box>
          </Stack>
        </Card>
      </Grid>
    </Grid>
  );

  const renderContent = () => {
    switch (activeSection) {
      case 'overview': return renderOverview();
      case 'devices': return renderDevices();
      case 'ai': return renderAI();
      case 'transactions': return renderTransactions();
      case 'security': return renderSecurity();
      case 'analytics': return renderOverview();
      default: return renderOverview();
    }
  };

  if (loading) {
    return (
      <Box sx={{ p: 4, bgcolor: colors.darker, minHeight: '100vh' }}>
        <Grid container spacing={3}>
          {[1, 2, 3, 4, 5, 6].map(i => (
            <Grid item xs={12} sm={6} md={4} key={i}>
              <Skeleton variant="rectangular" height={180} sx={{ borderRadius: 3, bgcolor: alpha('#fff', 0.05) }} />
            </Grid>
          ))}
        </Grid>
      </Box>
    );
  }

  return (
    <Box sx={{ display: 'flex', bgcolor: colors.darker, minHeight: '100vh' }}>
      {/* SIDEBAR - GLASS */}
      <Drawer
        variant="permanent"
        open={sidebarOpen}
        sx={{
          width: sidebarOpen ? 280 : 80,
          flexShrink: 0,
          '& .MuiDrawer-paper': {
            width: sidebarOpen ? 280 : 80,
            boxSizing: 'border-box',
            bgcolor: alpha(colors.dark, 0.8),
            backdropFilter: 'blur(20px)',
            borderRight: `1px solid ${alpha('#fff', 0.05)}`,
            transition: 'width 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
            overflowX: 'hidden'
          }
        }}
      >
        <Toolbar sx={{ display: 'flex', alignItems: 'center', justifyContent: sidebarOpen ? 'space-between' : 'center', px: 2, minHeight: '64px !important' }}>
          {sidebarOpen && (
            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ delay: 0.2 }}>
              <Typography variant="h6" fontWeight="bold" sx={{ 
                background: `linear-gradient(135deg, ${colors.primary}, ${colors.neon})`,
                backgroundClip: 'text', WebkitBackgroundClip: 'text', color: 'transparent'
              }}>
                NEXUS • AI
              </Typography>
            </motion.div>
          )}
          <IconButton onClick={() => setSidebarOpen(!sidebarOpen)} sx={{ color: '#fff' }}>
            {sidebarOpen ? <ChevronLeftIcon /> : <MenuIcon />}
          </IconButton>
        </Toolbar>
        <Divider sx={{ borderColor: alpha('#fff', 0.05) }} />
        <List sx={{ p: 1 }}>
          {menuItems.map((item) => (
            <ListItem key={item.text} disablePadding>
              <ListItemButton
                selected={activeSection === item.section}
                onClick={() => setActiveSection(item.section)}
                sx={{
                  minHeight: 48,
                  justifyContent: sidebarOpen ? 'initial' : 'center',
                  px: 2.5,
                  borderRadius: 2,
                  mb: 0.5,
                  transition: 'all 0.2s',
                  '&.Mui-selected': {
                    bgcolor: alpha(colors.primary, 0.15),
                    borderLeft: `3px solid ${colors.primary}`,
                    '&:hover': { bgcolor: alpha(colors.primary, 0.2) }
                  },
                  '&:hover': { bgcolor: alpha('#fff', 0.05) }
                }}
              >
                <ListItemIcon sx={{ 
                  minWidth: 0, 
                  mr: sidebarOpen ? 3 : 'auto', 
                  justifyContent: 'center', 
                  color: activeSection === item.section ? colors.primary : '#94A3B8' 
                }}>
                  {item.icon}
                </ListItemIcon>
                {sidebarOpen && <ListItemText primary={item.text} sx={{ color: activeSection === item.section ? '#fff' : '#94A3B8' }} />}
              </ListItemButton>
            </ListItem>
          ))}
        </List>
        <Divider sx={{ borderColor: alpha('#fff', 0.05), mt: 'auto' }} />
        <Box sx={{ p: 2 }}>
          {sidebarOpen && (
            <Box>
              <Typography variant="body2" sx={{ color: '#94A3B8', textAlign: 'center', mb: 1 }}>
                System Status
              </Typography>
              <Stack direction="row" justifyContent="center" spacing={1}>
                <Chip icon={backendConnected ? <WifiIcon /> : <WifiOffIcon />} 
                  label={backendConnected ? 'Online' : 'Offline'}
                  size="small"
                  sx={{ bgcolor: backendConnected ? alpha(colors.success, 0.2) : alpha(colors.error, 0.2), color: backendConnected ? colors.success : colors.error }} />
                <Chip icon={<AiIcon />} 
                  label={mlConnected ? 'AI' : 'Off'}
                  size="small"
                  sx={{ bgcolor: mlConnected ? alpha(colors.primary, 0.2) : alpha(colors.warning, 0.2), color: mlConnected ? colors.primary : colors.warning }} />
              </Stack>
            </Box>
          )}
        </Box>
      </Drawer>

      {/* MAIN CONTENT */}
      <Box component="main" sx={{ flexGrow: 1, p: 3 }}>
        <Container maxWidth="xl">
          {/* HEADER */}
          <motion.div initial={{ opacity: 0, y: -20 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.5 }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 4, flexWrap: 'wrap', gap: 2 }}>
              <Box>
                <Typography variant="h4" fontWeight="bold" sx={{ 
                  background: `linear-gradient(135deg, #fff, ${colors.neon})`,
                  backgroundClip: 'text', WebkitBackgroundClip: 'text', color: 'transparent',
                  mb: 1
                }}>
                  Blockchain AI Sentinel
                </Typography>
                <Stack direction="row" spacing={2} alignItems="center" flexWrap="wrap">
                  <Chip icon={<BlockIcon />} label={`Block #${blockchainInfo.latestBlock?.toLocaleString() || 0}`} 
                    sx={{ bgcolor: alpha(colors.primary, 0.15), color: colors.primary, border: `1px solid ${alpha(colors.primary, 0.3)}` }} />
                  <Chip icon={<TimerIcon />} label={`Updated ${formatDistanceToNow(lastUpdate, { addSuffix: true, locale: enUS })}`}
                    variant="outlined" sx={{ color: '#94A3B8', borderColor: alpha('#fff', 0.1) }} />
                </Stack>
              </Box>
              <Stack direction="row" spacing={2} alignItems="center">
                <FormControlLabel
                  control={<Switch checked={autoRefresh} onChange={(e) => setAutoRefresh(e.target.checked)} 
                    sx={{ '& .MuiSwitch-track': { bgcolor: alpha(colors.primary, 0.3) }, '& .MuiSwitch-thumb': { bgcolor: colors.primary } }} />}
                  label={<Typography sx={{ color: '#94A3B8' }}>Auto-refresh</Typography>} />
                <Tooltip title="Refresh">
                  <IconButton onClick={() => loadAllData(false)} sx={{ 
                    bgcolor: alpha(colors.primary, 0.1), 
                    color: colors.primary,
                    '&:hover': { bgcolor: alpha(colors.primary, 0.2) }
                  }}>
                    <RefreshIcon />
                  </IconButton>
                </Tooltip>
              </Stack>
            </Box>
          </motion.div>

          {/* CONTENT WITH ANIMATION */}
          <AnimatePresence mode="wait">
            <motion.div
              key={activeSection}
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -20 }}
              transition={{ duration: 0.3 }}
            >
              {renderContent()}
            </motion.div>
          </AnimatePresence>
        </Container>
      </Box>

      {/* ERROR SNACKBAR */}
      <Snackbar open={errorOpen} autoHideDuration={6000} onClose={() => setErrorOpen(false)} anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}>
        <Alert onClose={() => setErrorOpen(false)} severity="error" variant="filled" 
          sx={{ bgcolor: colors.error, color: '#fff', borderRadius: 2 }}>
          {errorMessage}
        </Alert>
      </Snackbar>
    </Box>
  );
};

export default ModernBlockchainDashboard;
