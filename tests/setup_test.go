package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
	"voting-app/app/controllers"
	"voting-app/app/middlewares"
	"voting-app/app/models"
	databases "voting-app/app"
	
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// TestSuite provides the base test suite for all end-to-end tests
type TestSuite struct {
	suite.Suite
	router   *gin.Engine
	db       *sql.DB
	testData *TestData
}

// TestData holds test fixtures
type TestData struct {
	TestUser1    models.SnapUser
	TestUser2    models.SnapUser
	TestCity     models.City
	TestCategory models.VenueCategory
	TestVenue1   models.Venue
	TestVenue2   models.Venue
	TestReview1  models.VenueReview
	TestReview2  models.VenueReview
}

// SetupSuite runs before all tests in the suite
func (suite *TestSuite) SetupSuite() {
	// Load test environment
	err := godotenv.Load("../test.env")
	if err != nil {
		// Fallback to local.env for testing
		godotenv.Load("../local.env")
	}
	
	// Set test environment
	os.Setenv("GIN_MODE", "test")
	gin.SetMode(gin.TestMode)
	
	// Setup test database connection
	suite.setupTestDatabase()
	
	// Setup router with all routes
	suite.setupRouter()
	
	// Create test data
	suite.createTestData()
}

// TearDownSuite runs after all tests in the suite
func (suite *TestSuite) TearDownSuite() {
	suite.cleanupTestData()
	if suite.db != nil {
		suite.db.Close()
	}
}

// SetupTest runs before each test
func (suite *TestSuite) SetupTest() {
	// Clean and recreate test data for each test
	suite.cleanupTestData()
	suite.createTestData()
}

// setupTestDatabase initializes the test database
func (suite *TestSuite) setupTestDatabase() {
	// Use a separate test database
	testDBName := "voting_app_test"
	
	// Connect to default database to create test database
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"))
	
	db, err := sql.Open("postgres", connStr)
	suite.Require().NoError(err)
	
	// Create test database if it doesn't exist
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", testDBName))
	if err != nil {
		// Database might already exist, which is fine
	}
	db.Close()
	
	// Connect to test database
	testConnStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		testDBName)
	
	suite.db, err = sql.Open("postgres", connStr)
	suite.Require().NoError(err)
	
	// Set the global database connection for the app
	databases.PostgresDB = suite.db
	
	// Run enhanced migrations
	suite.runEnhancedMigrations()
}

