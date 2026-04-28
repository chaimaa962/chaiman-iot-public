import React, { useState, useEffect } from 'react';
import {
  Paper,
  Grid,
  Typography,
  Box,
  CircularProgress,
} from '@mui/material';
import {
  Devices,
  History,
  CheckCircle,
  Timeline,
} from '@mui/icons-material';
import { getGlobalStats } from '../services/api';

const StatCard = ({ icon: Icon, title, value, color }) => (
  <Paper sx={{ p: 2, display: 'flex', alignItems: 'center', gap: 2 }}>
    <Icon sx={{ fontSize: 40, color }} />
    <Box>
      <Typography color="textSecondary" variant="caption">
        {title}
      </Typography>
      <Typography variant="h5">{value}</Typography>
    </Box>
  </Paper>
);

const GlobalStats = () => {
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadStats();
    const interval = setInterval(loadStats, 10000);
    return () => clearInterval(interval);
  }, []);

  const loadStats = async () => {
    const data = await getGlobalStats();
    setStats(data);
    setLoading(false);
  };

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Grid container spacing={2} sx={{ mb: 3 }}>
      <Grid item xs={12} sm={6} md={3}>
        <StatCard
          icon={Devices}
          title="Appareils"
          value={stats?.totalDevices || 0}
          color="#2196f3"
        />
      </Grid>
      <Grid item xs={12} sm={6} md={3}>
        <StatCard
          icon={History}
          title="Authentifications"
          value={stats?.totalAuths || 0}
          color="#4caf50"
        />
      </Grid>
      <Grid item xs={12} sm={6} md={3}>
        <StatCard
          icon={CheckCircle}
          title="Appareils actifs"
          value={stats?.activeDevices || 0}
          color="#ff9800"
        />
      </Grid>
      <Grid item xs={12} sm={6} md={3}>
        <StatCard
          icon={Timeline}
          title="Taux succès"
          value={stats?.totalAuths ? 
            `${Math.round((stats.activeDevices / stats.totalDevices) * 100)}%` : 
            '0%'}
          color="#9c27b0"
        />
      </Grid>
    </Grid>
  );
};

export default GlobalStats;
