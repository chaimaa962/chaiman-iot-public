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
  Badge,
  Avatar,
  Divider,
  Stack,
  useTheme
} from '@mui/material';
import {
  Security as SecurityIcon,
  CheckCircle as CheckIcon,
  Error as ErrorIcon,
  Verified as VerifiedIcon,
  Warning as WarningIcon,
  History as HistoryIcon,
  Speed as SpeedIcon,
  Devices as DevicesIcon,
  Refresh as RefreshIcon,
  Info as InfoIcon,
  Autorenew as RetryIcon,
  AccessTime as TimeIcon,
  TrendingUp as TrendingUpIcon,
  TrendingDown as TrendingDownIcon
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
  getDeviceAuthHistory,
  authenticateDevice 
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

const ZKPSecurity = () => {
  const theme = useTheme();
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [devices, setDevices] = useState([]);
  const [selectedDevice, setSelectedDevice] = useState(null);
  const [deviceHistory, setDeviceHistory] = useState([]);
  const [showHistory, setShowHistory] = useState(false);
  const [error, setError] = useState(null);
  const [authenticating, setAuthenticating] = useState({});

  // États simulés pour les échecs
  const [deviceStats, setDeviceStats] = useState({});

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
        label: 'Taux de réussite (%)',
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
      title: { display: true, text: 'Performance ZKP en temps réel' }
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
        title: { display: true, text: 'Taux (%)' }
      },
    },
  };

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
          attempts: Math.floor(Math.random() * 10),
          failures: Math.floor(Math.random() * 5),
          lastAttempt: new Date(Date.now() - Math.random() * 86400000).toISOString(),
          successRate: device.zkpValid ? 100 : Math.floor(Math.random() * 80)
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
        const successRates = devicesData.map(() => Math.floor(Math.random() * 30) + 70);

        setChartData({
          labels: labels,
          datasets: [
            { ...chartData.datasets[0], data: authCounts.slice(0, 12) },
            { ...chartData.datasets[1], data: successRates.slice(0, 12) },
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

  const handleViewHistory = async (address) => {
    try {
      const history = await getDeviceAuthHistory(address);
      setDeviceHistory(history || []);
      setSelectedDevice(address);
      setShowHistory(true);
    } catch (err) {
      console.error('Erreur historique:', err);
    }
  };

  const handleRetryAuth = async (address) => {
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

  // Calculer les métriques
  const totalDevices = devices.length;
  const zkpValidCount = devices.filter(d => d.zkpValid).length;
  const zkpInvalidCount = totalDevices - zkpValidCount;
  const successRate = totalDevices > 0 ? Math.round((zkpValidCount / totalDevices) * 100) : 0;
  const activeCount = devices.filter(d => d.isActive).length;

  const getStatusColor = (device) => {
    if (!device.zkpValid) return 'error';
    const stats = deviceStats[device.address];
    if (stats?.failures > 3) return 'warning';
    return 'success';
  };

  const getStatusIcon = (device) => {
    if (!device.zkpValid) return <ErrorIcon />;
    const stats = deviceStats[device.address];
    if (stats?.failures > 3) return <WarningIcon />;
    return <CheckIcon />;
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
            <SecurityIcon sx={{ fontSize: 40, color: theme.palette.primary.main }} />
            <Box>
              <Typography variant="h4" gutterBottom>
                ZKP Security
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Zero-Knowledge Proofs • Authentification sans révélation de secret
              </Typography>
            </Box>
          </Box>
          <Tooltip title="Rafraîchir">
            <IconButton onClick={loadData} disabled={refreshing}>
              <RefreshIcon />
            </IconButton>
          </Tooltip>
        </Box>

        {refreshing && <LinearProgress sx={{ mb: 3 }} />}
        {error && <Alert severity="error" sx={{ mb: 3 }}>{error}</Alert>}

        {/* Statistiques */}
        <Grid container spacing={3} sx={{ mb: 4 }}>
          <Grid item xs={12} md={6}>
            <Paper sx={{ p: 3 }}>
              <Typography variant="h6" gutterBottom>
                Résumé ZKP
              </Typography>
              <Grid container spacing={2}>
                <Grid item xs={6}>
                  <Box textAlign="center">
                    <Typography variant="h3" fontWeight="bold" color="primary">
                      {zkpValidCount}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Valides
                    </Typography>
                  </Box>
                </Grid>
                <Grid item xs={6}>
                  <Box textAlign="center">
                    <Typography variant="h3" fontWeight="bold" color="error">
                      {zkpInvalidCount}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Invalides
                    </Typography>
                  </Box>
                </Grid>
                <Grid item xs={12}>
                  <Box sx={{ mt: 2 }}>
                    <Box display="flex" justifyContent="space-between" mb={1}>
                      <Typography variant="body2" color="text.secondary">
                        Taux de succès
                      </Typography>
                      <Typography variant="body2" fontWeight="bold">
                        {successRate}%
                      </Typography>
                    </Box>
                    <LinearProgress
                      variant="determinate"
                      value={successRate}
                      color={successRate > 70 ? 'success' : successRate > 40 ? 'warning' : 'error'}
                      sx={{ height: 8, borderRadius: 4 }}
                    />
                  </Box>
                </Grid>
              </Grid>
            </Paper>
          </Grid>

          <Grid item xs={12} md={6}>
            <Paper sx={{ p: 3 }}>
              <Typography variant="h6" gutterBottom>
                Informations Blockchain
              </Typography>
              <Box display="flex" justifyContent="space-between" mb={2}>
                <Typography color="text.secondary">Total appareils:</Typography>
                <Typography fontWeight="bold">{totalDevices}</Typography>
              </Box>
              <Box display="flex" justifyContent="space-between" mb={2}>
                <Typography color="text.secondary">Appareils actifs:</Typography>
                <Typography fontWeight="bold" color="success.main">{activeCount}</Typography>
              </Box>
              <Box display="flex" justifyContent="space-between" mb={2}>
                <Typography color="text.secondary">Appareils inactifs:</Typography>
                <Typography fontWeight="bold" color="error.main">{totalDevices - activeCount}</Typography>
              </Box>
              <Box display="flex" justifyContent="space-between">
                <Typography color="text.secondary">Dernière mise à jour:</Typography>
                <Typography variant="body2">{new Date().toLocaleString()}</Typography>
              </Box>
            </Paper>
          </Grid>
        </Grid>

        {/* Graphique */}
        <Paper sx={{ p: 3, mb: 4 }}>
          <Typography variant="h6" gutterBottom>
            Performance ZKP
          </Typography>
          <Box sx={{ height: 300 }}>
            <Line data={chartData} options={chartOptions} />
          </Box>
        </Paper>

        {/* Tableau des appareils */}
        <Paper sx={{ p: 3 }}>
          <Typography variant="h6" gutterBottom>
            État des appareils
          </Typography>
          
          <TableContainer>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Appareil</TableCell>
                  <TableCell>Adresse</TableCell>
                  <TableCell align="center">Statut ZKP</TableCell>
                  <TableCell align="center">État</TableCell>
                  <TableCell align="center">Tentatives</TableCell>
                  <TableCell align="center">Dernière tentative</TableCell>
                  <TableCell align="center">Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {devices.slice(0, 5).map((device) => {
                  const stats = deviceStats[device.address] || {
                    attempts: 0,
                    failures: 0,
                    lastAttempt: null,
                    successRate: 0
                  };
                  const statusColor = getStatusColor(device);

                  return (
                    <TableRow key={device.address} hover>
                      <TableCell>
                        <Box display="flex" alignItems="center" gap={1}>
                          <DevicesIcon fontSize="small" color="primary" />
                          <Typography>{device.name || 'ESP32'}</Typography>
                        </Box>
                      </TableCell>
                      <TableCell>
                        <code style={{ 
                          backgroundColor: theme.palette.mode === 'dark' ? '#333' : '#f0f0f0',
                          padding: '2px 6px', 
                          borderRadius: '4px',
                          fontSize: '0.85rem'
                        }}>
                          {device.address?.slice(0, 8)}...{device.address?.slice(-6)}
                        </code>
                      </TableCell>
                      <TableCell align="center">
                        <Chip
                          size="small"
                          icon={getStatusIcon(device)}
                          label={device.zkpValid ? 'Valide' : 'Invalide'}
                          color={statusColor}
                          variant={device.zkpValid ? 'filled' : 'outlined'}
                          sx={{ minWidth: 80 }}
                        />
                      </TableCell>
                      <TableCell align="center">
                        <Chip
                          size="small"
                          icon={device.isActive ? <VerifiedIcon /> : <WarningIcon />}
                          label={device.isActive ? 'Actif' : 'Inactif'}
                          color={device.isActive ? 'success' : 'error'}
                          variant="outlined"
                        />
                      </TableCell>
                      <TableCell align="center">
                        <Box>
                          <Typography variant="body2">
                            {stats.attempts}
                          </Typography>
                        </Box>
                      </TableCell>
                      <TableCell align="center">
                        {stats.lastAttempt ? (
                          <Box display="flex" alignItems="center" justifyContent="center" gap={0.5}>
                            <TimeIcon fontSize="small" color="action" sx={{ fontSize: 16 }} />
                            <Typography variant="caption">
                              {new Date(stats.lastAttempt).toLocaleTimeString()}
                            </Typography>
                          </Box>
                        ) : (
                          <Typography variant="caption" color="text.secondary">
                            Jamais
                          </Typography>
                        )}
                      </TableCell>
                      <TableCell align="center">
                        <Stack direction="row" spacing={1} justifyContent="center">
                          <Tooltip title="Voir historique">
                            <IconButton size="small" onClick={() => handleViewHistory(device.address)}>
                              <HistoryIcon fontSize="small" />
                            </IconButton>
                          </Tooltip>
                          {!device.zkpValid && (
                            <Tooltip title="Réessayer">
                              <IconButton 
                                size="small" 
                                onClick={() => handleRetryAuth(device.address)}
                                color="warning"
                                disabled={authenticating[device.address]}
                              >
                                {authenticating[device.address] ? 
                                  <CircularProgress size={20} /> : 
                                  <RetryIcon fontSize="small" />
                                }
                              </IconButton>
                            </Tooltip>
                          )}
                          <Tooltip title="Détails">
                            <IconButton size="small">
                              <InfoIcon fontSize="small" />
                            </IconButton>
                          </Tooltip>
                        </Stack>
                      </TableCell>
                    </TableRow>
                  );
                })}
                {devices.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={7} align="center" sx={{ py: 3 }}>
                      <Typography color="text.secondary">
                        Aucun appareil trouvé dans la blockchain
                      </Typography>
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </TableContainer>
        </Paper>

        {/* Dialog Historique amélioré */}
        {showHistory && (
          <Paper sx={{ p: 3, mt: 3 }}>
            <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
              <Box>
                <Typography variant="h6">
                  Historique des authentifications
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  Appareil: {selectedDevice?.slice(0, 10)}...{selectedDevice?.slice(-4)}
                </Typography>
              </Box>
              <Button size="small" variant="outlined" onClick={() => setShowHistory(false)}>
                Fermer
              </Button>
            </Box>
            <Divider sx={{ mb: 2 }} />
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
                      <TableCell><strong>Preuve</strong></TableCell>
                      <TableCell><strong>Latence</strong></TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {deviceHistory.map((record, idx) => (
                      <TableRow key={idx}>
                        <TableCell>
                          {new Date(record.timestamp * 1000).toLocaleString()}
                        </TableCell>
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
                          <code>{record.proofHash?.slice(0, 15)}...</code>
                        </TableCell>
                        <TableCell>
                          <Box display="flex" alignItems="center" gap={0.5}>
                            <SpeedIcon fontSize="small" color="action" sx={{ fontSize: 16 }} />
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
          </Paper>
        )}
      </Container>
    </Layout>
  );
};

export default ZKPSecurity;
