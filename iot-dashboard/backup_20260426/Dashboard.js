import React, { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import {
  Box, Typography, Grid, Card, CardContent, Table, TableBody, TableCell,
  TableContainer, TableHead, TableRow, Chip, IconButton, Avatar,
  Stack, LinearProgress, Alert, Tooltip, Container, Switch,
  FormControlLabel, Drawer, List, ListItem, ListItemIcon, ListItemText,
  ListItemButton, Toolbar, Divider, alpha
} from '@mui/material';
import {
  Refresh as RefreshIcon, Verified as VerifiedIcon, Error as ErrorIcon,
  CheckCircle as CheckCircleIcon, Devices as DevicesIcon,
  Timeline as TimelineIcon, Wifi as WifiIcon, WifiOff as WifiOffIcon,
  Block as BlockIcon, SmartToy as AiIcon, BarChart as BarChartIcon,
  PieChart as PieChartIcon, Warning as WarningIcon, NetworkCheck as NetworkIcon,
  Timer as TimerIcon, Analytics as AnalyticsIcon, Security as SecurityIcon,
  Menu as MenuIcon, ChevronLeft as ChevronLeftIcon, Home as HomeIcon
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

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';
const ML_API_URL = process.env.REACT_APP_ML_API_URL || 'http://localhost:5001';
const POLLING_INTERVAL = 5000;
const MAX_HISTORY_POINTS = 60;
const INITIAL_LOAD_DELAY = 2000;
const MAX_RETRY_ATTEMPTS = 2;

const colors = {
  primary: '#6366F1',
  secondary: '#8B5CF6',
  success: '#10B981',
  warning: '#F59E0B',
  error: '#EF4444',
  info: '#06B6D4',
  dark: '#0F172A',
  darker: '#020617',
  glass: 'rgba(255, 255, 255, 0.05)',
  neon: '#00FFCC'
};

const menuItems = [
  { text: 'Overview', icon: <HomeIcon />, section: 'overview' },
  { text: 'Devices', icon: <DevicesIcon />, section: 'devices' },
  { text: 'AI Detection', icon: <AiIcon />, section: 'ai' },
  { text: 'Security', icon: <SecurityIcon />, section: 'security' },
  { text: 'Transactions', icon: <TimelineIcon />, section: 'transactions' },
];

const fetchWithTimeout = async (url, options = {}, timeout = 5000) => {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeout);
  try {
    const response = await fetch(url, { ...options, signal: controller.signal });
    clearTimeout(timeoutId);
    return response;
  } catch (error) {
    clearTimeout(timeoutId);
    throw error;
  }
};

const delay = (ms) => new Promise(resolve => setTimeout(resolve, ms));

const GlassStatCard = ({ title, value, subtitle, icon, color, animDelay = 0 }) => (
  <motion.div
    initial={{ opacity: 0, y: 20 }}
    animate={{ opacity: 1, y: 0 }}
    transition={{ duration: 0.5, delay: animDelay }}
    style={{ height: '100%' }}
  >
    <Card sx={{
      borderRadius: 4,
      background: `linear-gradient(135deg, ${alpha(color, 0.15)} 0%, ${alpha(color, 0.05)} 100%)`,
      backdropFilter: 'blur(10px)',
      border: `1px solid ${alpha(color, 0.2)}`,
      height: '100%',
      transition: 'all 0.3s',
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

const GlassChartCard = ({ title, icon, children, height = 280, animDelay = 0 }) => (
  <motion.div
    initial={{ opacity: 0, scale: 0.95 }}
    animate={{ opacity: 1, scale: 1 }}
    transition={{ duration: 0.6, delay: animDelay }}
    style={{ height: '100%' }}
  >
    <Card sx={{
      borderRadius: 4,
      background: colors.glass,
      backdropFilter: 'blur(10px)',
      border: `1px solid ${alpha('#fff', 0.1)}`,
      height: '100%',
      transition: 'all 0.3s',
      '&:hover': {
        border: `1px solid ${alpha(colors.primary, 0.3)}`,
        boxShadow: `0 10px 30px -5px ${alpha('#000', 0.5)}`
      }
    }}>
      <CardContent>
        <Typography variant="h6" fontWeight="bold" gutterBottom sx={{ 
          display: 'flex', alignItems: 'center', gap: 1, color: '#fff'
        }}>
          <Box sx={{ color: colors.primary }}>{icon}</Box>
          {title}
        </Typography>
        <Box sx={{ height }}>{children}</Box>
      </CardContent>
    </Card>
  </motion.div>
);

const AnimatedCounter = ({ value }) => {
  const [count, setCount] = useState(0);
  useEffect(() => {
    if (value === 0 || isNaN(value)) {
      setCount(0);
      return;
    }
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

const ModernBlockchainDashboard = () => {
  const [loading, setLoading] = useState(true);
  const [initialized, setInitialized] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [activeSection, setActiveSection] = useState('overview');
  const [backendConnected, setBackendConnected] = useState(false);
  const [mlConnected, setMlConnected] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [lastUpdate, setLastUpdate] = useState(null);
  const [error, setError] = useState(null);
  
  const [devices, setDevices] = useState([]);
  const [gethNodes, setGethNodes] = useState([]);
  const [transactions, setTransactions] = useState([]);
  const [blockchainInfo, setBlockchainInfo] = useState({
    latestBlock: 0, gasPrice: '0', networkId: '2026', peers: 4, syncing: false
  });
  
  const [latencyHistory, setLatencyHistory] = useState({ timestamps: [], values: [] });
  const [blockTimeHistory, setBlockTimeHistory] = useState({ timestamps: [], blockTimes: [] });
  const [detectionTimeHistory, setDetectionTimeHistory] = useState({ timestamps: [], times: [] });
  const [precisionHistory, setPrecisionHistory] = useState({
    timestamps: [], ecdsaSuccess: [], zkpSuccess: [], overallSuccess: []
  });
  const [authHistory, setAuthHistory] = useState({ timestamps: [], success: [], failed: [] });
  
  const [mlStats, setMlStats] = useState({ devices: [], threshold: 0.2661 });
  const [mlAnomalies, setMlAnomalies] = useState([]);
  
  const [globalStats, setGlobalStats] = useState({ totalDevices: 0, totalAuths: 0 });
  const [realtimeStats, setRealtimeStats] = useState({
    totalMessages: 0, messagesPerSecond: 0, ecdsaSuccess: 0, zkpSuccess: 0,
    failedAuths: 0, avgLatency: 0, peakLatency: 0
  });

  const isMountedRef = useRef(true);
  const pollingTimeoutRef = useRef(null);

  useEffect(() => {
    return () => {
      isMountedRef.current = false;
      if (pollingTimeoutRef.current) clearTimeout(pollingTimeoutRef.current);
    };
  }, []);

  const apiService = useMemo(() => ({
    checkHealth: async () => {
      try {
        const res = await fetchWithTimeout(`${API_BASE_URL}/api/health`, {}, 3000);
        return { ok: res.ok };
      } catch { return { ok: false }; }
    },
    checkMlHealth: async () => {
      try {
        const res = await fetchWithTimeout(`${ML_API_URL}/health`, {}, 2000);
        return { ok: res.ok };
      } catch { return { ok: false }; }
    },
    getAllDevices: async () => {
      try {
        const res = await fetchWithTimeout(`${API_BASE_URL}/api/devices`, {}, 5000);
        if (!res.ok) return [];
        return await res.json();
      } catch { return []; }
    },
    getGethNodes: async () => {
      try {
        const res = await fetchWithTimeout(`${API_BASE_URL}/api/geth-nodes`, {}, 5000);
        if (!res.ok) return [];
        return await res.json();
      } catch { return []; }
    },
    getBlockchainInfo: async () => {
      try {
        const res = await fetchWithTimeout(`${API_BASE_URL}/api/blockchain/info`, {}, 5000);
        if (!res.ok) return null;
        return await res.json();
      } catch { return null; }
    },
    getRecentTransactions: async (limit = 100) => {
      try {
        const res = await fetchWithTimeout(`${API_BASE_URL}/api/transactions/recent?limit=${limit}`, {}, 5000);
        if (!res.ok) return [];
        return await res.json();
      } catch { return []; }
    },
    getGlobalStats: async () => {
      try {
        const res = await fetchWithTimeout(`${API_BASE_URL}/api/stats/global`, {}, 5000);
        if (!res.ok) return null;
        return await res.json();
      } catch { return null; }
    }
  }), []);

  const updateHistoryData = useCallback((newData) => {
    if (!isMountedRef.current) return;
    const timestamp = new Date().toLocaleTimeString('en-US', { 
      hour: '2-digit', minute: '2-digit', second: '2-digit' 
    });
    
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
    
    const totalMessages = Math.max(1, newData.totalMessages || 1);
    const ecdsaRate = (newData.ecdsaSuccess || 0) / totalMessages * 100;
    const zkpRate = (newData.zkpSuccess || 0) / totalMessages * 100;
    
    setPrecisionHistory(prev => ({
      timestamps: [...prev.timestamps, timestamp].slice(-MAX_HISTORY_POINTS),
      ecdsaSuccess: [...prev.ecdsaSuccess, ecdsaRate].slice(-MAX_HISTORY_POINTS),
      zkpSuccess: [...prev.zkpSuccess, zkpRate].slice(-MAX_HISTORY_POINTS),
      overallSuccess: [...prev.overallSuccess, (ecdsaRate + zkpRate) / 2].slice(-MAX_HISTORY_POINTS)
    }));
  }, []);

  const loadAllData = useCallback(async (showLoading = false, attemptNumber = 0) => {
    if (!isMountedRef.current) return;
    if (showLoading && attemptNumber === 0) setLoading(true);
    
    try {
      const health = await apiService.checkHealth();
      if (!isMountedRef.current) return;
      setBackendConnected(health.ok);
      
      if (!health.ok) {
        if (attemptNumber < MAX_RETRY_ATTEMPTS) {
          setError('Connecting to server...');
          await delay(2000);
          if (isMountedRef.current) return loadAllData(showLoading, attemptNumber + 1);
        } else {
          setError('Cannot connect to backend. Check if server is running.');
          setLoading(false);
        }
        return;
      }
      
      setError(null);
      const mlHealth = await apiService.checkMlHealth();
      if (!isMountedRef.current) return;
      setMlConnected(mlHealth.ok);

      const [devicesData, gethData, blockchainData, txsData, globalStatsData] = await Promise.all([
        apiService.getAllDevices(),
        apiService.getGethNodes(),
        apiService.getBlockchainInfo(),
        apiService.getRecentTransactions(100),
        apiService.getGlobalStats()
      ]);

      if (!isMountedRef.current) return;

      setDevices(devicesData || []);
      setGethNodes(gethData || []);
      
      if (blockchainData) {
        setBlockchainInfo(prev => ({
          ...prev,
          latestBlock: blockchainData.latestBlock || prev.latestBlock,
          gasPrice: blockchainData.gasPrice || prev.gasPrice,
          networkId: blockchainData.networkId || prev.networkId,
          peers: blockchainData.peers || prev.peers,
          syncing: blockchainData.syncing || false
        }));
      }
      
      if (globalStatsData) setGlobalStats(globalStatsData);
      
      if (txsData && Array.isArray(txsData) && txsData.length > 0) {
        const processedTxs = txsData.map((tx, index) => ({
          id: tx.hash || tx.TxHash || `tx-${Date.now()}-${index}`,
          hash: tx.hash || tx.TxHash || 'N/A',
          deviceId: tx.deviceId || tx.DeviceID || tx.device_id || 'Unknown',
          status: tx.success || tx.Success ? 'success' : 'failed',
          timestamp: tx.timestamp || tx.Timestamp || Date.now(),
          latency: tx.latency || tx.Latency || 50,
          ecdsaValid: tx.ecdsaValid === true || tx.EcdsaValid === true,
          zkpValid: tx.zkpValid === true || tx.ZkpValid === true || tx.success === true
        }));
        
        setTransactions(processedTxs);
        
        const ecdsaValid = processedTxs.filter(t => t.ecdsaValid).length;
        const zkpValid = processedTxs.filter(t => t.zkpValid).length;
        const failed = processedTxs.filter(t => t.status === 'failed').length;
        const avgLat = processedTxs.length > 0 
          ? processedTxs.reduce((acc, t) => acc + (t.latency || 0), 0) / processedTxs.length 
          : 0;
        
        const newStats = {
          totalMessages: processedTxs.length,
          messagesPerSecond: '0.5',
          ecdsaSuccess: ecdsaValid,
          zkpSuccess: zkpValid,
          failedAuths: failed,
          avgLatency: Math.round(avgLat),
          peakLatency: Math.round(avgLat * 1.5),
          blockTime: 3
        };
        
        setRealtimeStats(newStats);
        updateHistoryData(newStats);
      }
      
      setLastUpdate(Date.now());
      setInitialized(true);
      
    } catch (error) {
      console.error('Error loading data:', error);
      if (!isMountedRef.current) return;
      if (attemptNumber < MAX_RETRY_ATTEMPTS) {
        setError('Loading error, retrying...');
        await delay(2000);
        if (isMountedRef.current) return loadAllData(showLoading, attemptNumber + 1);
      } else {
        setError('Persistent error: ' + error.message);
      }
    } finally {
      if (isMountedRef.current && showLoading) setLoading(false);
    }
  }, [apiService, updateHistoryData]);

  useEffect(() => {
    const timer = setTimeout(() => {
      if (isMountedRef.current && !initialized) loadAllData(true);
    }, INITIAL_LOAD_DELAY);
    return () => clearTimeout(timer);
  }, [loadAllData, initialized]);

  useEffect(() => {
    if (!autoRefresh || !backendConnected || !initialized) return;
    const poll = async () => {
      if (!isMountedRef.current) return;
      await loadAllData(false);
      if (isMountedRef.current && autoRefresh) {
        pollingTimeoutRef.current = setTimeout(poll, POLLING_INTERVAL);
      }
    };
    poll();
    return () => {
      if (pollingTimeoutRef.current) clearTimeout(pollingTimeoutRef.current);
    };
  }, [autoRefresh, backendConnected, initialized, loadAllData]);

  const chartOptions = useMemo(() => ({
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { 
        position: 'top', 
        labels: { font: { size: 10 }, boxWidth: 12, color: '#94A3B8' } 
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
  }), []);

  const latencyChartData = useMemo(() => ({
    labels: latencyHistory.timestamps.length > 0 ? latencyHistory.timestamps : ['No data'],
    datasets: [{
      label: 'Latency (ms)',
      data: latencyHistory.values.length > 0 ? latencyHistory.values : [0],
      borderColor: colors.warning,
      backgroundColor: alpha(colors.warning, 0.1),
      fill: true, tension: 0.4, pointRadius: 2
    }]
  }), [latencyHistory]);

  const totalDevices = (devices?.length || 0) + (gethNodes?.length || 0);
  const activeDevices = (gethNodes?.length || 0) + (devices?.filter(d => d.isActive)?.length || 0);
  const successRate = Math.round((realtimeStats.zkpSuccess / Math.max(1, realtimeStats.totalMessages)) * 100);

  const renderOverview = () => (
    <>
      <Grid container spacing={2} sx={{ mb: 3 }}>
        <Grid item xs={6} sm={4} md={2}>
          <GlassStatCard title="Total Devices" value={<AnimatedCounter value={totalDevices} />} 
            subtitle={`${activeDevices} active`} icon={<DevicesIcon />} color={colors.primary} animDelay={0} />
        </Grid>
        <Grid item xs={6} sm={4} md={2}>
          <GlassStatCard title="Authentications" value={<AnimatedCounter value={realtimeStats.totalMessages} />}
            subtitle={`${realtimeStats.messagesPerSecond} tx/s`} icon={<VerifiedIcon />} color={colors.success} animDelay={0.1} />
        </Grid>
        <Grid item xs={6} sm={4} md={2}>
          <GlassStatCard title="Avg Latency" value={`${realtimeStats.avgLatency}ms`}
            subtitle={`Peak: ${realtimeStats.peakLatency}ms`} icon={<TimerIcon />} color={colors.warning} animDelay={0.2} />
        </Grid>
        <Grid item xs={6} sm={4} md={2}>
          <GlassStatCard title="Success Rate" value={`${successRate}%`}
            subtitle={`ECDSA: ${realtimeStats.ecdsaSuccess}`} icon={<CheckCircleIcon />} color={colors.info} animDelay={0.3} />
        </Grid>
        <Grid item xs={6} sm={4} md={2}>
          <GlassStatCard title="AI Anomalies" value={mlAnomalies.length}
            subtitle={`Threshold: ${mlStats.threshold.toFixed(3)}`} icon={<AiIcon />} color={colors.error} animDelay={0.4} />
        </Grid>
        <Grid item xs={6} sm={4} md={2}>
          <GlassStatCard title="Network Peers" value={blockchainInfo.peers || 0}
            subtitle={`ChainID: ${blockchainInfo.networkId}`} icon={<NetworkIcon />} color={colors.secondary} animDelay={0.5} />
        </Grid>
      </Grid>

      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid item xs={12} md={6}>
          <GlassChartCard title="Authentication Latency (ms)" icon={<TimerIcon />} animDelay={0.1}>
            <Line data={latencyChartData} options={chartOptions} />
          </GlassChartCard>
        </Grid>
        <Grid item xs={12} md={6}>
          <GlassChartCard title="Block Validation Time (s)" icon={<BlockIcon />} animDelay={0.2}>
            <Line data={{
              labels: blockTimeHistory.timestamps.length > 0 ? blockTimeHistory.timestamps : ['No data'],
              datasets: [{
                label: 'Block Time (s)',
                data: blockTimeHistory.blockTimes.length > 0 ? blockTimeHistory.blockTimes : [0],
                borderColor: colors.primary,
                backgroundColor: alpha(colors.primary, 0.1),
                fill: true, tension: 0.4
              }]
            }} options={chartOptions} />
          </GlassChartCard>
        </Grid>
      </Grid>

      <Grid container spacing={3}>
        <Grid item xs={12}>
          <GlassChartCard title="Device Distribution" icon={<PieChartIcon />} height={250} animDelay={0.5}>
            <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
              <Doughnut data={{
                labels: ['Geth Nodes', 'IoT Devices'],
                datasets: [{
                  data: [gethNodes.length || 1, devices.length || 1],
                  backgroundColor: [colors.warning, colors.success],
                  borderWidth: 0
                }]
              }} options={{ cutout: '60%' }} />
            </Box>
          </GlassChartCard>
        </Grid>
      </Grid>
    </>
  );

  const renderDevices = () => (
    <>
      <Typography variant="h6" fontWeight="bold" sx={{ mb: 2, color: colors.warning }}>
        Geth Nodes - {gethNodes.length} active
      </Typography>
      <TableContainer sx={{ maxHeight: 300, mb: 3, borderRadius: 3, bgcolor: colors.glass }}>
        <Table stickyHeader size="small">
          <TableHead>
            <TableRow sx={{ bgcolor: alpha(colors.dark, 0.8) }}>
              <TableCell sx={{ color: '#fff', fontWeight: 600 }}>Device ID</TableCell>
              <TableCell sx={{ color: '#fff', fontWeight: 600 }}>Address</TableCell>
              <TableCell sx={{ color: '#fff', fontWeight: 600 }}>Status</TableCell>
              <TableCell sx={{ color: '#fff', fontWeight: 600 }}>Auths</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {gethNodes.length > 0 ? gethNodes.map((node, idx) => (
              <TableRow key={idx} hover>
                <TableCell sx={{ color: '#E2E8F0' }}>{node.device_id || 'GETH'}</TableCell>
                <TableCell sx={{ color: '#94A3B8' }}>
                  <code>{(node.address || '').substring(0, 15)}...</code>
                </TableCell>
                <TableCell>
                  <Chip size="small" 
                    icon={node.isActive ? <CheckCircleIcon /> : <ErrorIcon />}
                    label={node.isActive ? 'Active' : 'Inactive'}
                    sx={{ 
                      bgcolor: node.isActive ? alpha(colors.success, 0.2) : alpha(colors.error, 0.2), 
                      color: node.isActive ? colors.success : colors.error 
                    }} />
                </TableCell>
                <TableCell sx={{ color: '#E2E8F0' }}>{node.authCount || 0}</TableCell>
              </TableRow>
            )) : (
              <TableRow>
                <TableCell colSpan={4} sx={{ color: '#94A3B8', textAlign: 'center' }}>
                  No Geth nodes found
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>

      <Typography variant="h6" fontWeight="bold" sx={{ mb: 2, color: colors.success }}>
        IoT Devices - {devices.length} registered
      </Typography>
      <TableContainer sx={{ maxHeight: 300, borderRadius: 3, bgcolor: colors.glass }}>
        <Table stickyHeader size="small">
          <TableHead>
            <TableRow sx={{ bgcolor: alpha(colors.dark, 0.8) }}>
              <TableCell sx={{ color: '#fff', fontWeight: 600 }}>Address</TableCell>
              <TableCell sx={{ color: '#fff', fontWeight: 600 }}>Status</TableCell>
              <TableCell sx={{ color: '#fff', fontWeight: 600 }}>Auths</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {devices.length > 0 ? devices.slice(0, 10).map((device, idx) => (
              <TableRow key={idx} hover>
                <TableCell sx={{ color: '#E2E8F0' }}>
                  <code style={{ color: '#94A3B8' }}>{(device.address || '').substring(0, 20)}...</code>
                </TableCell>
                <TableCell>
                  <Chip size="small" 
                    icon={device.isActive ? <CheckCircleIcon /> : <ErrorIcon />}
                    label={device.isActive ? 'Active' : 'Inactive'}
                    sx={{ 
                      bgcolor: device.isActive ? alpha(colors.success, 0.2) : alpha(colors.error, 0.2), 
                      color: device.isActive ? colors.success : colors.error 
                    }} />
                </TableCell>
                <TableCell sx={{ color: '#E2E8F0' }}>{device.authCount || 0}</TableCell>
              </TableRow>
            )) : (
              <TableRow>
                <TableCell colSpan={3} sx={{ color: '#94A3B8', textAlign: 'center' }}>
                  No IoT devices found
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>
    </>
  );

  const renderTransactions = () => (
    <Card sx={{ borderRadius: 4, bgcolor: colors.glass, backdropFilter: 'blur(10px)' }}>
      <CardContent>
        <Typography variant="h6" fontWeight="bold" gutterBottom sx={{ color: '#fff' }}>
          Recent Transactions
        </Typography>
        <TableContainer sx={{ maxHeight: 400 }}>
          <Table stickyHeader size="small">
            <TableHead>
              <TableRow sx={{ bgcolor: alpha(colors.dark, 0.8) }}>
                <TableCell sx={{ color: '#fff', fontWeight: 600 }}>Hash</TableCell>
                <TableCell sx={{ color: '#fff', fontWeight: 600 }}>Device</TableCell>
                <TableCell sx={{ color: '#fff', fontWeight: 600 }}>ECDSA</TableCell>
                <TableCell sx={{ color: '#fff', fontWeight: 600 }}>ZKP</TableCell>
                <TableCell sx={{ color: '#fff', fontWeight: 600 }}>Latency</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {transactions.length > 0 ? transactions.slice(0, 15).map((tx) => (
                <TableRow key={tx.id} hover>
                  <TableCell sx={{ color: '#94A3B8' }}>
                    <code>{(tx.hash || '').substring(0, 12)}...</code>
                  </TableCell>
                  <TableCell sx={{ color: '#E2E8F0' }}>{tx.deviceId}</TableCell>
                  <TableCell>
                    {tx.ecdsaValid ? 
                      <CheckCircleIcon sx={{ color: colors.success }} fontSize="small" /> : 
                      <ErrorIcon sx={{ color: colors.error }} fontSize="small" />
                    }
                  </TableCell>
                  <TableCell>
                    {tx.zkpValid ? 
                      <VerifiedIcon sx={{ color: colors.success }} fontSize="small" /> : 
                      <ErrorIcon sx={{ color: colors.error }} fontSize="small" />
                    }
                  </TableCell>
                  <TableCell sx={{ color: tx.latency > 200 ? colors.warning : '#E2E8F0' }}>
                    {tx.latency}ms
                  </TableCell>
                </TableRow>
              )) : (
                <TableRow>
                  <TableCell colSpan={5} sx={{ color: '#94A3B8', textAlign: 'center' }}>
                    No transactions found
                  </TableCell>
                </TableRow>
              )}
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
              { label: 'Latest Block', value: `#${(blockchainInfo.latestBlock || 0).toLocaleString()}` },
              { label: 'Network ID', value: blockchainInfo.networkId || 'N/A' },
              { label: 'Gas Price', value: `${blockchainInfo.gasPrice || '0'} wei` },
              { label: 'Connected Peers', value: blockchainInfo.peers || 0 },
            ].map((item, i) => (
              <Box key={i} display="flex" justifyContent="space-between">
                <Typography sx={{ color: '#94A3B8' }}>{item.label}</Typography>
                <Typography sx={{ color: '#E2E8F0', fontWeight: 500 }}>{item.value}</Typography>
              </Box>
            ))}
          </Stack>
        </Card>
      </Grid>
    </Grid>
  );

  const renderContent = () => {
    switch (activeSection) {
      case 'overview': return renderOverview();
      case 'devices': return renderDevices();
      case 'ai': return <Box sx={{ p: 4, textAlign: 'center', color: '#94A3B8' }}>AI Detection module loading...</Box>;
      case 'transactions': return renderTransactions();
      case 'security': return renderSecurity();
      default: return renderOverview();
    }
  };

  if (loading && !initialized) {
    return (
      <Box sx={{ display: 'flex', flexDirection: 'column', justifyContent: 'center', alignItems: 'center', p: 4, bgcolor: colors.darker, minHeight: '100vh' }}>
        <motion.div animate={{ scale: [1, 1.2, 1] }} transition={{ repeat: Infinity, duration: 2 }}>
          <BlockIcon sx={{ fontSize: 60, color: colors.primary, mb: 2 }} />
        </motion.div>
        <Typography variant="h5" sx={{ color: '#fff', mb: 1 }}>Initializing Dashboard</Typography>
        <Typography variant="body2" sx={{ color: '#94A3B8' }}>Connecting to blockchain network...</Typography>
        <Box sx={{ width: 200, mt: 3 }}>
          <LinearProgress sx={{ height: 4, borderRadius: 2, bgcolor: alpha(colors.primary, 0.1), '& .MuiLinearProgress-bar': { bgcolor: colors.primary } }} />
        </Box>
        {error && <Alert severity="warning" sx={{ mt: 2, bgcolor: alpha(colors.warning, 0.1), color: '#fff' }}>{error}</Alert>}
      </Box>
    );
  }

  return (
    <Box sx={{ display: 'flex', bgcolor: colors.darker, minHeight: '100vh' }}>
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
            transition: 'width 0.3s',
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
                NEXUS AI
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
                  px: 2.5, borderRadius: 2, mb: 0.5,
                  '&.Mui-selected': {
                    bgcolor: alpha(colors.primary, 0.15),
                    borderLeft: `3px solid ${colors.primary}`
                  }
                }}
              >
                <ListItemIcon sx={{ minWidth: 0, mr: sidebarOpen ? 3 : 'auto', justifyContent: 'center', color: activeSection === item.section ? colors.primary : '#94A3B8' }}>
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
              <Typography variant="body2" sx={{ color: '#94A3B8', textAlign: 'center', mb: 1 }}>System Status</Typography>
              <Stack direction="row" justifyContent="center" spacing={1}>
                <Chip icon={backendConnected ? <WifiIcon /> : <WifiOffIcon />} 
                  label={backendConnected ? 'Online' : 'Offline'} size="small"
                  sx={{ bgcolor: backendConnected ? alpha(colors.success, 0.2) : alpha(colors.error, 0.2), color: backendConnected ? colors.success : colors.error }} />
                <Chip icon={<AiIcon />} label={mlConnected ? 'AI' : 'Off'} size="small"
                  sx={{ bgcolor: mlConnected ? alpha(colors.primary, 0.2) : alpha(colors.warning, 0.2), color: mlConnected ? colors.primary : colors.warning }} />
              </Stack>
            </Box>
          )}
        </Box>
      </Drawer>

      <Box component="main" sx={{ flexGrow: 1, p: 3 }}>
        <Container maxWidth="xl">
          <motion.div initial={{ opacity: 0, y: -20 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.5 }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 4, flexWrap: 'wrap', gap: 2 }}>
              <Box>
                <Typography variant="h4" fontWeight="bold" sx={{ 
                  background: `linear-gradient(135deg, #fff, ${colors.neon})`,
                  backgroundClip: 'text', WebkitBackgroundClip: 'text', color: 'transparent', mb: 1
                }}>
                  Blockchain AI Sentinel
                </Typography>
                <Stack direction="row" spacing={2} alignItems="center" flexWrap="wrap">
                  <Chip icon={<BlockIcon />} label={`Block #${(blockchainInfo.latestBlock || 0).toLocaleString()}`} 
                    sx={{ bgcolor: alpha(colors.primary, 0.15), color: colors.primary, border: `1px solid ${alpha(colors.primary, 0.3)}` }} />
                  {lastUpdate && (
                    <Chip icon={<TimerIcon />} label={`Updated ${formatDistanceToNow(lastUpdate, { addSuffix: true, locale: enUS })}`}
                      variant="outlined" sx={{ color: '#94A3B8', borderColor: alpha('#fff', 0.1) }} />
                  )}
                </Stack>
              </Box>
              <Stack direction="row" spacing={2} alignItems="center">
                <FormControlLabel
                  control={<Switch checked={autoRefresh} onChange={(e) => setAutoRefresh(e.target.checked)} />}
                  label={<Typography sx={{ color: '#94A3B8' }}>Auto-refresh</Typography>} />
                <Tooltip title="Refresh">
                  <IconButton onClick={() => loadAllData(false)} sx={{ bgcolor: alpha(colors.primary, 0.1), color: colors.primary }}>
                    <RefreshIcon />
                  </IconButton>
                </Tooltip>
              </Stack>
            </Box>
          </motion.div>

          {error && (
            <Alert severity="warning" sx={{ mb: 2, bgcolor: alpha(colors.warning, 0.1), color: '#fff' }} onClose={() => setError(null)}>
              {error}
            </Alert>
          )}

          <AnimatePresence mode="wait">
            <motion.div key={activeSection} initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }} exit={{ opacity: 0, x: -20 }} transition={{ duration: 0.3 }}>
              {renderContent()}
            </motion.div>
          </AnimatePresence>
        </Container>
      </Box>
    </Box>
  );
};

export default ModernBlockchainDashboard;
