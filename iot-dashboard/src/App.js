import React from 'react';
import { ThemeProvider } from './contexts/ThemeContext';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Dashboard from './components/dashboard/Dashboard';
import AIAnalytics from './components/ai/AIAnalytics';
import ZKPSecurity from './components/zkp/ZKPSecurity';
import Devices from './components/devices/Devices';  

function App() {
  return (
    <ThemeProvider>
      <Router>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/ai" element={<AIAnalytics />} />
          <Route path="/zkp" element={<ZKPSecurity />} />
          <Route path="/devices" element={<Devices />} />  
        </Routes>
      </Router>
    </ThemeProvider>
  );
}

export default App;
