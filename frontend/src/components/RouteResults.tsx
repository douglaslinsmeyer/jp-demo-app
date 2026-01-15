import type { Route } from '../types';

interface RouteResultsProps {
  routes: Route[];
}

export default function RouteResults({ routes }: RouteResultsProps) {
  const formatTime = (seconds: number): string => {
    const mins = Math.floor(seconds / 60);
    return `${mins} min`;
  };

  const formatDistance = (meters: number): string => {
    if (meters < 1000) {
      return `${meters}m`;
    }
    return `${(meters / 1000).toFixed(1)}km`;
  };

  const getScoreColor = (score: number): string => {
    if (score >= 90) return 'text-green-600';
    if (score >= 70) return 'text-yellow-600';
    return 'text-red-600';
  };

  const getScoreBadge = (score: number): string => {
    if (score >= 90) return 'bg-green-100 text-green-800';
    if (score >= 70) return 'bg-yellow-100 text-yellow-800';
    return 'bg-red-100 text-red-800';
  };

  return (
    <div className="w-full max-w-4xl mx-auto mt-8 space-y-4">
      <h3 className="text-xl font-bold text-gray-800 mb-4">Route Options</h3>

      {routes.map((route, index) => (
        <div
          key={index}
          className="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition-shadow"
        >
          {/* Header */}
          <div className="flex justify-between items-start mb-4">
            <div className="flex-1">
              <h4 className="text-lg font-semibold text-gray-800">{route.summary}</h4>
              <p className="text-3xl font-bold text-gray-900 mt-2">
                {formatTime(route.totalTime)}
              </p>
            </div>
            <div className="text-right">
              <span className={`inline-block px-3 py-1 rounded-full text-sm font-medium ${getScoreBadge(route.score)}`}>
                Score: {route.score.toFixed(1)}
              </span>
            </div>
          </div>

          {/* Time Breakdown */}
          <div className="grid grid-cols-2 md:grid-cols-3 gap-4 mb-6">
            <div className="bg-blue-50 p-3 rounded-lg">
              <p className="text-xs text-blue-600 font-medium">Transit</p>
              <p className="text-lg font-bold text-blue-900">{formatTime(route.breakdown.transitTime)}</p>
            </div>
            <div className="bg-green-50 p-3 rounded-lg">
              <p className="text-xs text-green-600 font-medium">Walking</p>
              <p className="text-lg font-bold text-green-900">{formatTime(route.breakdown.walkingTime)}</p>
            </div>
            <div className="bg-purple-50 p-3 rounded-lg">
              <p className="text-xs text-purple-600 font-medium">Waiting</p>
              <p className="text-lg font-bold text-purple-900">{formatTime(route.breakdown.waitingTime)}</p>
            </div>
            {route.breakdown.delayPenalty > 0 && (
              <div className="bg-red-50 p-3 rounded-lg">
                <p className="text-xs text-red-600 font-medium">Delays</p>
                <p className="text-lg font-bold text-red-900">+{formatTime(route.breakdown.delayPenalty)}</p>
              </div>
            )}
            {route.breakdown.weatherPenalty > 0 && (
              <div className="bg-orange-50 p-3 rounded-lg">
                <p className="text-xs text-orange-600 font-medium">Weather</p>
                <p className="text-lg font-bold text-orange-900">+{formatTime(route.breakdown.weatherPenalty)}</p>
              </div>
            )}
            {route.breakdown.crowdingPenalty > 0 && (
              <div className="bg-yellow-50 p-3 rounded-lg">
                <p className="text-xs text-yellow-600 font-medium">Crowding</p>
                <p className="text-lg font-bold text-yellow-900">+{formatTime(route.breakdown.crowdingPenalty)}</p>
              </div>
            )}
          </div>

          {/* Route Steps */}
          {route.steps.length > 0 && (
            <div className="border-t pt-4">
              <h5 className="text-sm font-semibold text-gray-700 mb-3">Journey Steps</h5>
              <div className="space-y-3">
                {route.steps.map((step, stepIndex) => (
                  <div key={stepIndex} className="flex items-start space-x-3">
                    <div className="flex-shrink-0 mt-1">
                      {step.type === 'WALK' && (
                        <svg className="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                        </svg>
                      )}
                      {step.type === 'TRANSIT' && (
                        <svg className="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
                        </svg>
                      )}
                    </div>
                    <div className="flex-1">
                      <p className="text-sm font-medium text-gray-800">{step.instructions}</p>
                      <div className="flex items-center space-x-4 mt-1">
                        <span className="text-xs text-gray-600">{formatTime(step.duration)}</span>
                        {step.distance && (
                          <span className="text-xs text-gray-600">{formatDistance(step.distance)}</span>
                        )}
                        {step.line && (
                          <span
                            className="text-xs font-medium px-2 py-1 rounded"
                            style={{
                              backgroundColor: step.line.color || '#666',
                              color: 'white'
                            }}
                          >
                            {step.line.shortName}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      ))}
    </div>
  );
}
