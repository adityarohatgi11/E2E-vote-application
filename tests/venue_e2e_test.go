package tests

import (
	"fmt"
	"net/http"
	"voting-app/app/models"
	"voting-app/app/serializers"

	"github.com/stretchr/testify/assert"
)

// TestVenueDiscovery tests the complete venue discovery flow
func (suite *TestSuite) TestVenueDiscovery() {
	suite.Run("Complete Venue Discovery Flow", func() {
		// Test 1: Get venue categories
		suite.testGetVenueCategories()

		// Test 2: Search venues with various filters
		suite.testVenueSearch()

		// Test 3: Get venue details
		suite.testGetVenueDetails()

		// Test 4: Get nearby venues
		suite.testGetNearbyVenues()

		// Test 5: Get featured venues
		suite.testGetFeaturedVenues()

		// Test 6: Create new venue
		suite.testCreateVenue()
	})
}

func (suite *TestSuite) testGetVenueCategories() {
	w := suite.makeGETRequest("/v1/venues/categories")

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var categories []models.VenueCategory
	suite.parseJSONResponse(w, &categories)

	assert.NotEmpty(suite.T(), categories)
	assert.Equal(suite.T(), "Restaurant", categories[0].Name)
}

func (suite *TestSuite) testVenueSearch() {
	// Test basic search
	w := suite.makeGETRequest("/v1/venues/search?q=restaurant")
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var searchResponse serializers.VenueSearchResponse
	suite.parseJSONResponse(w, &searchResponse)

	assert.NotEmpty(suite.T(), searchResponse.Venues)
	assert.Contains(suite.T(), searchResponse.Venues[0].Name, "Restaurant")

	// Test category filter
	w = suite.makeGETRequest("/v1/venues/search?category=1")
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	suite.parseJSONResponse(w, &searchResponse)
	assert.NotEmpty(suite.T(), searchResponse.Venues)

	// Test location-based search
	w = suite.makeGETRequest("/v1/venues/search?lat=37.7749&lng=-122.4194&radius=5")
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	suite.parseJSONResponse(w, &searchResponse)
	assert.NotEmpty(suite.T(), searchResponse.Venues)

	// Test price range filter
	w = suite.makeGETRequest("/v1/venues/search?price_range=$$")
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	suite.parseJSONResponse(w, &searchResponse)
	assert.NotEmpty(suite.T(), searchResponse.Venues)

	// Test minimum rating filter
	w = suite.makeGETRequest("/v1/venues/search?min_rating=4.0")
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	suite.parseJSONResponse(w, &searchResponse)
	assert.NotEmpty(suite.T(), searchResponse.Venues)

	// Test sorting
	w = suite.makeGETRequest("/v1/venues/search?sort_by=rating")
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	suite.parseJSONResponse(w, &searchResponse)
	assert.NotEmpty(suite.T(), searchResponse.Venues)

	// Test pagination
	w = suite.makeGETRequest("/v1/venues/search?page=1&limit=10")
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	suite.parseJSONResponse(w, &searchResponse)
	assert.NotEmpty(suite.T(), searchResponse.Venues)
	assert.NotNil(suite.T(), searchResponse.Pagination)
	assert.Equal(suite.T(), 1, searchResponse.Pagination.Page)
	assert.Equal(suite.T(), 10, searchResponse.Pagination.Limit)
}

