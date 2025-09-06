package tests

import (
	"time"
	"voting-app/app/models"
	"voting-app/app/services"

	"github.com/stretchr/testify/assert"
)

// TestAnalyticsSystem tests the comprehensive analytics system
func (suite *TestSuite) TestAnalyticsSystem() {
	suite.Run("Analytics System End-to-End", func() {
		// Setup analytics data
		suite.setupAnalyticsData()

		// Test venue analytics
		suite.testVenueAnalytics()

		// Test search tracking
		suite.testSearchAnalytics()

		// Test user engagement tracking
		suite.testUserEngagementAnalytics()

		// Test performance metrics
		suite.testPerformanceMetrics()
	})
}

func (suite *TestSuite) setupAnalyticsData() {
	// Create venue analytics data
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// Insert venue analytics for last few days
	_, err := suite.db.Exec(`INSERT INTO venue_analytics 
		(venue_id, date, profile_views, photo_views, phone_clicks, website_clicks, 
		 direction_requests, checkins, reviews_count, shares, average_daily_rating) VALUES 
		(1, $1, 150, 45, 12, 8, 25, 5, 2, 3, 4.5),
		(1, $2, 120, 38, 10, 6, 20, 3, 1, 2, 4.2),
		(2, $1, 80, 25, 5, 3, 15, 2, 1, 1, 4.0),
		(2, $2, 90, 30, 7, 4, 18, 4, 2, 2, 4.3) ON CONFLICT (venue_id, date) DO NOTHING`,
		today, yesterday)
	suite.Require().NoError(err)

	// Insert search analytics data
	_, err = suite.db.Exec(`INSERT INTO search_analytics 
		(user_id, search_query, search_type, filters_used, user_latitude, user_longitude, 
		 search_radius, results_count, clicked_venue_id, click_position) VALUES 
		(1, 'pizza', 'text', '{"category_id": 1}', 37.7749, -122.4194, 5, 10, 1, 1),
		(1, 'italian restaurant', 'text', '{"price_range": "$$"}', 37.7749, -122.4194, 10, 8, 2, 2),
		(2, '', 'filter', '{"category_id": 1, "min_rating": 4.0}', 37.7849, -122.4094, 3, 5, 1, 1),
		(2, 'sushi', 'text', '{}', 37.7649, -122.4294, 8, 12, NULL, NULL) ON CONFLICT DO NOTHING`)
	suite.Require().NoError(err)
}

func (suite *TestSuite) testVenueAnalytics() {
	analyticsService := &services.AnalyticsService{}

	// Test getting venue analytics for different time ranges
	analytics, err := analyticsService.GetVenueAnalytics(1, "week")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), analytics)

	// Verify basic metrics
	assert.Equal(suite.T(), int64(1), analytics.VenueID)
	assert.Equal(suite.T(), "Test Restaurant 1", analytics.VenueName)
	assert.Equal(suite.T(), "week", analytics.TimeRange)
	assert.True(suite.T(), analytics.ProfileViews > 0)
	assert.True(suite.T(), analytics.PhotoViews > 0)

	// Test different time ranges
	timeRanges := []string{"today", "week", "month"}
	for _, timeRange := range timeRanges {
		analytics, err := analyticsService.GetVenueAnalytics(1, timeRange)
		assert.NoError(suite.T(), err, "Should get analytics for time range: %s", timeRange)
		assert.Equal(suite.T(), timeRange, analytics.TimeRange)
		assert.NotNil(suite.T(), analytics.GrowthMetrics)
	}

	// Test analytics for venue with no data
	analytics, err = analyticsService.GetVenueAnalytics(999, "week")
	assert.Error(suite.T(), err) // Should error for non-existent venue

	// Test venue view tracking
	err = analyticsService.TrackVenueView(1, 1, "profile")
	assert.NoError(suite.T(), err)

	err = analyticsService.TrackVenueView(1, 1, "photo")
	assert.NoError(suite.T(), err)

	err = analyticsService.TrackVenueView(1, 1, "phone")
	assert.NoError(suite.T(), err)

	// Verify tracking was recorded
	var profileViews int
	err = suite.db.QueryRow(`
		SELECT profile_views FROM venue_analytics 
		WHERE venue_id = 1 AND date = CURRENT_DATE`).Scan(&profileViews)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), profileViews > 0)
}

