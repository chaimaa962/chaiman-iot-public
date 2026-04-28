import React, { useState } from 'react';
import { Box, Drawer, AppBar, Toolbar, IconButton, Typography, Avatar, Badge, Tooltip } from '@mui/material';
import {
  Menu as MenuIcon, Dashboard as DashboardIcon, Security as SecurityIcon,
  Psychology as AIIcon, DeviceHub as DevicesIcon, Assessment as ReportsIcon,
  Settings as SettingsIcon, Notifications as NotificationsIcon,
  Brightness4 as DarkIcon, Brightness7 as LightIcon,
} from '@mui/icons-material';
import { useThemeMode } from '../../contexts/ThemeContext';
import { Link, useLocation } from 'react-router-dom';

const drawerWidth = 280;

const menuItems = [
  { icon: DashboardIcon, label: 'Dashboard', path: '/' },
  { icon: SecurityIcon, label: 'ZKP Security', path: '/zkp' },
  { icon: AIIcon, label: 'AI Analytics', path: '/ai', badge: 'NEW' },
  { icon: DevicesIcon, label: 'Devices', path: '/devices', badge: 20 },
  { icon: ReportsIcon, label: 'Reports', path: '/reports' },
  { icon: SettingsIcon, label: 'Settings', path: '/settings' },
];

const Layout = ({ children }) => {
  const [mobileOpen, setMobileOpen] = useState(false);
  const { mode, toggleMode } = useThemeMode();
  const location = useLocation();

  const handleDrawerToggle = () => setMobileOpen(!mobileOpen);

  const drawer = (
    <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      <Box sx={{ p: 3, display: 'flex', alignItems: 'center', gap: 2 }}>
        <Box sx={{
          width: 48, height: 48, borderRadius: 3,
          background: 'linear-gradient(135deg, #00D4AA 0%, #00D4FF 100%)',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
        }}>
          <SecurityIcon sx={{ color: 'white', fontSize: 28 }} />
        </Box>
        <Box>
          <Typography variant="h6" fontWeight="bold">SecureIoT</Typography>
          <Typography variant="caption" color="text.secondary">ZKP + AI Powered</Typography>
        </Box>
      </Box>

      <Box sx={{ flex: 1, px: 2, py: 2 }}>
        {menuItems.map((item) => (
          <Box
            key={item.path}
            component={Link}
            to={item.path}
            sx={{
              display: 'flex', alignItems: 'center', gap: 2,
              p: 2, mb: 1, borderRadius: 3, textDecoration: 'none',
              background: location.pathname === item.path
                ? 'linear-gradient(135deg, rgba(0, 212, 170, 0.15) 0%, rgba(0, 212, 255, 0.15) 100%)'
                : 'transparent',
              border: location.pathname === item.path
                ? '1px solid rgba(0, 212, 170, 0.3)'
                : '1px solid transparent',
              '&:hover': { background: 'rgba(0, 212, 170, 0.08)' },
            }}
          >
            <item.icon sx={{ color: location.pathname === item.path ? '#00D4AA' : 'text.secondary', fontSize: 24 }} />
            <Typography sx={{
              flex: 1, textDecoration: 'none',
              color: location.pathname === item.path ? 'text.primary' : 'text.secondary',
              fontWeight: location.pathname === item.path ? 600 : 400,
            }}>
              {item.label}
            </Typography>
            {item.badge && (
              <Badge badgeContent={item.badge} color="error" sx={{ '& .MuiBadge-badge': { fontSize: 10 } }} />
            )}
          </Box>
        ))}
      </Box>

      <Box sx={{ p: 2, borderTop: 1, borderColor: 'divider' }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, p: 2, borderRadius: 3, bgcolor: 'background.paper' }}>
          <Avatar sx={{ bgcolor: 'primary.main' }}>A</Avatar>
          <Box sx={{ flex: 1 }}>
            <Typography variant="body2" fontWeight={600}>Admin User</Typography>
            <Typography variant="caption" color="text.secondary">Blockchain Admin</Typography>
          </Box>
        </Box>
      </Box>
    </Box>
  );

  return (
    <Box sx={{ display: 'flex', minHeight: '100vh', bgcolor: 'background.default' }}>
      <AppBar position="fixed" sx={{
        width: { sm: `calc(100% - ${drawerWidth}px)` },
        ml: { sm: `${drawerWidth}px` },
        bgcolor: 'background.paper',
        borderBottom: 1,
        borderColor: 'divider',
      }}>
        <Toolbar sx={{ justifyContent: 'space-between' }}>
          <Box sx={{ display: 'flex', alignItems: 'center' }}>
            <IconButton color="inherit" edge="start" onClick={handleDrawerToggle} sx={{ mr: 2, display: { sm: 'none' } }}>
              <MenuIcon />
            </IconButton>
            <Typography variant="h5" fontWeight="bold">Dashboard</Typography>
          </Box>

          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
            <Tooltip title={mode === 'dark' ? 'Light Mode' : 'Dark Mode'}>
              <IconButton onClick={toggleMode} sx={{ color: 'text.primary' }}>
                {mode === 'dark' ? <LightIcon /> : <DarkIcon />}
              </IconButton>
            </Tooltip>
            <Badge badgeContent={3} color="error">
              <IconButton sx={{ color: 'text.primary' }}><NotificationsIcon /></IconButton>
            </Badge>
            <Box sx={{
              px: 2, py: 0.5, borderRadius: 10, bgcolor: 'success.main',
              color: 'white', fontSize: 12, fontWeight: 600, display: 'flex', alignItems: 'center', gap: 1,
            }}>
              <Box sx={{ width: 8, height: 8, borderRadius: '50%', bgcolor: 'white' }} />
              Live
            </Box>
          </Box>
        </Toolbar>
      </AppBar>

      <Box component="nav" sx={{ width: { sm: drawerWidth }, flexShrink: { sm: 0 } }}>
        <Drawer
          variant="temporary"
          open={mobileOpen}
          onClose={handleDrawerToggle}
          ModalProps={{ keepMounted: true }}
          sx={{ display: { xs: 'block', sm: 'none' }, '& .MuiDrawer-paper': { width: drawerWidth } }}
        >
          {drawer}
        </Drawer>
        <Drawer
          variant="permanent"
          sx={{ display: { xs: 'none', sm: 'block' }, '& .MuiDrawer-paper': { width: drawerWidth } }}
          open
        >
          {drawer}
        </Drawer>
      </Box>

      <Box component="main" sx={{
        flexGrow: 1, p: 4,
        width: { sm: `calc(100% - ${drawerWidth}px)` },
        mt: 8, bgcolor: 'background.default', minHeight: '100vh',
      }}>
        {children}
      </Box>
    </Box>
  );
};

export default Layout;