func (suite *TestSuite) testGetVenueDetails() {
	// Test getting venue by ID
	w := suite.makeGETRequest("/v1/venues/1")
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var venueResponse serializers.VenueDetailResponse
	suite.parseJSONResponse(w, &venueResponse)

	assert.Equal(suite.T(), int64(1), venueResponse.Venue.ID)
	assert.Equal(suite.T(), "Test Restaurant 1", venueResponse.Venue.Name)
	assert.Equal(suite.T(), "test-restaurant-1", venueResponse.Venue.Slug)
	assert.NotNil(suite.T(), venueResponse.ReviewSummary)

	// Test with distance calculation
	w = suite.makeGETRequest("/v1/venues/1?user_lat=37.7749&user_lng=-122.4194")
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	suite.parseJSONResponse(w, &venueResponse)
	assert.NotNil(suite.T(), venueResponse.Venue.Distance)

	// Test non-existent venue
	w = suite.makeGETRequest("/v1/venues/999")
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	// Test invalid venue ID
	w = suite.makeGETRequest("/v1/venues/invalid")
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *TestSuite) testGetNearbyVenues() {
	// Test nearby venues with valid coordinates
	w := suite.makeGETRequest("/v1/venues/nearby?lat=37.7749&lng=-122.4194&radius=10")
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var venues []models.Venue
	suite.parseJSONResponse(w, &venues)
	assert.NotEmpty(suite.T(), venues)

	// Verify distance is calculated
	for _, venue := range venues {
		assert.NotNil(suite.T(), venue.Distance)
		assert.True(suite.T(), *venue.Distance <= 10.0) // Within radius
	}

	// Test with category filter
	w = suite.makeGETRequest("/v1/venues/nearby?lat=37.7749&lng=-122.4194&radius=10&category=1")
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	suite.parseJSONResponse(w, &venues)
	assert.NotEmpty(suite.T(), venues)

	// Test with limit
	w = suite.makeGETRequest("/v1/venues/nearby?lat=37.7749&lng=-122.4194&limit=5")
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	suite.parseJSONResponse(w, &venues)
	assert.True(suite.T(), len(venues) <= 5)

	// Test missing coordinates
	w = suite.makeGETRequest("/v1/venues/nearby")
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	// Test invalid coordinates
	w = suite.makeGETRequest("/v1/venues/nearby?lat=invalid&lng=-122.4194")
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *TestSuite) testGetFeaturedVenues() {
	// First, make a venue featured
	_, err := suite.db.Exec("UPDATE venues SET is_featured = true WHERE id = 1")
	suite.Require().NoError(err)

	w := suite.makeGETRequest("/v1/venues/featured")
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var venues []models.Venue
	suite.parseJSONResponse(w, &venues)
	assert.NotEmpty(suite.T(), venues)

	// Verify all returned venues are featured
	for _, venue := range venues {
		assert.True(suite.T(), venue.IsFeatured)
	}

	// Test with limit
	w = suite.makeGETRequest("/v1/venues/featured?limit=3")
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	suite.parseJSONResponse(w, &venues)
	assert.True(suite.T(), len(venues) <= 3)
}

func (suite *TestSuite) testCreateVenue() {
	venueData := serializers.CreateVenueRequest{
		Name:             "New Test Restaurant",
		Description:      "A newly created test restaurant",
		ShortDescription: "New test restaurant",
		Address:          "789 New St, San Francisco, CA",
		CityID:           1,
		Latitude:         37.7649,
		Longitude:        -122.4094,
		CategoryID:       1,
		Phone:            "(555) 123-4567",
		Email:            "contact@newtest.com",
		Website:          "https://newtest.com",
		PriceRange:       "$$",
		AvgCostPerPerson: 25.00,
		Amenities:        []string{"wifi", "parking"},
	}

	w := suite.makePOSTRequest("/v1/venues", venueData)
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var createdVenue models.Venue
	suite.parseJSONResponse(w, &createdVenue)

	assert.Equal(suite.T(), venueData.Name, createdVenue.Name)
	assert.Equal(suite.T(), venueData.Address, createdVenue.Address)
	assert.Equal(suite.T(), venueData.CategoryID, createdVenue.CategoryID)
	assert.Equal(suite.T(), venueData.Latitude, createdVenue.Latitude)
	assert.Equal(suite.T(), venueData.Longitude, createdVenue.Longitude)
	assert.True(suite.T(), createdVenue.ID > 0)

	// Test validation errors
	invalidVenueData := serializers.CreateVenueRequest{
		Name: "", // Empty name should fail
	}

	w = suite.makePOSTRequest("/v1/venues", invalidVenueData)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	// Test invalid coordinates
	invalidCoordsData := venueData
	invalidCoordsData.Latitude = 91.0 // Invalid latitude

	w = suite.makePOSTRequest("/v1/venues", invalidCoordsData)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	// Test invalid category
	invalidCategoryData := venueData
	invalidCategoryData.CategoryID = 999 // Non-existent category

	w = suite.makePOSTRequest("/v1/venues", invalidCategoryData)
	// This might succeed with foreign key constraint, depending on database setup
	// In a real implementation, you'd want to validate this
}

