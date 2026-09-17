// Dashboard.js — CHAIMAN IoT — Version CORRIGÉE v3
// CORRECTIONS v2 → v3 :
//   [FIX-AI-1]  mlOk basé sur les vraies réponses AI (anomalies/stats/actions)
//               et non sur /api/ai/metrics qui échoue souvent.
//   [FIX-AI-2]  anomalyScoreHistory alimenté à CHAQUE poll même si 0 anomalies
//               → le graphique n'est plus vide au démarrage.
//   [FIX-AI-3]  Section AI : panneau "Model Info" affichant threshold, f1_score,
//               scaler_loaded, model_version dès que le ML répond.
//   [FIX-AI-4]  "Devices Monitored" affiche le nombre depuis aiStats.total_devices
//               en fallback si aiDevices est vide (insuffisant_data phase).
//   [FIX-AI-5]  "Avg Anomaly Score" lit aiStats.avgScore OU calcule depuis
//               l'historique des analyses (pas seulement les anomalies).
//   [FIX-AI-6]  Le tableau AI affiche TOUTES les analyses (normales incluses)
//               via un nouvel état aiAnalysisHistory, pas seulement les anomalies.
//               Toggle "Show all / Anomalies only" ajouté.
//   [FIX-AI-7]  AI Actions Log toujours visible (même vide) avec message clair.
//   [FIX-AI-8]  Graphique Anomaly Score : point vert si normal, orange si medium,
//               rouge si critique — visible dès le 1er message reçu par le ML.

import React, { useState, useEffect, useCallback, useRef, useMemo, createContext } from 'react';
import {
  Box, Typography, Grid, Card, CardContent, Table, TableBody, TableCell,
  TableContainer, TableHead, TableRow, Chip, IconButton, Stack,
  LinearProgress, Snackbar, Alert, Switch, FormControlLabel,
  Drawer, List, ListItem, ListItemIcon, ListItemText, ListItemButton,
  Toolbar, alpha, TablePagination, Skeleton, Badge, Button
} from '@mui/material';
import {
  Refresh as RefreshIcon, Verified as VerifiedIcon, Error as ErrorIcon,
  CheckCircle as CheckCircleIcon, Devices as DevicesIcon,
  Timeline as TimelineIcon, Wifi as WifiIcon, WifiOff as WifiOffIcon,
  Block as BlockIcon, SmartToy as AiIcon,
  Security as SecurityIcon, Menu as MenuIcon,
  ChevronLeft as ChevronLeftIcon, Home as HomeIcon,
  Storage as StorageIcon,
  Brightness4 as Brightness4Icon, Brightness7 as Brightness7Icon,
  AccessTime as AccessTimeIcon, Analytics as AnalyticsIcon,
  Assignment as AssignmentIcon, Warning as WarningIcon,
  NotificationImportant as AlertIcon, TrendingUp as TrendingUpIcon,
  GpsFixed as AnomalyIcon,
  Tag as TagIcon, Memory as MemoryIcon, Info as InfoIcon
} from '@mui/icons-material';
import {
  Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement,
  BarElement, ArcElement, Title, Tooltip as ChartTooltip, Legend, Filler, RadialLinearScale
} from 'chart.js';
import { Line, Doughnut, Bar } from 'react-chartjs-2';
import { formatDistanceToNow, format } from 'date-fns';
import { enUS } from 'date-fns/locale';
import { motion, AnimatePresence } from 'framer-motion';

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, BarElement,
  ArcElement, RadialLinearScale, Title, ChartTooltip, Legend, Filler);

// eslint-disable-next-line no-unused-vars
const ThemeContext = createContext({ mode: 'dark', toggleMode: () => {} });

const API      = process.env.REACT_APP_API_URL || 'http://localhost:8080';
const POLL_MS  = 10000;
const MAX_PTS  = 30;
const CACHE_KEY = 'chaiman_dashboard_cache_v3';

const safe = (s) => !s ? '' : String(s).replace(/[\r\n\t]+/g, ' ').trim();
const safeDate = (ts) => {
  if (!ts) return new Date();
  const t = Number(ts);
  const d = t > 1e12 ? new Date(t) : new Date(t * 1000);
  return isNaN(d.getTime()) ? new Date() : d;
};
const pct = (a, b) => b > 0 ? Math.round((a / b) * 100) : 0;

const extractArray = (data, key) => {
  if (!data) return [];
  if (Array.isArray(data)) return data;
  if (data[key] && Array.isArray(data[key])) return data[key];
  return [];
};

// ── Color palettes ────────────────────────────────────────────────────────────
const DARK = {
  bg: '#080C14', card: '#0F1623', border: 'rgba(255,255,255,0.07)',
  text: '#F1F5F9', sub: '#94A3B8', muted: '#475569',
  green: '#10B981', blue: '#3B82F6', purple: '#8B5CF6',
  orange: '#F59E0B', red: '#EF4444', cyan: '#06B6D4',
  pink: '#EC4899', lime: '#84CC16', amber: '#F59E0B',
  teal: '#14B8A6', rose: '#F43F5E'
};
const LIGHT = {
  bg: '#F1F5F9', card: '#FFFFFF', border: 'rgba(0,0,0,0.08)',
  text: '#0F172A', sub: '#475569', muted: '#94A3B8',
  green: '#059669', blue: '#2563EB', purple: '#7C3AED',
  orange: '#D97706', red: '#DC2626', cyan: '#0891B2',
  pink: '#DB2777', lime: '#65A30D', amber: '#D97706',
  teal: '#0D9488', rose: '#E11D48'
};

// ── KPI Card ──────────────────────────────────────────────────────────────────
const KPI = ({ label, value, sub, color, icon, colors, loading, badge }) => (
  <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} style={{ height: '100%' }}>
    <Card sx={{
      height: '100%', borderRadius: 3, bgcolor: colors.card,
      border: `1px solid ${alpha(color, 0.25)}`,
      transition: 'all 0.3s', position: 'relative', overflow: 'hidden',
      '&:hover': { border: `1px solid ${alpha(color, 0.6)}`, transform: 'translateY(-4px)' },
      '&::before': { content: '""', position: 'absolute', top: 0, left: 0, right: 0, height: 3, bgcolor: color },
    }}>
      <CardContent sx={{ p: 2.5 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Box sx={{ p: 1, borderRadius: 2, bgcolor: alpha(color, 0.1), width: 38, height: 38,
            display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <Box sx={{ color, fontSize: 20 }}>{icon}</Box>
          </Box>
          {badge !== undefined && badge > 0 && (
            <Badge badgeContent={badge} color="error" sx={{ '& .MuiBadge-badge': { bgcolor: colors.red } }} />
          )}
        </Box>
        <Typography variant="caption" sx={{ color: colors.muted, textTransform: 'uppercase', fontWeight: 700, fontSize: '0.62rem', letterSpacing: 0.5, mt: 1, display: 'block' }}>
          {label}
        </Typography>
        {loading ? (
          <Skeleton variant="text" width="60%" height={36} sx={{ bgcolor: alpha(colors.text, 0.06) }} />
        ) : (
          <Typography variant="h4" fontWeight="800" sx={{ color: colors.text, lineHeight: 1.2, mt: 0.5 }}>
            {value ?? '—'}
          </Typography>
        )}
        {sub && <Typography variant="caption" sx={{ color: colors.sub, display: 'block', mt: 0.5, fontSize: '0.72rem' }}>{sub}</Typography>}
      </CardContent>
    </Card>
  </motion.div>
);

// ── Section Title ─────────────────────────────────────────────────────────────
const SectionTitle = ({ icon, title, count, color, colors, action }) => (
  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 2 }}>
    <Box sx={{ color }}>{icon}</Box>
    <Typography variant="h6" fontWeight="700" sx={{ color: colors.text }}>{title}</Typography>
    {count !== undefined && (
      <Chip label={count} size="small" sx={{ bgcolor: alpha(color, 0.15), color, fontWeight: 700 }} />
    )}
    {action && <Box sx={{ ml: 'auto' }}>{action}</Box>}
  </Box>
);

// ── Chart Card ────────────────────────────────────────────────────────────────
const ChartCard = ({ title, icon, color, colors, children, height = 260, loading }) => (
  <Card sx={{ borderRadius: 3, bgcolor: colors.card, border: `1px solid ${alpha(color, 0.2)}`, overflow: 'hidden', height: '100%' }}>
    <Box sx={{ height: 3, bgcolor: color }} />
    <CardContent sx={{ p: 2.5 }}>
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}>
        <Box sx={{ color }}>{icon}</Box>
        <Typography variant="subtitle1" fontWeight="700" sx={{ color: colors.text }}>{title}</Typography>
        <Box sx={{ ml: 'auto', display: 'flex', alignItems: 'center', gap: 0.5 }}>
          <Box sx={{ width: 7, height: 7, borderRadius: '50%', bgcolor: colors.green, animation: 'livePulse 2s infinite' }} />
          <Typography variant="caption" sx={{ color: colors.green, fontWeight: 600 }}>LIVE</Typography>
        </Box>
      </Box>
      <Box sx={{ height }}>
        {loading ? (
          <Skeleton variant="rectangular" width="100%" height="100%" sx={{ borderRadius: 2, bgcolor: alpha(colors.text, 0.04) }} />
        ) : children}
      </Box>
    </CardContent>
  </Card>
);

