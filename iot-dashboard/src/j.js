// src/components/dashboard/ModernBlockchainDashboard.js
// ULTRA MODERN DASHBOARD 2026 - GLASSMORPHISM + NEON + ANIMATIONS
// VERSION DYNAMIQUE - COMPTEURS SÉPARÉS + AI ANOMALY DETECTION

/* eslint-disable no-unused-vars */
/* eslint-disable no-unused-expressions */

import React, { useState, useEffect, useCallback, useMemo } from 'react';
import {
  Box, Typography, Grid, Card, CardContent, Table, TableBody, TableCell,
  TableContainer, TableHead, TableRow, Chip, IconButton, Avatar,
  Stack, LinearProgress, Alert, Tooltip, Skeleton,
  Container, Snackbar, Switch, FormControlLabel, Drawer, List, ListItem,
  ListItemIcon, ListItemText, ListItemButton, Toolbar, alpha,
  Button, Dialog, DialogTitle, DialogContent, DialogActions,
  FormControl, InputLabel, Select, MenuItem, TextField
} from '@mui/material';
import {
  Refresh as RefreshIcon, Verified as VerifiedIcon, Error as ErrorIcon,
  CheckCircle as CheckCircleIcon, Devices as DevicesIcon,
  Timeline as TimelineIcon, Wifi as WifiIcon, WifiOff as WifiOffIcon,
  Block as BlockIcon, SmartToy as AiIcon,
  Warning as WarningIcon, NetworkCheck as NetworkIcon,
  Timer as TimerIcon, Security as SecurityIcon,
  Menu as MenuIcon, ChevronLeft as ChevronLeftIcon, Home as HomeIcon,
  Add as AddIcon, Close as CloseIcon,
  Speed as SpeedIcon,
  BubbleChart as BubbleIcon,
  ArrowUpward as ArrowUpIcon,
  ArrowDownward as ArrowDownIcon,
  History as HistoryIcon,
  Memory as MemoryIcon,
  Storage as StorageIcon,
  Router as RouterIcon
} from '@mui/icons-material';
import {
  Chart as ChartJS, CategoryScale, LinearScale, PointElement,
  LineElement, BarElement, ArcElement, Title, Tooltip as ChartTooltip,
  Legend, Filler, RadialLinearScale
} from 'chart.js';
import { Line, Bar, Doughnut, Radar } from 'react-chartjs-2';
import { formatDistanceToNow, format } from 'date-fns';
import { enUS } from 'date-fns/locale';
import { motion, AnimatePresence } from 'framer-motion';

ChartJS.register(
  CategoryScale, LinearScale, PointElement, LineElement, BarElement, 
  ArcElement, Title, ChartTooltip, Legend, Filler, RadialLinearScale
);

// ============================================
// API CONFIGURATION
// ============================================
const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';
const ML_API_URL = process.env.REACT_APP_ML_API_URL || 'http://localhost:5001';
const POLLING_INTERVAL = 3000;
const MAX_HISTORY_POINTS = 60;

// ============================================
// SAFE HELPERS
// ============================================
const safeStr = (str) => {
  if (!str) return '';
  return String(str).replace(/\r\n/g, ' ').replace(/[\r\n\t]+/g, ' ').replace(/\s+/g, ' ').trim();
};

const safeDate = (ts) => {
  if (!ts) return Date.now();
  try {
    const t = Number(ts);
    const d = t > 1000000000000 ? new Date(t) : new Date(t * 1000);
    return isNaN(d.getTime()) ? new Date() : d;
  } catch { return new Date(); }
};

// ============================================
// PALETTE 2026 - DARK NEON PRO
// ============================================
const colors = {
  primary: '#6366F1', primaryLight: '#818CF8', primaryDark: '#4F46E5',
  secondary: '#8B5CF6', secondaryLight: '#A78BFA',
  success: '#10B981', successLight: '#34D399',
  warning: '#F59E0B', warningLight: '#FBBF24',
  error: '#EF4444', errorLight: '#F87171',
  info: '#06B6D4', infoLight: '#22D3EE',
  dark: '#0B0F19', darker: '#070A12', card: '#111827',
  textPrimary: '#F9FAFB', textSecondary: '#9CA3AF', textMuted: '#6B7280',
  neon: '#00FFCC', neonPurple: '#C084FC', neonPink: '#F472B6',
  neonCyan: '#22D3EE', neonLime: '#A3E635', neonOrange: '#FB923C',
  glass: 'rgba(17, 24, 39, 0.7)', glassBorder: 'rgba(255, 255, 255, 0.08)',
};

// ============================================
// MENU ITEMS
// ============================================
const menuItems = [
  { text: 'Overview', icon: <HomeIcon />, section: 'overview', color: colors.primary },
  { text: 'Devices', icon: <DevicesIcon />, section: 'devices', color: colors.neonCyan },
  { text: 'AI Detection', icon: <AiIcon />, section: 'ai', color: colors.neonPink },
  { text: 'Security', icon: <SecurityIcon />, section: 'security', color: colors.neonLime },
  { text: 'Transactions', icon: <TimelineIcon />, section: 'transactions', color: colors.neonOrange },
];

// ============================================
// KPI CARD COMPONENT
// ============================================
const KPICard = ({ title, value, subtitle, icon, color, trend, trendValue, delay = 0 }) => {
  const isPositive = trend === 'up';
  return (
    <motion.div initial={{ opacity: 0, y: 30, scale: 0.9 }} animate={{ opacity: 1, y: 0, scale: 1 }} transition={{ duration: 0.6, delay, type: "spring", stiffness: 100 }} style={{ height: '100%' }}>
      <Card sx={{ height: '100%', borderRadius: 3, background: `linear-gradient(135deg, ${alpha(color, 0.15)} 0%, ${alpha(colors.card, 0.9)} 100%)`, backdropFilter: 'blur(20px)', border: `1px solid ${alpha(color, 0.2)}`, position: 'relative', overflow: 'hidden', transition: 'all 0.4s', '&:hover': { transform: 'translateY(-8px)', border: `1px solid ${alpha(color, 0.5)}`, boxShadow: `0 20px 40px -10px ${alpha(color, 0.3)}` }, '&::before': { content: '""', position: 'absolute', top: 0, left: 0, right: 0, height: '3px', background: `linear-gradient(90deg, ${color}, ${alpha(color, 0.3)})` } }}>
        <CardContent sx={{ p: 3, position: 'relative', zIndex: 1 }}>
          <Box display="flex" justifyContent="space-between" alignItems="flex-start" mb={2}>
            <Box sx={{ p: 1.5, borderRadius: 2, bgcolor: alpha(color, 0.15), border: `1px solid ${alpha(color, 0.3)}` }}><Box sx={{ color: color, fontSize: 28 }}>{icon}</Box></Box>
            {trend && <Chip size="small" icon={isPositive ? <ArrowUpIcon fontSize="small" /> : <ArrowDownIcon fontSize="small" />} label={`${trendValue}%`} sx={{ bgcolor: isPositive ? alpha(colors.success, 0.2) : alpha(colors.error, 0.2), color: isPositive ? colors.success : colors.error, fontWeight: 600 }} />}
          </Box>
          <Typography variant="body2" sx={{ color: colors.textSecondary, mb: 1, fontWeight: 500, textTransform: 'uppercase', fontSize: '0.75rem' }}>{title}</Typography>
          <Typography variant="h3" fontWeight="bold" sx={{ color: colors.textPrimary, mb: 1, fontSize: '2.5rem', lineHeight: 1.2, textShadow: `0 0 20px ${alpha(color, 0.3)}` }}>{value}</Typography>
          {subtitle && <Typography variant="caption" sx={{ color: colors.textMuted, display: 'block', mt: 1 }}>{subtitle}</Typography>}
        </CardContent>
      </Card>
    </motion.div>
  );
};