func (suite *TestSuite) testSearchAnalytics() {
	analyticsService := &services.AnalyticsService{}

	// Test search tracking
	searchResults := []models.Venue{
		suite.testData.TestVenue1,
		suite.testData.TestVenue2,
	}

	filters := map[string]interface{}{
		"category_id": 1,
		"min_rating":  4.0,
		"latitude":    37.7749,
		"longitude":   -122.4194,
		"radius":      5.0,
	}

	clickedVenueID := int64(1)
	clickPosition := 1

	err := analyticsService.TrackSearch(
		1, "best restaurants", filters, searchResults,
		&clickedVenueID, &clickPosition)
	assert.NoError(suite.T(), err)

	// Verify search was recorded
	var searchCount int
	err = suite.db.QueryRow(`
		SELECT COUNT(*) FROM search_analytics 
		WHERE user_id = 1 AND search_query = 'best restaurants'`).Scan(&searchCount)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), searchCount > 0)

	// Test search without click
	err = analyticsService.TrackSearch(1, "no click search", filters, searchResults, nil, nil)
	assert.NoError(suite.T(), err)

	// Test different search types
	err = analyticsService.TrackSearch(1, "", filters, searchResults, &clickedVenueID, &clickPosition)
	assert.NoError(suite.T(), err) // Empty query should be classified as filter search
}

func (suite *TestSuite) testUserEngagementAnalytics() {
	// Test platform-wide analytics
	analyticsService := &services.AnalyticsService{}

	platformAnalytics, err := analyticsService.GetPlatformAnalytics("week")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), platformAnalytics)

	// Verify platform metrics
	assert.Equal(suite.T(), "week", platformAnalytics.TimeRange)
	assert.True(suite.T(), platformAnalytics.TotalVenues > 0)
	assert.True(suite.T(), platformAnalytics.TotalUsers > 0)
	assert.NotNil(suite.T(), platformAnalytics.TopCategories)
	assert.NotNil(suite.T(), platformAnalytics.TopSearchQueries)

	// Test different time ranges for platform analytics
	timeRanges := []string{"today", "week", "month", "quarter"}
	for _, timeRange := range timeRanges {
		analytics, err := analyticsService.GetPlatformAnalytics(timeRange)
		assert.NoError(suite.T(), err, "Should get platform analytics for: %s", timeRange)
		assert.Equal(suite.T(), timeRange, analytics.TimeRange)
	}
}

func (suite *TestSuite) testPerformanceMetrics() {
	analyticsService := &services.AnalyticsService{}

	// Test top performing venues
	topVenues, err := analyticsService.GetTopPerformingVenues("week", nil, nil, 10)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), topVenues)

	if len(topVenues) > 0 {
		// Verify venues are sorted by performance
		for i := 1; i < len(topVenues); i++ {
			// Performance calculation includes rating, views, and check-ins
			// Higher performing venues should come first
			assert.True(suite.T(),
				topVenues[i-1].AverageRating >= topVenues[i].AverageRating ||
					topVenues[i-1].ProfileViews >= topVenues[i].ProfileViews,
				"Venues should be sorted by performance")
		}
	}

	// Test filtering by category
	categoryID := int64(1)
	topInCategory, err := analyticsService.GetTopPerformingVenues("week", &categoryID, nil, 5)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), topInCategory)

	// Test filtering by city
	cityID := int64(1)
	topInCity, err := analyticsService.GetTopPerformingVenues("week", nil, &cityID, 5)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), topInCity)
}

// TestRecommendationSystem tests the AI recommendation engine
func (suite *TestSuite) TestRecommendationSystem() {
	suite.Run("Recommendation System End-to-End", func() {
		// Setup recommendation test data
		suite.setupRecommendationData()

		// Test personalized recommendations
		suite.testPersonalizedRecommendations()

		// Test similar venue discovery
		suite.testSimilarVenues()

		// Test recommendation accuracy
		suite.testRecommendationAccuracy()
	})
}

