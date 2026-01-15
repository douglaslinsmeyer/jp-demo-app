export interface RouteRequest {
  origin: string;
  destination: string;
  departureTime?: Date;
}

export interface RouteResponse {
  routes: Route[];
  optimalDepartureTime: string;
  factors: FactorInfo;
}

export interface Route {
  summary: string;
  totalTime: number; // in seconds
  breakdown: TimeBreakdown;
  steps: RouteStep[];
  score: number;
}

export interface TimeBreakdown {
  transitTime: number;
  walkingTime: number;
  waitingTime: number;
  delayPenalty: number;
  weatherPenalty: number;
  crowdingPenalty: number;
}

export interface RouteStep {
  type: 'WALK' | 'TRANSIT' | 'WAIT';
  instructions: string;
  distance?: number;
  duration: number;
  line?: LineInfo;
  departAt?: string;
  arriveAt?: string;
}

export interface LineInfo {
  name: string;
  shortName: string;
  color?: string;
  vehicle: string;
}

export interface FactorInfo {
  delays: DelayInfo[];
  weather: WeatherInfo;
  crowdingEstimate: CrowdingInfo;
}

export interface DelayInfo {
  lineName: string;
  delayMinutes: number;
  status: 'NORMAL' | 'DELAYED' | 'SUSPENDED';
  message?: string;
}

export interface WeatherInfo {
  temperature: number;
  precipitation: number;
  condition: string;
  lastUpdated: string;
}

export interface CrowdingInfo {
  level: 'LOW' | 'MODERATE' | 'HIGH' | 'VERY_HIGH';
  description: string;
}

export interface ErrorResponse {
  error: string;
  message: string;
  code: number;
}
