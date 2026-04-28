import React, { useState } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Typography,
  Box,
  Chip,
  Divider,
  Paper,
  Grid,
  IconButton,
  Tab,
  Tabs,
} from '@mui/material';
import {
  Verified,
  Warning,
  Security,
  QrCode,
  Close,
  ContentCopy,
  History,
} from '@mui/icons-material';
import DeviceHistory from './DeviceHistory';

const DeviceDetail = ({ device, open, onClose, onAuthenticate }) => {
  const [tabValue, setTabValue] = useState(0);

  if (!device) return null;

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
  };

  const formatDate = (timestamp) => {
    if (!timestamp) return 'Jamais';
    return new Date(timestamp).toLocaleString();
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <Security />
            <Typography variant="h6">Détails de l'appareil</Typography>
          </Box>
          <IconButton onClick={onClose}>
            <Close />
          </IconButton>
        </Box>
      </DialogTitle>

      <DialogContent dividers>
        <Tabs value={tabValue} onChange={(e, v) => setTabValue(v)} sx={{ mb: 2 }}>
          <Tab label="Informations" />
          <Tab label="Historique" icon={<History />} iconPosition="start" />
        </Tabs>

        {tabValue === 0 && (
          <Grid container spacing={3}>
            {/* Adresse */}
            <Grid item xs={12}>
              <Paper sx={{ p: 2, bgcolor: '#f8f9fa' }}>
                <Typography variant="subtitle2" color="textSecondary">
                  Adresse
                </Typography>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <Typography variant="body1" sx={{ fontFamily: 'monospace' }}>
                    {device.address}
                  </Typography>
                  <IconButton size="small" onClick={() => copyToClipboard(device.address)}>
                    <ContentCopy fontSize="small" />
                  </IconButton>
                </Box>
              </Paper>
            </Grid>

            {/* Statut et Dernière connexion */}
            <Grid item xs={6}>
              <Typography variant="subtitle2" color="textSecondary" gutterBottom>
                Statut
              </Typography>
              {device.isActive ? (
                <Chip
                  icon={<Verified />}
                  label="Actif"
                  color="success"
                  sx={{ fontWeight: 500 }}
                />
              ) : (
                <Chip
                  icon={<Warning />}
                  label="Inactif"
                  color="error"
                  sx={{ fontWeight: 500 }}
                />
              )}
            </Grid>

            <Grid item xs={6}>
              <Typography variant="subtitle2" color="textSecondary" gutterBottom>
                Dernière connexion
              </Typography>
              <Typography variant="body1">
                {formatDate(device.lastSeen)}
              </Typography>
            </Grid>

            {/* Séparateur ZKP */}
            <Grid item xs={12}>
              <Divider sx={{ my: 2 }} />
              <Typography variant="h6" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <Security color="primary" />
                Informations ZKP
              </Typography>
            </Grid>

            {/* Secret et Hash */}
            <Grid item xs={6}>
              <Typography variant="subtitle2" color="textSecondary" gutterBottom>
                Secret (privé)
              </Typography>
              <Typography variant="body1" sx={{ fontFamily: 'monospace' }}>
                {device.secret || 'N/A'}
              </Typography>
            </Grid>

            <Grid item xs={6}>
              <Typography variant="subtitle2" color="textSecondary" gutterBottom>
                Hash (public)
              </Typography>
              <Typography variant="body1" sx={{ fontFamily: 'monospace' }}>
                {device.hash || 'N/A'}
              </Typography>
            </Grid>

            {/* Clé publique */}
            <Grid item xs={12}>
              <Typography variant="subtitle2" color="textSecondary" gutterBottom>
                Clé publique
              </Typography>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <Typography variant="body2" sx={{ fontFamily: 'monospace', wordBreak: 'break-all' }}>
                  {device.publicKey || 'N/A'}
                </Typography>
                {device.publicKey && (
                  <IconButton size="small" onClick={() => copyToClipboard(device.publicKey)}>
                    <ContentCopy fontSize="small" />
                  </IconButton>
                )}
              </Box>
            </Grid>

            {/* Statut ZKP */}
            <Grid item xs={12}>
              <Typography variant="subtitle2" color="textSecondary" gutterBottom>
                Preuve ZKP
              </Typography>
              {device.zkpValid ? (
                <Chip
                  icon={<Verified />}
                  label="Valide"
                  color="success"
                  sx={{ fontWeight: 500 }}
                />
              ) : (
                <Chip
                  icon={<Warning />}
                  label="Invalide"
                  color="error"
                  sx={{ fontWeight: 500 }}
                />
              )}
            </Grid>

            {/* QR Code */}
            <Grid item xs={12}>
              <Box sx={{ display: 'flex', justifyContent: 'center', mt: 2 }}>
                <QrCode sx={{ fontSize: 128, color: '#666' }} />
              </Box>
            </Grid>
          </Grid>
        )}

        {tabValue === 1 && (
          <DeviceHistory deviceAddress={device.address} />
        )}
      </DialogContent>

      <DialogActions>
        <Button onClick={onClose}>Fermer</Button>
        <Button
          variant="contained"
          color="primary"
          onClick={() => {
            onAuthenticate(device.address);
            onClose();
          }}
          startIcon={<Security />}
        >
          Authentifier avec ZKP
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default DeviceDetail;