// runEnhancedMigrations creates all the enhanced tables
func (suite *TestSuite) runEnhancedMigrations() {
	// Read and execute the enhanced schema
	migrations := []string{
		// Cities table
		`CREATE TABLE IF NOT EXISTS cities (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			state VARCHAR(100),
			country VARCHAR(100) NOT NULL,
			latitude DECIMAL(10, 8),
			longitude DECIMAL(11, 8),
			timezone VARCHAR(50),
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// Venue categories
		`CREATE TABLE IF NOT EXISTS venue_categories (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL UNIQUE,
			description TEXT,
			icon VARCHAR(255),
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// Venue subcategories
		`CREATE TABLE IF NOT EXISTS venue_subcategories (
			id BIGSERIAL PRIMARY KEY,
			category_id BIGINT REFERENCES venue_categories(id),
			name VARCHAR(100) NOT NULL,
			description TEXT,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// Snapp users (from original schema)
		`CREATE TABLE IF NOT EXISTS snapp_users (
			id BIGSERIAL PRIMARY KEY,
			snapp_id VARCHAR NOT NULL UNIQUE
		)`,
		
		// Enhanced venues table
		`CREATE TABLE IF NOT EXISTS venues (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(255) UNIQUE NOT NULL,
			description TEXT,
			short_description VARCHAR(500),
			address TEXT NOT NULL,
			city_id BIGINT REFERENCES cities(id),
			latitude DECIMAL(10, 8) NOT NULL,
			longitude DECIMAL(11, 8) NOT NULL,
			postal_code VARCHAR(20),
			category_id BIGINT REFERENCES venue_categories(id),
			subcategory_id BIGINT REFERENCES venue_subcategories(id),
			phone VARCHAR(20),
			email VARCHAR(255),
			website VARCHAR(255),
			opening_hours JSONB,
			price_range VARCHAR(10),
			average_cost_per_person DECIMAL(8,2),
			cover_image VARCHAR(255),
			logo VARCHAR(255),
			average_rating DECIMAL(3,2) DEFAULT 0.00,
			total_ratings INTEGER DEFAULT 0,
			total_reviews INTEGER DEFAULT 0,
			amenities JSONB,
			is_active BOOLEAN DEFAULT true,
			is_verified BOOLEAN DEFAULT false,
			is_featured BOOLEAN DEFAULT false,
			owner_id BIGINT,
			claimed_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// Venue reviews
		`CREATE TABLE IF NOT EXISTS venue_reviews (
			id BIGSERIAL PRIMARY KEY,
			venue_id BIGINT REFERENCES venues(id),
			user_id BIGINT REFERENCES snapp_users(id),
			overall_rating DECIMAL(3,2) NOT NULL CHECK (overall_rating >= 1.0 AND overall_rating <= 5.0),
			detailed_ratings JSONB,
			title VARCHAR(255),
			review_text TEXT,
			visit_date DATE,
			visit_type VARCHAR(50),
			party_size INTEGER,
			photos JSONB,
			is_verified BOOLEAN DEFAULT false,
			is_featured BOOLEAN DEFAULT false,
			is_flagged BOOLEAN DEFAULT false,
			moderation_status VARCHAR(20) DEFAULT 'pending',
			helpful_votes INTEGER DEFAULT 0,
			unhelpful_votes INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(venue_id, user_id)
		)`,
		
		// Venue collections
		`CREATE TABLE IF NOT EXISTS venue_collections (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT REFERENCES snapp_users(id),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			is_public BOOLEAN DEFAULT true,
			cover_image VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// Venue collection items
		`CREATE TABLE IF NOT EXISTS venue_collection_items (
			id BIGSERIAL PRIMARY KEY,
			collection_id BIGINT REFERENCES venue_collections(id),
			venue_id BIGINT REFERENCES venues(id),
			note TEXT,
			added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(collection_id, venue_id)
		)`,
		
		// Venue checkins
		`CREATE TABLE IF NOT EXISTS venue_checkins (
			id BIGSERIAL PRIMARY KEY,
			venue_id BIGINT REFERENCES venues(id),
			user_id BIGINT REFERENCES snapp_users(id),
			message TEXT,
			photos JSONB,
			rating DECIMAL(3,2) CHECK (rating >= 1.0 AND rating <= 5.0),
			is_public BOOLEAN DEFAULT true,
			tagged_users JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// Voting campaigns
		`CREATE TABLE IF NOT EXISTS voting_campaigns (
			id BIGSERIAL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			campaign_type VARCHAR(50),
			city_id BIGINT REFERENCES cities(id),
			category_id BIGINT REFERENCES venue_categories(id),
			start_date TIMESTAMP NOT NULL,
			end_date TIMESTAMP NOT NULL,
			max_votes_per_user INTEGER DEFAULT 1,
			allow_multiple_categories BOOLEAN DEFAULT false,
			require_review BOOLEAN DEFAULT false,
			is_active BOOLEAN DEFAULT true,
			is_featured BOOLEAN DEFAULT false,
			winner_venue_id BIGINT REFERENCES venues(id),
			total_votes INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// Campaign votes
		`CREATE TABLE IF NOT EXISTS campaign_votes (
			id BIGSERIAL PRIMARY KEY,
			campaign_id BIGINT REFERENCES voting_campaigns(id),
			venue_id BIGINT REFERENCES venues(id),
			user_id BIGINT REFERENCES snapp_users(id),
			reason TEXT,
			confidence_score DECIMAL(3,2),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(campaign_id, user_id, venue_id)
		)`,
		
		// Venue analytics
		`CREATE TABLE IF NOT EXISTS venue_analytics (
			id BIGSERIAL PRIMARY KEY,
			venue_id BIGINT REFERENCES venues(id),
			date DATE NOT NULL,
			profile_views INTEGER DEFAULT 0,
			photo_views INTEGER DEFAULT 0,
			phone_clicks INTEGER DEFAULT 0,
			website_clicks INTEGER DEFAULT 0,
			direction_requests INTEGER DEFAULT 0,
			checkins INTEGER DEFAULT 0,
			reviews_count INTEGER DEFAULT 0,
			shares INTEGER DEFAULT 0,
			average_daily_rating DECIMAL(3,2),
			UNIQUE(venue_id, date)
		)`,
		
		// Search analytics
		`CREATE TABLE IF NOT EXISTS search_analytics (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT REFERENCES snapp_users(id),
			search_query VARCHAR(255),
			search_type VARCHAR(50),
			filters_used JSONB,
			user_latitude DECIMAL(10, 8),
			user_longitude DECIMAL(11, 8),
			search_radius INTEGER,
			results_count INTEGER,
			clicked_venue_id BIGINT REFERENCES venues(id),
			click_position INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	
	for _, migration := range migrations {
		_, err := suite.db.Exec(migration)
		suite.Require().NoError(err, "Failed to run migration: %s", migration)
	}
}

// setupRouter initializes the Gin router with all routes
func (suite *TestSuite) setupRouter() {
	suite.router = gin.New()
	suite.router.Use(gin.Recovery())
	
	// Add test middleware that bypasses authentication
	suite.router.Use(suite.testAuthMiddleware())
	
	v1 := suite.router.Group("/v1")
	
	// Venue routes
	venueRoutes := v1.Group("/venues")
	{
		venueController := new(controllers.VenueController)
		venueRoutes.GET("/search", venueController.Search)
		venueRoutes.GET("/nearby", venueController.GetNearby)
		venueRoutes.GET("/featured", venueController.GetFeatured)
		venueRoutes.GET("/categories", venueController.GetCategories)
		venueRoutes.GET("/:id", venueController.GetByID)
		venueRoutes.POST("/", venueController.CreateVenue)
	}
	
	// Review routes
	v1.GET("/venues/:venue_id/reviews", controllers.ReviewController{}.GetVenueReviews)
	v1.GET("/venues/:venue_id/reviews/summary", controllers.ReviewController{}.GetReviewSummary)
	
	userReviewRoutes := v1.Group("/reviews/:snapp_id")
	{
		reviewController := new(controllers.ReviewController)
		userReviewRoutes.POST("/", reviewController.CreateReview)
		userReviewRoutes.GET("/", reviewController.GetUserReviews)
		userReviewRoutes.POST("/:review_id/vote", reviewController.VoteReviewHelpful)
	}
	
	// Legacy vote routes for backwards compatibility
	voteRoutes := v1.Group("/vote/:snapp_id")
	{
		voteController := new(controllers.VoteController)
		voteRoutes.GET("/", voteController.Vote)
		voteRoutes.POST("/:voting_id/:vote_id", voteController.SubmitVote)
	}
}

// testAuthMiddleware provides a test authentication middleware
func (suite *TestSuite) testAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set test user ID for authenticated routes
		c.Set("snappUser_id", int64(1))
		c.Set("user_id", int64(1))
		c.Next()
	}
}

// createTestData creates test fixtures
func (suite *TestSuite) createTestData() {
	suite.testData = &TestData{}
	
	// Create test users
	_, err := suite.db.Exec("INSERT INTO snapp_users (id, snapp_id) VALUES (1, 'test_user_1') ON CONFLICT (id) DO NOTHING")
	suite.Require().NoError(err)
	_, err = suite.db.Exec("INSERT INTO snapp_users (id, snapp_id) VALUES (2, 'test_user_2') ON CONFLICT (id) DO NOTHING")
	suite.Require().NoError(err)
	
	suite.testData.TestUser1 = models.SnapUser{ID: 1, SnappId: "test_user_1"}
	suite.testData.TestUser2 = models.SnapUser{ID: 2, SnappId: "test_user_2"}
	
	// Create test city
	_, err = suite.db.Exec(`INSERT INTO cities (id, name, state, country, latitude, longitude) 
		VALUES (1, 'San Francisco', 'California', 'USA', 37.7749, -122.4194) ON CONFLICT (id) DO NOTHING`)
	suite.Require().NoError(err)
	
	suite.testData.TestCity = models.City{
		ID: 1, Name: "San Francisco", State: "California", Country: "USA",
		Latitude: 37.7749, Longitude: -122.4194,
	}
	
	// Create test category
	_, err = suite.db.Exec(`INSERT INTO venue_categories (id, name, description, icon) 
		VALUES (1, 'Restaurant', 'Restaurants and dining establishments', 'restaurant-icon') ON CONFLICT (id) DO NOTHING`)
	suite.Require().NoError(err)
	
	suite.testData.TestCategory = models.VenueCategory{
		ID: 1, Name: "Restaurant", Description: "Restaurants and dining establishments", Icon: "restaurant-icon",
	}
	
	// Create test venues
	_, err = suite.db.Exec(`INSERT INTO venues (id, name, slug, description, address, city_id, latitude, longitude, 
		category_id, price_range, average_rating, total_ratings, is_active) 
		VALUES (1, 'Test Restaurant 1', 'test-restaurant-1', 'A great test restaurant', 
		'123 Test St, San Francisco, CA', 1, 37.7849, -122.4094, 1, '$$', 4.5, 10, true) ON CONFLICT (id) DO NOTHING`)
	suite.Require().NoError(err)
	
	_, err = suite.db.Exec(`INSERT INTO venues (id, name, slug, description, address, city_id, latitude, longitude, 
		category_id, price_range, average_rating, total_ratings, is_active) 
		VALUES (2, 'Test Restaurant 2', 'test-restaurant-2', 'Another great test restaurant', 
		'456 Test Ave, San Francisco, CA', 1, 37.7749, -122.4194, 1, '$$$', 4.2, 8, true) ON CONFLICT (id) DO NOTHING`)
	suite.Require().NoError(err)
	
	suite.testData.TestVenue1 = models.Venue{
		ID: 1, Name: "Test Restaurant 1", Slug: "test-restaurant-1",
		Description: "A great test restaurant", Address: "123 Test St, San Francisco, CA",
		CityID: 1, Latitude: 37.7849, Longitude: -122.4094, CategoryID: 1,
		PriceRange: "$$", AverageRating: 4.5, TotalRatings: 10, IsActive: true,
	}
	
	suite.testData.TestVenue2 = models.Venue{
		ID: 2, Name: "Test Restaurant 2", Slug: "test-restaurant-2",
		Description: "Another great test restaurant", Address: "456 Test Ave, San Francisco, CA",
		CityID: 1, Latitude: 37.7749, Longitude: -122.4194, CategoryID: 1,
		PriceRange: "$$$", AverageRating: 4.2, TotalRatings: 8, IsActive: true,
	}
}

// cleanupTestData removes test data
func (suite *TestSuite) cleanupTestData() {
	tables := []string{
		"search_analytics", "venue_analytics", "campaign_votes", "voting_campaigns",
		"venue_checkins", "venue_collection_items", "venue_collections", "venue_reviews",
		"venues", "venue_subcategories", "venue_categories", "cities", "snapp_users",
	}
	
	for _, table := range tables {
		_, err := suite.db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			// Table might not exist yet, which is fine
		}
	}
}

// Helper methods for making HTTP requests

func (suite *TestSuite) makeGETRequest(url string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	return w
}

func (suite *TestSuite) makePOSTRequest(url string, payload interface{}) *httptest.ResponseRecorder {
	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	return w
}

func (suite *TestSuite) makePUTRequest(url string, payload interface{}) *httptest.ResponseRecorder {
	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	return w
}

func (suite *TestSuite) makeDELETERequest(url string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("DELETE", url, nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	return w
}

// Helper method to parse JSON response
func (suite *TestSuite) parseJSONResponse(w *httptest.ResponseRecorder, target interface{}) {
	err := json.Unmarshal(w.Body.Bytes(), target)
	suite.Require().NoError(err)
}

// Run the test suite
func TestVenueDiscoveryPlatform(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
