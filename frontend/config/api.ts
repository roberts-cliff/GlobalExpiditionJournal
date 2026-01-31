// API configuration
// In development (Expo web), the frontend runs on a different port than the backend
// In production, the frontend is served by the backend so relative URLs work

import { Platform } from 'react-native';

// Detect if running in web development mode
const isWebDev = Platform.OS === 'web' && __DEV__;

// Base URL for API calls
// - Web dev: absolute URL to backend (localhost:8080)
// - Production/native: relative URLs (backend serves frontend)
export const API_BASE_URL = isWebDev ? 'http://localhost:8080' : '';

// Helper to construct API URLs
export function apiUrl(path: string): string {
  return `${API_BASE_URL}${path}`;
}