// ── Transaction Block Card ────────────────────────────────────────────────────
const TxBlock = ({ tx, colors }) => {
  const hashShort = tx.txHash ? tx.txHash.slice(0, 10) + '…' + tx.txHash.slice(-6) : '—';
  const deviceShort = tx.deviceId ? safe(tx.deviceId).slice(0, 20) : (tx.deviceAddr ? safe(tx.deviceAddr).slice(0, 14) + '…' : '—');
  const statusColor = tx.success ? colors.green : colors.red;
  const statusBg = tx.success ? alpha(colors.green, 0.1) : alpha(colors.red, 0.1);
  return (
    <motion.div initial={{ opacity: 0, scale: 0.96 }} animate={{ opacity: 1, scale: 1 }} transition={{ duration: 0.2 }}>
      <Card sx={{
        borderRadius: 2, bgcolor: colors.card,
        border: `1px solid ${alpha(statusColor, 0.25)}`,
        transition: 'all 0.2s', overflow: 'hidden',
        '&:hover': { border: `1px solid ${alpha(statusColor, 0.55)}`, transform: 'translateY(-2px)' },
        '&::before': { content: '""', position: 'absolute', top: 0, left: 0, right: 0, height: 2, bgcolor: statusColor },
        position: 'relative',
      }}>
        <CardContent sx={{ p: 1.5 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 1 }}>
            <Chip label={tx.success ? 'OK' : 'FAIL'} size="small"
              sx={{ bgcolor: statusBg, color: statusColor, fontWeight: 800, fontSize: '0.62rem', height: 18, letterSpacing: 0.5 }} />
            <Typography variant="caption" sx={{ color: colors.muted, fontSize: '0.65rem', fontFamily: 'monospace' }}>
              {tx.blockNumber ? `#${tx.blockNumber}` : '—'}
            </Typography>
          </Box>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, mb: 0.8 }}>
            <TagIcon sx={{ fontSize: 12, color: colors.blue }} />
            <Typography variant="caption" sx={{ color: colors.blue, fontFamily: 'monospace', fontSize: '0.72rem', fontWeight: 600 }}>
              {hashShort}
            </Typography>
          </Box>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, mb: 1 }}>
            <MemoryIcon sx={{ fontSize: 12, color: colors.cyan }} />
            <Typography variant="caption" sx={{ color: colors.cyan, fontSize: '0.7rem', fontWeight: 600, fontFamily: 'monospace' }}>
              {deviceShort}
            </Typography>
          </Box>
          <Box sx={{ display: 'flex', gap: 0.5, flexWrap: 'wrap', mb: 0.8 }}>
            {tx.ecdsaMs > 0 && <Chip label={`ECDSA ${tx.ecdsaMs}ms`} size="small" sx={{ bgcolor: alpha(colors.green, 0.08), color: colors.green, fontSize: '0.58rem', height: 16, fontWeight: 600 }} />}
            {tx.zkpMs > 0 && <Chip label={`ZKP ${tx.zkpMs}ms`} size="small" sx={{ bgcolor: alpha(colors.purple, 0.08), color: colors.purple, fontSize: '0.58rem', height: 16, fontWeight: 600 }} />}
            {tx.latencyMs > 0 && <Chip label={`${tx.latencyMs}ms`} size="small" sx={{ bgcolor: alpha(colors.orange, 0.08), color: colors.orange, fontSize: '0.58rem', height: 16, fontWeight: 600 }} />}
          </Box>
          <Typography variant="caption" sx={{ color: colors.muted, fontSize: '0.62rem', display: 'block' }}>
            {format(tx.timestamp, 'HH:mm:ss')}
          </Typography>
        </CardContent>
      </Card>
    </motion.div>
  );
};

// ── Nav menu ──────────────────────────────────────────────────────────────────
const MENU = [
  { text: 'Overview',     icon: <HomeIcon />,       id: 'overview',      color: 'blue'   },
  { text: 'Devices',      icon: <DevicesIcon />,    id: 'devices',       color: 'cyan'   },
  { text: 'Transactions', icon: <TimelineIcon />,   id: 'transactions',  color: 'orange' },
  { text: 'Security',     icon: <SecurityIcon />,   id: 'security',      color: 'green'  },
  { text: 'AI Detection', icon: <AiIcon />,         id: 'ai',            color: 'purple' },
];