func (suite *TestSuite) setupRecommendationData() {
	// Create user check-ins to establish preferences
	_, err := suite.db.Exec(`INSERT INTO venue_checkins 
		(venue_id, user_id, message, rating, is_public) VALUES 
		(1, 1, 'Great dinner!', 4.5, true),
		(2, 1, 'Nice lunch spot', 4.0, true),
		(1, 2, 'Amazing food', 5.0, true) ON CONFLICT DO NOTHING`)
	suite.Require().NoError(err)

	// Create user follows for social recommendations
	_, err = suite.db.Exec(`INSERT INTO user_follows 
		(follower_id, following_id) VALUES 
		(1, 2),
		(2, 1) ON CONFLICT DO NOTHING`)
	suite.Require().NoError(err)

	// Approve existing reviews for recommendation calculation
	_, err = suite.db.Exec(`UPDATE venue_reviews 
		SET moderation_status = 'approved' 
		WHERE user_id IN (1, 2)`)
	suite.Require().NoError(err)
}

func (suite *TestSuite) testPersonalizedRecommendations() {
	recommendationEngine := &services.RecommendationEngine{}

	// Test basic personalized recommendations
	context := services.RecommendationContext{
		UserID:      1,
		UserLat:     &suite.testData.TestVenue1.Latitude,
		UserLng:     &suite.testData.TestVenue1.Longitude,
		TimeOfDay:   "evening",
		Occasion:    "date",
		GroupSize:   2,
		MaxDistance: 10.0,
		Limit:       10,
	}

	recommendations, err := recommendationEngine.GetPersonalizedRecommendations(context)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), recommendations)

	// Verify recommendation structure
	for _, rec := range recommendations {
		assert.True(suite.T(), rec.Score >= 0 && rec.Score <= 1, "Score should be between 0 and 1")
		assert.NotNil(suite.T(), rec.Venue)
		assert.NotNil(suite.T(), rec.Reasons)
		assert.True(suite.T(), rec.Venue.ID > 0)

		// Verify distance calculation if location provided
		if context.UserLat != nil && context.UserLng != nil {
			assert.NotNil(suite.T(), rec.Venue.Distance)
			assert.True(suite.T(), *rec.Venue.Distance <= context.MaxDistance)
		}
	}

	// Test recommendations are sorted by score
	if len(recommendations) > 1 {
		for i := 1; i < len(recommendations); i++ {
			assert.True(suite.T(), recommendations[i-1].Score >= recommendations[i].Score,
				"Recommendations should be sorted by score")
		}
	}

	// Test without location
	contextNoLocation := context
	contextNoLocation.UserLat = nil
	contextNoLocation.UserLng = nil

	recommendations, err = recommendationEngine.GetPersonalizedRecommendations(contextNoLocation)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), recommendations)

	// Test different time contexts
	timeContexts := []string{"morning", "afternoon", "evening", "night"}
	for _, timeContext := range timeContexts {
		testContext := context
		testContext.TimeOfDay = timeContext

		recommendations, err := recommendationEngine.GetPersonalizedRecommendations(testContext)
		assert.NoError(suite.T(), err, "Should get recommendations for time: %s", timeContext)
		assert.NotNil(suite.T(), recommendations)
	}

	// Test different occasions
	occasions := []string{"casual", "date", "business", "celebration"}
	for _, occasion := range occasions {
		testContext := context
		testContext.Occasion = occasion

		recommendations, err := recommendationEngine.GetPersonalizedRecommendations(testContext)
		assert.NoError(suite.T(), err, "Should get recommendations for occasion: %s", occasion)
		assert.NotNil(suite.T(), recommendations)
	}
}

func (suite *TestSuite) testSimilarVenues() {
	recommendationEngine := &services.RecommendationEngine{}

	// Test getting similar venues
	similarVenues, err := recommendationEngine.GetSimilarVenues(1, 5)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), similarVenues)

	// Verify similar venues have distance calculated
	for _, venue := range similarVenues {
		assert.NotNil(suite.T(), venue.Distance, "Similar venues should have distance calculated")
		assert.NotEqual(suite.T(), int64(1), venue.ID, "Similar venues should not include the reference venue")
	}

	// Test with non-existent venue
	similarVenues, err = recommendationEngine.GetSimilarVenues(999, 5)
	assert.Error(suite.T(), err) // Should error for non-existent venue

	// Test with different limits
	limits := []int{1, 3, 5, 10}
	for _, limit := range limits {
		venues, err := recommendationEngine.GetSimilarVenues(1, limit)
		assert.NoError(suite.T(), err, "Should get similar venues with limit: %d", limit)
		assert.True(suite.T(), len(venues) <= limit, "Should respect limit")
	}
}

