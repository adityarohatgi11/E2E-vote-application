import axios from 'axios';
import type {
  Venue,
  VenueSearchParams,
  VenueSearchResponse,
  VenueDetailResponse,
  VenueCategory,
  VenueReview,
  ReviewSummary,
  CreateReviewRequest,
  PaginationInfo,
} from '@/types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/v1';

// Create axios instance with default config
const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add auth token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Handle unauthorized - redirect to login
      localStorage.removeItem('auth_token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// Venue API services
export const venueService = {
  // Search venues with filters
  async search(params: Partial<VenueSearchParams>): Promise<VenueSearchResponse> {
    const response = await api.get('/venues/search', { params });
    return response.data;
  },

  // Get venue details by ID
  async getById(id: number, userLat?: number, userLng?: number): Promise<VenueDetailResponse> {
    const params = userLat && userLng ? { user_lat: userLat, user_lng: userLng } : {};
    const response = await api.get(`/venues/${id}`, { params });
    return response.data;
  },

  // Get nearby venues
  async getNearby(lat: number, lng: number, radius = 10, categoryId?: number): Promise<Venue[]> {
    const params = { lat, lng, radius, category: categoryId };
    const response = await api.get('/venues/nearby', { params });
    return response.data;
  },

  // Get featured venues
  async getFeatured(limit = 10): Promise<Venue[]> {
    const response = await api.get('/venues/featured', { params: { limit } });
    return response.data;
  },

  // Get venue categories
  async getCategories(): Promise<VenueCategory[]> {
    const response = await api.get('/venues/categories');
    return response.data;
  },

  // Create new venue (authenticated)
  async create(venueData: Partial<Venue>): Promise<Venue> {
    const response = await api.post('/venues', venueData);
    return response.data;
  },
};

// Review API services
export const reviewService = {
  // Get venue reviews
  async getVenueReviews(
    venueId: number,
    params: {
      minRating?: number;
      maxRating?: number;
      visitType?: string;
      hasPhotos?: boolean;
      sortBy?: string;
      page?: number;
      limit?: number;
    } = {}
  ): Promise<{ reviews: VenueReview[]; pagination: PaginationInfo }> {
    const response = await api.get(`/venues/${venueId}/reviews`, { params });
    return response.data;
  },

  // Get venue review summary
  async getVenueReviewSummary(venueId: number): Promise<ReviewSummary> {
    const response = await api.get(`/venues/${venueId}/reviews/summary`);
    return response.data;
  },

  // Create review (authenticated)
  async create(userId: string, reviewData: CreateReviewRequest): Promise<VenueReview> {
    const response = await api.post(`/reviews/${userId}`, reviewData);
    return response.data;
  },

  // Get user reviews (authenticated)
  async getUserReviews(
    userId: string,
    params: { sortBy?: string; page?: number; limit?: number } = {}
  ): Promise<{ reviews: VenueReview[]; pagination: PaginationInfo }> {
    const response = await api.get(`/users/${userId}/reviews`, { params });
    return response.data;
  },

  // Vote on review helpfulness (authenticated)
  async voteHelpful(userId: string, reviewId: number, isHelpful: boolean): Promise<void> {
    await api.post(`/reviews/${userId}/${reviewId}/vote`, { isHelpful });
  },
};

// Authentication services
export const authService = {
  async login(email: string, password: string) {
    const response = await api.post('/auth/login', { email, password });
    const { access: token } = response.data;
    localStorage.setItem('auth_token', token);
    return response.data;
  },

  async register(email: string, password: string) {
    const response = await api.post('/auth/register', { email, password });
    return response.data;
  },

  async logout() {
    localStorage.removeItem('auth_token');
  },

  async resetPassword(password: string) {
    const response = await api.post('/auth/reset-pass', { password });
    return response.data;
  },
};

// Legacy voting services (for backwards compatibility)
export const votingService = {
  // Get voting data
  async getVotingData(userId: string) {
    const response = await api.get(`/vote/${userId}`);
    return response.data;
  },

  // Submit vote
  async submitVote(userId: string, votingId: number, voteId: number) {
    const response = await api.post(`/vote/${userId}/${votingId}/${voteId}`);
    return response.data;
  },
};

// Utility functions
export const geoService = {
  // Get user's current location
  async getCurrentLocation(): Promise<{ lat: number; lng: number }> {
    return new Promise((resolve, reject) => {
      if (!navigator.geolocation) {
        reject(new Error('Geolocation is not supported'));
        return;
      }

      navigator.geolocation.getCurrentPosition(
        (position) => {
          resolve({
            lat: position.coords.latitude,
            lng: position.coords.longitude,
          });
        },
        (error: GeolocationPositionError) => {
          reject(error);
        },
        {
          enableHighAccuracy: true,
          timeout: 10000,
          maximumAge: 300000, // 5 minutes
        }
      );
    });
  },

  // Calculate distance between two points
  calculateDistance(lat1: number, lng1: number, lat2: number, lng2: number): number {
    const R = 6371; // Earth's radius in kilometers
    const dLat = (lat2 - lat1) * (Math.PI / 180);
    const dLng = (lng2 - lng1) * (Math.PI / 180);
    const a =
      Math.sin(dLat / 2) * Math.sin(dLat / 2) +
      Math.cos(lat1 * (Math.PI / 180)) *
        Math.cos(lat2 * (Math.PI / 180)) *
        Math.sin(dLng / 2) *
        Math.sin(dLng / 2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
    return R * c;
  },
};

export default api;
