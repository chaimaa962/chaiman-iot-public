// src/components/dashboard/StatCard.js
import React, { useState, useEffect } from 'react';
import { Card, CardContent, Typography, Box, Chip } from '@mui/material';
import { motion } from 'framer-motion';
import {
  DevicesOther as DevicesIcon,
  Security as SecurityIcon,
  CheckCircle as CheckIcon,
  Warning as WarningIcon,
  TrendingUp as TrendUpIcon,
  TrendingDown as TrendDownIcon,
} from '@mui/icons-material';

const iconMap = {
  devices: DevicesIcon,
  security: SecurityIcon,
  check: CheckIcon,
  warning: WarningIcon,
};

const StatCard = ({ title, value, subtitle, icon, trend, trendUp, color }) => {
  const [displayValue, setDisplayValue] = useState(0);
  const Icon = iconMap[icon] || DevicesIcon;

  useEffect(() => {
    const numValue = typeof value === 'string' ? parseFloat(value) : value;
    if (isNaN(numValue)) {
      setDisplayValue(value);
      return;
    }
    
    let start = 0;
    const duration = 1500;
    const increment = numValue / (duration / 16);
    
    const timer = setInterval(() => {
      start += increment;
      if (start >= numValue) {
        setDisplayValue(numValue);
        clearInterval(timer);
      } else {
        setDisplayValue(Math.floor(start));
      }
    }, 16);
    
    return () => clearInterval(timer);
  }, [value]);

  return (
    <motion.div whileHover={{ y: -5, transition: { duration: 0.2 } }}>
      <Card
        sx={{
          height: '100%',
          position: 'relative',
          overflow: 'hidden',
          '&::before': {
            content: '""',
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            height: 4,
            background: `linear-gradient(90deg, ${color}.main, ${color}.light)`,
          },
        }}
      >
        <CardContent sx={{ p: 3 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 2 }}>
            <Box
              sx={{
                p: 1.5,
                borderRadius: 3,
                bgcolor: `${color}.main`,
                opacity: 0.1,
              }}
            >
              <Icon sx={{ color: `${color}.main`, fontSize: 28 }} />
            </Box>
            <Chip
              icon={trendUp ? <TrendUpIcon /> : <TrendDownIcon />}
              label={trend}
              size="small"
              sx={{
                bgcolor: trendUp ? 'success.main' : 'error.main',
                color: 'white',
                fontWeight: 600,
                '& .MuiChip-icon': { color: 'white', fontSize: 16 },
              }}
            />
          </Box>
          
          <Typography variant="h3" fontWeight="bold" sx={{ mb: 0.5 }}>
            {typeof displayValue === 'number' ? displayValue.toLocaleString() : displayValue}
          </Typography>
          
          <Typography variant="body2" color="text.secondary" gutterBottom>
            {title}
          </Typography>
          
          <Typography variant="caption" color="text.secondary" sx={{ opacity: 0.7 }}>
            {subtitle}
          </Typography>
        </CardContent>
      </Card>
    </motion.div>
  );
};

export default StatCard;
