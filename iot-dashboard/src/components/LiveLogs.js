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
  return (
    <Paper sx={{ p: 3, borderRadius: 2, boxShadow: 3 }}>
      <Typography variant="h6" gutterBottom>
        📋 Dernières authentifications
      </Typography>
      <List sx={{ maxHeight: 400, overflow: 'auto' }}>
        {logs.map((log, index) => (
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
                  <strong>{log.deviceId}</strong>
                  <Chip
                    size="small"
                    icon={<Fingerprint />}
                    label={`secret: ${log.secret}`}
                    variant="outlined"
                  />
                </Box>
              }
              secondary={
                <Box sx={{ display: 'flex', gap: 2, mt: 0.5 }}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                    <AccessTime fontSize="small" color="action" />
                    <span>{format(log.timestamp, 'HH:mm:ss')}</span>
                  </Box>
                  <Chip
                    size="small"
                    label={`${log.latency}ms`}
                    color={log.latency < 50 ? 'success' : log.latency < 100 ? 'warning' : 'error'}
                  />
                  <Chip
                    size="small"
                    label={`hash: ${log.hash}`}
                    variant="outlined"
                  />
                </Box>
              }
            />
          </ListItem>
        ))}
      </List>
    </Paper>
  );
};

export default LiveLogs;