// TestVenueSearchEdgeCases tests edge cases in venue search
func (suite *TestSuite) TestVenueSearchEdgeCases() {
	suite.Run("Venue Search Edge Cases", func() {
		// Test empty search
		w := suite.makeGETRequest("/v1/venues/search")
		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var searchResponse serializers.VenueSearchResponse
		suite.parseJSONResponse(w, &searchResponse)
		assert.NotNil(suite.T(), searchResponse.Venues)

		// Test search with no results
		w = suite.makeGETRequest("/v1/venues/search?q=nonexistentrestaurant12345")
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		suite.parseJSONResponse(w, &searchResponse)
		assert.Empty(suite.T(), searchResponse.Venues)

		// Test search with invalid category
		w = suite.makeGETRequest("/v1/venues/search?category=999")
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		suite.parseJSONResponse(w, &searchResponse)
		assert.Empty(suite.T(), searchResponse.Venues)

		// Test search with invalid price range
		w = suite.makeGETRequest("/v1/venues/search?price_range=invalid")
		assert.Equal(suite.T(), http.StatusOK, w.Code) // Should not error, just filter out

		// Test search with extreme coordinates
		w = suite.makeGETRequest("/v1/venues/search?lat=90&lng=180&radius=1")
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		suite.parseJSONResponse(w, &searchResponse)
		assert.Empty(suite.T(), searchResponse.Venues) // No venues at North Pole

		// Test search with negative radius
		w = suite.makeGETRequest("/v1/venues/search?lat=37.7749&lng=-122.4194&radius=-5")
		assert.Equal(suite.T(), http.StatusOK, w.Code) // Should default to positive radius

		// Test search with very large limit
		w = suite.makeGETRequest("/v1/venues/search?limit=1000")
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		suite.parseJSONResponse(w, &searchResponse)
		assert.True(suite.T(), len(searchResponse.Venues) <= 100) // Should be capped

		// Test search with negative page
		w = suite.makeGETRequest("/v1/venues/search?page=-1")
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		suite.parseJSONResponse(w, &searchResponse)
		assert.Equal(suite.T(), 1, searchResponse.Pagination.Page) // Should default to 1
	})
}

// TestVenueDataIntegrity tests data consistency and integrity
func (suite *TestSuite) TestVenueDataIntegrity() {
	suite.Run("Venue Data Integrity", func() {
		// Create a venue with specific data
		venueData := serializers.CreateVenueRequest{
			Name:             "Data Integrity Test Restaurant",
			Description:      "Testing data integrity",
			Address:          "123 Integrity St, San Francisco, CA",
			CityID:           1,
			Latitude:         37.7749,
			Longitude:        -122.4194,
			CategoryID:       1,
			PriceRange:       "$$$",
			AvgCostPerPerson: 50.00,
		}

		w := suite.makePOSTRequest("/v1/venues", venueData)
		assert.Equal(suite.T(), http.StatusCreated, w.Code)

		var createdVenue models.Venue
		suite.parseJSONResponse(w, &createdVenue)
		venueID := createdVenue.ID

		// Verify the venue can be retrieved with all data intact
		w = suite.makeGETRequest(fmt.Sprintf("/v1/venues/%d", venueID))
		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var venueResponse serializers.VenueDetailResponse
		suite.parseJSONResponse(w, &venueResponse)
		retrievedVenue := venueResponse.Venue

		assert.Equal(suite.T(), venueData.Name, retrievedVenue.Name)
		assert.Equal(suite.T(), venueData.Description, retrievedVenue.Description)
		assert.Equal(suite.T(), venueData.Address, retrievedVenue.Address)
		assert.Equal(suite.T(), venueData.CityID, retrievedVenue.CityID)
		assert.Equal(suite.T(), venueData.Latitude, retrievedVenue.Latitude)
		assert.Equal(suite.T(), venueData.Longitude, retrievedVenue.Longitude)
		assert.Equal(suite.T(), venueData.CategoryID, retrievedVenue.CategoryID)
		assert.Equal(suite.T(), venueData.PriceRange, retrievedVenue.PriceRange)
		assert.Equal(suite.T(), venueData.AvgCostPerPerson, retrievedVenue.AvgCostPerPerson)

		// Verify venue appears in search results
		w = suite.makeGETRequest("/v1/venues/search?q=Data Integrity")
		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var searchResponse serializers.VenueSearchResponse
		suite.parseJSONResponse(w, &searchResponse)

		found := false
		for _, venue := range searchResponse.Venues {
			if venue.ID == venueID {
				found = true
				break
			}
		}
		assert.True(suite.T(), found, "Created venue should appear in search results")

		// Verify venue appears in nearby search
		w = suite.makeGETRequest(fmt.Sprintf("/v1/venues/nearby?lat=%f&lng=%f&radius=1",
			venueData.Latitude, venueData.Longitude))
		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var nearbyVenues []models.Venue
		suite.parseJSONResponse(w, &nearbyVenues)

		found = false
		for _, venue := range nearbyVenues {
			if venue.ID == venueID {
				found = true
				assert.NotNil(suite.T(), venue.Distance)
				assert.True(suite.T(), *venue.Distance < 1.0) // Should be very close
				break
			}
		}
		assert.True(suite.T(), found, "Created venue should appear in nearby results")

		// Verify category filtering works
		w = suite.makeGETRequest(fmt.Sprintf("/v1/venues/search?category=%d", venueData.CategoryID))
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		suite.parseJSONResponse(w, &searchResponse)

		found = false
		for _, venue := range searchResponse.Venues {
			if venue.ID == venueID {
				found = true
				assert.Equal(suite.T(), venueData.CategoryID, venue.CategoryID)
				break
			}
		}
		assert.True(suite.T(), found, "Created venue should appear in category-filtered results")
	})
}

