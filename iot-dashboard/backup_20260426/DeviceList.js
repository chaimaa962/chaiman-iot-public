import React, { useState } from 'react';
import {
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TablePagination,
  Chip,
  IconButton,
  Tooltip,
  Typography,
  Box,
  Button,
  TextField,
  InputAdornment,
  Menu,
  MenuItem,
} from '@mui/material';
import {
  Verified,
  Warning,
  Refresh,
  Info,
  Security,
  Search,
  FilterList,
  Download,
  CheckCircle,
  Cancel,
} from '@mui/icons-material';

const DeviceList = ({ devices, onRefresh, onSelect, onAuthenticate }) => {
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(5);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterAnchor, setFilterAnchor] = useState(null);
  const [statusFilter, setStatusFilter] = useState('all'); // 'all', 'active', 'inactive'
  const [zkpFilter, setZkpFilter] = useState('all'); // 'all', 'valid', 'invalid'

  // Filtrer les appareils
  const filteredDevices = devices.filter(device => {
    // Filtre par recherche
    const matchesSearch = 
      device.address.toLowerCase().includes(searchTerm.toLowerCase()) ||
      (device.name && device.name.toLowerCase().includes(searchTerm.toLowerCase()));
    
    // Filtre par statut
    const matchesStatus = 
      statusFilter === 'all' ? true :
      statusFilter === 'active' ? device.isActive :
      !device.isActive;
    
    // Filtre par ZKP
    const matchesZKP = 
      zkpFilter === 'all' ? true :
      zkpFilter === 'valid' ? device.zkpValid :
      !device.zkpValid;
    
    return matchesSearch && matchesStatus && matchesZKP;
  });

  // Pagination
  const paginatedDevices = filteredDevices.slice(
    page * rowsPerPage, 
    page * rowsPerPage + rowsPerPage
  );

  const getStatusChip = (isActive) => {
    return isActive ? (
      <Chip
        icon={<Verified />}
        label="Actif"
        color="success"
        size="small"
        sx={{ fontWeight: 500 }}
      />
    ) : (
      <Chip
        icon={<Warning />}
        label="Inactif"
        color="error"
        size="small"
        sx={{ fontWeight: 500 }}
      />
    );
  };

  const getZKPStatus = (zkpValid) => {
    return zkpValid ? (
      <Tooltip title="Preuve ZKP valide">
        <CheckCircle sx={{ color: '#4caf50', fontSize: 20 }} />
      </Tooltip>
    ) : (
      <Tooltip title="Preuve ZKP invalide">
        <Cancel sx={{ color: '#f44336', fontSize: 20 }} />
      </Tooltip>
    );
  };

  const formatAddress = (address) => {
    if (!address) return '';
    return `${address.substring(0, 6)}...${address.substring(address.length - 4)}`;
  };

  const formatDate = (timestamp) => {
    if (!timestamp) return 'Jamais';
    const date = new Date(timestamp);
    return date.toLocaleString();
  };

  const handleFilterClick = (event) => {
    setFilterAnchor(event.currentTarget);
  };

  const handleFilterClose = () => {
    setFilterAnchor(null);
  };

  const handleStatusFilterChange = (filter) => {
    setStatusFilter(filter);
    setPage(0);
    handleFilterClose();
  };

  const handleZkpFilterChange = (filter) => {
    setZkpFilter(filter);
    setPage(0);
    handleFilterClose();
  };

  const handleResetFilters = () => {
    setStatusFilter('all');
    setZkpFilter('all');
    setSearchTerm('');
    setPage(0);
    handleFilterClose();
  };

  const handleExportCSV = () => {
    const csvContent = [
      ['Adresse', 'Nom', 'Clé Publique', 'Statut', 'Dernière Connexion', 'ZKP Valide'],
      ...filteredDevices.map(d => [
        d.address,
        d.name || 'ESP32',
        d.publicKey || '',
        d.isActive ? 'Actif' : 'Inactif',
        formatDate(d.lastSeen),
        d.zkpValid ? 'Oui' : 'Non'
      ])
    ].map(row => row.join(',')).join('\n');

    const blob = new Blob([csvContent], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `appareils_${new Date().toISOString().slice(0,10)}.csv`;
    a.click();
  };

  return (
    <Paper sx={{ p: 3, borderRadius: 2, boxShadow: 3 }}>
      {/* En-tête avec recherche et filtres */}
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2, flexWrap: 'wrap', gap: 2 }}>
        <Typography variant="h6">Appareils connectés</Typography>
        
        <Box sx={{ display: 'flex', gap: 2, alignItems: 'center', flexWrap: 'wrap' }}>
          {/* Barre de recherche */}
          <TextField
            size="small"
            placeholder="Rechercher..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <Search fontSize="small" />
                </InputAdornment>
              ),
            }}
            sx={{ minWidth: 200 }}
          />

          {/* Filtres */}
          <Button
            variant="outlined"
            startIcon={<FilterList />}
            onClick={handleFilterClick}
            size="small"
          >
            Filtres
            {(statusFilter !== 'all' || zkpFilter !== 'all') && (
              <Chip 
                size="small" 
                label="!" 
                color="primary" 
                sx={{ ml: 1, width: 20, height: 20 }} 
              />
            )}
          </Button>

          {/* Export CSV */}
          <Tooltip title="Exporter en CSV">
            <IconButton onClick={handleExportCSV} size="small">
              <Download />
            </IconButton>
          </Tooltip>

          {/* Rafraîchir */}
          <Tooltip title="Rafraîchir">
            <IconButton onClick={onRefresh} size="small">
              <Refresh />
            </IconButton>
          </Tooltip>
        </Box>
      </Box>

      {/* Menu des filtres */}
      <Menu
        anchorEl={filterAnchor}
        open={Boolean(filterAnchor)}
        onClose={handleFilterClose}
      >
        <MenuItem disabled><Typography variant="caption">Filtre par statut</Typography></MenuItem>
        <MenuItem onClick={() => handleStatusFilterChange('all')} selected={statusFilter === 'all'}>
          Tous les statuts
        </MenuItem>
        <MenuItem onClick={() => handleStatusFilterChange('active')} selected={statusFilter === 'active'}>
          <Verified fontSize="small" sx={{ color: '#4caf50', mr: 1 }} /> Actifs seulement
        </MenuItem>
        <MenuItem onClick={() => handleStatusFilterChange('inactive')} selected={statusFilter === 'inactive'}>
          <Warning fontSize="small" sx={{ color: '#f44336', mr: 1 }} /> Inactifs seulement
        </MenuItem>
        
        <MenuItem disabled><Typography variant="caption" sx={{ mt: 1 }}>Filtre ZKP</Typography></MenuItem>
        <MenuItem onClick={() => handleZkpFilterChange('all')} selected={zkpFilter === 'all'}>
          Tous
        </MenuItem>
        <MenuItem onClick={() => handleZkpFilterChange('valid')} selected={zkpFilter === 'valid'}>
          <CheckCircle fontSize="small" sx={{ color: '#4caf50', mr: 1 }} /> ZKP valide
        </MenuItem>
        <MenuItem onClick={() => handleZkpFilterChange('invalid')} selected={zkpFilter === 'invalid'}>
          <Cancel fontSize="small" sx={{ color: '#f44336', mr: 1 }} /> ZKP invalide
        </MenuItem>
        
        <Box sx={{ p: 1, mt: 1 }}>
          <Button size="small" fullWidth variant="outlined" onClick={handleResetFilters}>
            Réinitialiser
          </Button>
        </Box>
      </Menu>

      {/* Tableau */}
      <TableContainer>
        <Table>
          <TableHead>
            <TableRow sx={{ backgroundColor: '#f5f5f5' }}>
              <TableCell><strong>Adresse</strong></TableCell>
              <TableCell><strong>Nom</strong></TableCell>
              <TableCell><strong>Clé publique</strong></TableCell>
              <TableCell><strong>Statut</strong></TableCell>
              <TableCell><strong>Dernière connexion</strong></TableCell>
              <TableCell align="center"><strong>ZKP</strong></TableCell>
              <TableCell align="center"><strong>Dernier résultat</strong></TableCell>
              <TableCell align="center"><strong>Actions</strong></TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {paginatedDevices.map((device) => (
              <TableRow key={device.address} hover>
                <TableCell>
                  <code style={{ fontSize: '0.9rem' }}>
                    {formatAddress(device.address)}
                  </code>
                </TableCell>
                <TableCell>{device.name || 'ESP32'}</TableCell>
                <TableCell>
                  <code>{device.publicKey?.substring(0, 8)}...</code>
                </TableCell>
                <TableCell>{getStatusChip(device.isActive)}</TableCell>
                <TableCell>{formatDate(device.lastSeen)}</TableCell>
                <TableCell align="center">
                  {getZKPStatus(device.zkpValid)}
                </TableCell>
                <TableCell align="center">
                  {device.lastAuth ? (
                    device.lastAuth.success ? (
                      <Chip label="Succès" color="success" size="small" />
                    ) : (
                      <Chip label="Échec" color="error" size="small" />
                    )
                  ) : (
                    <Typography variant="caption" color="textSecondary">-</Typography>
                  )}
                </TableCell>
                <TableCell align="center">
                  <Box sx={{ display: 'flex', gap: 1, justifyContent: 'center' }}>
                    <Tooltip title="Voir détails">
                      <IconButton size="small" onClick={() => onSelect(device)}>
                        <Info fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="Authentifier avec ZKP">
                      <Button
                        size="small"
                        variant="contained"
                        color="primary"
                        startIcon={<Security />}
                        onClick={() => onAuthenticate(device.address)}
                      >
                        ZKP
                      </Button>
                    </Tooltip>
                  </Box>
                </TableCell>
              </TableRow>
            ))}
            {paginatedDevices.length === 0 && (
              <TableRow>
                <TableCell colSpan={8} align="center" sx={{ py: 3 }}>
                  <Typography color="textSecondary">
                    Aucun appareil trouvé
                  </Typography>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>

      {/* Pagination et informations */}
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mt: 2 }}>
        <Typography variant="caption" color="textSecondary">
          {filteredDevices.length} appareil(s) sur {devices.length} total
        </Typography>
        <TablePagination
          rowsPerPageOptions={[5, 10, 25, 50]}
          component="div"
          count={filteredDevices.length}
          rowsPerPage={rowsPerPage}
          page={page}
          onPageChange={(e, newPage) => setPage(newPage)}
          onRowsPerPageChange={(e) => {
            setRowsPerPage(parseInt(e.target.value, 10));
            setPage(0);
          }}
        />
      </Box>
    </Paper>
  );
};

export default DeviceList;
