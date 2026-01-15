import axios from 'axios';
import type { RouteRequest, RouteResponse, DelayInfo, WeatherInfo } from '../types';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:3000';

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const api = {
  // Calculate optimized routes
  calculateRoute: async (request: RouteRequest): Promise<RouteResponse> => {
    const response = await apiClient.post<RouteResponse>('/api/calculate-route', request);
    return response.data;
  },

  // Get current train delays
  getDelays: async (): Promise<DelayInfo[]> => {
    const response = await apiClient.get<{ delays: DelayInfo[] }>('/api/delays');
    return response.data.delays;
  },

  // Get current weather
  getWeather: async (): Promise<WeatherInfo> => {
    const response = await apiClient.get<WeatherInfo>('/api/weather');
    return response.data;
  },

  // Health check
  healthCheck: async (): Promise<{ status: string; version: string }> => {
    const response = await apiClient.get('/health');
    return response.data;
  },
};

export default apiClient;
