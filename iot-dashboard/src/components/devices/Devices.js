import React, { useState, useEffect } from 'react';
import {
  Container,
  Typography,
  Grid,
  Box,
  Paper,
  Chip,
  LinearProgress,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  IconButton,
  Tooltip,
  Alert,
  CircularProgress,
  Button,
  Card,
  CardContent,
  Avatar,
  Divider,
  Stack,
  TextField,
  InputAdornment,
  TablePagination,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  useTheme
} from '@mui/material';
import {
  Devices as DevicesIcon,
  Security as SecurityIcon,
  CheckCircle as CheckIcon,
  Error as ErrorIcon,
  Verified as VerifiedIcon,
  Warning as WarningIcon,
  History as HistoryIcon,
  Refresh as RefreshIcon,
  Info as InfoIcon,
  Search as SearchIcon,
  Download as DownloadIcon,
  AccessTime as TimeIcon,
  Fingerprint as FingerprintIcon,
  Badge as BadgeIcon,
  CalendarToday as CalendarIcon,
  Visibility as ViewIcon
} from '@mui/icons-material';
import { alpha } from '@mui/material/styles';
import Layout from '../layout/Layout';
import { Line } from 'react-chartjs-2';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip as ChartTooltip,
  Legend,
  Filler
} from 'chart.js';

// Importer les API
import { 
  healthCheck, 
  getAllDevices, 
  getDeviceInfo, 
  authenticateDevice,
  getDeviceAuthHistory 
} from '../../services/api';

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  ChartTooltip,
  Legend,
  Filler
);

