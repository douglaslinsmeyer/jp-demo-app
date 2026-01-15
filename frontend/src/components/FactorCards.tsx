import type { FactorInfo } from '../types';

interface FactorCardsProps {
  factors: FactorInfo;
}

export default function FactorCards({ factors }: FactorCardsProps) {
  const { delays, weather, crowdingEstimate } = factors;

  const getWeatherIcon = (condition: string) => {
    switch (condition.toUpperCase()) {
      case 'CLEAR':
        return '☀️';
      case 'PARTLY_CLOUDY':
        return '⛅';
      case 'RAIN':
      case 'RAIN_SHOWERS':
      case 'DRIZZLE':
        return '🌧️';
      case 'SNOW':
      case 'SNOW_SHOWERS':
        return '🌨️';
      case 'THUNDERSTORM':
      case 'THUNDERSTORM_WITH_HAIL':
        return '⛈️';
      case 'FOG':
        return '🌫️';
      default:
        return '☁️';
    }
  };

  const getCrowdingColor = (level: string) => {
    switch (level) {
      case 'LOW':
        return 'bg-green-100 text-green-800 border-green-200';
      case 'MODERATE':
        return 'bg-yellow-100 text-yellow-800 border-yellow-200';
      case 'HIGH':
        return 'bg-orange-100 text-orange-800 border-orange-200';
      case 'VERY_HIGH':
        return 'bg-red-100 text-red-800 border-red-200';
      default:
        return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  const getDelayStatusColor = (status: string) => {
    switch (status) {
      case 'NORMAL':
        return 'text-green-600';
      case 'DELAYED':
        return 'text-yellow-600';
      case 'SUSPENDED':
        return 'text-red-600';
      default:
        return 'text-gray-600';
    }
  };

  return (
    <div className="w-full max-w-4xl mx-auto mt-8">
      <h3 className="text-xl font-bold text-gray-800 mb-4">Current Conditions</h3>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {/* Weather Card */}
        <div className="bg-white rounded-lg shadow-md p-6">
          <div className="flex items-center justify-between mb-4">
            <h4 className="text-sm font-semibold text-gray-700">Weather</h4>
            <span className="text-3xl">{getWeatherIcon(weather.condition)}</span>
          </div>

          <div className="space-y-2">
            <div>
              <p className="text-3xl font-bold text-gray-900">{weather.temperature.toFixed(1)}°C</p>
              <p className="text-sm text-gray-600 capitalize">
                {weather.condition.toLowerCase().replace(/_/g, ' ')}
              </p>
            </div>

            {weather.precipitation > 0 && (
              <div className="flex items-center text-sm text-blue-600">
                <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z" />
                </svg>
                {weather.precipitation.toFixed(1)} mm/h
              </div>
            )}

            <p className="text-xs text-gray-500 mt-2">
              Updated: {new Date(weather.lastUpdated).toLocaleTimeString()}
            </p>
          </div>
        </div>

        {/* Delays Card */}
        <div className="bg-white rounded-lg shadow-md p-6">
          <div className="flex items-center justify-between mb-4">
            <h4 className="text-sm font-semibold text-gray-700">Train Status</h4>
            <svg className="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
            </svg>
          </div>

          <div className="space-y-3">
            {delays.length === 0 ? (
              <p className="text-sm text-gray-600">No delays reported</p>
            ) : (
              delays.slice(0, 3).map((delay, index) => (
                <div key={index} className="border-l-4 border-gray-200 pl-3">
                  <div className="flex items-center justify-between">
                    <p className="text-sm font-medium text-gray-800">{delay.lineName}</p>
                    {delay.delayMinutes > 0 && (
                      <span className="text-xs font-semibold text-red-600">+{delay.delayMinutes}m</span>
                    )}
                  </div>
                  <p className={`text-xs ${getDelayStatusColor(delay.status)}`}>
                    {delay.status === 'NORMAL' ? 'On time' : delay.message || delay.status}
                  </p>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Crowding Card */}
        <div className="bg-white rounded-lg shadow-md p-6">
          <div className="flex items-center justify-between mb-4">
            <h4 className="text-sm font-semibold text-gray-700">Crowding</h4>
            <svg className="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
            </svg>
          </div>

          <div className="space-y-3">
            <div className={`px-4 py-2 rounded-lg border-2 ${getCrowdingColor(crowdingEstimate.level)}`}>
              <p className="text-lg font-bold text-center">{crowdingEstimate.level}</p>
            </div>
            <p className="text-sm text-gray-600">{crowdingEstimate.description}</p>
          </div>
        </div>
      </div>
    </div>
  );
}
