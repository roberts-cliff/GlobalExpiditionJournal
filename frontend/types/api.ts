// API response types

export interface User {
  id: number;
  canvasUserId: string;
  canvasInstanceUrl: string;
  displayName: string;
  email: string;
}

export interface Country {
  id: number;
  name: string;
  isoCode: string;
  region: string;
}

export interface Visit {
  id: number;
  userId: number;
  countryId: number;
  visitedAt: string;
  notes: string;
  country?: Country;
}

export interface ApiError {
  error: string;
  message?: string;
}

export interface HealthResponse {
  status: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
}

// Scrapbook types
export interface ScrapbookEntry {
  id: number;
  countryId: number;
  title: string;
  notes?: string;
  mediaUrl?: string;
  mediaType?: string;
  tags?: string;
  visitedAt?: string;
  createdAt: string;
  updatedAt: string;
  country?: Country;
}

export interface ScrapbookStats {
  totalEntries: number;
  countriesDocumented: number;
  photosUploaded: number;
}

// List response types
export interface CountryListResponse {
  countries: Country[];
  total: number;
}

export interface VisitListResponse {
  visits: Visit[];
  total: number;
}

export interface ScrapbookEntryListResponse {
  entries: ScrapbookEntry[];
  total: number;
}

// Request types
export interface CreateVisitRequest {
  countryId: number;
  visitedAt?: string;
  notes?: string;
}

export interface CreateScrapbookEntryRequest {
  countryId: number;
  title: string;
  notes?: string;
  mediaUrl?: string;
  mediaType?: string;
  tags?: string;
  visitedAt?: string;
}

export interface UpdateScrapbookEntryRequest {
  title?: string;
  notes?: string;
  mediaUrl?: string;
  mediaType?: string;
  tags?: string;
  visitedAt?: string;
}

// Upload types
export interface UploadResponse {
  url: string;
  filename: string;
}
