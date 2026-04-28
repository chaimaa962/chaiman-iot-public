import React, { useState, useEffect } from 'react';
import { Line } from 'react-chartjs-2';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js';
import { Paper, Typography, Box } from '@mui/material';

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
);

const RealTimeChart = ({ data }) => {
  const [chartData, setChartData] = useState({
    labels: [],
    datasets: [
      {
        label: 'Authentifications réussies',
        data: [],
        borderColor: '#4caf50',
        backgroundColor: 'rgba(76, 175, 80, 0.1)',
        tension: 0.4,
        fill: true,
      },
      {
        label: 'Temps de réponse (ms)',
        data: [],
        borderColor: '#2196f3',
        backgroundColor: 'rgba(33, 150, 243, 0.1)',
        tension: 0.4,
        fill: true,
        yAxisID: 'y1',
      },
    ],
  });

  const options = {
    responsive: true,
    interaction: {
      mode: 'index',
      intersect: false,
    },
    plugins: {
      legend: {
        position: 'top',
      },
      title: {
        display: true,
        text: 'Authentifications ZKP en temps réel',
      },
    },
    scales: {
      y: {
        type: 'linear',
        display: true,
        position: 'left',
        title: {
          display: true,
          text: 'Nombre d\'auth',
        },
      },
      y1: {
        type: 'linear',
        display: true,
        position: 'right',
        title: {
          display: true,
          text: 'Temps (ms)',
        },
        grid: {
          drawOnChartArea: false,
        },
      },
    },
  };

  useEffect(() => {
    if (data && data.length > 0) {
      setChartData({
        labels: data.map(d => d.time),
        datasets: [
          {
            ...chartData.datasets[0],
            data: data.map(d => d.success),
          },
          {
            ...chartData.datasets[1],
            data: data.map(d => d.latency),
          },
        ],
      });
    }
  }, [data]);

  return (
    <Paper sx={{ p: 3, borderRadius: 2, boxShadow: 3, mb: 3 }}>
      <Typography variant="h6" gutterBottom>
        📈 Authentifications en temps réel
      </Typography>
      <Box sx={{ height: 400 }}>
        <Line data={chartData} options={options} />
      </Box>
    </Paper>
  );
};

export default RealTimeChart;