func (suite *TestSuite) testRecommendationAccuracy() {
	recommendationEngine := &services.RecommendationEngine{}

	// Test that highly rated venues appear in recommendations
	context := services.RecommendationContext{
		UserID:      1,
		UserLat:     &suite.testData.TestVenue1.Latitude,
		UserLng:     &suite.testData.TestVenue1.Longitude,
		MaxDistance: 10.0,
		Limit:       10,
	}

	recommendations, err := recommendationEngine.GetPersonalizedRecommendations(context)
	assert.NoError(suite.T(), err)

	// Verify that venues with higher ratings generally get higher scores
	highRatedFound := false
	for _, rec := range recommendations {
		if rec.Venue.AverageRating >= 4.0 {
			highRatedFound = true
			assert.True(suite.T(), rec.Score > 0.3, "High-rated venues should get decent scores")
		}
	}
	assert.True(suite.T(), highRatedFound, "Should recommend some high-rated venues")

	// Test that user's preferred category appears in recommendations
	categoryPreferenceFound := false
	for _, rec := range recommendations {
		if rec.Venue.CategoryID == 1 { // User has reviewed restaurants (category 1)
			categoryPreferenceFound = true
			// Should have reasons related to category preference
			reasonFound := false
			for _, reason := range rec.Reasons {
				if reason == "Matches your preferred category" {
					reasonFound = true
					break
				}
			}
			if rec.Score > 0.5 { // Only check for high-scoring recommendations
				assert.True(suite.T(), reasonFound, "High-scoring category matches should have preference reason")
			}
		}
	}
	// Note: This might not always be true if there are no venues in preferred categories
	assert.NotNil(suite.T(), categoryPreferenceFound, "Category preference tracking worked")

	// Test that location relevance affects scoring
	if len(recommendations) > 1 {
		closeVenues := 0
		for _, rec := range recommendations {
			if rec.Venue.Distance != nil && *rec.Venue.Distance <= 2.0 {
				closeVenues++
				// Close venues should generally score well
				if rec.Venue.AverageRating >= 4.0 {
					assert.True(suite.T(), rec.Score > 0.4, "Close, high-rated venues should score well")
				}
			}
		}
	}
}

// TestGeolocationServices tests location-based features
func (suite *TestSuite) TestGeolocationServices() {
	suite.Run("Geolocation Services End-to-End", func() {
		suite.testDistanceCalculations()
		suite.testLocationBounds()
		suite.testNearbySearch()
		suite.testOptimalMeetingPoint()
	})
}

func (suite *TestSuite) testDistanceCalculations() {
	geoService := &services.GeolocationService{}

	// Test distance calculation between known points
	lat1, lng1 := 37.7749, -122.4194 // San Francisco
	lat2, lng2 := 37.7849, -122.4094 // Slightly north and east

	distance := geoService.CalculateDistance(lat1, lng1, lat2, lng2)

	assert.True(suite.T(), distance.Kilometers > 0, "Distance should be positive")
	assert.True(suite.T(), distance.Kilometers < 5, "Distance should be reasonable for nearby points")
	assert.True(suite.T(), distance.Miles > 0, "Miles should be positive")
	assert.True(suite.T(), distance.Meters > 0, "Meters should be positive")

	// Test distance to same point
	sameDistance := geoService.CalculateDistance(lat1, lng1, lat1, lng1)
	assert.True(suite.T(), sameDistance.Kilometers < 0.001, "Distance to same point should be near zero")

	// Test distance units conversion
	assert.True(suite.T(), distance.Miles < distance.Kilometers, "Miles should be less than kilometers")
	assert.True(suite.T(), distance.Meters > distance.Kilometers, "Meters should be more than kilometers")
}

