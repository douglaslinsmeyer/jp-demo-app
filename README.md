# Tokyo Commute Optimizer

A web application to help determine the optimal time to leave work in Tokyo to get home as quickly as possible, considering train schedules, real-time delays, weather conditions, train crowding, and transfer convenience.

## Features

- **Smart Departure Time Recommendations**: Calculate the best time to leave based on multiple factors
- **Real-time Transit Delays**: Integration with ODPT API for Tokyo train delay information
- **Weather Integration**: Factor in current and forecasted weather conditions
- **Crowding Estimates**: Algorithm-based predictions for train crowding levels
- **Route Visualization**: Interactive maps showing your journey with step-by-step directions

## Technology Stack

### Frontend
- React 18+ with Vite
- TypeScript
- TailwindCSS
- React Query (TanStack Query)
- Axios

### Backend
- Go 1.21+
- Gin web framework
- Redis for caching

### APIs
- Google Maps Routes API
- ODPT API (Tokyo transit data)
- Open-Meteo JMA API (Weather)

## Prerequisites

Before you begin, you need to register for API keys:

1. **Google Maps Platform** (https://console.cloud.google.com/)
   - Create a new project
   - Enable APIs: Routes API, Places API, Maps JavaScript API
   - Create an API key
   - Free tier includes $200/month credit

2. **ODPT API** (https://developer.odpt.org/)
   - Register for a developer account
   - Request an API key (takes up to 2 business days)
   - Review API guidelines and terms

3. **Open-Meteo JMA API** (https://open-meteo.com/en/docs/jma-api)
   - No registration required
   - Free to use

## Getting Started

### 1. Clone and Setup

```bash
# Clone the repository
git clone <your-repo-url>
cd japan-ai-demo

# Copy environment variables
cp .env.example .env

# Edit .env and add your API keys
nano .env  # or use your preferred editor
```

### 2. Run with Docker Compose

```bash
# Start all services
docker-compose up

# The application will be available at:
# - Frontend: http://localhost:5173
# - Backend API: http://localhost:3000
# - Redis: localhost:6379
```

### 3. Development Without Docker

#### Frontend

```bash
cd frontend
npm install
npm run dev
```

#### Backend

```bash
cd backend
go mod download
go run cmd/api/main.go
```

#### Redis

```bash
# Install and start Redis locally
redis-server
```

## API Endpoints

### Health Check
```
GET /health
```

### Calculate Route
```
POST /api/calculate-route
{
  "origin": "Shibuya Station, Tokyo",
  "destination": "Shinjuku Station, Tokyo",
  "departureTime": "2026-01-15T18:00:00Z" // optional
}
```

### Get Delays
```
GET /api/delays
```

### Get Weather
```
GET /api/weather
```

## Project Structure

```
.
├── frontend/                 # React frontend
│   ├── src/
│   │   ├── components/      # React components
│   │   ├── hooks/           # Custom React hooks
│   │   ├── types/           # TypeScript types
│   │   └── App.tsx          # Main app component
│   ├── Dockerfile
│   └── package.json
│
├── backend/                  # Go backend
│   ├── cmd/
│   │   └── api/
│   │       └── main.go      # Application entry point
│   ├── internal/
│   │   ├── handlers/        # HTTP handlers
│   │   ├── services/        # Business logic & API integrations
│   │   ├── models/          # Data models
│   │   ├── cache/           # Redis caching
│   │   └── middleware/      # HTTP middleware
│   ├── Dockerfile
│   └── go.mod
│
└── docker-compose.yml        # Docker composition
```

## Testing

### Test Backend API

```bash
# Health check
curl http://localhost:3000/health

# Calculate route
curl -X POST http://localhost:3000/api/calculate-route \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "Shibuya Station, Tokyo",
    "destination": "Shinjuku Station, Tokyo"
  }'
```

### Check Redis Cache

```bash
docker-compose exec redis redis-cli
> KEYS *
> GET <some-key>
```

## Features in Development

- [ ] Google Maps API integration for routing
- [ ] ODPT API integration for real-time delays
- [ ] Weather API integration
- [ ] Crowding estimation algorithm
- [ ] Route optimization logic
- [ ] Frontend UI components
- [ ] Interactive map visualization

## Future Enhancements

- Save favorite routes
- Historical analysis of best departure times
- Push notifications for delays
- Integration with calendar
- Seat availability predictions
- Cost comparison between routes
- Accessibility preferences

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
