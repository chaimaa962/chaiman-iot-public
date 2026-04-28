import React from 'react';
import {
  AppBar,
  Toolbar,
  Typography,
  Button,
  Box,
  IconButton,
  Badge,
  Tooltip,
} from '@mui/material';
import {
  Biotech,
  Refresh,
  Notifications,
  Settings,
  Dashboard,
} from '@mui/icons-material';

const Navbar = ({ onRefresh }) => {
  return (
    <AppBar position="static" sx={{ mb: 3 }}>
      <Toolbar>
        <Biotech sx={{ mr: 2, fontSize: 28 }} />
        <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
          IoT Dashboard avec ZKP
        </Typography>

        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <Tooltip title="Dashboard">
            <IconButton color="inherit">
              <Dashboard />
            </IconButton>
          </Tooltip>

          <Tooltip title="Notifications">
            <IconButton color="inherit">
              <Badge badgeContent={3} color="error">
                <Notifications />
              </Badge>
            </IconButton>
          </Tooltip>

          <Tooltip title="Paramètres">
            <IconButton color="inherit">
              <Settings />
            </IconButton>
          </Tooltip>

          <Tooltip title="Rafraîchir">
            <Button
              variant="outlined"
              color="inherit"
              onClick={onRefresh}
              startIcon={<Refresh />}
              sx={{ ml: 2 }}
            >
              Rafraîchir
            </Button>
          </Tooltip>
        </Box>
      </Toolbar>
    </AppBar>
  );
};

export default Navbar;