// ============================================
// GRAND CHART CARD COMPONENT
// ============================================
const LargeChartCard = ({ title, icon, children, color, delay = 0 }) => {
  return (
    <motion.div initial={{ opacity: 0, y: 30, scale: 0.95 }} animate={{ opacity: 1, y: 0, scale: 1 }} transition={{ duration: 0.5, delay, type: "spring", stiffness: 100 }} style={{ height: '100%', width: '100%' }}>
      <Card sx={{ height: 500, width: '100%', borderRadius: 4, background: `linear-gradient(145deg, ${alpha(colors.card, 0.95)} 0%, ${alpha(color, 0.08)} 100%)`, backdropFilter: 'blur(20px)', border: `1px solid ${alpha(color, 0.25)}`, transition: 'all 0.4s', '&:hover': { transform: 'translateY(-8px)', border: `1px solid ${alpha(color, 0.6)}`, boxShadow: `0 30px 50px -20px ${alpha(color, 0.4)}` }, display: 'flex', flexDirection: 'column', overflow: 'hidden', position: 'relative' }}>
        <Box sx={{ position: 'absolute', top: 0, left: 0, right: 0, height: '3px', background: `linear-gradient(90deg, ${color}, ${alpha(color, 0.5)}, transparent)` }} />
        <CardContent sx={{ height: '100%', width: '100%', p: 3, display: 'flex', flexDirection: 'column', position: 'relative', zIndex: 1 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2.5, flexShrink: 0 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
              <Avatar sx={{ bgcolor: alpha(color, 0.15), width: 48, height: 48, border: `1px solid ${alpha(color, 0.4)}`, boxShadow: `0 0 15px ${alpha(color, 0.3)}` }}><Box sx={{ color: color, fontSize: 28 }}>{icon}</Box></Avatar>
              <Box><Typography variant="h5" fontWeight="bold" sx={{ color: colors.textPrimary, fontSize: '1.3rem' }}>{title}</Typography><Typography variant="caption" sx={{ color: alpha(color, 0.9), fontWeight: 500 }}>Live data from blockchain</Typography></Box>
            </Box>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, px: 1.5, py: 0.5, borderRadius: 2, bgcolor: alpha(color, 0.1), border: `1px solid ${alpha(color, 0.2)}` }}>
              <Box sx={{ width: 10, height: 10, borderRadius: '50%', bgcolor: color, boxShadow: `0 0 12px ${color}`, animation: 'pulse 2s ease-in-out infinite' }} />
              <Typography variant="caption" sx={{ color: alpha(color, 0.9), fontWeight: 500 }}>LIVE</Typography>
            </Box>
          </Box>
          <Box sx={{ position: 'relative', flex: 1, width: '100%', minHeight: 0, mt: 1 }}>{children}</Box>
        </CardContent>
      </Card>
    </motion.div>
  );
};

// ============================================
// LIVE TRANSACTION FEED
// ============================================
const LiveTransactionFeed = ({ transactions }) => {
  const [visibleTxs, setVisibleTxs] = useState([]);
  useEffect(() => { const valid = (transactions || []).filter(tx => tx && tx.timestamp); setVisibleTxs(valid.slice(0, 6)); }, [transactions]);
  if (!visibleTxs || visibleTxs.length === 0) return <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}><Typography variant="caption" sx={{ color: colors.textMuted, textAlign: 'center', mt: 4 }}>Waiting for transactions...</Typography></Box>;
  return (
    <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}>
        <Box sx={{ width: 8, height: 8, borderRadius: '50%', bgcolor: colors.error, animation: 'pulse 1s infinite', boxShadow: `0 0 8px ${colors.error}` }} />
        <Typography variant="caption" sx={{ color: colors.error, fontWeight: 'bold', letterSpacing: 1 }}>LIVE FEED</Typography>
        <Typography variant="caption" sx={{ color: colors.textMuted, ml: 'auto' }}>Last {visibleTxs.length} transactions</Typography>
      </Box>
      <Stack spacing={1.5} sx={{ flex: 1, overflowY: 'auto', pr: 0.5 }}>
        {visibleTxs.map((tx, idx) => {
          const sid = safeStr(tx.deviceId || 'Unknown').substring(0, 12);
          const sl = Number(tx.latency) || 0;
          const sz = tx.zkpValid === true;
          return (
            <motion.div key={tx.id || idx} initial={{ opacity: 0, x: -20 }} animate={{ opacity: 1, x: 0 }} transition={{ delay: idx * 0.05 }} style={{ width: '100%' }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, p: 1, borderRadius: 2, bgcolor: alpha(colors.card, 0.4), borderLeft: `3px solid ${sz ? colors.success : colors.error}`, transition: 'all 0.2s', '&:hover': { bgcolor: alpha(colors.primary, 0.1), transform: 'translateX(4px)' } }}>
                <Chip size="small" label={sz ? 'OK' : 'FAIL'} sx={{ bgcolor: sz ? alpha(colors.success, 0.2) : alpha(colors.error, 0.2), color: sz ? colors.success : colors.error, height: 20, minWidth: 45, fontSize: '0.65rem', fontWeight: 'bold' }} />
                <Typography variant="caption" sx={{ color: colors.textPrimary, flex: 1, fontFamily: 'monospace', fontSize: '0.7rem' }}>{sid}...</Typography>
                <Typography variant="caption" sx={{ color: sl > 200 ? colors.warning : colors.textMuted, fontWeight: sl > 200 ? 'bold' : 'normal', minWidth: 45, textAlign: 'right' }}>{sl}ms</Typography>
                <TimerIcon sx={{ fontSize: 12, color: colors.textMuted }} />
              </Box>
            </motion.div>
          );
        })}
      </Stack>
    </Box>
  );
};