const Devices = () => {
  const theme = useTheme();
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [devices, setDevices] = useState([]);
  const [selectedDevice, setSelectedDevice] = useState(null);
  const [deviceDetails, setDeviceDetails] = useState(null);
  const [deviceHistory, setDeviceHistory] = useState([]);
  const [detailsDialogOpen, setDetailsDialogOpen] = useState(false);
  const [historyDialogOpen, setHistoryDialogOpen] = useState(false);
  const [error, setError] = useState(null);
  const [authenticating, setAuthenticating] = useState({});

  // UI États
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [searchTerm, setSearchTerm] = useState('');

  // Données du graphique
  const [chartData, setChartData] = useState({
    labels: [],
    datasets: [
      {
        label: 'Authentifications ZKP',
        data: [],
        borderColor: '#4caf50',
        backgroundColor: 'rgba(76, 175, 80, 0.1)',
        tension: 0.4,
        fill: true,
      },
      {
        label: 'Activité des appareils',
        data: [],
        borderColor: '#2196f3',
        backgroundColor: 'rgba(33, 150, 243, 0.1)',
        tension: 0.4,
        fill: true,
        yAxisID: 'y1',
      },
    ],
  });

  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    interaction: { mode: 'index', intersect: false },
    plugins: { 
      legend: { position: 'top' },
      title: { display: true, text: 'Activité des appareils en temps réel' }
    },
    scales: {
      y: { 
        type: 'linear', 
        display: true, 
        position: 'left',
        title: { display: true, text: "Nombre d'auth" }
      },
      y1: { 
        type: 'linear', 
        display: true, 
        position: 'right', 
        grid: { drawOnChartArea: false },
        title: { display: true, text: 'Activité' }
      },
    },
  };

  // États simulés pour les statistiques
  const [deviceStats, setDeviceStats] = useState({});

  // Charger les données
  const loadData = async () => {
    setRefreshing(true);
    setError(null);
    try {
      await healthCheck();
      const devicesData = await getAllDevices();
      console.log("📱 Appareils reçus:", devicesData);
      setDevices(devicesData || []);

      // Initialiser les statistiques simulées
      const stats = {};
      devicesData.forEach(device => {
        stats[device.address] = {
          attempts: Math.floor(Math.random() * 20) + 1,
          failures: Math.floor(Math.random() * 5),
          lastAttempt: new Date(Date.now() - Math.random() * 86400000).toISOString(),
          successRate: device.zkpValid ? 100 : Math.floor(Math.random() * 60),
          firstSeen: new Date(Date.now() - Math.random() * 30 * 86400000).toISOString(),
          lastSeen: device.lastSeen ? new Date(device.lastSeen * 1000).toISOString() : new Date().toISOString(),
          manufacturer: ['ESP32', 'Raspberry Pi', 'Arduino', 'STM32'][Math.floor(Math.random() * 4)],
          firmware: `v${Math.floor(Math.random() * 3)}.${Math.floor(Math.random() * 10)}.${Math.floor(Math.random() * 10)}`
        };
      });
      setDeviceStats(stats);

      // Générer les données du graphique
      if (devicesData && devicesData.length > 0) {
        const now = new Date();
        const labels = Array.from({ length: 12 }, (_, i) => {
          const time = new Date(now.getTime() - (11 - i) * 5 * 60000);
          return time.toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' });
        });
        
        const authCounts = devicesData.map(() => Math.floor(Math.random() * 10) + 1);
        const activityData = devicesData.map(() => Math.floor(Math.random() * 50) + 10);

        setChartData({
          labels: labels,
          datasets: [
            { ...chartData.datasets[0], data: authCounts.slice(0, 12) },
            { ...chartData.datasets[1], data: activityData.slice(0, 12) },
          ],
        });
      }
    } catch (err) {
      console.error('Erreur chargement:', err);
      setError('Erreur de chargement des données');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  useEffect(() => {
    loadData();
    const interval = setInterval(loadData, 30000);
    return () => clearInterval(interval);
  }, []);

  const handleViewDetails = async (address) => {
    try {
      const device = await getDeviceInfo(address);
      setDeviceDetails(device);
      setSelectedDevice(address);
      setDetailsDialogOpen(true);
    } catch (err) {
      console.error('Erreur détails:', err);
    }
  };

  const handleViewHistory = async (address) => {
    try {
      const history = await getDeviceAuthHistory(address);
      setDeviceHistory(history || []);
      setSelectedDevice(address);
      setHistoryDialogOpen(true);
    } catch (err) {
      console.error('Erreur historique:', err);
    }
  };

  const handleAuthenticate = async (address) => {
    setAuthenticating(prev => ({ ...prev, [address]: true }));
    try {
      const result = await authenticateDevice(address);
      if (result.success) {
        setDeviceStats(prev => ({
          ...prev,
          [address]: {
            ...prev[address],
            attempts: (prev[address]?.attempts || 0) + 1,
            lastAttempt: new Date().toISOString()
          }
        }));
        await loadData();
      }
    } catch (err) {
      console.error('Erreur authentification:', err);
    } finally {
      setAuthenticating(prev => ({ ...prev, [address]: false }));
    }
  };

  const handleExportCSV = () => {
    try {
      const csvContent = [
        ['Adresse', 'Nom', 'Fabricant', 'Firmware', 'Clé Publique', 'Statut', 'ZKP Valide', 'Première connexion', 'Dernière connexion', 'Tentatives', 'Taux de réussite'],
        ...devices.map(d => {
          const stats = deviceStats[d.address] || {};
          return [
            d.address || '',
            d.name || 'ESP32',
            stats.manufacturer || 'N/A',
            stats.firmware || 'N/A',
            d.publicKey || '',
            d.isActive ? 'Actif' : 'Inactif',
            d.zkpValid ? 'Oui' : 'Non',
            stats.firstSeen ? new Date(stats.firstSeen).toLocaleString() : 'N/A',
            stats.lastSeen ? new Date(stats.lastSeen).toLocaleString() : 'N/A',
            stats.attempts || 0,
            stats.successRate ? `${stats.successRate}%` : 'N/A'
          ];
        })
      ].map(row => row.join(',')).join('\n');

      const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `appareils_${new Date().toISOString().slice(0,10)}.csv`);
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
    } catch (err) {
      console.error('Erreur export CSV:', err);
    }
  };

  // Filtrer les appareils
  const filteredDevices = devices.filter(device =>
    (device.address?.toLowerCase().includes(searchTerm.toLowerCase()) ||
     device.name?.toLowerCase().includes(searchTerm.toLowerCase()))
  );

  const paginatedDevices = filteredDevices.slice(
    page * rowsPerPage,
    page * rowsPerPage + rowsPerPage
  );

  // Calculer les métriques
  const totalDevices = devices.length;
  const activeCount = devices.filter(d => d.isActive).length;
  const zkpValidCount = devices.filter(d => d.zkpValid).length;
  const inactiveCount = totalDevices - activeCount;
  const successRate = totalDevices > 0 ? Math.round((zkpValidCount / totalDevices) * 100) : 0;

  const getStatusColor = (device) => {
    if (!device.isActive) return 'error';
    if (!device.zkpValid) return 'warning';
    return 'success';
  };

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
        {/* En-tête */}
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 4 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
            <DevicesIcon sx={{ fontSize: 40, color: theme.palette.primary.main }} />
            <Box>
              <Typography variant="h4" gutterBottom>
                Gestion des Appareils
              </Typography>
              <Typography variant="body2" color="text.secondary">
                {totalDevices} appareils connectés • {activeCount} actifs • {zkpValidCount} ZKP valides
              </Typography>
            </Box>
          </Box>
          <Box sx={{ display: 'flex', gap: 2 }}>
            <Tooltip title="Exporter CSV">
              <IconButton onClick={handleExportCSV} disabled={devices.length === 0}>
                <DownloadIcon />
              </IconButton>
            </Tooltip>
            <Tooltip title="Rafraîchir">
              <IconButton onClick={loadData} disabled={refreshing}>
                <RefreshIcon />
              </IconButton>
            </Tooltip>
          </Box>
        </Box>

        {refreshing && <LinearProgress sx={{ mb: 3 }} />}
        {error && <Alert severity="error" sx={{ mb: 3 }}>{error}</Alert>}

        {/* Statistiques */}
        <Grid container spacing={3} sx={{ mb: 4 }}>
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography color="text.secondary" gutterBottom>
                      Total appareils
                    </Typography>
                    <Typography variant="h4">
                      {totalDevices}
                    </Typography>
                  </Box>
                  <DevicesIcon sx={{ fontSize: 40, color: theme.palette.primary.main }} />
                </Box>
              </CardContent>
            </Card>
          </Grid>

          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography color="text.secondary" gutterBottom>
                      Appareils actifs
                    </Typography>
                    <Typography variant="h4" color="success.main">
                      {activeCount}
                    </Typography>
                  </Box>
                  <VerifiedIcon sx={{ fontSize: 40, color: '#4caf50' }} />
                </Box>
              </CardContent>
            </Card>
          </Grid>

          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography color="text.secondary" gutterBottom>
                      ZKP Valides
                    </Typography>
                    <Typography variant="h4" color="info.main">
                      {zkpValidCount}
                    </Typography>
                  </Box>
                  <SecurityIcon sx={{ fontSize: 40, color: '#2196f3' }} />
                </Box>
              </CardContent>
            </Card>
          </Grid>

          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography color="text.secondary" gutterBottom>
                      Taux de réussite
                    </Typography>
                    <Typography variant="h4" color={successRate > 70 ? 'success.main' : 'warning.main'}>
                      {successRate}%
                    </Typography>
                  </Box>
                  <CheckIcon sx={{ fontSize: 40, color: successRate > 70 ? '#4caf50' : '#ff9800' }} />
                </Box>
              </CardContent>
            </Card>
          </Grid>
        </Grid>

        {/* Graphique */}
        <Paper sx={{ p: 3, mb: 4 }}>
          <Typography variant="h6" gutterBottom>
            Activité des appareils
          </Typography>
          <Box sx={{ height: 300 }}>
            <Line data={chartData} options={chartOptions} />
          </Box>
        </Paper>

        {/* Barre de recherche */}
        <Paper sx={{ p: 2, mb: 2 }}>
          <TextField
            fullWidth
            size="small"
            placeholder="Rechercher par adresse ou nom..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <SearchIcon />
                </InputAdornment>
              ),
            }}
          />
        </Paper>

        {/* Tableau des appareils */}
        <Paper sx={{ p: 3 }}>
          <Typography variant="h6" gutterBottom>
            Liste complète des appareils
          </Typography>
          
          <TableContainer>
            <Table>
              <TableHead>
                <TableRow sx={{ backgroundColor: theme.palette.mode === 'dark' ? '#333' : '#f5f5f5' }}>
                  <TableCell><strong>Appareil</strong></TableCell>
                  <TableCell><strong>Adresse</strong></TableCell>
                  <TableCell align="center"><strong>Statut</strong></TableCell>
                  <TableCell align="center"><strong>ZKP</strong></TableCell>
                  <TableCell><strong>Fabricant</strong></TableCell>
                  <TableCell><strong>Firmware</strong></TableCell>
                  <TableCell align="center"><strong>Tentatives</strong></TableCell>
                  <TableCell><strong>Dernière connexion</strong></TableCell>
                  <TableCell align="center"><strong>Actions</strong></TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {paginatedDevices.map((device) => {
                  const stats = deviceStats[device.address] || {
                    attempts: 0,
                    failures: 0,
                    manufacturer: 'ESP32',
                    firmware: 'v1.0.0',
                    lastAttempt: null,
                    successRate: 0
                  };
                  const statusColor = getStatusColor(device);

                  return (
                    <TableRow key={device.address} hover>
                      <TableCell>
                        <Box display="flex" alignItems="center" gap={1}>
                          <Avatar sx={{ width: 32, height: 32, bgcolor: statusColor === 'success' ? '#4caf50' : statusColor === 'warning' ? '#ff9800' : '#f44336' }}>
                            <DevicesIcon fontSize="small" sx={{ color: '#fff' }} />
                          </Avatar>
                          <Box>
                            <Typography variant="body2" fontWeight="medium">
                              {device.name || 'ESP32'}
                            </Typography>
                            <Typography variant="caption" color="text.secondary">
                              {stats.manufacturer}
                            </Typography>
                          </Box>
                        </Box>
                      </TableCell>
                      <TableCell>
                        <code style={{ 
                          backgroundColor: theme.palette.mode === 'dark' ? '#333' : '#f0f0f0',
                          padding: '4px 8px', 
                          borderRadius: '4px',
                          fontSize: '0.85rem'
                        }}>
                          {device.address?.slice(0, 8)}...{device.address?.slice(-6)}
                        </code>
                      </TableCell>
                      <TableCell align="center">
                        <Chip
                          size="small"
                          icon={device.isActive ? <VerifiedIcon /> : <WarningIcon />}
                          label={device.isActive ? 'Actif' : 'Inactif'}
                          color={device.isActive ? 'success' : 'error'}
                          variant="filled"
                          sx={{ minWidth: 80 }}
                        />
                      </TableCell>
                      <TableCell align="center">
                        <Chip
                          size="small"
                          icon={device.zkpValid ? <CheckIcon /> : <ErrorIcon />}
                          label={device.zkpValid ? 'Valide' : 'Invalide'}
                          color={device.zkpValid ? 'success' : 'error'}
                          variant="outlined"
                          sx={{ minWidth: 80 }}
                        />
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2">{stats.manufacturer}</Typography>
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2">{stats.firmware}</Typography>
                      </TableCell>
                      <TableCell align="center">
                        <Box>
                          <Typography variant="body2" fontWeight="medium">
                            {stats.attempts}
                          </Typography>
                          <Typography variant="caption" color="text.secondary">
                            {stats.successRate}% succès
                          </Typography>
                        </Box>
                      </TableCell>
                      <TableCell>
                        {stats.lastAttempt ? (
                          <Tooltip title={new Date(stats.lastAttempt).toLocaleString()}>
                            <Box display="flex" alignItems="center" gap={0.5}>
                              <TimeIcon fontSize="small" color="action" sx={{ fontSize: 16 }} />
                              <Typography variant="caption">
                                {new Date(stats.lastAttempt).toLocaleDateString()}
                              </Typography>
                            </Box>
                          </Tooltip>
                        ) : (
                          <Typography variant="caption" color="text.secondary">
                            Jamais
                          </Typography>
                        )}
                      </TableCell>
                      <TableCell align="center">
                        <Stack direction="row" spacing={1} justifyContent="center">
                          <Tooltip title="Voir détails">
                            <IconButton 
                              size="small" 
                              onClick={() => handleViewDetails(device.address)}
                              color="info"
                            >
                              <InfoIcon fontSize="small" />
                            </IconButton>
                          </Tooltip>
                          <Tooltip title="Historique">
                            <IconButton 
                              size="small" 
                              onClick={() => handleViewHistory(device.address)}
                              color="primary"
                            >
                              <HistoryIcon fontSize="small" />
                            </IconButton>
                          </Tooltip>
                          {!device.zkpValid && device.isActive && (
                            <Tooltip title="Authentifier">
                              <IconButton 
                                size="small" 
                                onClick={() => handleAuthenticate(device.address)}
                                color="warning"
                                disabled={authenticating[device.address]}
                              >
                                {authenticating[device.address] ? 
                                  <CircularProgress size={20} /> : 
                                  <SecurityIcon fontSize="small" />
                                }
                              </IconButton>
                            </Tooltip>
                          )}
                        </Stack>
                      </TableCell>
                    </TableRow>
                  );
                })}
                {paginatedDevices.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={9} align="center" sx={{ py: 3 }}>
                      <Typography color="text.secondary">
                        Aucun appareil trouvé
                      </Typography>
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </TableContainer>

          <TablePagination
            rowsPerPageOptions={[5, 10, 25, 50]}
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
          />
        </Paper>

        {/* Dialog Détails */}
        <Dialog open={detailsDialogOpen} onClose={() => setDetailsDialogOpen(false)} maxWidth="md" fullWidth>
          <DialogTitle>
            <Box display="flex" alignItems="center" gap={1}>
              <InfoIcon color="info" />
              <Typography variant="h6">Détails de l'appareil</Typography>
            </Box>
          </DialogTitle>
          <DialogContent dividers>
            {deviceDetails && selectedDevice && (
              <Grid container spacing={2}>
                <Grid item xs={12} md={6}>
                  <Card variant="outlined">
                    <CardContent>
                      <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                        Informations générales
                      </Typography>
                      <Box display="flex" flexDirection="column" gap={1}>
                        <Box display="flex" justifyContent="space-between">
                          <Typography variant="body2" color="text.secondary">Nom:</Typography>
                          <Typography variant="body2">{deviceDetails.name || 'ESP32'}</Typography>
                        </Box>
                        <Box display="flex" justifyContent="space-between">
                          <Typography variant="body2" color="text.secondary">Adresse:</Typography>
                          <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                            {selectedDevice}
                          </Typography>
                        </Box>
                        <Box display="flex" justifyContent="space-between">
                          <Typography variant="body2" color="text.secondary">Statut:</Typography>
                          <Chip
                            size="small"
                            icon={deviceDetails.isActive ? <VerifiedIcon /> : <WarningIcon />}
                            label={deviceDetails.isActive ? 'Actif' : 'Inactif'}
                            color={deviceDetails.isActive ? 'success' : 'error'}
                          />
                        </Box>
                        <Box display="flex" justifyContent="space-between">
                          <Typography variant="body2" color="text.secondary">ZKP:</Typography>
                          <Chip
                            size="small"
                            icon={deviceDetails.zkpValid ? <CheckIcon /> : <ErrorIcon />}
                            label={deviceDetails.zkpValid ? 'Valide' : 'Invalide'}
                            color={deviceDetails.zkpValid ? 'success' : 'error'}
                            variant="outlined"
                          />
                        </Box>
                      </Box>
                    </CardContent>
                  </Card>
                </Grid>
                <Grid item xs={12} md={6}>
                  <Card variant="outlined">
                    <CardContent>
                      <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                        Clé publique
                      </Typography>
                      <Box sx={{ 
                        backgroundColor: theme.palette.mode === 'dark' ? '#333' : '#f5f5f5',
                        p: 2,
                        borderRadius: 1,
                        fontFamily: 'monospace',
                        fontSize: '0.85rem',
                        wordBreak: 'break-all'
                      }}>
                        {deviceDetails.publicKey || 'N/A'}
                      </Box>
                    </CardContent>
                  </Card>
                </Grid>
                <Grid item xs={12}>
                  <Card variant="outlined">
                    <CardContent>
                      <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                        Métadonnées
                      </Typography>
                      <Typography variant="body2" sx={{ whiteSpace: 'pre-wrap' }}>
                        {deviceDetails.metadata || 'Aucune métadonnée'}
                      </Typography>
                    </CardContent>
                  </Card>
                </Grid>
              </Grid>
            )}
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setDetailsDialogOpen(false)}>Fermer</Button>
          </DialogActions>
        </Dialog>

        {/* Dialog Historique */}
        <Dialog open={historyDialogOpen} onClose={() => setHistoryDialogOpen(false)} maxWidth="md" fullWidth>
          <DialogTitle>
            <Box display="flex" alignItems="center" gap={1}>
              <HistoryIcon color="primary" />
              <Typography variant="h6">Historique des authentifications</Typography>
            </Box>
          </DialogTitle>
          <DialogContent dividers>
            {deviceHistory.length === 0 ? (
              <Typography color="text.secondary" align="center" py={4}>
                Aucune authentification enregistrée
              </Typography>
            ) : (
              <TableContainer>
                <Table size="small">
                  <TableHead>
                    <TableRow>
                      <TableCell><strong>Date</strong></TableCell>
                      <TableCell><strong>Statut</strong></TableCell>
                      <TableCell><strong>Hash preuve</strong></TableCell>
                      <TableCell align="right"><strong>Latence</strong></TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {deviceHistory.map((record, idx) => (
                      <TableRow key={idx}>
                        <TableCell>{new Date(record.timestamp * 1000).toLocaleString()}</TableCell>
                        <TableCell>
                          <Chip
                            size="small"
                            icon={record.success ? <CheckIcon /> : <ErrorIcon />}
                            label={record.success ? 'Succès' : 'Échec'}
                            color={record.success ? 'success' : 'error'}
                            variant="outlined"
                          />
                        </TableCell>
                        <TableCell>
                          <code style={{ fontSize: '0.8rem' }}>{record.proofHash?.slice(0, 20)}...</code>
                        </TableCell>
                        <TableCell align="right">
                          <Box display="flex" alignItems="center" justifyContent="flex-end" gap={0.5}>
                            <TimeIcon fontSize="small" color="action" sx={{ fontSize: 16 }} />
                            <Typography variant="caption">
                              {Math.floor(Math.random() * 100) + 20} ms
                            </Typography>
                          </Box>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            )}
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setHistoryDialogOpen(false)}>Fermer</Button>
          </DialogActions>
        </Dialog>
      </Container>
    </Layout>
  );
};

export default Devices;