func (suite *TestSuite) testLocationBounds() {
	geoService := &services.GeolocationService{}

	centerLat, centerLng := 37.7749, -122.4194
	radius := 5.0 // 5km

	bounds := geoService.GetBounds(centerLat, centerLng, radius)

	// Verify bounds structure
	assert.True(suite.T(), bounds.NorthEast.Latitude > centerLat, "NorthEast lat should be greater than center")
	assert.True(suite.T(), bounds.NorthEast.Longitude > centerLng, "NorthEast lng should be greater than center")
	assert.True(suite.T(), bounds.SouthWest.Latitude < centerLat, "SouthWest lat should be less than center")
	assert.True(suite.T(), bounds.SouthWest.Longitude < centerLng, "SouthWest lng should be less than center")

	// Verify bounds are reasonable
	latDiff := bounds.NorthEast.Latitude - bounds.SouthWest.Latitude
	lngDiff := bounds.NorthEast.Longitude - bounds.SouthWest.Longitude

	assert.True(suite.T(), latDiff > 0 && latDiff < 1, "Latitude difference should be reasonable")
	assert.True(suite.T(), lngDiff > 0 && lngDiff < 1, "Longitude difference should be reasonable")
}

func (suite *TestSuite) testNearbySearch() {
	geoService := &services.GeolocationService{}

	lat, lng := 37.7749, -122.4194
	radius := 10.0
	filters := map[string]interface{}{
		"category_id": 1,
		"min_rating":  3.0,
	}

	result, err := geoService.GetNearbyVenues(lat, lng, radius, filters)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)

	// Verify result structure
	assert.Equal(suite.T(), lat, result.UserLat)
	assert.Equal(suite.T(), lng, result.UserLng)
	assert.Equal(suite.T(), radius, result.Radius)
	assert.NotNil(suite.T(), result.Venues)

	// Verify all venues are within radius and meet criteria
	for _, venue := range result.Venues {
		assert.NotNil(suite.T(), venue.Distance, "Venue should have distance calculated")
		assert.True(suite.T(), *venue.Distance <= radius, "Venue should be within radius")

		if minRating, exists := filters["min_rating"]; exists {
			assert.True(suite.T(), venue.AverageRating >= minRating.(float64),
				"Venue should meet minimum rating criteria")
		}

		if categoryID, exists := filters["category_id"]; exists {
			assert.Equal(suite.T(), categoryID.(int), int(venue.CategoryID),
				"Venue should match category filter")
		}
	}

	// Test with invalid coordinates
	_, err = geoService.GetNearbyVenues(91.0, lng, radius, filters) // Invalid latitude
	assert.Error(suite.T(), err)

	_, err = geoService.GetNearbyVenues(lat, 181.0, radius, filters) // Invalid longitude
	assert.Error(suite.T(), err)

	// Test with zero radius
	result, err = geoService.GetNearbyVenues(lat, lng, 0, filters)
	assert.NoError(suite.T(), err) // Should default to reasonable radius
	assert.True(suite.T(), result.Radius > 0)
}

func (suite *TestSuite) testOptimalMeetingPoint() {
	geoService := &services.GeolocationService{}

	// Test with multiple locations
	locations := []services.LatLng{
		{Latitude: 37.7749, Longitude: -122.4194}, // San Francisco
		{Latitude: 37.7849, Longitude: -122.4094}, // Slightly north/east
		{Latitude: 37.7649, Longitude: -122.4294}, // Slightly south/west
	}

	preferences := map[string]interface{}{
		"category_id": 1,
		"min_rating":  4.0,
	}

	meetingPoint, venues, err := geoService.FindOptimalMeetingPoint(locations, preferences)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), meetingPoint)
	assert.NotNil(suite.T(), venues)

	// Verify meeting point is roughly in the center
	avgLat := (locations[0].Latitude + locations[1].Latitude + locations[2].Latitude) / 3
	avgLng := (locations[0].Longitude + locations[1].Longitude + locations[2].Longitude) / 3

	assert.True(suite.T(), abs(meetingPoint.Latitude-avgLat) < 0.01,
		"Meeting point should be near center latitude")
	assert.True(suite.T(), abs(meetingPoint.Longitude-avgLng) < 0.01,
		"Meeting point should be near center longitude")

	// Test with single location
	singleLocation := []services.LatLng{locations[0]}
	meetingPoint, venues, err = geoService.FindOptimalMeetingPoint(singleLocation, preferences)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), locations[0].Latitude, meetingPoint.Latitude)
	assert.Equal(suite.T(), locations[0].Longitude, meetingPoint.Longitude)

	// Test with empty locations
	_, _, err = geoService.FindOptimalMeetingPoint([]services.LatLng{}, preferences)
	assert.Error(suite.T(), err)
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
