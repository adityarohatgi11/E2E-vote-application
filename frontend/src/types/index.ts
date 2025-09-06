// Core API types matching the Go backend

export interface Venue {
  id: number;
  name: string;
  slug: string;
  description?: string;
  shortDescription?: string;
  address: string;
  cityId: number;
  city?: City;
  latitude: number;
  longitude: number;
  postalCode?: string;
  categoryId: number;
  category?: VenueCategory;
  subcategoryId?: number;
  subcategory?: VenueSubcategory;
  phone?: string;
  email?: string;
  website?: string;
  openingHours?: Record<string, unknown>;
  priceRange?: string;
  averageCostPerPerson?: number;
  coverImage?: string;
  logo?: string;
  averageRating: number;
  totalRatings: number;
  totalReviews: number;
  amenities?: string[];
  isActive: boolean;
  isVerified: boolean;
  isFeatured: boolean;
  distance?: number;
  isOpen?: boolean;
  nextOpenTime?: string;
  createdAt: string;
  updatedAt: string;
}

export interface VenueCategory {
  id: number;
  name: string;
  description?: string;
  icon?: string;
  isActive: boolean;
}

export interface VenueSubcategory {
  id: number;
  categoryId: number;
  name: string;
  description?: string;
  isActive: boolean;
}

export interface City {
  id: number;
  name: string;
  state?: string;
  country: string;
  latitude?: number;
  longitude?: number;
  timezone?: string;
}

export interface VenueReview {
  id: number;
  venueId: number;
  userId: number;
  overallRating: number;
  detailedRatings?: Record<string, number>;
  title?: string;
  reviewText?: string;
  visitDate?: string;
  visitType?: string;
  partySize?: number;
  photos?: string[];
  isVerified: boolean;
  isFeatured: boolean;
  isFlagged: boolean;
  moderationStatus: string;
  helpfulVotes: number;
  unhelpfulVotes: number;
  userName?: string;
  venueName?: string;
  createdAt: string;
  updatedAt: string;
}

export interface ReviewSummary {
  venueId: number;
  averageRating: number;
  totalReviews: number;
  ratingBreakdown: Record<string, number>;
  detailedAverage?: Record<string, number>;
  recentReviews: VenueReview[];
  topReviews: VenueReview[];
}

export interface VenueSearchParams {
  query?: string;
  categoryId?: number;
  subcategoryId?: number;
  cityId?: number;
  latitude?: number;
  longitude?: number;
  radius?: number;
  priceRange?: string[];
  minRating?: number;
  amenities?: string[];
  isOpen?: boolean;
  isFeatured?: boolean;
  sortBy?: string;
  page: number;
  limit: number;
}

export interface VenueSearchResponse {
  venues: Venue[];
  pagination: PaginationInfo;
  searchParams: VenueSearchParams;
  suggestions?: string[];
  filters: VenueFilterOptions;
}

export interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
  hasNext: boolean;
  hasPrev: boolean;
}

export interface VenueFilterOptions {
  categories: VenueCategory[];
  subcategories: VenueSubcategory[];
  priceRanges: string[];
  amenities: string[];
  cities: City[];
}

export interface VenueDetailResponse {
  venue: Venue;
  reviewSummary: ReviewSummary;
  events: VenueEvent[];
  similarVenues?: Venue[];
  checkinCount?: number;
}

export interface VenueEvent {
  id: number;
  venueId: number;
  title: string;
  description?: string;
  eventType: string;
  startDatetime: string;
  endDatetime: string;
  isRecurring: boolean;
  ticketPrice?: number;
  bookingRequired: boolean;
  eventImage?: string;
  isActive: boolean;
  isFeatured: boolean;
}

export interface CreateReviewRequest {
  venueId: number;
  overallRating: number;
  detailedRatings?: Record<string, number>;
  title?: string;
  reviewText?: string;
  visitDate?: string;
  visitType?: string;
  partySize?: number;
  photos?: string[];
}

export interface User {
  id: number;
  email: string;
  name?: string;
  avatar?: string;
  isVerified: boolean;
  createdAt: string;
}

export interface AuthResponse {
  user: User;
  token: string;
  refreshToken: string;
}

// UI-specific types
export interface MapBounds {
  northEast: { lat: number; lng: number };
  southWest: { lat: number; lng: number };
}

export interface FilterState {
  categories: number[];
  priceRanges: string[];
  rating: number;
  distance: number;
  amenities: string[];
  isOpen: boolean;
}

export interface SearchState {
  query: string;
  location: { lat: number; lng: number } | null;
  filters: FilterState;
  sortBy: string;
  viewMode: 'grid' | 'list' | 'map';
}
