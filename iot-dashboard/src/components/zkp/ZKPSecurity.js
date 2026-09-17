import React, { useState, useEffect } from 'react';
import {
  Container, Typography, Grid, Box, Paper, Chip, LinearProgress,
  Table, TableBody, TableCell, TableContainer, TableHead, TableRow,
  IconButton, Tooltip, Alert, CircularProgress, Button, Divider, useTheme
} from '@mui/material';
import {
  Security as SecurityIcon, CheckCircle as CheckIcon,
  Error as ErrorIcon, Verified as VerifiedIcon,
  History as HistoryIcon, Devices as DevicesIcon,
  Refresh as RefreshIcon, AccessTime as TimeIcon
} from '@mui/icons-material';
import Layout from '../layout/Layout';
import { Line } from 'react-chartjs-2';
import {
  Chart as ChartJS, CategoryScale, LinearScale, PointElement,
  LineElement, Title, Tooltip as ChartTooltip, Legend, Filler
} from 'chart.js';

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, ChartTooltip, Legend, Filler);

// ============================================
// CONFIGURATION
// ============================================
const API_BASE = 'http://localhost:8080';

// ============================================
// HELPERS
// ============================================
const safeDate = (ts) => {
  if (!ts) return new Date();
  try {
    const t = Number(ts);
    const d = t > 1000000000000 ? new Date(t) : new Date(t * 1000);
    return isNaN(d.getTime()) ? new Date() : d;
  } catch { return new Date(); }
};

