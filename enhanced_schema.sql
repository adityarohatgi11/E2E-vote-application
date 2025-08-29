-- Enhanced Database Schema for Venue Discovery & Rating Platform
-- This replaces the simple voting system with a comprehensive venue rating platform

-- ===============================
-- CORE VENUE & LOCATION TABLES
-- ===============================

-- Venue Categories (Restaurants, Bars, Cafes, etc.)
CREATE TABLE venue_categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    icon VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Venue Subcategories (Italian Restaurant, Sports Bar, Coffee Shop, etc.)
CREATE TABLE venue_subcategories (
    id BIGSERIAL PRIMARY KEY,
    category_id BIGINT REFERENCES venue_categories(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Cities/Locations
CREATE TABLE cities (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    state VARCHAR(100),
    country VARCHAR(100) NOT NULL,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    timezone VARCHAR(50),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Enhanced Venues Table (replacing participants)
CREATE TABLE venues (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL, -- SEO friendly URLs
    description TEXT,
    short_description VARCHAR(500),
    
    -- Location Information
    address TEXT NOT NULL,
    city_id BIGINT REFERENCES cities(id),
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    postal_code VARCHAR(20),
    
    -- Categorization
    category_id BIGINT REFERENCES venue_categories(id),
    subcategory_id BIGINT REFERENCES venue_subcategories(id),
    
    -- Contact & Business Info
    phone VARCHAR(20),
    email VARCHAR(255),
    website VARCHAR(255),
    
    -- Business Hours (JSON format)
    opening_hours JSONB, -- {"monday": {"open": "09:00", "close": "22:00"}, ...}
    
    -- Pricing Information
    price_range VARCHAR(10), -- $, $$, $$$, $$$$
    average_cost_per_person DECIMAL(8,2),
    
    -- Media
    cover_image VARCHAR(255),
    logo VARCHAR(255),
    
    -- Ratings Cache (for performance)
    average_rating DECIMAL(3,2) DEFAULT 0.00,
    total_ratings INTEGER DEFAULT 0,
    total_reviews INTEGER DEFAULT 0,
    
    -- Features & Amenities (JSON)
    amenities JSONB, -- ["wifi", "parking", "outdoor_seating", "live_music"]
    
    -- Status & Verification
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,
    is_featured BOOLEAN DEFAULT false,
    
    -- Owner Information
    owner_id BIGINT REFERENCES users(id),
    claimed_at TIMESTAMP,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ===============================
-- ENHANCED RATING & REVIEW SYSTEM
-- ===============================

-- Rating Criteria (Food Quality, Service, Ambiance, etc.)
CREATE TABLE rating_criteria (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category_id BIGINT REFERENCES venue_categories(id), -- Different criteria for different venue types
    weight DECIMAL(3,2) DEFAULT 1.00, -- For weighted average calculations
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Enhanced User Reviews (replacing simple votes)
CREATE TABLE venue_reviews (
    id BIGSERIAL PRIMARY KEY,
    venue_id BIGINT REFERENCES venues(id),
    user_id BIGINT REFERENCES snapp_users(id),
    
    -- Overall Rating
    overall_rating DECIMAL(3,2) NOT NULL CHECK (overall_rating >= 1.0 AND overall_rating <= 5.0),
    
    -- Detailed Ratings (JSON for flexibility)
    detailed_ratings JSONB, -- {"food_quality": 4.5, "service": 4.0, "ambiance": 5.0, "value": 3.5}
    
    -- Review Content
    title VARCHAR(255),
    review_text TEXT,
    
    -- Visit Information
    visit_date DATE,
    visit_type VARCHAR(50), -- "dinner", "lunch", "drinks", "event"
    party_size INTEGER,
    
    -- Media Attachments
    photos JSONB, -- Array of photo URLs
    
    -- Moderation
    is_verified BOOLEAN DEFAULT false,
    is_featured BOOLEAN DEFAULT false,
    is_flagged BOOLEAN DEFAULT false,
    moderation_status VARCHAR(20) DEFAULT 'pending', -- pending, approved, rejected
    
    -- Engagement
    helpful_votes INTEGER DEFAULT 0,
    unhelpful_votes INTEGER DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    UNIQUE(venue_id, user_id) -- One review per user per venue
);

-- Review Helpfulness Voting
CREATE TABLE review_votes (
    id BIGSERIAL PRIMARY KEY,
    review_id BIGINT REFERENCES venue_reviews(id),
    user_id BIGINT REFERENCES snapp_users(id),
    is_helpful BOOLEAN NOT NULL, -- true for helpful, false for unhelpful
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(review_id, user_id)
);

-- ===============================
-- COLLECTIONS & LISTS
-- ===============================

-- User-created venue lists
CREATE TABLE venue_collections (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES snapp_users(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_public BOOLEAN DEFAULT true,
    cover_image VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Venues in collections
CREATE TABLE venue_collection_items (
    id BIGSERIAL PRIMARY KEY,
    collection_id BIGINT REFERENCES venue_collections(id),
    venue_id BIGINT REFERENCES venues(id),
    note TEXT,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(collection_id, venue_id)
);

-- ===============================
-- EVENTS & PROMOTIONS
-- ===============================

-- Venue Events (Happy Hours, Live Music, Special Menus)
CREATE TABLE venue_events (
    id BIGSERIAL PRIMARY KEY,
    venue_id BIGINT REFERENCES venues(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Event Details
    event_type VARCHAR(50), -- "happy_hour", "live_music", "special_menu", "party"
    start_datetime TIMESTAMP NOT NULL,
    end_datetime TIMESTAMP NOT NULL,
    
    -- Recurrence (for recurring events)
    is_recurring BOOLEAN DEFAULT false,
    recurrence_pattern VARCHAR(50), -- "daily", "weekly", "monthly"
    recurrence_days JSONB, -- ["monday", "wednesday", "friday"]
    
    -- Pricing & Booking
    ticket_price DECIMAL(8,2),
    booking_required BOOLEAN DEFAULT false,
    booking_url VARCHAR(255),
    max_capacity INTEGER,
    
    -- Media
    event_image VARCHAR(255),
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    is_featured BOOLEAN DEFAULT false,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ===============================
-- SOCIAL FEATURES
-- ===============================

-- User Check-ins
CREATE TABLE venue_checkins (
    id BIGSERIAL PRIMARY KEY,
    venue_id BIGINT REFERENCES venues(id),
    user_id BIGINT REFERENCES snapp_users(id),
    
    -- Check-in Details
    message TEXT,
    photos JSONB,
    rating DECIMAL(3,2) CHECK (rating >= 1.0 AND rating <= 5.0),
    
    -- Social Features
    is_public BOOLEAN DEFAULT true,
    tagged_users JSONB, -- Array of user IDs
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User Followers/Following for social features
CREATE TABLE user_follows (
    id BIGSERIAL PRIMARY KEY,
    follower_id BIGINT REFERENCES snapp_users(id),
    following_id BIGINT REFERENCES snapp_users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(follower_id, following_id)
);

-- ===============================
-- ENHANCED VOTING SYSTEM
-- ===============================

-- Voting Campaigns (Best Restaurant 2024, Top Bars in City, etc.)
CREATE TABLE voting_campaigns (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    campaign_type VARCHAR(50), -- "best_restaurant", "top_bars", "hidden_gems"
    
    -- Geographic Scope
    city_id BIGINT REFERENCES cities(id),
    category_id BIGINT REFERENCES venue_categories(id),
    
    -- Campaign Duration
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    
    -- Voting Rules
    max_votes_per_user INTEGER DEFAULT 1,
    allow_multiple_categories BOOLEAN DEFAULT false,
    require_review BOOLEAN DEFAULT false, -- Must write review to vote
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    is_featured BOOLEAN DEFAULT false,
    
    -- Results
    winner_venue_id BIGINT REFERENCES venues(id),
    total_votes INTEGER DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Campaign Votes (enhanced from user_voting)
CREATE TABLE campaign_votes (
    id BIGSERIAL PRIMARY KEY,
    campaign_id BIGINT REFERENCES voting_campaigns(id),
    venue_id BIGINT REFERENCES venues(id),
    user_id BIGINT REFERENCES snapp_users(id),
    
    -- Vote Details
    reason TEXT, -- Why they voted for this venue
    confidence_score DECIMAL(3,2), -- How confident they are (1-5)
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(campaign_id, user_id, venue_id) -- Prevent duplicate votes
);

-- ===============================
-- ANALYTICS & INSIGHTS
-- ===============================

-- Venue Performance Metrics
CREATE TABLE venue_analytics (
    id BIGSERIAL PRIMARY KEY,
    venue_id BIGINT REFERENCES venues(id),
    date DATE NOT NULL,
    
    -- Engagement Metrics
    profile_views INTEGER DEFAULT 0,
    photo_views INTEGER DEFAULT 0,
    phone_clicks INTEGER DEFAULT 0,
    website_clicks INTEGER DEFAULT 0,
    direction_requests INTEGER DEFAULT 0,
    
    -- Social Metrics
    checkins INTEGER DEFAULT 0,
    reviews_count INTEGER DEFAULT 0,
    shares INTEGER DEFAULT 0,
    
    -- Ratings
    average_daily_rating DECIMAL(3,2),
    
    UNIQUE(venue_id, date)
);

-- Search & Discovery Tracking
CREATE TABLE search_analytics (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES snapp_users(id),
    
    -- Search Details
    search_query VARCHAR(255),
    search_type VARCHAR(50), -- "text", "category", "location", "filter"
    filters_used JSONB,
    
    -- Location Context
    user_latitude DECIMAL(10, 8),
    user_longitude DECIMAL(11, 8),
    search_radius INTEGER, -- in kilometers
    
    -- Results & Interactions
    results_count INTEGER,
    clicked_venue_id BIGINT REFERENCES venues(id),
    click_position INTEGER, -- Position in search results
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ===============================
-- INDEXES FOR PERFORMANCE
-- ===============================

-- Geospatial indexes for location-based queries
CREATE INDEX idx_venues_location ON venues USING GIST (ST_Point(longitude, latitude));
CREATE INDEX idx_venues_category ON venues(category_id);
CREATE INDEX idx_venues_rating ON venues(average_rating DESC);
CREATE INDEX idx_venues_active ON venues(is_active) WHERE is_active = true;

-- Review indexes
CREATE INDEX idx_reviews_venue ON venue_reviews(venue_id);
CREATE INDEX idx_reviews_user ON venue_reviews(user_id);
CREATE INDEX idx_reviews_rating ON venue_reviews(overall_rating DESC);
CREATE INDEX idx_reviews_date ON venue_reviews(created_at DESC);

-- Search indexes
CREATE INDEX idx_venues_text_search ON venues USING GIN(to_tsvector('english', name || ' ' || coalesce(description, '')));
CREATE INDEX idx_venues_city ON venues(city_id);

-- Analytics indexes
CREATE INDEX idx_venue_analytics_date ON venue_analytics(venue_id, date DESC);
CREATE INDEX idx_search_analytics_user ON search_analytics(user_id, created_at DESC);