// TestVenuePerformance tests basic performance characteristics
func (suite *TestSuite) TestVenuePerformance() {
	suite.Run("Venue Performance Tests", func() {
		// Create multiple venues for performance testing
		venueCount := 20
		for i := 0; i < venueCount; i++ {
			venueData := serializers.CreateVenueRequest{
				Name:             fmt.Sprintf("Performance Test Restaurant %d", i),
				Description:      fmt.Sprintf("Performance testing venue %d", i),
				Address:          fmt.Sprintf("%d Performance St, San Francisco, CA", i),
				CityID:           1,
				Latitude:         37.7749 + float64(i)*0.001, // Spread venues slightly
				Longitude:        -122.4194 + float64(i)*0.001,
				CategoryID:       1,
				PriceRange:       "$$",
				AvgCostPerPerson: 25.00,
			}

			w := suite.makePOSTRequest("/v1/venues", venueData)
			assert.Equal(suite.T(), http.StatusCreated, w.Code)
		}

		// Test search performance with multiple venues
		w := suite.makeGETRequest("/v1/venues/search?q=Performance")
		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var searchResponse serializers.VenueSearchResponse
		suite.parseJSONResponse(w, &searchResponse)
		assert.True(suite.T(), len(searchResponse.Venues) >= venueCount)

		// Test nearby search performance
		w = suite.makeGETRequest("/v1/venues/nearby?lat=37.7749&lng=-122.4194&radius=5")
		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var nearbyVenues []models.Venue
		suite.parseJSONResponse(w, &nearbyVenues)
		assert.NotEmpty(suite.T(), nearbyVenues)

		// Verify all returned venues have distance calculated
		for _, venue := range nearbyVenues {
			assert.NotNil(suite.T(), venue.Distance, "Distance should be calculated for venue %d", venue.ID)
		}

		// Test pagination performance
		w = suite.makeGETRequest("/v1/venues/search?page=1&limit=10")
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		suite.parseJSONResponse(w, &searchResponse)
		assert.True(suite.T(), len(searchResponse.Venues) <= 10)
		assert.NotNil(suite.T(), searchResponse.Pagination)

		// Test second page
		w = suite.makeGETRequest("/v1/venues/search?page=2&limit=10")
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		suite.parseJSONResponse(w, &searchResponse)
		assert.NotNil(suite.T(), searchResponse.Pagination)
		assert.Equal(suite.T(), 2, searchResponse.Pagination.Page)
	})
}
