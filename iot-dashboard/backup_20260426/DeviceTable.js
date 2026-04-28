import React, { useState } from 'react';
import {
  Card, CardContent, Typography, Table, TableBody, TableCell,
  TableContainer, TableHead, TableRow, Chip, IconButton, Button,
  Box, TextField, InputAdornment, Avatar, Tooltip, Dialog,
  DialogTitle, DialogContent, DialogActions, List, ListItem, ListItemText,
  Grid
} from '@mui/material';
import {
  Search as SearchIcon, Security as ZKPIcon, MoreVert as MoreIcon,
  FilterList as FilterIcon, Download as DownloadIcon, Circle as StatusIcon,
  CheckCircle as CheckIcon, Error as ErrorIcon, History as HistoryIcon
} from '@mui/icons-material';
import { getDeviceAuthHistory } from '../../services/api';

const DeviceTable = ({ devices, onAuthenticate, backendConnected }) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedDevice, setSelectedDevice] = useState(null);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [deviceHistory, setDeviceHistory] = useState([]);

  const handleShowHistory = async (device) => {
    setSelectedDevice(device);
    setHistoryOpen(true);
    try {
      const history = await getDeviceAuthHistory(device.address);
      setDeviceHistory(history);
    } catch (err) {
      setDeviceHistory([]);
    }
  };

  const getZKPStatus = (device) => {
    if (device.zkpVerified) {
      return { 
        label: '✅ Vérifié', 
        color: 'success',
        icon: CheckIcon,
        detail: `Dernière auth: ${device.lastAuthTime || 'inconnue'}`
      };
    } else if (device.authCount > 0) {
      return { 
        label: '❌ Échec', 
        color: 'error',
        icon: ErrorIcon,
        detail: 'Échec ZKP - Réessayer'
      };
    } else {
      return { 
        label: '⏳ En attente', 
        color: 'warning',
        icon: ErrorIcon,
        detail: 'Jamais authentifié'
      };
    }
  };

  const filteredDevices = devices.filter(d => 
    (d.name?.toLowerCase() || '').includes(searchTerm.toLowerCase()) || 
    (d.address?.toLowerCase() || '').includes(searchTerm.toLowerCase())
  );

  return (
    <>
      <Card sx={{ mt: 3 }}>
        <CardContent sx={{ p: 3 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
            <Box>
              <Typography variant="h6" fontWeight="bold">📱 Connected Devices</Typography>
              <Typography variant="body2" color="text.secondary">
                {devices.length} devices • {devices.filter(d => d.zkpVerified).length} ZKP vérifiés • {devices.filter(d => !d.zkpVerified && d.authCount > 0).length} échecs
              </Typography>
            </Box>
            <Box sx={{ display: 'flex', gap: 2 }}>
              <TextField
                placeholder="Rechercher un device..."
                size="small"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                InputProps={{
                  startAdornment: (
                    <InputAdornment position="start">
                      <SearchIcon sx={{ color: 'text.secondary' }} />
                    </InputAdornment>
                  ),
                }}
                sx={{ width: 300 }}
              />
              <Button variant="outlined" startIcon={<FilterIcon />} sx={{ borderRadius: 3 }}>
                Filtrer
              </Button>
              <Button variant="outlined" startIcon={<DownloadIcon />} sx={{ borderRadius: 3 }}>
                Export
              </Button>
            </Box>
          </Box>

          <TableContainer>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Device</TableCell>
                  <TableCell>Statut</TableCell>
                  <TableCell>🔐 ZKP Status</TableCell>
                  <TableCell>📊 Auths</TableCell>
                  <TableCell>⏱️ Latence</TableCell>
                  <TableCell>🕐 Dernière connexion</TableCell>
                  <TableCell align="right">Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {filteredDevices.map((device) => {
                  const zkpStatus = getZKPStatus(device);
                  return (
                    <TableRow key={device.address || device.id} hover>
                      <TableCell>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                          <Avatar sx={{ 
                            bgcolor: device.zkpVerified ? 'success.main' : 'warning.main', 
                            width: 48, 
                            height: 48 
                          }}>
                            {device.name?.[0] || 'D'}
                          </Avatar>
                          <Box>
                            <Typography variant="body2" fontWeight={600}>{device.name || 'ESP32'}</Typography>
                            <Typography variant="caption" color="text.secondary" sx={{ fontFamily: 'monospace', fontSize: 10 }}>
                              {device.address?.substring(0, 20)}...
                            </Typography>
                          </Box>
                        </Box>
                      </TableCell>
                      
                      <TableCell>
                        <Chip
                          icon={<StatusIcon sx={{ fontSize: 12 }} />}
                          label={device.isActive ? 'Actif' : 'Inactif'}
                          size="small"
                          sx={{
                            bgcolor: device.isActive ? 'success.main' : 'grey.500',
                            color: 'white', 
                            textTransform: 'capitalize', 
                            fontWeight: 600,
                          }}
                        />
                      </TableCell>
                      
                      <TableCell>
                        <Tooltip title={zkpStatus.detail}>
                          <Box sx={{
                            display: 'inline-flex', 
                            alignItems: 'center', 
                            gap: 1,
                            px: 2, 
                            py: 0.75, 
                            borderRadius: 2,
                            bgcolor: `${zkpStatus.color}.main`,
                            color: 'white', 
                            fontSize: 12, 
                            fontWeight: 600,
                          }}>
                            <zkpStatus.icon sx={{ fontSize: 16 }} />
                            {zkpStatus.label}
                          </Box>
                        </Tooltip>
                      </TableCell>
                      
                      <TableCell>
                        <Typography variant="h6" fontWeight="bold" color={device.authCount > 0 ? 'primary.main' : 'text.secondary'}>
                          {device.authCount || 0}
                        </Typography>
                      </TableCell>
                      
                      <TableCell>
                        <Typography variant="body2" fontWeight={600}>
                          {device.latency || '-'}
                        </Typography>
                      </TableCell>
                      
                      <TableCell>
                        <Typography variant="body2" color="text.primary">
                          {device.lastSeen || 'Inconnu'}
                        </Typography>
                        {device.lastAuthTime && device.lastAuthTime !== 'Jamais' && (
                          <Typography variant="caption" color="text.secondary" sx={{ display: 'block' }}>
                            Auth: {device.lastAuthTime}
                          </Typography>
                        )}
                      </TableCell>
                      
                      <TableCell align="right">
                        <Tooltip title="Voir l'historique">
                          <IconButton 
                            size="small" 
                            sx={{ color: 'info.main', mr: 1 }}
                            onClick={() => handleShowHistory(device)}
                          >
                            <HistoryIcon />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="Authentifier ZKP">
                          <IconButton 
                            size="small" 
                            sx={{ color: device.zkpVerified ? 'success.main' : 'primary.main' }}
                            onClick={() => onAuthenticate && onAuthenticate(device.address)}
                            disabled={!backendConnected}
                          >
                            <ZKPIcon />
                          </IconButton>
                        </Tooltip>
                        <IconButton size="small">
                          <MoreIcon />
                        </IconButton>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </TableContainer>
          
          {filteredDevices.length === 0 && (
            <Box sx={{ textAlign: 'center', py: 4 }}>
              <Typography variant="body1" color="text.secondary">
                Aucun device trouvé
              </Typography>
            </Box>
          )}
        </CardContent>
      </Card>

      {/* Dialog Historique */}
      <Dialog 
        open={historyOpen} 
        onClose={() => setHistoryOpen(false)}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
            <HistoryIcon color="primary" />
            <Box>
              <Typography variant="h6">Historique ZKP</Typography>
              <Typography variant="caption" color="text.secondary">
                {selectedDevice?.name} - {selectedDevice?.address?.substring(0, 20)}...
              </Typography>
            </Box>
          </Box>
        </DialogTitle>
        <DialogContent>
          {deviceHistory.length === 0 ? (
            <Box sx={{ textAlign: 'center', py: 4 }}>
              <Typography variant="body1" color="text.secondary">
                Aucun historique d'authentification
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                Ce device n'a jamais été authentifié avec ZKP
              </Typography>
            </Box>
          ) : (
            <List>
              {deviceHistory.map((record, index) => (
                <ListItem 
                  key={index}
                  divider={index < deviceHistory.length - 1}
                  sx={{
                    bgcolor: record.success ? 'success.main' : 'error.main',
                    opacity: 0.1,
                    borderRadius: 2,
                    mb: 1,
                  }}
                >
                  <ListItemText
                    primary={
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                        {record.success ? (
                          <CheckIcon color="success" />
                        ) : (
                          <ErrorIcon color="error" />
                        )}
                        <Typography variant="body1" fontWeight={600}>
                          {record.success ? 'Authentification réussie' : 'Échec d\'authentification'}
                        </Typography>
                      </Box>
                    }
                    secondary={
                      <Box sx={{ mt: 1 }}>
                        <Typography variant="body2" color="text.secondary">
                          Date: {record.timestamp ? new Date(record.timestamp * 1000).toLocaleString() : 'Inconnue'}
                        </Typography>
                        <Typography variant="caption" color="text.secondary" sx={{ fontFamily: 'monospace' }}>
                          Hash: {record.proofHash || 'N/A'}
                        </Typography>
                      </Box>
                    }
                  />
                </ListItem>
              ))}
            </List>
          )}
          
          {/* Résumé */}
          {deviceHistory.length > 0 && (
            <Box sx={{ mt: 3, p: 2, borderRadius: 2, bgcolor: 'background.default' }}>
              <Typography variant="h6" gutterBottom>Résumé</Typography>
              <Grid container spacing={2}>
                <Grid item xs={4}>
                  <Typography variant="h4" color="success.main">
                    {deviceHistory.filter(h => h.success).length}
                  </Typography>
                  <Typography variant="caption">Succès</Typography>
                </Grid>
                <Grid item xs={4}>
                  <Typography variant="h4" color="error.main">
                    {deviceHistory.filter(h => !h.success).length}
                  </Typography>
                  <Typography variant="caption">Échecs</Typography>
                </Grid>
                <Grid item xs={4}>
                  <Typography variant="h4" color="primary.main">
                    {deviceHistory.length}
                  </Typography>
                  <Typography variant="caption">Total</Typography>
                </Grid>
              </Grid>
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setHistoryOpen(false)}>Fermer</Button>
          <Button 
            variant="contained" 
            onClick={() => {
              setHistoryOpen(false);
              onAuthenticate && onAuthenticate(selectedDevice?.address);
            }}
            disabled={!backendConnected}
          >
            Réauthentifier
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
};

export default DeviceTable;
