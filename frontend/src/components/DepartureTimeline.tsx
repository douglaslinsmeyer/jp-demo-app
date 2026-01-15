interface DepartureTimelineProps {
  optimalDepartureTime: string;
}

export default function DepartureTimeline({ optimalDepartureTime }: DepartureTimelineProps) {
  const formatTime = (dateString: string): string => {
    const date = new Date(dateString);
    return date.toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit',
      hour12: true
    });
  };

  const formatDate = (dateString: string): string => {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      weekday: 'short',
      month: 'short',
      day: 'numeric'
    });
  };

  return (
    <div className="w-full max-w-2xl mx-auto mt-8">
      <div className="bg-gradient-to-r from-blue-500 to-blue-600 rounded-lg shadow-lg p-8 text-white">
        <div className="text-center">
          <p className="text-sm font-medium opacity-90 mb-2">Optimal Departure Time</p>
          <p className="text-5xl font-bold mb-2">{formatTime(optimalDepartureTime)}</p>
          <p className="text-sm opacity-75">{formatDate(optimalDepartureTime)}</p>
        </div>

        <div className="mt-6 flex items-center justify-center space-x-2">
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p className="text-sm">
            Based on current traffic, weather, and crowding conditions
          </p>
        </div>
      </div>
    </div>
  );
}
