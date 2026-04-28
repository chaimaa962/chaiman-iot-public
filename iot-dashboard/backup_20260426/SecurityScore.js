import React from 'react';
import { Card, CardContent, Typography, Box, LinearProgress } from '@mui/material';
import { Security } from '@mui/icons-material';

const SecurityScore = ({ score, threats, lastUpdate }) => {
  return (
    <Card sx={{ mb: 3, background: 'linear-gradient(135deg, rgba(0, 212, 170, 0.1) 0%, rgba(0, 212, 255, 0.1) 100%)' }}>
      <CardContent>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
          <Security sx={{ fontSize: 48, color: 'primary.main' }} />
          <Box sx={{ flex: 1 }}>
            <Typography variant="h6">Security Score</Typography>
            <Typography variant="h3" fontWeight="bold">{score}/100</Typography>
            <LinearProgress variant="determinate" value={score} sx={{ mt: 1, height: 8, borderRadius: 4 }} />
          </Box>
          <Box sx={{ textAlign: 'right' }}>
            <Typography variant="body2" color="text.secondary">Threats: {threats}</Typography>
            <Typography variant="caption" color="text.secondary">Updated {lastUpdate}</Typography>
          </Box>
        </Box>
      </CardContent>
    </Card>
  );
};

export default SecurityScore;
