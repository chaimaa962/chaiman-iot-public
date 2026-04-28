import React, { useState, useEffect } from 'react';
import {
  Paper,
  Typography,
  List,
  ListItem,
  ListItemText,
  ListItemAvatar,
  Avatar,
  Chip,
  Box,
} from '@mui/material';
import {
  CheckCircle,
  Error,
  AccessTime,
  Fingerprint,
} from '@mui/icons-material';
import { format } from 'date-fns';

const LiveLogs = ({ logs }) => {
  // Fonction sécurisée pour formater le timestamp
  const safeFormat = (timestamp) => {
    if (!timestamp) return '--:--:--';
    try {
      const ts = Number(timestamp);
      // Si le timestamp est en secondes (inférieur à l'an 2000 en ms), le convertir en ms
      const date = ts > 1000000000000 ? new Date(ts) : new Date(ts * 1000);
      // Vérifier que la date est valide
      if (isNaN(date.getTime())) return '--:--:--';
      return format(date, 'HH:mm:ss');
    } catch (e) {
      return '--:--:--';
    }
  };

  return (
    <Paper sx={{ p: 3, borderRadius: 2, boxShadow: 3 }}>
      <Typography variant="h6" gutterBottom>
        📋 Dernières authentifications
      </Typography>
      <List sx={{ maxHeight: 400, overflow: 'auto' }}>
        {(logs || []).map((log, index) => (
          <ListItem
            key={index}
            sx={{
              borderBottom: '1px solid #e0e0e0',
              '&:hover': { backgroundColor: '#f5f5f5' },
            }}
          >
            <ListItemAvatar>
              <Avatar sx={{ bgcolor: log.success ? '#4caf50' : '#f44336' }}>
                {log.success ? <CheckCircle /> : <Error />}
              </Avatar>
            </ListItemAvatar>
            <ListItemText
              primary={
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <strong>{log.deviceId || 'Unknown'}</strong>
                  <Chip
                    size="small"
                    icon={<Fingerprint />}
                    label={`secret: ${log.secret || 'N/A'}`}
                    variant="outlined"
                  />
                </Box>
              }
              secondary={
                <Box sx={{ display: 'flex', gap: 2, mt: 0.5 }}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                    <AccessTime fontSize="small" color="action" />
                    <span>{safeFormat(log.timestamp)}</span>
                  </Box>
                  <Chip
                    size="small"
                    label={`${log.latency || 0}ms`}
                    color={log.latency < 50 ? 'success' : log.latency < 100 ? 'warning' : 'error'}
                  />
                  <Chip
                    size="small"
                    label={`hash: ${log.hash ? log.hash.substring(0, 10) : 'N/A'}`}
                    variant="outlined"
                  />
                </Box>
              }
            />
          </ListItem>
        ))}
        {(!logs || logs.length === 0) && (
          <ListItem>
            <ListItemText primary="Aucune authentification récente" sx={{ textAlign: 'center', color: 'text.secondary' }} />
          </ListItem>
        )}
      </List>
    </Paper>
  );
};

export default LiveLogs;
