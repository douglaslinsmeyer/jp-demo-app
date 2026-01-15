import { useState } from 'react';
import AddressInput from './components/AddressInput';
import DepartureTimeline from './components/DepartureTimeline';
import RouteResults from './components/RouteResults';
import FactorCards from './components/FactorCards';
import { useRouteCalculation } from './hooks/useRouteCalculation';
import type { RouteResponse } from './types';

function App() {
  const [routeData, setRouteData] = useState<RouteResponse | null>(null);
  const calculateRoute = useRouteCalculation();

  const handleCalculate = (origin: string, destination: string) => {
    calculateRoute.mutate(
      { origin, destination },
      {
        onSuccess: (data) => {
          setRouteData(data);
        },
        onError: (error) => {
          console.error('Failed to calculate route:', error);
          alert('Failed to calculate route. Please try again.');
        },
      }
    );
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      <div className="container mx-auto px-4 py-8">
        {/* Header */}
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold text-gray-900 mb-2">
            Tokyo Commute Optimizer
          </h1>
          <p className="text-gray-600">
            Find the best time to leave work based on real-time conditions
          </p>
        </div>

        {/* Address Input */}
        <AddressInput
          onCalculate={handleCalculate}
          isLoading={calculateRoute.isPending}
        />

        {/* Error Message */}
        {calculateRoute.isError && (
          <div className="max-w-2xl mx-auto mt-4">
            <div className="bg-red-50 border border-red-200 rounded-lg p-4">
              <div className="flex items-center">
                <svg
                  className="w-5 h-5 text-red-600 mr-2"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
                <p className="text-red-800">
                  Failed to calculate route. Please try again.
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Results */}
        {routeData && (
          <>
            {/* Optimal Departure Time */}
            <DepartureTimeline optimalDepartureTime={routeData.optimalDepartureTime} />

            {/* Current Conditions */}
            <FactorCards factors={routeData.factors} />

            {/* Route Options */}
            <RouteResults routes={routeData.routes} />
          </>
        )}

        {/* Footer */}
        <div className="mt-16 text-center text-sm text-gray-500">
          <p>Data sources: Google Maps, ODPT Transit API, Open-Meteo Weather</p>
          <p className="mt-1">Updates automatically based on real-time conditions</p>
        </div>
      </div>
    </div>
  );
}

export default App;
