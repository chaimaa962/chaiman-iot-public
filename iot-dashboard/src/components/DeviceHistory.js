import React, { useState, useEffect } from 'react';
import {
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
  Chip,
  Box,
  CircularProgress,
  Alert,
} from '@mui/material';
import { Verified, Warning, History } from '@mui/icons-material';
import { getDeviceAuthHistory } from '../services/api';

const DeviceHistory = ({ deviceAddress }) => {
  const [history, setHistory] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    loadHistory();
    // Rafraîchir toutes les 30 secondes
    const interval = setInterval(loadHistory, 30000);
    return () => clearInterval(interval);
  }, [deviceAddress]);

  const loadHistory = async () => {
    try {
      setLoading(true);
      const data = await getDeviceAuthHistory(deviceAddress);
      setHistory(data);
      setError(null);
    } catch (err) {
      setError("Erreur chargement historique");
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (timestamp) => {
    if (!timestamp) return 'N/A';
    return new Date(timestamp * 1000).toLocaleString();
  };

  const formatProofHash = (hash) => {
    if (!hash) return 'N/A';
    return `${hash.substring(0, 10)}...${hash.substring(hash.length - 8)}`;
  };

  if (loading && history.length === 0) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return <Alert severity="error">{error}</Alert>;
  }

  return (
    <Paper sx={{ p: 2, mt: 2 }}>
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}>
        <History color="primary" />
        <Typography variant="h6">Historique des authentifications</Typography>
        <Chip 
          label={`${history.length} enregistrements`} 
          size="small" 
          color="primary" 
          variant="outlined"
        />
      </Box>

      {history.length === 0 ? (
        <Alert severity="info">Aucune authentification enregistrée</Alert>
      ) : (
        <TableContainer>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Date</TableCell>
                <TableCell>Statut</TableCell>
                <TableCell>Preuve ZKP</TableCell>
                <TableCell>Hash</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {history.map((record, index) => (
                <TableRow key={index} hover>
                  <TableCell>{formatDate(record.timestamp)}</TableCell>
                  <TableCell>
                    {record.success ? (
                      <Chip
                        icon={<Verified />}
                        label="Succès"
                        color="success"
                        size="small"
                      />
                    ) : (
                      <Chip
                        icon={<Warning />}
                        label="Échec"
                        color="error"
                        size="small"
                      />
                    )}
                  </TableCell>
                  <TableCell>
                    <code>{record.proofHash ? formatProofHash(record.proofHash) : 'N/A'}</code>
                  </TableCell>
                  <TableCell>
                    <code>{record.proofHash ? record.proofHash.substring(0, 8) : 'N/A'}...</code>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}
    </Paper>
  );
};

export default DeviceHistory;