// ============================================
// ANIMATED COUNTER
// ============================================
const AnimatedCounter = ({ value, duration = 1500 }) => {
  const [count, setCount] = useState(0);
  useEffect(() => {
    const num = typeof value === 'string' ? parseInt(value.replace(/[^0-9]/g, '')) || 0 : value;
    const steps = 40; let c = 0, s = 0;
    const t = setInterval(() => { s++; c = Math.min(num, num * (1 - Math.pow(1 - s / steps, 4))); setCount(Math.floor(c)); if (s >= steps) { clearInterval(t); setCount(num); } }, duration / steps);
    return () => clearInterval(t);
  }, [value, duration]);
  const fmt = typeof value === 'string' && value.includes('%') ? `${count}%` : typeof value === 'string' && value.includes('ms') ? `${count}ms` : count.toLocaleString();
  return <span>{fmt}</span>;
};

// ============================================
// MAIN COMPONENT
// ============================================
const ModernBlockchainDashboard = () => {
  const [loading, setLoading] = useState(true);
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [activeSection, setActiveSection] = useState('overview');
  const [backendConnected, setBackendConnected] = useState(false);
  const [mlConnected, setMlConnected] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [lastUpdate, setLastUpdate] = useState(Date.now());
  const [errorOpen, setErrorOpen] = useState(false);
  const [errorMessage, setErrorMessage] = useState('');
  const [addNodeOpen, setAddNodeOpen] = useState(false);
  const [isValidator, setIsValidator] = useState(true);
  const [numValidators, setNumValidators] = useState(1);
  const [numFullnodes, setNumFullnodes] = useState(0);
  const [scriptRunning, setScriptRunning] = useState(false);
  const [devices, setDevices] = useState([]);
  const [gethNodes, setGethNodes] = useState([]);
  const [transactions, setTransactions] = useState([]);
  const [blockchainInfo, setBlockchainInfo] = useState({ latestBlock: 0, gasPrice: '0', networkId: '2026', peers: 4 });
  const [latencyHistory, setLatencyHistory] = useState({ timestamps: [], values: [] });
  const [precisionHistory, setPrecisionHistory] = useState({ timestamps: [], ecdsaSuccess: [], zkpSuccess: [] });
  const [authHistory, setAuthHistory] = useState({ timestamps: [], success: [], failed: [] });
  const [mlStats, setMlStats] = useState({ f1_score: 96.45, precision: 94.32, recall: 98.69, threshold: 0.1437, total_analyzed: 0, anomalies_detected: 0, revocations: 0 });
  const [mlAnomalies, setMlAnomalies] = useState([]);
  const [mlActionLog, setMlActionLog] = useState([]);
  const [globalStats, setGlobalStats] = useState({ totalDevices: 0, totalAuths: 0 });
  const [realtimeStats, setRealtimeStats] = useState({ totalMessages: 0, messagesPerSecond: 0, ecdsaSuccess: 0, zkpSuccess: 0, failedAuths: 0, avgLatency: 0, peakLatency: 0 });

  const iotDevicesCount = useMemo(() => devices.filter(d => !d.isGethNode).length, [devices]);
  const gethValidatorsCount = useMemo(() => gethNodes.filter(n => n.type === 'validator').length, [gethNodes]);
  const gethFullnodesCount = useMemo(() => gethNodes.filter(n => n.type === 'fullnode').length, [gethNodes]);
  const totalDevices = iotDevicesCount + gethValidatorsCount + gethFullnodesCount;

  const apiService = useMemo(() => ({
    checkHealth: async () => { try { const r = await fetch(`${API_BASE_URL}/api/health`); return { ok: r.ok }; } catch { return { ok: false }; } },
    checkMlHealth: async () => { try { const r = await fetch(`${ML_API_URL}/v4/health`); return { ok: r.ok, data: r.ok ? await r.json() : null }; } catch { return { ok: false, data: null }; } },
    getAllDevices: async () => { try { const r = await fetch(`${API_BASE_URL}/api/devices`); if (!r.ok) return []; return await r.json(); } catch { return []; } },
    getGethNodes: async () => { try { const r = await fetch(`${API_BASE_URL}/api/geth-nodes`); if (!r.ok) return []; return await r.json(); } catch { return []; } },
    getBlockchainInfo: async () => { try { const r = await fetch(`${API_BASE_URL}/api/blockchain/info`); if (!r.ok) return null; return await r.json(); } catch { return null; } },
    getRecentTransactions: async (limit = 200) => { try { const r = await fetch(`${API_BASE_URL}/api/transactions/recent?limit=${limit}`); if (!r.ok) return []; return await r.json(); } catch { return []; } },
    getGlobalStats: async () => { try { const r = await fetch(`${API_BASE_URL}/api/stats/global`); if (!r.ok) return null; return await r.json(); } catch { return null; } },
    getMlDevices: async () => { try { const r = await fetch(`${ML_API_URL}/v4/devices`); if (!r.ok) return []; return await r.json(); } catch { return []; } },
    getMlAnalysis: async (addr) => { try { const r = await fetch(`${ML_API_URL}/v4/analyze/${addr}`); if (!r.ok) return null; return await r.json(); } catch { return null; } },
  }), []);

  const updateHistoryData = useCallback((newData) => {
    const ts = new Date().toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    setLatencyHistory(p => ({ timestamps: [...p.timestamps.slice(-MAX_HISTORY_POINTS), ts], values: [...p.values.slice(-MAX_HISTORY_POINTS), newData.avgLatency || 0] }));
    const er = (newData.ecdsaSuccess / Math.max(1, newData.totalMessages)) * 100;
    const zr = (newData.zkpSuccess / Math.max(1, newData.totalMessages)) * 100;
    setPrecisionHistory(p => ({ timestamps: [...p.timestamps.slice(-MAX_HISTORY_POINTS), ts], ecdsaSuccess: [...p.ecdsaSuccess.slice(-MAX_HISTORY_POINTS), er], zkpSuccess: [...p.zkpSuccess.slice(-MAX_HISTORY_POINTS), zr] }));
    setAuthHistory(p => ({ timestamps: [...p.timestamps.slice(-MAX_HISTORY_POINTS), ts], success: [...p.success.slice(-MAX_HISTORY_POINTS), newData.ecdsaSuccess || 0], failed: [...p.failed.slice(-MAX_HISTORY_POINTS), newData.failedAuths || 0] }));
  }, []);

  const loadAllData = useCallback(async (showLoading = false) => {
    if (showLoading) setLoading(true);
    try {
      const health = await apiService.checkHealth();
      setBackendConnected(health.ok);
      const mlHealth = await apiService.checkMlHealth();
      setMlConnected(mlHealth.ok);
      if (mlHealth.data) setMlStats(p => ({ ...p, threshold: mlHealth.data.threshold || p.threshold, f1_score: mlHealth.data.f1_score || p.f1_score }));
      if (!health.ok) { setErrorMessage('Backend disconnected'); setErrorOpen(true); if (showLoading) setLoading(false); return; }

      const [devicesData, gethData, blockchainData, txsData, globalStatsData] = await Promise.all([
        apiService.getAllDevices(), apiService.getGethNodes(), apiService.getBlockchainInfo(),
        apiService.getRecentTransactions(200), apiService.getGlobalStats()
      ]);
      setDevices(devicesData || []); setGethNodes(gethData || []);
      if (blockchainData) setBlockchainInfo(p => ({ ...p, ...blockchainData }));
      if (globalStatsData) setGlobalStats(globalStatsData);

      // Traiter les transactions
      if (txsData && txsData.length > 0) {
        const processedTxs = txsData.map((tx, i) => ({
          id: safeStr(tx.hash || `tx-${Date.now()}-${i}`),
          hash: safeStr(tx.hash || 'N/A'), deviceId: safeStr(tx.deviceId || 'Unknown'),
          status: tx.success ? 'success' : 'failed', timestamp: safeDate(tx.timestamp),
          latency: Number(tx.latency) || 0,
          ecdsaValid: tx.ecdsaValid === true, zkpValid: tx.zkpValid === true || tx.success === true
        }));
        setTransactions(processedTxs);
        const ev = processedTxs.filter(t => t.ecdsaValid).length;
        const zv = processedTxs.filter(t => t.zkpValid).length;
        const al = Math.round(processedTxs.reduce((a, t) => a + t.latency, 0) / Math.max(1, processedTxs.length));
        const pl = Math.round(Math.max(...processedTxs.map(t => t.latency), 0));
        setRealtimeStats({ totalMessages: processedTxs.length, messagesPerSecond: (processedTxs.filter(t => t.timestamp > Date.now() - 60000).length / 60).toFixed(2), ecdsaSuccess: ev, zkpSuccess: zv, failedAuths: processedTxs.filter(t => t.status === 'failed').length, avgLatency: al, peakLatency: pl });
        updateHistoryData({ avgLatency: al, ecdsaSuccess: ev, zkpSuccess: zv, totalMessages: processedTxs.length, failedAuths: processedTxs.filter(t => t.status === 'failed').length });
      }

      // 🔥 Récupérer les anomalies IA depuis le backend
      if (mlHealth.ok) {
        try {
          const mlDevices = await apiService.getMlDevices();
          if (mlDevices && mlDevices.devices) {
            const anomalies = [];
            for (const addr of mlDevices.devices.slice(0, 10)) {
              try {
                const analysis = await apiService.getMlAnalysis(addr);
                if (analysis && analysis.is_anomaly) {
                  anomalies.push({ id: Date.now() + Math.random(), deviceAddress: addr, score: analysis.score || 0, severity: analysis.severity || 'LOW', action: analysis.action || 'MONITOR', status: analysis.status || 'normal', timestamp: analysis.timestamp || new Date().toISOString(), mse: analysis.mse || 0 });
                  if (analysis.action === 'REVOKE' || analysis.action === 'IMMEDIATE_REVOKE') {
                    setMlActionLog(p => [{ id: Date.now(), deviceAddress: addr, action: analysis.action, score: analysis.score, timestamp: new Date().toISOString() }, ...p].slice(0, 20));
                  }
                }
              } catch {}
            }
            setMlAnomalies(anomalies.slice(0, 20));
            setMlStats(p => ({ ...p, total_analyzed: mlDevices.devices.length, anomalies_detected: anomalies.length, revocations: anomalies.filter(a => a.action === 'REVOKE' || a.action === 'IMMEDIATE_REVOKE').length }));
          }
        } catch {}
      }
      setLastUpdate(Date.now());
    } catch (error) { console.error(error); setErrorMessage('Error: ' + (error.message || 'Unknown')); setErrorOpen(true); }
    finally { if (showLoading) setLoading(false); }
  }, [apiService, updateHistoryData]);

  useEffect(() => { loadAllData(true); }, [loadAllData]);
  useEffect(() => { if (!autoRefresh) return; const i = setInterval(() => loadAllData(false), POLLING_INTERVAL); return () => clearInterval(i); }, [autoRefresh, loadAllData]);

  // Charts
  const latencyChartData = { labels: latencyHistory.timestamps, datasets: [{ data: latencyHistory.values.filter(v => !isNaN(v)), borderColor: colors.warning, backgroundColor: (ctx) => { const g = ctx.chart.ctx.createLinearGradient(0, 0, 0, 400); g.addColorStop(0, alpha(colors.warning, 0.4)); g.addColorStop(1, alpha(colors.warning, 0)); return g; }, fill: true, borderWidth: 3, tension: 0.4 }] };
  const latencyOptions = { responsive: true, maintainAspectRatio: false, plugins: { legend: { display: false } }, scales: { x: { grid: { display: false }, ticks: { color: colors.textMuted } }, y: { beginAtZero: true, grid: { color: alpha('#fff', 0.03) }, ticks: { color: colors.textMuted } } } };
  const precisionChartData = { labels: precisionHistory.timestamps.slice(-12), datasets: [{ label: 'ECDSA', data: precisionHistory.ecdsaSuccess.slice(-12), backgroundColor: alpha(colors.success, 0.8), borderRadius: 8 }, { label: 'ZKP', data: precisionHistory.zkpSuccess.slice(-12), backgroundColor: alpha(colors.info, 0.8), borderRadius: 8 }] };
  const authChartData = { labels: authHistory.timestamps.slice(-12), datasets: [{ label: 'Success', data: authHistory.success.slice(-12), backgroundColor: alpha(colors.success, 0.8), borderRadius: 8 }, { label: 'Failed', data: authHistory.failed.slice(-12), backgroundColor: alpha(colors.error, 0.8), borderRadius: 8 }] };
  const barOptions = { responsive: true, maintainAspectRatio: false, plugins: { legend: { position: 'top', labels: { color: colors.textSecondary } } }, scales: { x: { grid: { display: false }, ticks: { color: colors.textMuted } }, y: { beginAtZero: true, grid: { color: alpha('#fff', 0.03) }, ticks: { color: colors.textMuted } } } };

  const successRate = Math.round((realtimeStats.zkpSuccess / Math.max(1, realtimeStats.totalMessages)) * 100);
  const ecdsaRate = Math.round((realtimeStats.ecdsaSuccess / Math.max(1, realtimeStats.totalMessages)) * 100);

  if (loading) return <Box sx={{ p: 4, bgcolor: colors.darker, minHeight: '100vh' }}><Grid container spacing={3}>{[1,2,3,4].map(i => <Grid item xs={12} lg={3} key={i}><Skeleton variant="rectangular" height={500} sx={{ borderRadius: 4, bgcolor: alpha('#fff', 0.05) }} /></Grid>)}</Grid></Box>;

  return (
    <Box sx={{ display: 'flex', bgcolor: colors.darker, minHeight: '100vh' }}>
      <Drawer variant="permanent" open={sidebarOpen} sx={{ width: sidebarOpen?280:80, '& .MuiDrawer-paper': { width: sidebarOpen?280:80, bgcolor: alpha(colors.dark, 0.85), borderRight: `1px solid ${alpha('#fff', 0.05)}`, transition: 'width 0.3s', overflowX: 'hidden' } }}>
        <Toolbar sx={{ display: 'flex', alignItems: 'center', justifyContent: sidebarOpen?'space-between':'center', px: 2, minHeight: '72px!important', borderBottom: `1px solid ${alpha('#fff', 0.05)}` }}>
          {sidebarOpen && <Typography variant="h6" fontWeight="bold" sx={{ background: `linear-gradient(135deg, ${colors.primary}, ${colors.neon})`, backgroundClip: 'text', WebkitBackgroundClip: 'text', color: 'transparent' }}>CHAIMAN IOT</Typography>}
          <IconButton onClick={() => setSidebarOpen(!sidebarOpen)} sx={{ color: colors.textSecondary }}>{sidebarOpen ? <ChevronLeftIcon/> : <MenuIcon/>}</IconButton>
        </Toolbar>
        <List sx={{ p: 1.5 }}>{menuItems.map(item => <ListItem key={item.text} disablePadding sx={{ mb: 0.5 }}><ListItemButton selected={activeSection===item.section} onClick={()=>setActiveSection(item.section)} sx={{ minHeight:52, borderRadius:2, '&.Mui-selected':{bgcolor:alpha(item.color,0.15), borderLeft:`3px solid ${item.color}`}, '&:hover':{bgcolor:alpha('#fff',0.05)} }}><ListItemIcon sx={{ color: activeSection===item.section?item.color:colors.textSecondary }}>{item.icon}</ListItemIcon>{sidebarOpen && <ListItemText primary={item.text} sx={{ color: activeSection===item.section?colors.textPrimary:colors.textSecondary }} />}</ListItemButton></ListItem>)}</List>
        <Box sx={{ mt:'auto', p:2, borderTop:`1px solid ${alpha('#fff',0.05)}` }}>{sidebarOpen && <Stack direction="row" spacing={1}><Chip icon={backendConnected?<WifiIcon/>:<WifiOffIcon/>} label={backendConnected?'Online':'Offline'} size="small" sx={{ bgcolor:backendConnected?alpha(colors.success,0.15):alpha(colors.error,0.15), color:backendConnected?colors.success:colors.error }} /><Chip icon={<AiIcon/>} label={mlConnected?'AI Ready':'AI Off'} size="small" sx={{ bgcolor:mlConnected?alpha(colors.primary,0.15):alpha(colors.warning,0.15), color:mlConnected?colors.primary:colors.warning }} /></Stack>}</Box>
      </Drawer>

      <Box component="main" sx={{ flexGrow:1, p:3 }}>
        <Container maxWidth={false}>
          <Box sx={{ display:'flex', justifyContent:'space-between', alignItems:'center', mb:4, pb:3, borderBottom:`1px solid ${alpha('#fff',0.05)}`, flexWrap:'wrap', gap:2 }}>
            <Box><Typography variant="h4" fontWeight="800" sx={{ background:`linear-gradient(135deg,#fff,${colors.neon})`, backgroundClip:'text', WebkitBackgroundClip:'text', color:'transparent', mb:1 }}>CHAIMAN IOT • AI Sentinel</Typography><Stack direction="row" spacing={2}><Chip icon={<BlockIcon/>} label={`Block #${blockchainInfo.latestBlock?.toLocaleString()||0}`} sx={{ bgcolor:alpha(colors.primary,0.15), color:colors.primary }} /><Chip icon={<TimerIcon/>} label={`Updated ${formatDistanceToNow(lastUpdate,{addSuffix:true,locale:enUS})}`} variant="outlined" sx={{ color:colors.textSecondary }} /></Stack></Box>
            <Stack direction="row" spacing={2}><FormControlLabel control={<Switch checked={autoRefresh} onChange={e=>setAutoRefresh(e.target.checked)} />} label="Auto-refresh" sx={{ color:colors.textSecondary }} /><IconButton onClick={()=>loadAllData(false)} sx={{ bgcolor:alpha(colors.primary,0.1), color:colors.primary }}><RefreshIcon/></IconButton></Stack>
          </Box>

          <AnimatePresence mode="wait"><motion.div key={activeSection} initial={{ opacity:0,x:20 }} animate={{ opacity:1,x:0 }} exit={{ opacity:0,x:-20 }}>
            {activeSection === 'overview' && (
              <Box>
                <Grid container spacing={2} sx={{ mb: 4 }}>
                  {[{t:'IoT',v:iotDevicesCount,c:colors.success,i:<DevicesIcon/>},{t:'Validators',v:gethValidatorsCount,c:colors.warning,i:<SecurityIcon/>},{t:'Fullnodes',v:gethFullnodesCount,c:colors.info,i:<StorageIcon/>},{t:'Total',v:totalDevices,c:colors.primary,i:<RouterIcon/>},{t:'Auths',v:globalStats.totalAuths||realtimeStats.totalMessages,c:colors.neonLime,i:<VerifiedIcon/>},{t:'Latency',v:`${realtimeStats.avgLatency}ms`,c:colors.neonOrange,i:<SpeedIcon/>},{t:'Success',v:`${successRate}%`,c:colors.neonCyan,i:<CheckCircleIcon/>},{t:'AI Anomalies',v:mlAnomalies.length,c:colors.neonPink,i:<AiIcon/>}].map((k,i)=><Grid item xs={12} sm={6} md={3} lg={1.5} key={i}><KPICard title={k.t} value={<AnimatedCounter value={k.v}/>} subtitle="" icon={k.i} color={k.c} delay={i*0.05}/></Grid>)}
                </Grid>
                <Grid container spacing={3} sx={{ mb: 3 }}>
                  <Grid item xs={12} lg={4}><LargeChartCard title="Latency" icon={<TimerIcon/>} color={colors.warning}><Line data={latencyChartData} options={latencyOptions}/></LargeChartCard></Grid>
                  <Grid item xs={12} lg={4}><LargeChartCard title="Precision" icon={<VerifiedIcon/>} color={colors.success}><Bar data={precisionChartData} options={barOptions}/></LargeChartCard></Grid>
                  <Grid item xs={12} lg={4}><LargeChartCard title="Auths/min" icon={<TimelineIcon/>} color={colors.info}><Bar data={authChartData} options={barOptions}/></LargeChartCard></Grid>
                </Grid>
                <Grid container spacing={3}><Grid item xs={12}><Card sx={{ borderRadius:4, bgcolor:alpha(colors.card,0.95), border:`1px solid ${alpha(colors.neonOrange,0.25)}`, overflow:'hidden' }}><CardContent sx={{ p:3 }}><Typography variant="h6" sx={{ color:colors.textPrimary, mb:2 }}>Live Feed</Typography><LiveTransactionFeed transactions={transactions}/></CardContent></Card></Grid></Grid>
              </Box>
            )}
            {activeSection === 'devices' && (
              <Box>
                <Box sx={{ display:'flex', justifyContent:'space-between', alignItems:'center', mb:3 }}>
                  <Typography variant="h5" fontWeight="bold" sx={{ color:colors.warning }}>Geth Nodes - {gethNodes.length}</Typography>
                  <Button variant="contained" startIcon={<AddIcon/>} onClick={()=>setAddNodeOpen(true)} sx={{ bgcolor:colors.primary, borderRadius:2, textTransform:'none', fontWeight:600 }}>Add Geth Node</Button>
                </Box>
                <Card sx={{ mb:4, borderRadius:3, bgcolor:alpha(colors.card,0.8), border:`1px solid ${alpha(colors.warning,0.2)}` }}><TableContainer sx={{ maxHeight:400 }}><Table stickyHeader><TableHead><TableRow>{['ID','Address','Type','Status','Auths','Last Seen'].map(h=><TableCell key={h} sx={{ color:colors.textPrimary, fontWeight:600, bgcolor:alpha(colors.darker,0.9) }}>{h}</TableCell>)}</TableRow></TableHead><TableBody>{gethNodes.map((node,i)=>(
                  <TableRow key={node.address||i} hover><TableCell sx={{ color:colors.textPrimary }}><Box sx={{ display:'flex',alignItems:'center',gap:1 }}><Avatar sx={{ width:28,height:28,bgcolor:node.type==='validator'?colors.warning:colors.info,fontSize:12 }}>{node.type==='validator'?'V':'F'}</Avatar><Typography variant="body2">{safeStr(node.device_id||'GETH')}</Typography></Box></TableCell><TableCell sx={{ color:colors.textSecondary }}><code>{safeStr(node.address||'').substring(0,15)}...</code></TableCell><TableCell><Chip size="small" label={node.type||'validator'} sx={{ bgcolor:alpha(node.type==='fullnode'?colors.info:colors.warning,0.2), color:node.type==='fullnode'?colors.info:colors.warning }}/></TableCell><TableCell><Chip size="small" icon={node.isActive?<CheckCircleIcon fontSize="small"/>:<ErrorIcon fontSize="small"/>} label={node.isActive?'Active':'Inactive'} sx={{ bgcolor:alpha(node.isActive?colors.success:colors.error,0.2), color:node.isActive?colors.success:colors.error }}/></TableCell><TableCell sx={{ color:colors.textPrimary }}>{node.authCount||0}</TableCell><TableCell sx={{ color:colors.textSecondary }}>{node.lastSeen?formatDistanceToNow(node.lastSeen*1000,{addSuffix:true,locale:enUS}):'Never'}</TableCell></TableRow>
                ))}</TableBody></Table></TableContainer></Card>
                <Box sx={{ display:'flex', justifyContent:'space-between', alignItems:'center', mb:3 }}>
                  <Typography variant="h5" fontWeight="bold" sx={{ color:colors.success }}>IoT Devices - {iotDevicesCount}</Typography>
                </Box>
                <Card sx={{ borderRadius:3, bgcolor:alpha(colors.card,0.8), border:`1px solid ${alpha(colors.success,0.2)}` }}><TableContainer sx={{ maxHeight:400 }}><Table stickyHeader><TableHead><TableRow>{['Address','Status','Auths','Last Seen'].map(h=><TableCell key={h} sx={{ color:colors.textPrimary, fontWeight:600, bgcolor:alpha(colors.darker,0.9) }}>{h}</TableCell>)}</TableRow></TableHead><TableBody>{devices.filter(d=>!d.isGethNode).slice(0,50).map((device,i)=>(
                  <TableRow key={device.address||i} hover><TableCell sx={{ color:colors.textPrimary }}><Box sx={{ display:'flex',alignItems:'center',gap:1 }}><Avatar sx={{ width:28,height:28,bgcolor:colors.success,fontSize:12 }}>I</Avatar><code style={{ color:colors.textSecondary }}>{safeStr(device.address||'').substring(0,20)}...</code></Box></TableCell><TableCell><Chip size="small" icon={device.isActive?<CheckCircleIcon fontSize="small"/>:<ErrorIcon fontSize="small"/>} label={device.isActive?'Active':'Inactive'} sx={{ bgcolor:alpha(device.isActive?colors.success:colors.error,0.2), color:device.isActive?colors.success:colors.error }}/></TableCell><TableCell sx={{ color:colors.textPrimary }}>{device.authCount||0}</TableCell><TableCell sx={{ color:colors.textSecondary }}>{device.lastSeen?formatDistanceToNow(device.lastSeen*1000,{addSuffix:true,locale:enUS}):'Never'}</TableCell></TableRow>
                ))}</TableBody></Table></TableContainer></Card>
              </Box>
            )}
            {activeSection === 'ai' && (
              <Box>
                <Grid container spacing={2} sx={{ mb:4 }}>
                  {[{t:'F1-Score',v:`${mlStats.f1_score}%`,c:colors.neon,i:<AiIcon/>},{t:'Precision',v:`${mlStats.precision}%`,c:colors.neonCyan,i:<CheckCircleIcon/>},{t:'Recall',v:`${mlStats.recall}%`,c:colors.neonLime,i:<SpeedIcon/>},{t:'Threshold',v:mlStats.threshold.toFixed(4),c:colors.neonOrange,i:<TimerIcon/>},{t:'Analyzed',v:mlStats.total_analyzed,c:colors.primary,i:<DevicesIcon/>},{t:'Anomalies',v:mlStats.anomalies_detected,c:colors.error,i:<WarningIcon/>},{t:'Revocations',v:mlStats.revocations,c:colors.errorLight,i:<ErrorIcon/>},{t:'Model',v:'V4 LSTM',c:colors.neonPink,i:<MemoryIcon/>}].map((k,i)=><Grid item xs={12} sm={6} md={3} lg={1.5} key={i}><KPICard title={k.t} value={k.v} subtitle="" icon={k.i} color={k.c} delay={i*0.05}/></Grid>)}
                </Grid>
                <Grid container spacing={3} sx={{ mb:3 }}>
                  <Grid item xs={12} lg={8}>
                    <LargeChartCard title="AI Anomaly Detection (Real-time MSE)" icon={<AiIcon/>} color={colors.error}>
                      <Line data={{ labels: mlAnomalies.slice(-20).map(a => { try { return format(safeDate(a.timestamp), 'HH:mm:ss'); } catch { return '--:--:--'; } }), datasets: [{ label: 'Anomaly Score', data: mlAnomalies.slice(-20).map(a => Number(a.score)||0), borderColor: colors.error, backgroundColor: (ctx) => { const g = ctx.chart.ctx.createLinearGradient(0,0,0,400); g.addColorStop(0,alpha(colors.error,0.3)); g.addColorStop(1,alpha(colors.error,0)); return g; }, fill: true, tension:0.4, borderWidth:3 }] }} options={{ ...latencyOptions, plugins: { legend: { display: true, labels: { color: colors.textSecondary } } } }} />
                    </LargeChartCard>
                  </Grid>
                  <Grid item xs={12} lg={4}>
                    <LargeChartCard title="Model Performance" icon={<BubbleIcon/>} color={colors.secondary}>
                      <Box sx={{ display:'flex',alignItems:'center',justifyContent:'center',height:'100%' }}><Box sx={{ width:'90%',maxWidth:350 }}><Radar data={{ labels:['F1','Precision','Recall','Speed','Accuracy','Score'], datasets:[{ data:[mlStats.f1_score,mlStats.precision,mlStats.recall,92,88,85], backgroundColor:alpha(colors.secondary,0.3), borderColor:colors.secondary, borderWidth:2 }] }} options={{ responsive:true,maintainAspectRatio:false,plugins:{legend:{display:false}},scales:{r:{beginAtZero:true,max:100,grid:{color:alpha('#fff',0.1)},ticks:{color:colors.textMuted},pointLabels:{color:colors.textSecondary,font:{size:10}}}} }} /></Box></Box>
                    </LargeChartCard>
                  </Grid>
                </Grid>
                <Grid container spacing={3}>
                  <Grid item xs={12} lg={6}>
                    <Card sx={{ borderRadius:3, bgcolor:alpha(colors.card,0.8), border:`1px solid ${alpha(colors.error,0.2)}`, p:2 }}>
                      <Typography variant="h6" sx={{ color:colors.textPrimary, mb:2, display:'flex',alignItems:'center',gap:1 }}><WarningIcon sx={{ color:colors.error }}/>Anomaly History ({mlAnomalies.length})</Typography>
                      <TableContainer sx={{ maxHeight:300 }}><Table size="small"><TableHead><TableRow>{['Time','Device','Score','Severity','Status'].map(h=><TableCell key={h} sx={{ color:colors.textSecondary }}>{h}</TableCell>)}</TableRow></TableHead><TableBody>{mlAnomalies.slice(0,10).map((a,i)=>(
                        <TableRow key={a.id||i} hover><TableCell sx={{ color:colors.textPrimary }}>{a.timestamp?format(safeDate(a.timestamp),'HH:mm:ss'):'N/A'}</TableCell><TableCell><Chip size="small" label={safeStr(a.deviceAddress||'').substring(0,15)} sx={{ bgcolor:alpha(colors.primary,0.2), color:colors.primaryLight }}/></TableCell><TableCell sx={{ color:colors.warning, fontWeight:600 }}>{Number(a.score).toFixed(3)}</TableCell><TableCell><Chip size="small" label={a.severity||'LOW'} sx={{ bgcolor:a.severity==='CRITICAL'?alpha(colors.error,0.3):a.severity==='HIGH'?alpha(colors.warning,0.3):alpha(colors.success,0.2), color:a.severity==='CRITICAL'?colors.error:a.severity==='HIGH'?colors.warning:colors.success }}/></TableCell><TableCell><Box sx={{ width:10,height:10,borderRadius:'50%',bgcolor:a.status==='anomaly'?colors.error:colors.success,display:'inline-block',mr:1 }}/>{a.status||'normal'}</TableCell></TableRow>
                      ))}</TableBody></Table></TableContainer>
                    </Card>
                  </Grid>
                  <Grid item xs={12} lg={6}>
                    <Card sx={{ borderRadius:3, bgcolor:alpha(colors.card,0.8), border:`1px solid ${alpha(colors.warning,0.2)}`, p:2 }}>
                      <Typography variant="h6" sx={{ color:colors.textPrimary, mb:2, display:'flex',alignItems:'center',gap:1 }}><HistoryIcon sx={{ color:colors.warning }}/>AI Action Log ({mlActionLog.length})</Typography>
                      <TableContainer sx={{ maxHeight:300 }}><Table size="small"><TableHead><TableRow>{['Time','Device','Action','Score'].map(h=><TableCell key={h} sx={{ color:colors.textSecondary }}>{h}</TableCell>)}</TableRow></TableHead><TableBody>{mlActionLog.slice(0,10).map((a,i)=>(
                        <TableRow key={a.id||i} hover><TableCell sx={{ color:colors.textPrimary }}>{a.timestamp?format(safeDate(a.timestamp),'HH:mm:ss'):'N/A'}</TableCell><TableCell><Chip size="small" label={safeStr(a.deviceAddress||'').substring(0,15)} sx={{ bgcolor:alpha(colors.primary,0.2), color:colors.primaryLight }}/></TableCell><TableCell><Chip size="small" label={a.action||'MONITOR'} sx={{ bgcolor:a.action==='REVOKE'||a.action==='IMMEDIATE_REVOKE'?alpha(colors.error,0.3):alpha(colors.warning,0.2), color:a.action==='REVOKE'||a.action==='IMMEDIATE_REVOKE'?colors.error:colors.warning }}/></TableCell><TableCell sx={{ color:colors.warning, fontWeight:600 }}>{Number(a.score).toFixed(2)}</TableCell></TableRow>
                      ))}</TableBody></Table></TableContainer>
                    </Card>
                  </Grid>
                </Grid>
              </Box>
            )}
            {activeSection === 'transactions' && (
              <LargeChartCard title="Transactions" icon={<TimelineIcon/>} color={colors.info}>
                <Box sx={{ height:'100%',display:'flex',flexDirection:'column' }}><TableContainer sx={{ flex:1 }}><Table stickyHeader size="small"><TableHead><TableRow>{['Hash','Device','ECDSA','ZKP','Latency','Time'].map(h=><TableCell key={h} sx={{ color:colors.textPrimary,fontWeight:600 }}>{h}</TableCell>)}</TableRow></TableHead><TableBody>{transactions.slice(0,15).map((tx,i)=>(<TableRow key={tx.id||i} hover><TableCell><code style={{ color:colors.textSecondary }}>{safeStr(tx.hash||'').substring(0,12)}...</code></TableCell><TableCell sx={{ color:colors.textPrimary }}>{safeStr(tx.deviceId||'').substring(0,15)}</TableCell><TableCell>{tx.ecdsaValid?<CheckCircleIcon sx={{ color:colors.success }} fontSize="small"/>:<ErrorIcon sx={{ color:colors.error }} fontSize="small"/>}</TableCell><TableCell>{tx.zkpValid?<VerifiedIcon sx={{ color:colors.success }} fontSize="small"/>:<ErrorIcon sx={{ color:colors.error }} fontSize="small"/>}</TableCell><TableCell sx={{ color:tx.latency>200?colors.warning:colors.textPrimary }}>{tx.latency}ms</TableCell><TableCell sx={{ color:colors.textSecondary }}>{formatDistanceToNow(safeDate(tx.timestamp),{addSuffix:true,locale:enUS})}</TableCell></TableRow>))}</TableBody></Table></TableContainer></Box>
              </LargeChartCard>
            )}
            {activeSection === 'security' && (
              <Grid container spacing={3}>
                <Grid item xs={12} lg={6}><LargeChartCard title="Security" icon={<SecurityIcon/>} color={colors.primary}><Box sx={{ display:'flex',flexDirection:'column',height:'100%' }}><Box sx={{ flex:1,display:'flex',alignItems:'center',justifyContent:'center' }}><Box sx={{ width:'70%',maxWidth:280 }}><Doughnut data={{ labels:['ECDSA OK','ECDSA Fail','ZKP OK','ZKP Fail'], datasets:[{ data:[realtimeStats.ecdsaSuccess,Math.max(0,realtimeStats.totalMessages-realtimeStats.ecdsaSuccess),realtimeStats.zkpSuccess,Math.max(0,realtimeStats.totalMessages-realtimeStats.zkpSuccess)], backgroundColor:[colors.success,alpha(colors.error,0.5),colors.info,alpha(colors.warning,0.5)] }] }} options={{ responsive:true,maintainAspectRatio:false,cutout:'65%' }}/></Box></Box><Stack spacing={2} sx={{ mt:2 }}><Box><Box display="flex" justifyContent="space-between" mb={1}><Typography sx={{ color:colors.textPrimary }}>ECDSA</Typography><Typography sx={{ color:colors.success,fontWeight:700 }}>{Math.round((realtimeStats.ecdsaSuccess/Math.max(1,realtimeStats.totalMessages))*100)}%</Typography></Box><LinearProgress variant="determinate" value={(realtimeStats.ecdsaSuccess/Math.max(1,realtimeStats.totalMessages))*100} sx={{ height:8,borderRadius:4,bgcolor:alpha('#fff',0.05),'& .MuiLinearProgress-bar':{bgcolor:colors.success,borderRadius:4}}}/></Box><Box><Box display="flex" justifyContent="space-between" mb={1}><Typography sx={{ color:colors.textPrimary }}>ZKP</Typography><Typography sx={{ color:colors.info,fontWeight:700 }}>{Math.round((realtimeStats.zkpSuccess/Math.max(1,realtimeStats.totalMessages))*100)}%</Typography></Box><LinearProgress variant="determinate" value={(realtimeStats.zkpSuccess/Math.max(1,realtimeStats.totalMessages))*100} sx={{ height:8,borderRadius:4,bgcolor:alpha('#fff',0.05),'& .MuiLinearProgress-bar':{bgcolor:colors.info,borderRadius:4}}}/></Box></Stack></Box></LargeChartCard></Grid>
                <Grid item xs={12} lg={6}><LargeChartCard title="Blockchain" icon={<BlockIcon/>} color={colors.warning}><Stack spacing={3} sx={{ height:'100%',justifyContent:'center' }}><Box sx={{ textAlign:'center',p:2,borderRadius:3,bgcolor:alpha(colors.darker,0.5),border:`1px solid ${alpha(colors.neon,0.2)}` }}><Typography variant="h2" fontWeight="bold" sx={{ color:colors.neon,textShadow:`0 0 20px ${alpha(colors.neon,0.5)}` }}>#{blockchainInfo.latestBlock?.toLocaleString()||0}</Typography><Typography sx={{ color:colors.textSecondary,textTransform:'uppercase',letterSpacing:2,fontSize:'0.75rem' }}>Latest Block</Typography></Box><Grid container spacing={2}>{[{l:'Network ID',v:blockchainInfo.networkId,c:colors.info},{l:'Gas',v:`${blockchainInfo.gasPrice} wei`,c:colors.warning},{l:'Peers',v:blockchainInfo.peers||12,c:colors.success},{l:'Status',v:blockchainInfo.syncing?'Syncing':'Synced',c:blockchainInfo.syncing?colors.warning:colors.success}].map((it,i)=>(<Grid item xs={6} key={i}><Card sx={{ bgcolor:alpha(colors.darker,0.6),p:2,borderRadius:3,border:`1px solid ${alpha(it.c,0.2)}` }}><Typography sx={{ color:colors.textSecondary,fontSize:'0.7rem',textTransform:'uppercase' }}>{it.l}</Typography><Typography variant="h6" fontWeight="bold" sx={{ color:it.c }}>{it.v}</Typography></Card></Grid>))}</Grid></Stack></LargeChartCard></Grid>
              </Grid>
            )}
          </motion.div></AnimatePresence>
        </Container>
      </Box>

      <Dialog open={addNodeOpen} onClose={()=>setAddNodeOpen(false)} maxWidth="sm" fullWidth PaperProps={{ sx:{bgcolor:colors.card,border:`1px solid ${alpha('#fff',0.1)}`,borderRadius:3} }}>
        <DialogTitle sx={{ bgcolor:alpha(colors.darker,0.5), color:colors.textPrimary }}>Add Geth Nodes<IconButton onClick={()=>setAddNodeOpen(false)} sx={{ position:'absolute',right:8,top:8,color:colors.textSecondary }}><CloseIcon/></IconButton></DialogTitle>
        <DialogContent sx={{ bgcolor:alpha(colors.darker,0.5), pt:3 }}><Stack spacing={3}>
          <FormControl fullWidth><InputLabel>Node Type</InputLabel><Select value={isValidator} onChange={e=>{setIsValidator(e.target.value);e.target.value?(setNumValidators(1),setNumFullnodes(0)):(setNumValidators(0),setNumFullnodes(1));}}><MenuItem value={true}>Validator</MenuItem><MenuItem value={false}>Full Node</MenuItem></Select></FormControl>
          {isValidator ? <FormControl fullWidth><InputLabel>Number</InputLabel><Select value={numValidators} onChange={e=>setNumValidators(e.target.value)}>{[1,2,3,4,5].map(n=><MenuItem key={n} value={n}>{n}</MenuItem>)}</Select></FormControl> : <FormControl fullWidth><InputLabel>Number</InputLabel><Select value={numFullnodes} onChange={e=>setNumFullnodes(e.target.value)}>{[1,2,3,4,5].map(n=><MenuItem key={n} value={n}>{n}</MenuItem>)}</Select></FormControl>}
        </Stack></DialogContent>
        <DialogActions sx={{ bgcolor:alpha(colors.darker,0.5), px:3, pb:2 }}><Button onClick={()=>setAddNodeOpen(false)}>Cancel</Button><Button variant="contained" disabled={scriptRunning} sx={{ bgcolor:colors.primary }}>OK</Button></DialogActions>
      </Dialog>

      <Snackbar open={errorOpen} autoHideDuration={6000} onClose={()=>setErrorOpen(false)} anchorOrigin={{ vertical:'bottom',horizontal:'right' }}>
        <Alert severity={errorMessage.includes('✅')?'success':'error'} variant="filled" sx={{ bgcolor:errorMessage.includes('✅')?colors.success:colors.error,borderRadius:2 }}>{errorMessage}</Alert>
      </Snackbar>
      <style>{`@keyframes pulse{0%,100%{opacity:1}50%{opacity:0.5}}@keyframes shimmer{0%{transform:translateX(-100%)}100%{transform:translateX(100%)}}`}</style>
    </Box>
  );
};

export default ModernBlockchainDashboard;
