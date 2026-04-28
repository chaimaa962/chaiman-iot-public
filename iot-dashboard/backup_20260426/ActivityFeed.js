import React, { useState, useEffect } from 'react';
import { 
  Card, CardContent, Typography, List, ListItem, 
  ListItemText, Chip, Box, IconButton, Badge 
} from '@mui/material';
import { CheckCircle, Error, Refresh, RadioButtonChecked } from '@mui/icons-material';

const ActivityFeed = ({ backendConnected, devices }) => {
  const [activities, setActivities] = useState([]);

  useEffect(() => {
    if (devices && devices.length > 0) {
      // Générer activités depuis les devices réels
      const newActivities = devices
        .filter(d => d.authCount > 0 || d.zkp)
        .map((device, index) => ({
          id: index,
          device: device.name,
          address: device.address,
          action: device.zkp ? 'ZKP Auth Success' : 'ZKP Auth Failed',
          time: device.lastAuthTime,
          success: device.zkp,
          authCount: device.authCount,
        }))
        .sort((a, b) => b.authCount - a.authCount)
        .slice(0, 10); // Top 10

      setActivities(newActivities);
    }
  }, [devices]);

  return (
    <Card sx={{ height: 500 }}>
      <CardContent>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
          <Typography variant="h6">🔴 Activité ZKP Temps Réel</Typography>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            {backendConnected && (
              <Badge badgeContent={activities.length} color="primary">
                <Chip 
                  icon={<RadioButtonChecked sx={{ fontSize: 12, animation: 'pulse 1s infinite' }} />}
                  label="LIVE" 
                  size="small" 
                  color="success" 
                  sx={{ color: 'white', fontWeight: 'bold' }} 
                />
              </Badge>
            )}
            <IconButton size="small">
              <Refresh />
            </IconButton>
          </Box>
        </Box>

        <List sx={{ maxHeight: 420, overflow: 'auto' }}>
          {activities.length === 0 ? (
            <Box sx={{ textAlign: 'center', py: 4 }}>
              <Typography color="text.secondary">Aucune activité ZKP</Typography>
              <Typography variant="caption" color="text.secondary">
                Les authentifications apparaîtront ici
              </Typography>
            </Box>
          ) : (
            activities.map((activity) => (
              <ListItem 
                key={activity.id} 
                divider
                sx={{
                  borderLeft: 3,
                  borderColor: activity.success ? 'success.main' : 'error.main',
                  mb: 1,
                  borderRadius: 1,
                }}
              >
                {activity.success ? (
                  <CheckCircle color="success" sx={{ mr: 2 }} />
                ) : (
                  <Error color="error" sx={{ mr: 2 }} />
                )}
                <ListItemText
                  primary={
                    <Box>
                      <Typography variant="body2" fontWeight={600}>
                        {activity.device}
                      </Typography>
                      <Typography variant="caption" color="text.secondary" sx={{ fontFamily: 'monospace' }}>
                        {activity.address?.substring(0, 15)}...
                      </Typography>
                    </Box>
                  }
                  secondary={
                    <Box>
                      <Typography variant="body2" color={activity.success ? 'success.main' : 'error.main'}>
                        {activity.action}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        Auths: {activity.authCount} • {activity.time}
                      </Typography>
                    </Box>
                  }
                />
                <Chip 
                  label={activity.success ? '✓' : '✗'} 
                  size="small" 
                  color={activity.success ? 'success' : 'error'}
                  sx={{ ml: 1 }}
                />
              </ListItem>
            ))
          )}
        </List>
      </CardContent>
    </Card>
  );
};

export default ActivityFeed;
