import React, { useState, useEffect } from 'react';
import {
  Container, Typography, Grid, Box, Paper, Chip, LinearProgress,
  Table, TableBody, TableCell, TableContainer, TableHead, TableRow,
  IconButton, Tooltip, Alert, CircularProgress, Button, Card, CardContent,
  Stack, TextField, InputAdornment, TablePagination,
  Dialog, DialogTitle, DialogContent, DialogActions, useTheme
} from '@mui/material';
import {
  Devices as DevicesIcon, Security as SecurityIcon,
  CheckCircle as CheckIcon, Error as ErrorIcon,
  Verified as VerifiedIcon,
  History as HistoryIcon, Refresh as RefreshIcon,
  Info as InfoIcon, Search as SearchIcon,
  Download as DownloadIcon, AccessTime as TimeIcon
} from '@mui/icons-material';
import Layout from '../layout/Layout';
import { Line } from 'react-chartjs-2';
import {
  Chart as ChartJS, CategoryScale, LinearScale, PointElement,
  LineElement, Title, Tooltip as ChartTooltip, Legend, Filler
} from 'chart.js';

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, ChartTooltip, Legend, Filler);

const API_BASE = 'http://localhost:8080';

// Fonction securisee pour les dates
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
  return String(str).replace(/\r\n/g, ' ').replace(/[\r\n\t]+/g, ' ').replace(/\s+/g, ' ').trim();
};
const Devices = () => {
  const theme = useTheme();
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [devices, setDevices] = useState([]);
  const [gethNodes, setGethNodes] = useState([]);
  const [, setTransactions] = useState([]);
  const [globalStats, setGlobalStats] = useState({ totalDevices: 0, totalAuths: 0, latestBlock: 0 });
  const [selectedDevice, setSelectedDevice] = useState(null);
  const [deviceDetails, setDeviceDetails] = useState(null);
  const [deviceHistory, setDeviceHistory] = useState([]);
  const [detailsDialogOpen, setDetailsDialogOpen] = useState(false);
  const [historyDialogOpen, setHistoryDialogOpen] = useState(false);
  const [error, setError] = useState(null);
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [searchTerm, setSearchTerm] = useState('');

  const [chartData, setChartData] = useState({
    labels: [],
    datasets: [
      { label: 'Latence (ms)', data: [], borderColor: '#ff9800', backgroundColor: 'rgba(255,152,0,0.1)', tension: 0.4, fill: true },
      { label: 'Transactions', data: [], borderColor: '#2196f3', backgroundColor: 'rgba(33,150,243,0.1)', tension: 0.4, fill: true, yAxisID: 'y1' },
    ],
  });

  const chartOptions = {
    responsive: true, maintainAspectRatio: false,
    plugins: { legend: { position: 'top' } },
    scales: {
      y: { type: 'linear', display: true, position: 'left', title: { display: true, text: 'ms' } },
      y1: { type: 'linear', display: true, position: 'right', grid: { drawOnChartArea: false }, title: { display: true, text: 'Nb' } },
    },
  };

  // [FIRE] CHARGER LES VRAIES DONNEES DU BACKEND
  const loadData = async () => {
    setRefreshing(true);
    setError(null);
    try {
      // 1. Health check
      const healthRes = await fetch(`${API_BASE}/api/health`);
      if (!healthRes.ok) throw new Error('Backend offline');

      // 2. Recuperer les devices IoT (filtres - sans Geth)
      const devicesRes = await fetch(`${API_BASE}/api/devices`);
      const devicesData = await devicesRes.json();
      setDevices(Array.isArray(devicesData) ? devicesData : []);

      // 3. Recuperer les noeuds Geth
      const gethRes = await fetch(`${API_BASE}/api/geth-nodes`);
      const gethData = await gethRes.json();
      setGethNodes(Array.isArray(gethData) ? gethData : []);

      // 4. Recuperer les stats globales
      const statsRes = await fetch(`${API_BASE}/api/stats/global`);
      const statsData = await statsRes.json();
      setGlobalStats({
        totalDevices: statsData.totalDevices || 0,
        totalAuths: statsData.totalAuths || 0,
        latestBlock: statsData.latestBlock || 0,
      });

      // 5. Recuperer les transactions recentes pour le graphique
      const txRes = await fetch(`${API_BASE}/api/transactions/recent?limit=50`);
      const txData = await txRes.json();
      setTransactions(Array.isArray(txData) ? txData : []);

      // 6. Construire le graphique avec les VRAIES transactions
      if (txData && txData.length > 0) {
        const last24 = txData.slice(-24);
        const labels = last24.map(tx => {
          try {
            return safeDate(tx.timestamp).toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' });
          } catch { return '--:--'; }
        });
        const latencies = last24.map(tx => tx.latency || 0);
        const txCounts = [];
        for (let i = 0; i < last24.length; i++) {
          txCounts.push(last24.slice(Math.max(0, i - 3), i + 1).length);
        }
        setChartData({
          labels,
          datasets: [
            { ...chartData.datasets[0], data: latencies },
            { ...chartData.datasets[1], data: txCounts },
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
    const interval = setInterval(loadData, 15000); // Rafraichir toutes les 15s
    return () => clearInterval(interval);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleViewDetails = (address) => {
    const device = devices.find(d => d.address === address) || gethNodes.find(n => n.address === address);
    setDeviceDetails(device || {});
    setSelectedDevice(address);
    setDetailsDialogOpen(true);
  };

  const handleViewHistory = async (address) => {
    try {
      const res = await fetch(`${API_BASE}/api/device/${address}/history?limit=50`);
      const data = await res.json();
      setDeviceHistory(Array.isArray(data) ? data : []);
      setSelectedDevice(address);
      setHistoryDialogOpen(true);
    } catch (err) {
      setDeviceHistory([]);
      setHistoryDialogOpen(true);
    }
  };

  const handleExportCSV = () => {
    try {
      const allDevices = [...devices, ...gethNodes];
      const rows = allDevices.map(d => [
        safeStr(d.address || ''), d.isActive ? 'Actif' : 'Inactif',
        d.authCount || 0, d.entityType || (d.isGethNode ? 'geth_node' : 'iot_device')
      ]);
      const csv = [['Adresse','Statut','Auths','Type'], ...rows].map(r => r.join(',')).join('\n');
      const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `devices_${new Date().toISOString().slice(0, 10)}.csv`;
      link.click();
      URL.revokeObjectURL(url);
    } catch (err) {}
  };

  // Fusionner IoT + Geth pour l'affichage complet
  const allDevices = [
    ...devices.map(d => ({ ...d, isGethNode: false, type: 'IoT', entityType: 'iot_device' })),
    ...gethNodes.map(n => ({ ...n, isGethNode: true, type: n.type === 'validator' ? 'Validator' : 'Fullnode', address: n.address, entityType: n.type || 'geth' }))
  ];

  const filteredDevices = allDevices.filter(d =>
    safeStr(d.address).toLowerCase().includes(searchTerm.toLowerCase()) ||
    safeStr(d.type).toLowerCase().includes(searchTerm.toLowerCase())
  );

  const paginatedDevices = filteredDevices.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage);

  const totalDevices = allDevices.length;
  const iotCount = devices.length;
  const gethCount = gethNodes.length;

  if (loading) return (
    <Layout><Box display="flex" justifyContent="center" alignItems="center" minHeight="80vh"><CircularProgress /></Box></Layout>
  );

  return (
    <Layout>
      <Container maxWidth="xl" sx={{ py: 4 }}>
        {/* En-tete */}
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 4 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
            <DevicesIcon sx={{ fontSize: 40, color: theme.palette.primary.main }} />
            <Box>
              <Typography variant="h4">Gestion des Appareils</Typography>
              <Typography variant="body2" color="text.secondary">
                {totalDevices} total - {iotCount} IoT - {gethCount} Geth - Bloc #{globalStats.latestBlock}
              </Typography>
            </Box>
          </Box>
          <Box sx={{ display: 'flex', gap: 2 }}>
            <Tooltip title="Rafraichir"><IconButton onClick={loadData} disabled={refreshing}><RefreshIcon /></IconButton></Tooltip>
            <Button variant="outlined" startIcon={<DownloadIcon />} onClick={handleExportCSV}>Exporter</Button>
          </Box>
        </Box>

        {refreshing && <LinearProgress sx={{ mb: 3 }} />}
        {error && <Alert severity="error" sx={{ mb: 3 }}>{error}</Alert>}

        {/* KPI Cards */}
        <Grid container spacing={3} sx={{ mb: 4 }}>
          {[
            { t: 'Total', v: totalDevices, c: theme.palette.primary.main, i: <DevicesIcon sx={{ fontSize: 40 }} /> },
            { t: 'IoT Devices', v: iotCount, c: '#4caf50', i: <DevicesIcon sx={{ fontSize: 40, color: '#4caf50' }} /> },
            { t: 'Geth Nodes', v: gethCount, c: '#ff9800', i: <SecurityIcon sx={{ fontSize: 40, color: '#ff9800' }} /> },
            { t: 'Authentications', v: globalStats.totalAuths, c: '#2196f3', i: <VerifiedIcon sx={{ fontSize: 40, color: '#2196f3' }} /> },
          ].map((k, i) => (
            <Grid item xs={12} sm={6} md={3} key={i}>
              <Card><CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box><Typography color="text.secondary" gutterBottom>{k.t}</Typography><Typography variant="h4" sx={{ color: k.c }}>{k.v}</Typography></Box>
                  <Box sx={{ color: k.c }}>{k.i}</Box>
                </Box>
              </CardContent></Card>
            </Grid>
          ))}
        </Grid>

        {/* Graphique dynamique */}
        <Paper sx={{ p: 3, mb: 4 }}>
          <Typography variant="h6" gutterBottom>Activite en temps reel</Typography>
          <Box sx={{ height: 300 }}><Line data={chartData} options={chartOptions} /></Box>
        </Paper>

        {/* Recherche */}
        <Paper sx={{ p: 2, mb: 2 }}>
          <TextField fullWidth size="small" placeholder="Rechercher par adresse ou type..."
            value={searchTerm} onChange={e => setSearchTerm(e.target.value)}
            InputProps={{ startAdornment: <InputAdornment position="start"><SearchIcon /></InputAdornment> }}
          />
        </Paper>

        {/* Tableau */}
        <Paper sx={{ p: 3 }}>
          <Typography variant="h6" gutterBottom>Tous les appareils ({filteredDevices.length})</Typography>
          <TableContainer><Table>
            <TableHead><TableRow sx={{ bgcolor: theme.palette.mode === 'dark' ? '#333' : '#f5f5f5' }}>
              {['Type','Adresse','Statut','Auths','Derniere connexion','Actions'].map(h => <TableCell key={h}><strong>{h}</strong></TableCell>)}
            </TableRow></TableHead>
            <TableBody>
              {paginatedDevices.map((device, idx) => (
                <TableRow key={device.address || idx} hover>
                  <TableCell>
                    <Chip size="small"
                      icon={device.isGethNode ? <SecurityIcon /> : <DevicesIcon />}
                      label={device.type || 'IoT'}
                      color={device.isGethNode ? 'warning' : 'success'}
                      variant="filled"
                    />
                  </TableCell>
                  <TableCell>
                    <code style={{ fontSize: '0.8rem' }}>{safeStr(device.address).slice(0, 10)}...{safeStr(device.address).slice(-6)}</code>
                  </TableCell>
                  <TableCell>
                    <Chip size="small"
                      icon={device.isActive ? <CheckIcon /> : <ErrorIcon />}
                      label={device.isActive ? 'Actif' : 'Inactif'}
                      color={device.isActive ? 'success' : 'error'}
                    />
                  </TableCell>
                  <TableCell>{device.authCount || 0}</TableCell>
                  <TableCell>
                    <Box display="flex" alignItems="center" gap={0.5}>
                      <TimeIcon fontSize="small" color="action" sx={{ fontSize: 16 }} />
                      <Typography variant="caption">
                        {device.lastSeen ? safeDate(device.lastSeen).toLocaleDateString() : 'Jamais'}
                      </Typography>
                    </Box>
                  </TableCell>
                  <TableCell>
                    <Stack direction="row" spacing={1}>
                      <Tooltip title="Details"><IconButton size="small" onClick={() => handleViewDetails(device.address)} color="info"><InfoIcon fontSize="small" /></IconButton></Tooltip>
                      <Tooltip title="Historique"><IconButton size="small" onClick={() => handleViewHistory(device.address)} color="primary"><HistoryIcon fontSize="small" /></IconButton></Tooltip>
                    </Stack>
                  </TableCell>
                </TableRow>
              ))}
              {paginatedDevices.length === 0 && (
                <TableRow><TableCell colSpan={6} align="center" sx={{ py: 3 }}><Typography color="text.secondary">Aucun appareil trouve</Typography></TableCell></TableRow>
              )}
            </TableBody>
          </Table></TableContainer>
          <TablePagination rowsPerPageOptions={[5, 10, 25, 50]} component="div" count={filteredDevices.length}
            rowsPerPage={rowsPerPage} page={page} onPageChange={(e, np) => setPage(np)}
            onRowsPerPageChange={e => { setRowsPerPage(parseInt(e.target.value, 10)); setPage(0); }}
          />
        </Paper>

        {/* Dialog Details */}
        <Dialog open={detailsDialogOpen} onClose={() => setDetailsDialogOpen(false)} maxWidth="sm" fullWidth>
          <DialogTitle>Details de l'appareil</DialogTitle>
          <DialogContent dividers>
            {deviceDetails && (
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                <Typography variant="body2"><strong>Adresse:</strong> <code>{selectedDevice}</code></Typography>
                <Typography variant="body2"><strong>Statut:</strong> {deviceDetails.isActive ? 'Actif' : 'Inactif'}</Typography>
                <Typography variant="body2"><strong>Authentifications:</strong> {deviceDetails.authCount || 0}</Typography>
                <Typography variant="body2"><strong>Type:</strong> {deviceDetails.entityType || deviceDetails.type || 'IoT'}</Typography>
                <Box sx={{ bgcolor: theme.palette.mode === 'dark' ? '#333' : '#f5f5f5', p: 2, borderRadius: 1 }}>
                  <Typography variant="body2" sx={{ wordBreak: 'break-all', fontFamily: 'monospace', fontSize: '0.8rem' }}>
                    {safeStr(deviceDetails.publicKey || 'N/A')}
                  </Typography>
                </Box>
              </Box>
            )}
          </DialogContent>
          <DialogActions><Button onClick={() => setDetailsDialogOpen(false)}>Fermer</Button></DialogActions>
        </Dialog>

        {/* Dialog Historique */}
        <Dialog open={historyDialogOpen} onClose={() => setHistoryDialogOpen(false)} maxWidth="md" fullWidth>
          <DialogTitle>Historique des authentifications</DialogTitle>
          <DialogContent dividers>
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
                        <Chip size="small"
                          icon={record.success ? <CheckIcon /> : <ErrorIcon />}
                          label={record.success ? 'Succes' : 'Echec'}
                          color={record.success ? 'success' : 'error'}
                          variant="outlined"
                        />
                      </TableCell>
                      <TableCell><code style={{ fontSize: '0.8rem' }}>{safeStr(record.proofHash).slice(0, 30)}...</code></TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table></TableContainer>
            )}
          </DialogContent>
          <DialogActions><Button onClick={() => setHistoryDialogOpen(false)}>Fermer</Button></DialogActions>
        </Dialog>
      </Container>
    </Layout>
  );
};

export default Devices;

