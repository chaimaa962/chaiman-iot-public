// src/contexts/ThemeContext.js
import React, { createContext, useContext, useState, useEffect } from 'react';
import { ThemeProvider as MUIThemeProvider, createTheme } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';

const ThemeContext = createContext();

export const useThemeMode = () => useContext(ThemeContext);

const getDesignTokens = (mode) => ({
  palette: {
    mode,
    primary: {
      main: mode === 'dark' ? '#00D4AA' : '#0066FF',
      light: mode === 'dark' ? '#33E0BB' : '#3385FF',
      dark: mode === 'dark' ? '#00AA88' : '#0052CC',
    },
    secondary: {
      main: mode === 'dark' ? '#FF6B6B' : '#FF4757',
    },
    background: {
      default: mode === 'dark' ? '#0A0E27' : '#F0F4F8',
      paper: mode === 'dark' ? '#151B3D' : '#FFFFFF',
    },
    text: {
      primary: mode === 'dark' ? '#E8ECF4' : '#1A202C',
      secondary: mode === 'dark' ? '#8B9DC3' : '#4A5568',
    },
  },
  typography: {
    fontFamily: '"Inter", "Roboto", "Helvetica", "Arial", sans-serif',
    h1: { fontWeight: 800 },
    h2: { fontWeight: 700 },
    h3: { fontWeight: 700 },
    h6: { fontWeight: 600 },
    button: { textTransform: 'none' },
  },
  shape: { borderRadius: 16 },
});

export const ThemeProvider = ({ children }) => {
  const [mode, setMode] = useState(() => {
    const saved = localStorage.getItem('theme-mode');
    return saved || 'dark';
  });

  useEffect(() => {
    localStorage.setItem('theme-mode', mode);
  }, [mode]);

  const toggleMode = () => {
    setMode((prev) => (prev === 'light' ? 'dark' : 'light'));
  };

  const theme = createTheme(getDesignTokens(mode));

  return (
    <ThemeContext.Provider value={{ mode, toggleMode }}>
      <MUIThemeProvider theme={theme}>
        <CssBaseline />
        {children}
      </MUIThemeProvider>
    </ThemeContext.Provider>
  );
};