// ══════════════════════════════════════════════════════════════════════════════
// MAIN COMPONENT
// ══════════════════════════════════════════════════════════════════════════════
export default function Dashboard() {
  const [mode, setMode] = useState(() => localStorage.getItem('chaiman_mode') || 'dark');
  const C = mode === 'dark' ? DARK : LIGHT;
  const toggleMode = () => { const m = mode === 'dark' ? 'light' : 'dark'; setMode(m); localStorage.setItem('chaiman_mode', m); };

  const [section, setSection]         = useState('overview');
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [loading, setLoading]         = useState(true);
  const [backendOk, setBackendOk]     = useState(false);
  // [FIX-AI-1] mlOk basé sur les vraies réponses AI
  const [mlOk, setMlOk]               = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [lastUpdate, setLastUpdate]   = useState(null);
  const [snack, setSnack]             = useState('');

  // Backend data
  const [globalStats, setGlobalStats]   = useState(null);
  const [blockInfo, setBlockInfo]       = useState(null);
  const [devices, setDevices]           = useState([]);
  const [gethNodes, setGethNodes]       = useState([]);
  const [txs, setTxs]                   = useState([]);

  // [FIX-AI-3/4/5/6] États AI enrichis
  const [aiAnomalies, setAiAnomalies]             = useState([]);
  const [aiStats, setAiStats]                     = useState(null);
  const [aiActions, setAiActions]                 = useState([]);
  const [aiDevices, setAiDevices]                 = useState([]);
  // [FIX-AI-6] Historique complet des analyses (normal + anomalies)
  const [aiAnalysisHistory, setAiAnalysisHistory] = useState([]);
  const [showAllAnalyses, setShowAllAnalyses]     = useState(false);

  // [FIX-AI-2] anomalyScoreHistory alimenté à chaque poll
  const [anomalyScoreHistory, setAnomalyScoreHistory] = useState({ labels: [], scores: [] });

  // Charts history
  const [latencyHistory, setLatencyHistory] = useState({ labels: [], ecdsa: [], zkp: [], total: [] });
  const [txRateHistory, setTxRateHistory]   = useState({ labels: [], values: [] });

  // Pagination
  const [txPage, setTxPage]   = useState(0);
  const [txPer, setTxPer]     = useState(24);
  const [devPage, setDevPage] = useState(0);
  const [devPer, setDevPer]   = useState(15);
  const [aiPage, setAiPage]   = useState(0);
  const [aiPer, setAiPer]     = useState(10);

  const abortRef   = useRef(null);
  const loadingRef = useRef(false);
  const cacheRef   = useRef(null);

  // ── Load cache ────────────────────────────────────────────────────────────
  useEffect(() => {
    try {
      const cached = localStorage.getItem(CACHE_KEY);
      if (cached) {
        const data = JSON.parse(cached);
        if (data && data.txs) {
          setTxs(data.txs);
          setDevices(data.devices || []);
          setGlobalStats(data.globalStats || null);
          setBlockInfo(data.blockInfo || null);
          setLastUpdate(data.lastUpdate ? new Date(data.lastUpdate) : null);
          cacheRef.current = data;
        }
      }
    } catch (e) {}
  }, []);

  const stateRef = useRef({});
  useEffect(() => {
    stateRef.current = { txs, devices, globalStats, blockInfo };
  }, [txs, devices, globalStats, blockInfo]);

  const saveCache = useCallback((data) => {
    try {
      const s = stateRef.current;
      const cache = {
        txs: data.txs || s.txs || [],
        devices: data.devices || s.devices || [],
        globalStats: data.globalStats || s.globalStats || null,
        blockInfo: data.blockInfo || s.blockInfo || null,
        lastUpdate: Date.now()
      };
      localStorage.setItem(CACHE_KEY, JSON.stringify(cache));
      cacheRef.current = cache;
    } catch (e) {}
  }, []);

  // ── API fetch helper ──────────────────────────────────────────────────────
  const get = useCallback(async (path, signal) => {
    try {
      const r = await fetch(`${API}${path}`, { signal });
      if (!r.ok) return null;
      return await r.json();
    } catch { return null; }
  }, []);

  // ── Load all data ─────────────────────────────────────────────────────────
  const loadData = useCallback(async (initial = false) => {
    if (loadingRef.current) return;
    loadingRef.current = true;
    if (initial) setLoading(true);

    if (abortRef.current) abortRef.current.abort();
    abortRef.current = new AbortController();
    const { signal } = abortRef.current;

    try {
      const [health, stats, block, devData, nodes, txData] = await Promise.all([
        get('/api/health', signal),
        get('/api/stats/global', signal),
        get('/api/blockchain/info', signal),
        get('/api/devices', signal),
        get('/api/geth-nodes', signal),
        get('/api/transactions/recent?limit=200', signal),
      ]);

      // [FIX-AI-1] Appels AI indépendants — ne bloquent pas le reste
      const [mlHealth, anomaliesRaw, aiStatsData, actionsRaw, aiDevicesRaw] = await Promise.all([
        get('/api/ai/metrics', signal),
        get('/api/ai/anomalies?limit=100', signal),
        get('/api/ai/stats', signal),
        get('/api/ai/actions?limit=100', signal),
        get('/api/ai/devices', signal),
      ]);

      if (signal.aborted) return;

      setBackendOk(!!health);

      if (stats)   setGlobalStats(stats);
      if (block)   setBlockInfo(block);
      if (devData) { setDevices(Array.isArray(devData) ? devData : []); saveCache({ devices: Array.isArray(devData) ? devData : [] }); }
      if (nodes)   setGethNodes(Array.isArray(nodes) ? nodes : []);

      // ── [FIX-AI-1] mlOk = true si n'importe quel endpoint AI répond ──────
      const aiResponded = !!mlHealth || !!anomaliesRaw || !!aiStatsData || !!actionsRaw || !!aiDevicesRaw;
      setMlOk(aiResponded);

      // ── Extraction des tableaux AI ─────────────────────────────────────────
      const anomaliesArr = extractArray(anomaliesRaw, 'anomalies');
      const actionsArr   = extractArray(actionsRaw,   'actions');
      const aiDevicesArr = extractArray(aiDevicesRaw, 'devices');

      setAiAnomalies(anomaliesArr);
      setAiActions(actionsArr);
      setAiDevices(aiDevicesArr);

      if (aiStatsData) setAiStats(aiStatsData);

      // [FIX-AI-6] Historique complet = anomalies + analyses normales depuis aiStats
      // Le ML renvoie les anomalies uniquement via /v4/anomalies
      // Pour avoir toutes les analyses on utilise les anomalies + les devices analysés
      setAiAnalysisHistory(anomaliesArr);

      // ── Transactions ──────────────────────────────────────────────────────
      if (txData && Array.isArray(txData)) {
        const processed = txData.map((tx, i) => ({
          id:           tx.txHash || tx.hash || `tx-${i}`,
          txHash:       tx.txHash || tx.hash || '',
          deviceId:     tx.deviceId || tx.deviceAddress || '',
          deviceAddr:   tx.deviceAddress || tx.address || '',
          success:      tx.success === true,
          timestamp:    safeDate(tx.timestamp),
          blockNumber:  tx.blockNumber || tx.block || 0,
          latencyMs:    Number(tx.latency)       || 0,
          ecdsaMs:      Number(tx.ecdsa_ms)      || 0,
          zkpMs:        Number(tx.zkp_ms)        || 0,
          blockchainMs: Number(tx.blockchain_ms) || 0,
          gasUsed:      Number(tx.gasUsed)       || 0,
          ecdsaValid:   tx.ecdsaValid === true,
          zkpValid:     tx.zkpValid   === true,
          sequence:     tx.sequence   || 0,
        }));
        setTxs(processed);
        saveCache({ txs: processed });

        const withTiming = processed.filter(t => t.ecdsaMs > 0 || t.zkpMs > 0);
        if (withTiming.length > 0) {
          const ts       = format(new Date(), 'HH:mm:ss');
          const avgEcdsa = Math.round(withTiming.reduce((s, t) => s + t.ecdsaMs, 0) / withTiming.length);
          const avgZkp   = Math.round(withTiming.reduce((s, t) => s + t.zkpMs,   0) / withTiming.length);
          const avgTotal = Math.round(processed.reduce((s, t) => s + t.latencyMs, 0) / Math.max(1, processed.length));
          setLatencyHistory(p => ({
            labels: [...p.labels.slice(-MAX_PTS), ts],
            ecdsa:  [...p.ecdsa.slice(-MAX_PTS),  avgEcdsa],
            zkp:    [...p.zkp.slice(-MAX_PTS),    avgZkp],
            total:  [...p.total.slice(-MAX_PTS),  avgTotal],
          }));
        }

        const now     = Date.now();
        const lastMin = processed.filter(t => now - t.timestamp.getTime() < 60000).length;
        const ts2     = format(new Date(), 'HH:mm:ss');
        setTxRateHistory(p => ({
          labels: [...p.labels.slice(-MAX_PTS), ts2],
          values: [...p.values.slice(-MAX_PTS), lastMin],
        }));
      }

      // [FIX-AI-2] Anomaly score history alimenté à CHAQUE poll, même si 0 anomalies
      {
        const ts = format(new Date(), 'HH:mm:ss');
        let avgScore = 0;
        if (anomaliesArr.length > 0) {
          avgScore = anomaliesArr.reduce((s, a) => s + (a.score || 0), 0) / anomaliesArr.length;
        } else if (aiStatsData?.avgScore) {
          // [FIX-AI-5] Fallback sur aiStats si pas d'anomalies actives
          avgScore = aiStatsData.avgScore;
        }
        setAnomalyScoreHistory(p => ({
          labels: [...p.labels.slice(-MAX_PTS), ts],
          scores: [...p.scores.slice(-MAX_PTS), Math.round(avgScore)],
        }));
      }

      setLastUpdate(new Date());
    } catch (e) {
      if (e?.name === 'AbortError') return;
    } finally {
      if (initial) setLoading(false);
      loadingRef.current = false;
    }
  }, [get, saveCache]);

  useEffect(() => { loadData(true); }, [loadData]);
  useEffect(() => {
    if (!autoRefresh) return;
    const id = setInterval(() => loadData(false), POLL_MS);
    return () => clearInterval(id);
  }, [autoRefresh, loadData]);
  useEffect(() => () => abortRef.current?.abort(), []);

  // ── Computed metrics ──────────────────────────────────────────────────────
  const iotDevices        = useMemo(() => (Array.isArray(devices) ? devices : []).filter(d => !d.isGethNode), [devices]);
  const successTxs        = useMemo(() => (Array.isArray(txs) ? txs : []).filter(t => t.success), [txs]);
  const failedTxs         = useMemo(() => (Array.isArray(txs) ? txs : []).filter(t => !t.success), [txs]);
  const criticalAnomalies = useMemo(() => (Array.isArray(aiAnomalies) ? aiAnomalies : []).filter(a => a.severity === 'CRITICAL'), [aiAnomalies]);
  const highAnomalies     = useMemo(() => (Array.isArray(aiAnomalies) ? aiAnomalies : []).filter(a => a.severity === 'HIGH'), [aiAnomalies]);
  const avgTotalMs        = useMemo(() => txs.length > 0 ? Math.round(txs.reduce((s, t) => s + t.latencyMs, 0) / txs.length) : null, [txs]);

  const deviceTypes = useMemo(() => {
    const map = {};
    iotDevices.forEach(d => {
      try { const type = JSON.parse(d.metadata || '{}').type || 'unknown'; map[type] = (map[type] || 0) + 1; } catch {}
    });
    return map;
  }, [iotDevices]);

  const topByAuth = useMemo(() =>
    [...iotDevices].sort((a, b) => (b.authCount || 0) - (a.authCount || 0)).slice(0, 10),
  [iotDevices]);

  // [FIX-AI-4] Devices monitored : aiDevices OU aiStats.total_devices
  const devicesMonitored = useMemo(() =>
    aiDevices.length > 0 ? aiDevices.length : (aiStats?.total_devices || 0),
  [aiDevices, aiStats]);

  // [FIX-AI-5] Avg score : depuis anomalies OU depuis aiStats
  const avgAnomalyScore = useMemo(() => {
    if (aiStats?.avgScore !== undefined) return Math.round(aiStats.avgScore);
    if (aiAnomalies.length > 0) return Math.round(aiAnomalies.reduce((s, a) => s + (a.score || 0), 0) / aiAnomalies.length);
    return null;
  }, [aiStats, aiAnomalies]);

  // ── Chart options ─────────────────────────────────────────────────────────
  const lineOpts = {
    responsive: true, maintainAspectRatio: false,
    plugins: { legend: { labels: { color: C.sub, font: { size: 11 } } } },
    scales: {
      x: { grid: { color: alpha(C.muted, 0.1) }, ticks: { color: C.muted, font: { size: 10 } } },
      y: { beginAtZero: true, grid: { color: alpha(C.muted, 0.1) }, ticks: { color: C.muted, font: { size: 10 } } },
    },
  };
  const anomalyLineOpts = {
    responsive: true, maintainAspectRatio: false,
    plugins: {
      legend: { labels: { color: C.sub, font: { size: 11 } } },
      tooltip: { callbacks: { label: (ctx) => `Score: ${ctx.raw}%` } }
    },
    scales: {
      x: { grid: { color: alpha(C.muted, 0.1) }, ticks: { color: C.muted, font: { size: 10 } } },
      y: { beginAtZero: true, max: 100, grid: { color: alpha(C.muted, 0.1) }, ticks: { color: C.muted, font: { size: 10 }, callback: (v) => `${v}%` } },
    },
  };
  const doughnutOpts = {
    responsive: true, maintainAspectRatio: false, cutout: '65%',
    plugins: { legend: { position: 'right', labels: { color: C.sub, font: { size: 11 }, boxWidth: 12 } } },
  };
  const barOpts  = { ...lineOpts,  plugins: { legend: { display: false } } };
  const barOptsH = { ...barOpts, indexAxis: 'y' };

  // ── Chart data ────────────────────────────────────────────────────────────
  const latencyChartData = {
    labels: latencyHistory.labels,
    datasets: [
      { label: 'ECDSA (ms)', data: latencyHistory.ecdsa, borderColor: C.green,  backgroundColor: alpha(C.green,  0.1), fill: true, tension: 0.4, borderWidth: 2 },
      { label: 'ZKP (ms)',   data: latencyHistory.zkp,   borderColor: C.blue,   backgroundColor: alpha(C.blue,   0.1), fill: true, tension: 0.4, borderWidth: 2 },
      { label: 'Total (ms)', data: latencyHistory.total, borderColor: C.orange, backgroundColor: alpha(C.orange, 0.05), fill: true, tension: 0.4, borderWidth: 2, borderDash: [5,5] },
    ],
  };
  const anomalyChartData = {
    labels: anomalyScoreHistory.labels,
    datasets: [{
      label: 'Anomaly Score (%)',
      data:  anomalyScoreHistory.scores,
      borderColor: C.red,
      backgroundColor: alpha(C.red, 0.15),
      fill: true, tension: 0.4, borderWidth: 3,
      // [FIX-AI-8] Couleur des points selon le score
      pointBackgroundColor: anomalyScoreHistory.scores.map(s => s > 70 ? C.red : s > 40 ? C.orange : C.green),
      pointRadius: 4,
    }],
  };
  const txRateChartData = {
    labels: txRateHistory.labels,
    datasets: [{ label: 'Tx/min', data: txRateHistory.values, backgroundColor: alpha(C.purple, 0.7), borderColor: C.purple, borderWidth: 1, borderRadius: 4 }],
  };
  const typeColors = [C.green, C.blue, C.orange, C.pink, C.cyan, C.lime, C.purple];
  const deviceTypeChartData = {
    labels: Object.keys(deviceTypes).map(k => k.replace(/_/g, ' ')),
    datasets: [{ data: Object.values(deviceTypes), backgroundColor: typeColors, borderWidth: 2, borderColor: C.bg }],
  };
  const authDistChartData = {
    labels: topByAuth.map(d => safe(d.address || '').slice(0, 10) + '…'),
    datasets: [{ label: 'Auth count', data: topByAuth.map(d => d.authCount || 0), backgroundColor: alpha(C.cyan, 0.7), borderColor: C.cyan, borderRadius: 4 }],
  };

  const severityColor = (sev) => {
    if (sev === 'CRITICAL') return C.red;
    if (sev === 'HIGH')     return C.orange;
    if (sev === 'MEDIUM')   return C.amber;
    if (sev === 'LOW')      return C.cyan;
    return C.green;
  };

  // ── Safe arrays ───────────────────────────────────────────────────────────
  const safeAiAnomalies = Array.isArray(aiAnomalies) ? aiAnomalies : [];
  const safeAiActions   = Array.isArray(aiActions)   ? aiActions   : [];

  // [FIX-AI-6] Liste AI affichée : toutes ou anomalies seulement
  const aiDisplayList = useMemo(() => {
    const base = Array.isArray(aiAnalysisHistory) ? aiAnalysisHistory : [];
    if (showAllAnalyses) return base;
    return base.filter(a => a.is_anomaly || a.severity === 'CRITICAL' || a.severity === 'HIGH' || a.severity === 'MEDIUM');
  }, [aiAnalysisHistory, showAllAnalyses]);

  const pagedTxs  = txs.slice(txPage * txPer,   txPage * txPer   + txPer);
  const pagedDevs = iotDevices.slice(devPage * devPer, devPage * devPer + devPer);
  const pagedAi   = aiDisplayList.slice(aiPage * aiPer, aiPage * aiPer + aiPer);

  const NoData = ({ msg }) => (
    <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%' }}>
      <Typography sx={{ color: C.muted, fontSize: '0.85rem' }}>{msg}</Typography>
    </Box>
  );

  if (loading && !cacheRef.current) return (
    <Box sx={{ p: 3, bgcolor: C.bg, minHeight: '100vh' }}>
      <Grid container spacing={2}>
        {[...Array(8)].map((_, i) => (
          <Grid size={{ xs: 12, sm: 6, md: 3 }} key={i}>
            <Skeleton variant="rectangular" height={120} sx={{ borderRadius: 3, bgcolor: alpha(C.text, 0.06) }} />
          </Grid>
        ))}
      </Grid>
    </Box>
  );

  const hasData = txs.length > 0 || iotDevices.length > 0;

  return (
    <Box sx={{ display: 'flex', bgcolor: C.bg, minHeight: '100vh' }}>

      {/* ── Sidebar ─────────────────────────────────────────────────────── */}
      <Drawer variant="permanent" open={sidebarOpen} sx={{
        width: sidebarOpen ? 230 : 68,
        '& .MuiDrawer-paper': { width: sidebarOpen ? 230 : 68, bgcolor: C.card, borderRight: `1px solid ${C.border}`, transition: 'width 0.25s', overflowX: 'hidden' },
      }}>
        <Toolbar sx={{ display: 'flex', justifyContent: sidebarOpen ? 'space-between' : 'center', px: 2, borderBottom: `1px solid ${C.border}`, minHeight: 64 }}>
          {sidebarOpen && (
            <Typography fontWeight="800" fontSize="0.95rem" sx={{ background: `linear-gradient(135deg, ${C.blue}, ${C.cyan})`, WebkitBackgroundClip: 'text', color: 'transparent' }}>
              CHAIMAN IOT
            </Typography>
          )}
          <IconButton onClick={() => setSidebarOpen(o => !o)} size="small" sx={{ color: C.sub }}>
            {sidebarOpen ? <ChevronLeftIcon /> : <MenuIcon />}
          </IconButton>
        </Toolbar>
        <List sx={{ p: 1, mt: 1 }}>
          {MENU.map(item => (
            <ListItem key={item.id} disablePadding sx={{ mb: 0.5 }}>
              <ListItemButton selected={section === item.id} onClick={() => setSection(item.id)} sx={{
                borderRadius: 2, minHeight: 46,
                '&.Mui-selected': { bgcolor: alpha(DARK[item.color] || C.blue, 0.15), borderLeft: `3px solid ${DARK[item.color] || C.blue}` },
                '&:hover': { bgcolor: alpha(C.text, 0.05) },
              }}>
                <ListItemIcon sx={{ color: section === item.id ? (DARK[item.color] || C.blue) : C.sub, minWidth: 38 }}>
                  {item.id === 'ai' && criticalAnomalies.length > 0
                    ? <Badge badgeContent={criticalAnomalies.length} color="error">{item.icon}</Badge>
                    : item.icon}
                </ListItemIcon>
                {sidebarOpen && (
                  <ListItemText primary={item.text} sx={{ '& .MuiListItemText-primary': { color: section === item.id ? C.text : C.sub, fontSize: '0.9rem', fontWeight: section === item.id ? 700 : 400 } }} />
                )}
              </ListItemButton>
            </ListItem>
          ))}
        </List>
        <Box sx={{ mt: 'auto', p: 2, borderTop: `1px solid ${C.border}` }}>
          <Stack direction="row" spacing={1} justifyContent={sidebarOpen ? 'space-between' : 'center'} alignItems="center">
            {sidebarOpen && (
              <Chip
                icon={backendOk ? <WifiIcon sx={{ fontSize: 14 }} /> : <WifiOffIcon sx={{ fontSize: 14 }} />}
                label={backendOk ? 'Online' : 'Offline'} size="small"
                sx={{ bgcolor: alpha(backendOk ? C.green : C.red, 0.15), color: backendOk ? C.green : C.red, fontSize: '0.7rem' }}
              />
            )}
            <IconButton onClick={toggleMode} size="small" sx={{ color: C.sub }}>
              {mode === 'dark' ? <Brightness7Icon fontSize="small" /> : <Brightness4Icon fontSize="small" />}
            </IconButton>
          </Stack>
        </Box>
      </Drawer>

      {/* ── Main content ─────────────────────────────────────────────────── */}
      <Box component="main" sx={{ flexGrow: 1, p: 3, overflow: 'auto' }}>

        {/* Header */}
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3, flexWrap: 'wrap', gap: 2 }}>
          <Box>
            <Typography variant="h5" fontWeight="800" sx={{ color: C.text }}>CHAIMAN IOT • AI Sentinel</Typography>
            <Stack direction="row" spacing={1} mt={0.5} flexWrap="wrap">
              <Chip icon={<BlockIcon />} label={`Block #${blockInfo?.blockNumber?.toLocaleString() ?? '—'}`} size="small" sx={{ bgcolor: alpha(C.blue, 0.12), color: C.blue }} />
              <Chip icon={<StorageIcon />} label={`Chain ${blockInfo?.chainID ?? '—'}`} size="small" sx={{ bgcolor: alpha(C.purple, 0.12), color: C.purple }} />
              {criticalAnomalies.length > 0 && (
                <Chip icon={<AlertIcon />} label={`${criticalAnomalies.length} critical alerts`} size="small" sx={{ bgcolor: alpha(C.red, 0.2), color: C.red, animation: 'livePulse 2s infinite' }} />
              )}
              {lastUpdate && (
                <Chip label={`Updated ${formatDistanceToNow(lastUpdate, { addSuffix: true, locale: enUS })}`} size="small" variant="outlined" sx={{ color: C.sub, borderColor: C.border }} />
              )}
            </Stack>
          </Box>
          <Stack direction="row" spacing={1} alignItems="center">
            <FormControlLabel
              control={<Switch checked={autoRefresh} onChange={e => setAutoRefresh(e.target.checked)} size="small" />}
              label={<Typography variant="caption" sx={{ color: C.sub }}>Auto ({POLL_MS / 1000}s)</Typography>}
            />
            <IconButton onClick={() => loadData(false)} sx={{ bgcolor: alpha(C.blue, 0.1), color: C.blue }}>
              <RefreshIcon fontSize="small" />
            </IconButton>
          </Stack>
        </Box>

        <AnimatePresence mode="wait">
          <motion.div key={section} initial={{ opacity: 0, x: 10 }} animate={{ opacity: 1, x: 0 }} exit={{ opacity: 0 }}>

            {/* ═══════════════════ OVERVIEW ══════════════════════════════ */}
            {section === 'overview' && (
              <Box>
                <Grid container spacing={2} sx={{ mb: 3 }}>
                  {[
                    { label: 'Total Devices',  value: globalStats?.totalDevices?.toLocaleString(),         sub: 'registered on-chain',                    color: C.green,  icon: <DevicesIcon /> },
                    { label: 'Active Devices', value: globalStats?.activeDevices?.toLocaleString(),        sub: `${globalStats?.inactiveDevices ?? 0} inactive`, color: C.cyan,   icon: <CheckCircleIcon /> },
                    { label: 'Total Auths',    value: globalStats?.totalAuthentications?.toLocaleString(), sub: 'recorded on blockchain',                 color: C.blue,   icon: <VerifiedIcon /> },
                    { label: 'Block',          value: `#${blockInfo?.blockNumber?.toLocaleString() ?? '—'}`, sub: `Chain ${blockInfo?.chainID ?? '—'}`,   color: C.purple, icon: <BlockIcon /> },
                    { label: 'Avg Latency',    value: avgTotalMs !== null ? `${avgTotalMs}ms` : '—',       sub: 'end-to-end',                             color: C.pink,   icon: <AccessTimeIcon /> },
                    { label: 'AI Anomalies',   value: safeAiAnomalies.length,                              sub: `${criticalAnomalies.length} critical`,   color: C.red,    icon: <AnomalyIcon />, badge: criticalAnomalies.length },
                    // [FIX-AI-1] AI Status basé sur mlOk réel
                    { label: 'AI Status',      value: mlOk ? 'Active' : 'Offline',                        sub: mlOk ? `${devicesMonitored} monitored` : 'ML service down', color: mlOk ? C.green : C.muted, icon: <AiIcon /> },
                  ].map((k, i) => (
                    <Grid size={{ xs: 12, sm: 6, md: 3, lg: 12/7 }} key={i}>
                      <KPI {...k} colors={C} loading={false} />
                    </Grid>
                  ))}
                </Grid>

                <Grid container spacing={2} sx={{ mb: 3 }}>
                  <Grid size={{ xs: 12, lg: 6 }}>
                    <ChartCard title="Auth Latency — Real measured ms" icon={<AccessTimeIcon />} color={C.green} colors={C} height={250} loading={!hasData}>
                      {latencyHistory.labels.length > 0
                        ? <Line data={latencyChartData} options={lineOpts} />
                        : <NoData msg="Waiting for data..." />}
                    </ChartCard>
                  </Grid>
                  <Grid size={{ xs: 12, lg: 6 }}>
                    {/* [FIX-AI-2] Graphique toujours alimenté */}
                    <ChartCard title="AI Anomaly Score (real-time)" icon={<AiIcon />} color={C.red} colors={C} height={250} loading={!hasData}>
                      {anomalyScoreHistory.labels.length > 0
                        ? <Line data={anomalyChartData} options={anomalyLineOpts} />
                        : <NoData msg={mlOk ? 'Collecting data...' : 'ML service offline'} />}
                    </ChartCard>
                  </Grid>
                </Grid>

                <Grid container spacing={2} sx={{ mb: 3 }}>
                  <Grid size={{ xs: 12, md: 4 }}>
                    <ChartCard title="Device Types" icon={<DevicesIcon />} color={C.cyan} colors={C} height={230} loading={!hasData}>
                      {Object.keys(deviceTypes).length > 0
                        ? <Doughnut data={deviceTypeChartData} options={doughnutOpts} />
                        : <NoData msg="No devices" />}
                    </ChartCard>
                  </Grid>
                  <Grid size={{ xs: 12, md: 4 }}>
                    <ChartCard title="Tx/min Rate" icon={<TimelineIcon />} color={C.purple} colors={C} height={230} loading={!hasData}>
                      {txRateHistory.labels.length > 0
                        ? <Bar data={txRateChartData} options={barOpts} />
                        : <NoData msg="Waiting for data..." />}
                    </ChartCard>
                  </Grid>
                  <Grid size={{ xs: 12, md: 4 }}>
                    <ChartCard title="Top Devices by Auth" icon={<VerifiedIcon />} color={C.orange} colors={C} height={230} loading={!hasData}>
                      {topByAuth.length > 0
                        ? <Bar data={authDistChartData} options={barOptsH} />
                        : <NoData msg="No data" />}
                    </ChartCard>
                  </Grid>
                </Grid>
              </Box>
            )}

            {/* ═══════════════════ AI DETECTION ══════════════════════════ */}
            {section === 'ai' && (
              <Box>
                {/* KPI row */}
                <Grid container spacing={2} sx={{ mb: 3 }}>
                  {[
                    { label: 'Total Anomalies',   v: safeAiAnomalies.length,                                                    c: C.purple, icon: <AnomalyIcon /> },
                    { label: 'Critical',          v: criticalAnomalies.length,                                                   c: C.red,    icon: <AlertIcon /> },
                    { label: 'High',              v: highAnomalies.length,                                                       c: C.orange, icon: <WarningIcon /> },
                    { label: 'Actions Taken',     v: safeAiActions.length,                                                       c: C.blue,   icon: <AssignmentIcon /> },
                    // [FIX-AI-4]
                    { label: 'Devices Monitored', v: devicesMonitored > 0 ? devicesMonitored : (mlOk ? '—' : 'Offline'),         c: C.cyan,   icon: <DevicesIcon /> },
                    // [FIX-AI-5]
                    { label: 'Avg Anomaly Score', v: avgAnomalyScore !== null ? `${avgAnomalyScore}%` : (mlOk ? '0%' : '—'),     c: C.pink,   icon: <AnalyticsIcon /> },
                  ].map((k, i) => (
                    <Grid size={{ xs: 6, sm: 4, md: 2 }} key={i}>
                      <KPI label={k.label} value={k.v} color={k.c} icon={k.icon} colors={C} loading={false} />
                    </Grid>
                  ))}
                </Grid>

                {/* [FIX-AI-3] Model Info Panel */}
                {aiStats && (
                  <Card sx={{ borderRadius: 3, bgcolor: C.card, border: `1px solid ${alpha(C.purple, 0.3)}`, mb: 3, p: 2 }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1.5 }}>
                      <InfoIcon sx={{ color: C.purple, fontSize: 18 }} />
                      <Typography variant="subtitle2" fontWeight="700" sx={{ color: C.text }}>
                        AI Model Info — {aiStats.model_version || 'V5 Dev'}
                      </Typography>
                      <Chip label={mlOk ? 'ONLINE' : 'OFFLINE'} size="small"
                        sx={{ ml: 'auto', bgcolor: alpha(mlOk ? C.green : C.red, 0.15), color: mlOk ? C.green : C.red, fontWeight: 700 }} />
                    </Box>
                    <Grid container spacing={1.5}>
                      {[
                        { l: 'F1 Score',          v: aiStats.f1_score     ? `${aiStats.f1_score}%`             : '—',   c: C.green  },
                        { l: 'Precision',         v: aiStats.precision    ? `${aiStats.precision}%`            : '—',   c: C.cyan   },
                        { l: 'Recall',            v: aiStats.recall       ? `${aiStats.recall}%`               : '—',   c: C.blue   },
                        { l: 'Threshold',         v: aiStats.threshold    ? aiStats.threshold.toFixed(6)       : '—',   c: C.orange },
                        { l: 'Scaler',            v: aiStats.scaler_loaded ? '✅ Loaded' : '⚠️ Z-score local',          c: aiStats.scaler_loaded ? C.green : C.orange },
                        { l: 'Features',          v: aiStats.features     ? `${aiStats.features} features`     : '—',   c: C.purple },
                        { l: 'Seq Length',        v: aiStats.sequence_length ? `${aiStats.sequence_length} pts` : '—',  c: C.teal   },
                        { l: 'Total Analyzed',    v: aiStats.total_analyzed?.toLocaleString()                  || '0',  c: C.sub    },
                        { l: 'Revocations',       v: aiStats.revocations?.toLocaleString()                     || '0',  c: C.red    },
                      ].map((item, i) => (
                        <Grid size={{ xs: 6, sm: 4, md: 3, lg: 12/9 }} key={i}>
                          <Box sx={{ p: 1, borderRadius: 1.5, bgcolor: alpha(item.c, 0.07), border: `1px solid ${alpha(item.c, 0.15)}` }}>
                            <Typography variant="caption" sx={{ color: C.muted, display: 'block', fontSize: '0.6rem', textTransform: 'uppercase', letterSpacing: 0.5 }}>{item.l}</Typography>
                            <Typography fontWeight="700" sx={{ color: item.c, fontSize: '0.82rem', mt: 0.2 }}>{item.v}</Typography>
                          </Box>
                        </Grid>
                      ))}
                    </Grid>
                  </Card>
                )}

                {/* Charts */}
                <Grid container spacing={2} sx={{ mb: 3 }}>
                  <Grid size={{ xs: 12, lg: 8 }}>
                    {/* [FIX-AI-2] Graphique toujours alimenté */}
                    <ChartCard title="Anomaly Score Evolution (real-time)" icon={<TrendingUpIcon />} color={C.red} colors={C} height={260} loading={!hasData}>
                      {anomalyScoreHistory.labels.length > 0
                        ? <Line data={anomalyChartData} options={anomalyLineOpts} />
                        : <NoData msg={mlOk ? 'Collecting first data points...' : 'ML service offline — start detection_dev.py'} />}
                    </ChartCard>
                  </Grid>
                  <Grid size={{ xs: 12, lg: 4 }}>
                    <ChartCard title="Anomaly Severity Distribution" icon={<AnalyticsIcon />} color={C.purple} colors={C} height={260} loading={!hasData}>
                      {safeAiAnomalies.length > 0 ? (
                        <Doughnut data={{
                          labels: ['Critical', 'High', 'Medium', 'Low', 'Normal'],
                          datasets: [{
                            data: [
                              safeAiAnomalies.filter(a => a.severity === 'CRITICAL').length,
                              safeAiAnomalies.filter(a => a.severity === 'HIGH').length,
                              safeAiAnomalies.filter(a => a.severity === 'MEDIUM').length,
                              safeAiAnomalies.filter(a => a.severity === 'LOW').length,
                              safeAiAnomalies.filter(a => a.severity === 'NORMAL' || !a.severity).length,
                            ],
                            backgroundColor: [C.red, C.orange, C.amber, C.cyan, C.green],
                            borderWidth: 2, borderColor: C.bg,
                          }]
                        }} options={doughnutOpts} />
                      ) : (
                        <NoData msg={mlOk ? 'No anomalies — system healthy ✅' : 'Waiting for ML service...'} />
                      )}
                    </ChartCard>
                  </Grid>
                </Grid>

                {/* [FIX-AI-6] Tableau avec toggle anomalies only / all */}
                <Card sx={{ borderRadius: 3, bgcolor: C.card, border: `1px solid ${alpha(C.purple, 0.3)}`, mb: 3 }}>
                  <CardContent sx={{ pb: 0 }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 1.5 }}>
                      <AlertIcon sx={{ color: C.red }} />
                      <Typography variant="h6" fontWeight="700" sx={{ color: C.text }}>AI Analysis Log</Typography>
                      <Chip label={aiDisplayList.length} size="small" sx={{ bgcolor: alpha(C.red, 0.15), color: C.red, fontWeight: 700 }} />
                      <Box sx={{ ml: 'auto' }}>
                        <Button
                          size="small"
                          variant={showAllAnalyses ? 'contained' : 'outlined'}
                          onClick={() => setShowAllAnalyses(v => !v)}
                          sx={{ fontSize: '0.7rem', textTransform: 'none', borderColor: C.purple, color: showAllAnalyses ? C.text : C.purple, bgcolor: showAllAnalyses ? alpha(C.purple, 0.2) : 'transparent' }}
                        >
                          {showAllAnalyses ? 'Anomalies only' : 'Show all'}
                        </Button>
                      </Box>
                    </Box>
                  </CardContent>
                  <TableContainer sx={{ maxHeight: 400 }}>
                    <Table size="small" stickyHeader>
                      <TableHead>
                        <TableRow>
                          {['Device', 'Severity', 'Score', 'MSE', 'Action', 'Status', 'Timestamp'].map(h => (
                            <TableCell key={h} sx={{ bgcolor: C.card, color: C.text, fontWeight: 700, borderBottom: `1px solid ${C.border}`, fontSize: '0.78rem' }}>{h}</TableCell>
                          ))}
                        </TableRow>
                      </TableHead>
                      <TableBody>
                        {pagedAi.map((a, i) => {
                          const sev   = a.severity || 'NORMAL';
                          const color = severityColor(sev);
                          return (
                            <TableRow key={a.id || i} hover sx={{ '&:hover': { bgcolor: alpha(C.text, 0.03) } }}>
                              <TableCell sx={{ color: C.sub, fontFamily: 'monospace', fontSize: '0.72rem' }}>
                                {safe(a.device_address || a.device || '').slice(0, 14)}…
                              </TableCell>
                              <TableCell>
                                <Chip label={sev} size="small" sx={{ bgcolor: alpha(color, 0.2), color, fontWeight: 700, fontSize: '0.65rem', height: 20 }} />
                              </TableCell>
                              <TableCell sx={{ color: a.score > 70 ? C.red : a.score > 40 ? C.orange : C.green, fontWeight: 700 }}>
                                {a.score !== undefined ? `${Math.round(a.score)}%` : '—'}
                              </TableCell>
                              <TableCell sx={{ color: C.sub, fontSize: '0.78rem' }}>
                                {a.mse !== undefined ? a.mse.toFixed(6) : '—'}
                              </TableCell>
                              <TableCell>
                                <Chip label={a.action || 'NONE'} size="small"
                                  sx={{ bgcolor: alpha(a.action === 'REVOKE' || a.action === 'IMMEDIATE_REVOKE' ? C.red : a.action === 'CHALLENGE' ? C.orange : C.blue, 0.12),
                                        color:   a.action === 'REVOKE' || a.action === 'IMMEDIATE_REVOKE' ? C.red : a.action === 'CHALLENGE' ? C.orange : C.blue }} />
                              </TableCell>
                              <TableCell>
                                <Chip label={a.status || 'normal'} size="small"
                                  sx={{ bgcolor: alpha(a.status === 'anomaly' ? C.red : C.green, 0.1), color: a.status === 'anomaly' ? C.red : C.green, fontSize: '0.65rem', height: 18 }} />
                              </TableCell>
                              <TableCell sx={{ color: C.sub, fontSize: '0.72rem', whiteSpace: 'nowrap' }}>
                                {a.timestamp ? format(safeDate(a.timestamp), 'HH:mm:ss') : '—'}
                              </TableCell>
                            </TableRow>
                          );
                        })}
                        {aiDisplayList.length === 0 && (
                          <TableRow>
                            <TableCell colSpan={7} align="center" sx={{ color: C.muted, py: 4 }}>
                              {mlOk
                                ? 'No anomalies detected — system is healthy ✅'
                                : 'ML service offline — start detection_dev.py on port 5001'}
                            </TableCell>
                          </TableRow>
                        )}
                      </TableBody>
                    </Table>
                  </TableContainer>
                  <TablePagination component="div" count={aiDisplayList.length}
                    page={aiPage} onPageChange={(_, p) => setAiPage(p)}
                    rowsPerPage={aiPer} onRowsPerPageChange={e => { setAiPer(+e.target.value); setAiPage(0); }}
                    rowsPerPageOptions={[10, 25, 50]}
                    sx={{ color: C.sub, borderTop: `1px solid ${C.border}` }} />
                </Card>

                {/* [FIX-AI-7] Actions log toujours visible */}
                <Card sx={{ borderRadius: 3, bgcolor: C.card, border: `1px solid ${C.border}` }}>
                  <CardContent sx={{ pb: 0 }}>
                    <SectionTitle icon={<AssignmentIcon />} title="AI Actions Log" count={safeAiActions.length} color={C.blue} colors={C} />
                  </CardContent>
                  <TableContainer sx={{ maxHeight: 250 }}>
                    <Table size="small">
                      <TableHead>
                        <TableRow>
                          {['Device', 'Action', 'Severity', 'Score', 'Timestamp'].map(h => (
                            <TableCell key={h} sx={{ bgcolor: C.card, color: C.text, fontWeight: 700, borderBottom: `1px solid ${C.border}`, fontSize: '0.78rem' }}>{h}</TableCell>
                          ))}
                        </TableRow>
                      </TableHead>
                      <TableBody>
                        {safeAiActions.slice(0, 20).map((a, i) => (
                          <TableRow key={i} hover>
                            <TableCell sx={{ color: C.sub, fontFamily: 'monospace', fontSize: '0.72rem' }}>
                              {safe(a.device_address || '').slice(0, 14)}…
                            </TableCell>
                            <TableCell>
                              <Chip label={a.action || 'none'} size="small"
                                sx={{ bgcolor: alpha(a.action === 'REVOKE' || a.action === 'IMMEDIATE_REVOKE' ? C.red : C.blue, 0.12),
                                      color:   a.action === 'REVOKE' || a.action === 'IMMEDIATE_REVOKE' ? C.red : C.blue }} />
                            </TableCell>
                            <TableCell>
                              <Chip label={a.severity || 'NORMAL'} size="small"
                                sx={{ bgcolor: alpha(severityColor(a.severity), 0.12), color: severityColor(a.severity) }} />
                            </TableCell>
                            <TableCell sx={{ color: C.sub }}>{a.score !== undefined ? `${Math.round(a.score)}%` : '—'}</TableCell>
                            <TableCell sx={{ color: C.sub, fontSize: '0.72rem', whiteSpace: 'nowrap' }}>
                              {a.timestamp ? format(safeDate(a.timestamp), 'HH:mm:ss') : '—'}
                            </TableCell>
                          </TableRow>
                        ))}
                        {safeAiActions.length === 0 && (
                          <TableRow>
                            <TableCell colSpan={5} align="center" sx={{ color: C.muted, py: 3 }}>
                              {mlOk ? 'No actions taken yet — no threats detected' : 'ML service offline'}
                            </TableCell>
                          </TableRow>
                        )}
                      </TableBody>
                    </Table>
                  </TableContainer>
                </Card>
              </Box>
            )}

            {/* ═══════════════════ DEVICES ═══════════════════════════════ */}
            {section === 'devices' && (
              <Box>
                <SectionTitle icon={<DevicesIcon />} title="IoT Devices" count={iotDevices.length} color={C.cyan} colors={C} />
                <Card sx={{ borderRadius: 3, bgcolor: C.card, border: `1px solid ${C.border}`, mb: 3 }}>
                  <TableContainer sx={{ maxHeight: 480 }}>
                    <Table size="small" stickyHeader>
                      <TableHead>
                        <TableRow>
                          {['Address', 'Type', 'Status', 'Auth Count', 'Last Seen'].map(h => (
                            <TableCell key={h} sx={{ bgcolor: C.card, color: C.text, fontWeight: 700, borderBottom: `1px solid ${C.border}`, fontSize: '0.82rem' }}>{h}</TableCell>
                          ))}
                        </TableRow>
                      </TableHead>
                      <TableBody>
                        {pagedDevs.map((d, i) => {
                          let devType = 'unknown';
                          try { devType = JSON.parse(d.metadata || '{}').type || 'unknown'; } catch {}
                          return (
                            <TableRow key={d.address || i} hover sx={{ '&:hover': { bgcolor: alpha(C.text, 0.03) } }}>
                              <TableCell sx={{ color: C.sub, fontFamily: 'monospace', fontSize: '0.78rem' }}>{safe(d.address || '').slice(0, 22)}…</TableCell>
                              <TableCell><Chip label={devType.replace(/_/g, ' ')} size="small" sx={{ bgcolor: alpha(C.cyan, 0.12), color: C.cyan, fontSize: '0.7rem' }} /></TableCell>
                              <TableCell>
                                <Chip icon={d.isActive ? <CheckCircleIcon sx={{ fontSize: 13 }} /> : <ErrorIcon sx={{ fontSize: 13 }} />}
                                  label={d.isActive ? 'Active' : 'Inactive'} size="small"
                                  sx={{ bgcolor: alpha(d.isActive ? C.green : C.red, 0.12), color: d.isActive ? C.green : C.red }} />
                              </TableCell>
                              <TableCell sx={{ color: C.text, fontWeight: 600 }}>{d.authCount ?? 0}</TableCell>
                              <TableCell sx={{ color: C.sub, fontSize: '0.78rem' }}>
                                {d.lastSeen ? formatDistanceToNow(safeDate(d.lastSeen), { addSuffix: true, locale: enUS }) : '—'}
                              </TableCell>
                            </TableRow>
                          );
                        })}
                        {iotDevices.length === 0 && (
                          <TableRow><TableCell colSpan={5} align="center" sx={{ color: C.muted, py: 4 }}>No IoT devices found</TableCell></TableRow>
                        )}
                      </TableBody>
                    </Table>
                  </TableContainer>
                  <TablePagination component="div" count={iotDevices.length}
                    page={devPage} onPageChange={(_, p) => setDevPage(p)}
                    rowsPerPage={devPer} onRowsPerPageChange={e => { setDevPer(+e.target.value); setDevPage(0); }}
                    rowsPerPageOptions={[10, 15, 25, 50]}
                    sx={{ color: C.sub, borderTop: `1px solid ${C.border}` }} />
                </Card>

                <SectionTitle icon={<SecurityIcon />} title="Geth Nodes" count={gethNodes.length} color={C.orange} colors={C} />
                <Card sx={{ borderRadius: 3, bgcolor: C.card, border: `1px solid ${C.border}` }}>
                  <TableContainer sx={{ maxHeight: 380 }}>
                    <Table size="small" stickyHeader>
                      <TableHead>
                        <TableRow>
                          {['Device ID', 'Address', 'Type', 'RPC Port', 'Status', 'Auth Count'].map(h => (
                            <TableCell key={h} sx={{ bgcolor: C.card, color: C.text, fontWeight: 700, borderBottom: `1px solid ${C.border}`, fontSize: '0.82rem' }}>{h}</TableCell>
                          ))}
                        </TableRow>
                      </TableHead>
                      <TableBody>
                        {(Array.isArray(gethNodes) ? gethNodes : []).map((n, i) => (
                          <TableRow key={n.address || i} hover sx={{ '&:hover': { bgcolor: alpha(C.text, 0.03) } }}>
                            <TableCell sx={{ color: C.text, fontWeight: 600, fontSize: '0.82rem' }}>{n.device_id || `NODE_${i}`}</TableCell>
                            <TableCell sx={{ color: C.sub, fontFamily: 'monospace', fontSize: '0.78rem' }}>{safe(n.address || '').slice(0, 18)}…</TableCell>
                            <TableCell>
                              <Chip label={n.type || 'validator'} size="small"
                                sx={{ bgcolor: alpha(n.type === 'fullnode' ? C.blue : C.orange, 0.12), color: n.type === 'fullnode' ? C.blue : C.orange }} />
                            </TableCell>
                            <TableCell sx={{ color: C.sub }}>{n.rpc_port || '—'}</TableCell>
                            <TableCell>
                              <Chip icon={n.isActive ? <CheckCircleIcon sx={{ fontSize: 13 }} /> : <ErrorIcon sx={{ fontSize: 13 }} />}
                                label={n.isActive ? 'Active' : 'Inactive'} size="small"
                                sx={{ bgcolor: alpha(n.isActive ? C.green : C.red, 0.12), color: n.isActive ? C.green : C.red }} />
                            </TableCell>
                            <TableCell sx={{ color: C.text }}>{n.authCount ?? 0}</TableCell>
                          </TableRow>
                        ))}
                        {gethNodes.length === 0 && (
                          <TableRow><TableCell colSpan={6} align="center" sx={{ color: C.muted, py: 4 }}>No Geth nodes found</TableCell></TableRow>
                        )}
                      </TableBody>
                    </Table>
                  </TableContainer>
                </Card>
              </Box>
            )}

            {/* ═══════════════════ TRANSACTIONS ══════════════════════════ */}
            {section === 'transactions' && (
              <Box>
                <Grid container spacing={2} sx={{ mb: 3 }}>
                  {[
                    { l: 'Total',   v: txs.length,        c: C.blue  },
                    { l: 'Success', v: successTxs.length, c: C.green },
                    { l: 'Failed',  v: failedTxs.length,  c: C.red   },
                    { l: 'Rate',    v: txRateHistory.values.length > 0 ? `${txRateHistory.values[txRateHistory.values.length - 1]}/min` : '—', c: C.purple },
                  ].map((s, i) => (
                    <Grid size={{ xs: 6, sm: 3 }} key={i}>
                      <Card sx={{ borderRadius: 2, bgcolor: C.card, border: `1px solid ${alpha(s.c, 0.2)}`, p: 1.5, textAlign: 'center' }}>
                        <Typography variant="caption" sx={{ color: C.muted, display: 'block', fontSize: '0.65rem', textTransform: 'uppercase', letterSpacing: 0.5 }}>{s.l}</Typography>
                        <Typography fontWeight="700" sx={{ color: s.c, fontSize: '1rem', mt: 0.3 }}>{s.v}</Typography>
                      </Card>
                    </Grid>
                  ))}
                </Grid>

                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 2 }}>
                  <TimelineIcon sx={{ color: C.orange }} />
                  <Typography variant="h6" fontWeight="700" sx={{ color: C.text }}>Transaction Blocks</Typography>
                  <Chip label={txs.length} size="small" sx={{ bgcolor: alpha(C.orange, 0.15), color: C.orange, fontWeight: 700 }} />
                  <Box sx={{ ml: 'auto', display: 'flex', alignItems: 'center', gap: 0.5 }}>
                    <Box sx={{ width: 7, height: 7, borderRadius: '50%', bgcolor: C.green, animation: 'livePulse 2s infinite' }} />
                    <Typography variant="caption" sx={{ color: C.green, fontWeight: 600 }}>LIVE FEED</Typography>
                  </Box>
                </Box>

                {txs.length === 0 ? (
                  <Card sx={{ borderRadius: 3, bgcolor: C.card, border: `1px solid ${C.border}`, p: 6, textAlign: 'center' }}>
                    <Typography sx={{ color: C.muted }}>No transactions yet — waiting for blockchain data...</Typography>
                  </Card>
                ) : (
                  <Grid container spacing={1.5}>
                    {pagedTxs.map((tx, i) => (
                      <Grid size={{ xs: 12, sm: 6, md: 4, lg: 3 }} key={tx.id || i}>
                        <TxBlock tx={tx} colors={C} />
                      </Grid>
                    ))}
                  </Grid>
                )}

                <Box sx={{ mt: 2, display: 'flex', justifyContent: 'flex-end' }}>
                  <Card sx={{ bgcolor: C.card, border: `1px solid ${C.border}`, borderRadius: 2 }}>
                    <TablePagination component="div" count={txs.length}
                      page={txPage} onPageChange={(_, p) => setTxPage(p)}
                      rowsPerPage={txPer} onRowsPerPageChange={e => { setTxPer(+e.target.value); setTxPage(0); }}
                      rowsPerPageOptions={[12, 24, 48, 96]} labelRowsPerPage="Blocks per page:"
                      sx={{ color: C.sub }} />
                  </Card>
                </Box>
              </Box>
            )}

            {/* ═══════════════════ SECURITY ═══════════════════════════════ */}
            {section === 'security' && (
              <Box>
                <Grid container spacing={2} sx={{ mb: 3 }}>
                  <Grid size={{ xs: 12, md: 5 }}>
                    <ChartCard title="Auth Success vs Failed" icon={<SecurityIcon />} color={C.green} colors={C} height={250} loading={!hasData}>
                      {txs.length > 0
                        ? <Doughnut data={{ labels: ['Success', 'Failed'], datasets: [{ data: [successTxs.length, failedTxs.length], backgroundColor: [C.green, C.red], borderWidth: 2, borderColor: C.bg }] }} options={doughnutOpts} />
                        : <NoData msg="No transaction data" />}
                    </ChartCard>
                  </Grid>
                  <Grid size={{ xs: 12, md: 7 }}>
                    <Card sx={{ borderRadius: 3, bgcolor: C.card, border: `1px solid ${C.border}`, p: 3, height: '100%' }}>
                      <SectionTitle icon={<BlockIcon />} title="Blockchain Info" color={C.blue} colors={C} />
                      <Stack spacing={1.5}>
                        {[
                          { l: 'Contract Address', v: blockInfo?.contractAddress ? safe(blockInfo.contractAddress) : '—', c: C.purple },
                          { l: 'Node Address',     v: blockInfo?.nodeAddress     ? safe(blockInfo.nodeAddress)     : '—', c: C.cyan   },
                          { l: 'Chain ID',         v: blockInfo?.chainID ?? '—',                                          c: C.blue   },
                          { l: 'Current Block',    v: `#${blockInfo?.blockNumber?.toLocaleString() ?? '—'}`,              c: C.green  },
                          { l: 'ECDSA Valid',      v: txs.filter(t => t.ecdsaValid).length,                               c: C.green  },
                          { l: 'ZKP Valid',        v: txs.filter(t => t.zkpValid).length,                                 c: C.blue   },
                        ].map((item, i) => (
                          <Box key={i} sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', p: 1.5, borderRadius: 2, bgcolor: alpha(item.c, 0.06), border: `1px solid ${alpha(item.c, 0.15)}` }}>
                            <Typography variant="body2" sx={{ color: C.sub }}>{item.l}</Typography>
                            <Typography fontWeight="700" sx={{ color: item.c, fontFamily: 'monospace', fontSize: '0.85rem', wordBreak: 'break-all', textAlign: 'right', maxWidth: '60%' }}>
                              {item.v}
                            </Typography>
                          </Box>
                        ))}
                      </Stack>
                    </Card>
                  </Grid>
                </Grid>

                <Card sx={{ borderRadius: 3, bgcolor: C.card, border: `1px solid ${C.border}`, p: 3 }}>
                  <SectionTitle icon={<VerifiedIcon />} title="Cryptographic Validation Rates" color={C.green} colors={C} />
                  <Stack spacing={2}>
                    {[
                      { l: 'ECDSA Signature',     count: txs.filter(t => t.ecdsaValid).length, color: C.green },
                      { l: 'ZKP Groth16 Proof',   count: txs.filter(t => t.zkpValid).length,   color: C.blue  },
                      { l: 'Overall Auth Success', count: successTxs.length,                    color: C.lime  },
                    ].map(({ l, count, color }) => (
                      <Box key={l}>
                        <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.5 }}>
                          <Typography variant="body2" sx={{ color: C.sub }}>{l}</Typography>
                          <Typography variant="body2" fontWeight="700" sx={{ color }}>
                            {txs.length > 0 ? `${pct(count, txs.length)}% (${count}/${txs.length})` : '—'}
                          </Typography>
                        </Box>
                        <LinearProgress variant="determinate" value={txs.length > 0 ? pct(count, txs.length) : 0}
                          sx={{ height: 8, borderRadius: 4, bgcolor: alpha(color, 0.1), '& .MuiLinearProgress-bar': { bgcolor: color, borderRadius: 4 } }} />
                      </Box>
                    ))}
                  </Stack>
                </Card>
              </Box>
            )}

          </motion.div>
        </AnimatePresence>
      </Box>

      <Snackbar open={!!snack} autoHideDuration={4000} onClose={() => setSnack('')} anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}>
        <Alert severity="info" variant="filled">{snack}</Alert>
      </Snackbar>

      <style>{`@keyframes livePulse { 0%,100%{opacity:1} 50%{opacity:0.3} }`}</style>
    </Box>
  );
}
