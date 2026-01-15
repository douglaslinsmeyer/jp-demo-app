import { useMutation, useQuery } from '@tanstack/react-query';
import { api } from '../api/client';
import type { RouteRequest } from '../types';

export const useRouteCalculation = () => {
  return useMutation({
    mutationFn: (request: RouteRequest) => api.calculateRoute(request),
  });
};

export const useDelays = () => {
  return useQuery({
    queryKey: ['delays'],
    queryFn: () => api.getDelays(),
    refetchInterval: 2 * 60 * 1000, // Refetch every 2 minutes
    staleTime: 1 * 60 * 1000, // Consider data stale after 1 minute
  });
};

export const useWeather = () => {
  return useQuery({
    queryKey: ['weather'],
    queryFn: () => api.getWeather(),
    refetchInterval: 15 * 60 * 1000, // Refetch every 15 minutes
    staleTime: 10 * 60 * 1000, // Consider data stale after 10 minutes
  });
};