const safeStr = (str) => {
  if (!str) return '';
  // Nettoyer uniquement les caracteres vraiment problematiques
  return String(str)
    .replace(/\r\n/g, ' ')
    .replace(/[\r\n\t]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim();
};
// ============================================
// COMPOSANT PRINCIPAL
// ============================================
const ZKPSecurity = () => {
  const theme = useTheme();
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [devices, setDevices] = useState([]);
  const [gethNodes, setGethNodes] = useState([]);
  const [transactions, setTransactions] = useState([]);
  const [globalStats, setGlobalStats] = useState({ totalDevices: 0, totalAuths: 0, latestBlock: 0 });
  const [selectedDevice, setSelectedDevice] = useState(null);
  const [deviceHistory, setDeviceHistory] = useState([]);
  const [showHistory, setShowHistory] = useState(false);
  const [error, setError] = useState(null);

  // Donnees du graphique
  const [chartData, setChartData] = useState({
    labels: [],
    datasets: [
      { label: 'ZKP Valides', data: [], borderColor: '#4caf50', backgroundColor: 'rgba(76,175,80,0.1)', tension: 0.4, fill: true },
      { label: 'Taux succes (%)', data: [], borderColor: '#2196f3', backgroundColor: 'rgba(33,150,243,0.1)', tension: 0.4, fill: true, yAxisID: 'y1' },
    ],
  });

  const chartOptions = {
    responsive: true, maintainAspectRatio: false,
    plugins: { legend: { position: 'top' }, title: { display: true, text: 'Performance ZKP en temps reel' } },
    scales: {
      y: { type: 'linear', display: true, position: 'left', title: { display: true, text: "Nombre" } },
      y1: { type: 'linear', display: true, position: 'right', grid: { drawOnChartArea: false }, title: { display: true, text: 'Taux (%)' }, min: 0, max: 100 },
    },
  };

  // ============================================
  // CHARGER LES VRAIES DONNEES
  // ============================================
  const loadData = async () => {
    setRefreshing(true);
    setError(null);
    try {
      // 1. Health check
      const healthRes = await fetch(`${API_BASE}/api/health`);
      if (!healthRes.ok) throw new Error('Backend offline');

      // 2. Devices IoT
      const devicesRes = await fetch(`${API_BASE}/api/devices`);
      const devicesData = await devicesRes.json();
      setDevices(Array.isArray(devicesData) ? devicesData.slice(0, 10) : []);

      // 3. Noeuds Geth
      const gethRes = await fetch(`${API_BASE}/api/geth-nodes`);
      const gethData = await gethRes.json();
      setGethNodes(Array.isArray(gethData) ? gethData : []);

      // 4. Stats globales
      const statsRes = await fetch(`${API_BASE}/api/stats/global`);
      const statsData = await statsRes.json();
      setGlobalStats({
        totalDevices: statsData.totalDevices || 0,
        totalAuths: statsData.totalAuths || 0,
        latestBlock: statsData.latestBlock || 0,
        activeDevices: statsData.activeDevices || 0,
        inactiveDevices: (statsData.totalDevices || 0) - (statsData.activeDevices || 0),
      });

      // 5. Transactions recentes pour le graphique
      const txRes = await fetch(`${API_BASE}/api/transactions/recent?limit=60`);
      const txData = await txRes.json();
      setTransactions(Array.isArray(txData) ? txData : []);

      // 6. Construire le graphique avec les VRAIES donnees
      if (txData && txData.length > 0) {
        const last12 = txData.slice(-12);
        const labels = last12.map(tx => {
          try { return safeDate(tx.timestamp).toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' }); }
          catch { return '--:--'; }
        });
        const zkpOk = last12.map(tx => tx.ZKPValid || tx.zkpValid ? 1 : 0);
        const successRates = last12.map((_, i) => {
          const batch = last12.slice(0, i + 1);
          const ok = batch.filter(tx => tx.ZKPValid || tx.zkpValid).length;
          return Math.round((ok / batch.length) * 100);
        });

        setChartData({
          labels,
          datasets: [
            { ...chartData.datasets[0], data: zkpOk },
            { ...chartData.datasets[1], data: successRates },
          ],
        });
      }
    } catch (err) {
      setError('Erreur: ' + err.message);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  useEffect(() => {
    loadData();
    const interval = setInterval(loadData, 15000);
    return () => clearInterval(interval);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // ============================================
  // VOIR L'HISTORIQUE D'UN APPAREIL
  // ============================================
  const handleViewHistory = async (address) => {
    try {
      const res = await fetch(`${API_BASE}/api/device/${address}/history?limit=20`);
      const data = await res.json();
      setDeviceHistory(Array.isArray(data) ? data : []);
      setSelectedDevice(address);
      setShowHistory(true);
    } catch (err) {
      setDeviceHistory([]);
      setSelectedDevice(address);
      setShowHistory(true);
    }
  };

  // ============================================
  // CALCULS
  // ============================================
  const allDevices = [...devices, ...gethNodes];
  const activeDevices = globalStats.activeDevices || allDevices.filter(d => d.isActive).length;
  const totalAuths = globalStats.totalAuths || transactions.length;
  const zkpSuccess = transactions.filter(tx => tx.ZKPValid === true || tx.zkpValid === true).length;
  const successRate = transactions.length > 0 ? Math.round((zkpSuccess / transactions.length) * 100) : 0;

  if (loading) {
    return (
      <Layout>
        <Box display="flex" justifyContent="center" alignItems="center" minHeight="80vh">
          <CircularProgress />
        </Box>
      </Layout>
    );
  }

  return (
    <Layout>
      <Container maxWidth="xl" sx={{ py: 4 }}>
        {/* En-tete */}
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 4 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
            <SecurityIcon sx={{ fontSize: 40, color: theme.palette.primary.main }} />
            <Box>
              <Typography variant="h4">ZKP Security</Typography>
              <Typography variant="body2" color="text.secondary">
                Zero-Knowledge Proofs - Bloc #{globalStats.latestBlock} - {totalAuths} authentifications
              </Typography>
            </Box>
          </Box>
          <Tooltip title="Rafraichir">
            <IconButton onClick={loadData} disabled={refreshing}><RefreshIcon /></IconButton>
          </Tooltip>
        </Box>

        {refreshing && <LinearProgress sx={{ mb: 3 }} />}
        {error && <Alert severity="error" sx={{ mb: 3 }}>{error}</Alert>}

        {/* KPI Cards */}
        <Grid container spacing={3} sx={{ mb: 4 }}>
          {[
            { t: 'Total Auths', v: totalAuths, c: theme.palette.primary.main, i: <SecurityIcon sx={{ fontSize: 40 }} /> },
            { t: 'ZKP Valides', v: zkpSuccess, c: '#4caf50', i: <CheckIcon sx={{ fontSize: 40, color: '#4caf50' }} /> },
            { t: 'Taux succes', v: `${successRate}%`, c: '#2196f3', i: <VerifiedIcon sx={{ fontSize: 40, color: '#2196f3' }} /> },
            { t: 'Appareils actifs', v: activeDevices, c: '#ff9800', i: <DevicesIcon sx={{ fontSize: 40, color: '#ff9800' }} /> },
          ].map((k, i) => (
            <Grid item xs={6} md={3} key={i}>
              <Paper sx={{ p: 3, textAlign: 'center' }}>
                <Box sx={{ color: k.c, mb: 1 }}>{k.i}</Box>
                <Typography variant="h4" fontWeight="bold" sx={{ color: k.c }}>{k.v}</Typography>
                <Typography variant="body2" color="text.secondary">{k.t}</Typography>
              </Paper>
            </Grid>
          ))}
        </Grid>

        {/* Graphique dynamique */}
        <Paper sx={{ p: 3, mb: 4 }}>
          <Typography variant="h6" gutterBottom>Performance ZKP (temps reel)</Typography>
          <Box sx={{ height: 300 }}><Line data={chartData} options={chartOptions} /></Box>
        </Paper>

        {/* Tableau des appareils */}
        <Paper sx={{ p: 3 }}>
          <Typography variant="h6" gutterBottom>Appareils ({allDevices.length} total)</Typography>
          <TableContainer>
            <Table>
              <TableHead>
                <TableRow sx={{ bgcolor: theme.palette.mode === 'dark' ? '#333' : '#f5f5f5' }}>
                  {['Adresse','Statut','Auths','Type','Derniere connexion','Actions'].map(h => <TableCell key={h}><strong>{h}</strong></TableCell>)}
                </TableRow>
              </TableHead>
              <TableBody>
                {allDevices.slice(0, 10).map((device, idx) => (
                  <TableRow key={device.address || idx} hover>
                    <TableCell>
                      <code style={{ fontSize: '0.8rem' }}>{safeStr(device.address).slice(0, 10)}...{safeStr(device.address).slice(-6)}</code>
                    </TableCell>
                    <TableCell>
                      <Chip size="small" icon={device.isActive ? <CheckIcon /> : <ErrorIcon />}
                        label={device.isActive ? 'Actif' : 'Inactif'}
                        color={device.isActive ? 'success' : 'error'} />
                    </TableCell>
                    <TableCell>{device.authCount || 0}</TableCell>
                    <TableCell>
                      <Chip size="small" label={device.isGethNode ? (device.type || 'Geth') : 'IoT'}
                        color={device.isGethNode ? 'warning' : 'info'} variant="outlined" />
                    </TableCell>
                    <TableCell>
                      <Box display="flex" alignItems="center" gap={0.5}>
                        <TimeIcon fontSize="small" color="action" sx={{ fontSize: 16 }} />
                        <Typography variant="caption">
                          {device.lastSeen ? safeDate(device.lastSeen).toLocaleDateString() : 'Jamais'}
                        </Typography>
                      </Box>
                    </TableCell>
                    <TableCell>
                      <Tooltip title="Historique">
                        <IconButton size="small" onClick={() => handleViewHistory(device.address)} color="primary">
                          <HistoryIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </Paper>

        {/* Historique dialog */}
        {showHistory && (
          <Paper sx={{ p: 3, mt: 3 }}>
            <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
              <Typography variant="h6">Historique: {safeStr(selectedDevice).slice(0, 15)}...</Typography>
              <Button size="small" variant="outlined" onClick={() => setShowHistory(false)}>Fermer</Button>
            </Box>
            <Divider sx={{ mb: 2 }} />
            {deviceHistory.length === 0 ? (
              <Typography color="text.secondary" align="center" py={4}>Aucune authentification</Typography>
            ) : (
              <TableContainer><Table size="small">
                <TableHead><TableRow>
                  {['Date','Statut','Hash'].map(h => <TableCell key={h}><strong>{h}</strong></TableCell>)}
                </TableRow></TableHead>
                <TableBody>
                  {deviceHistory.map((record, idx) => (
                    <TableRow key={idx}>
                      <TableCell>{record.timestamp ? safeDate(record.timestamp).toLocaleString() : 'N/A'}</TableCell>
                      <TableCell>
                        <Chip size="small" icon={record.success ? <CheckIcon /> : <ErrorIcon />}
                          label={record.success ? 'Succes' : 'Echec'}
                          color={record.success ? 'success' : 'error'} variant="outlined" />
                      </TableCell>
                      <TableCell><code style={{ fontSize: '0.8rem' }}>{safeStr(record.proofHash).slice(0, 20)}...</code></TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table></TableContainer>
            )}
          </Paper>
        )}
      </Container>
    </Layout>
  );
};

export default ZKPSecurity;
