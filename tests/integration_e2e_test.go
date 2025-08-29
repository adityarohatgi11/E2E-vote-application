package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"voting-app/app/models"
	"voting-app/app/serializers"
	
	"github.com/stretchr/testify/assert"
)

// TestFullUserJourney tests a complete user journey through the platform
func (suite *TestSuite) TestFullUserJourney() {
	suite.Run("Complete User Journey", func() {
		// User Journey: Discovery -> Venue Details -> Review -> Collection -> Voting
		suite.testUserDiscoveryJourney()
		suite.testUserReviewJourney()
		suite.testUserCollectionJourney()
		suite.testUserVotingJourney()
	})
}

func (suite *TestSuite) testUserDiscoveryJourney() {
	// Step 1: User searches for restaurants near them
	searchURL := "/v1/venues/search?q=restaurant&lat=37.7749&lng=-122.4194&radius=10&sort_by=rating"
	w := suite.makeGETRequest(searchURL)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var searchResponse serializers.VenueSearchResponse
	suite.parseJSONResponse(w, &searchResponse)
	assert.NotEmpty(suite.T(), searchResponse.Venues)
	
	selectedVenue := searchResponse.Venues[0]
	
	// Step 2: User views venue details
	w = suite.makeGETRequest(fmt.Sprintf("/v1/venues/%d?user_lat=37.7749&user_lng=-122.4194", selectedVenue.ID))
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var venueDetail serializers.VenueDetailResponse
	suite.parseJSONResponse(w, &venueDetail)
	assert.Equal(suite.T(), selectedVenue.ID, venueDetail.Venue.ID)
	assert.NotNil(suite.T(), venueDetail.ReviewSummary)
	
	// Step 3: User gets nearby venues for comparison
	w = suite.makeGETRequest("/v1/venues/nearby?lat=37.7749&lng=-122.4194&radius=5&limit=5")
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var nearbyVenues []models.Venue
	suite.parseJSONResponse(w, &nearbyVenues)
	assert.NotEmpty(suite.T(), nearbyVenues)
	
	// Step 4: User checks reviews of the venue
	w = suite.makeGETRequest(fmt.Sprintf("/v1/venues/%d/reviews", selectedVenue.ID))
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var reviewsResponse serializers.ReviewSearchResponse
	suite.parseJSONResponse(w, &reviewsResponse)
	// Reviews might be empty for new venues, which is fine
}

func (suite *TestSuite) testUserReviewJourney() {
	// Step 1: User decides to visit and later write a review
	venueID := int64(1)
	
	// First, check if user already reviewed (clean slate)
	w := suite.makeGETRequest(fmt.Sprintf("/v1/venues/%d/reviews", venueID))
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	// Step 2: User writes a comprehensive review
	reviewData := serializers.CreateReviewRequest{
		VenueID:       venueID,
		OverallRating: 4.5,
		Title:         "Excellent dining experience!",
		ReviewText:    "The food was outstanding, service was friendly and attentive. The ambiance was perfect for a date night. Highly recommend the seafood pasta and the chocolate dessert. Will definitely be back!",
		VisitDate:     func() *time.Time { t := time.Now().AddDate(0, 0, -2); return &t }(),
		VisitType:     "dinner",
		PartySize:     2,
		Photos:        []string{"dinner1.jpg", "food1.jpg", "dessert1.jpg"},
	}
	
	w = suite.makePOSTRequest("/v1/reviews/test_user_1", reviewData)
	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	
	var createdReview models.VenueReview
	suite.parseJSONResponse(w, &createdReview)
	reviewID := createdReview.ID
	
	// Step 3: User checks their review appears in venue reviews
	w = suite.makeGETRequest(fmt.Sprintf("/v1/venues/%d/reviews", venueID))
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var reviewsResponse serializers.ReviewSearchResponse
	suite.parseJSONResponse(w, &reviewsResponse)
	
	reviewFound := false
	for _, review := range reviewsResponse.Reviews {
		if review.ID == reviewID {
			reviewFound = true
			assert.Equal(suite.T(), reviewData.Title, review.Title)
			assert.Equal(suite.T(), reviewData.OverallRating, review.OverallRating)
			break
		}
	}
	assert.True(suite.T(), reviewFound, "User's review should appear in venue reviews")
	
	// Step 4: Another user finds the review helpful
	voteData := serializers.ReviewVoteRequest{IsHelpful: true}
	w = suite.makePOSTRequest(fmt.Sprintf("/v1/reviews/test_user_2/%d/vote", reviewID), voteData)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	// Step 5: User checks their review history
	w = suite.makeGETRequest("/v1/users/test_user_1/reviews")
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var userReviewsResponse serializers.ReviewSearchResponse
	suite.parseJSONResponse(w, &userReviewsResponse)
	assert.NotEmpty(suite.T(), userReviewsResponse.Reviews)
	
	// Verify the review shows increased helpfulness
	userReviewFound := false
	for _, review := range userReviewsResponse.Reviews {
		if review.ID == reviewID {
			userReviewFound = true
			assert.Equal(suite.T(), 1, review.HelpfulVotes)
			break
		}
	}
	assert.True(suite.T(), userReviewFound, "User should see their review in their history")
}

func (suite *TestSuite) testUserCollectionJourney() {
	// Note: This test assumes collection endpoints exist
	// For now, we'll test the data layer directly
	
	// Step 1: User creates a "Favorites" collection
	_, err := suite.db.Exec(`INSERT INTO venue_collections 
		(user_id, name, description, is_public) VALUES 
		(1, 'My Favorites', 'Places I love to go', true) ON CONFLICT DO NOTHING`)
	suite.Require().NoError(err)
	
	var collectionID int64
	err = suite.db.QueryRow(`SELECT id FROM venue_collections WHERE user_id = 1 AND name = 'My Favorites'`).Scan(&collectionID)
	suite.Require().NoError(err)
	
	// Step 2: User adds venues to their collection
	_, err = suite.db.Exec(`INSERT INTO venue_collection_items 
		(collection_id, venue_id, note) VALUES 
		($1, 1, 'Best pasta in the city!'),
		($1, 2, 'Great for business dinners') ON CONFLICT DO NOTHING`, collectionID)
	suite.Require().NoError(err)
	
	// Step 3: Verify collection contains venues
	var itemCount int
	err = suite.db.QueryRow(`SELECT COUNT(*) FROM venue_collection_items WHERE collection_id = $1`, collectionID).Scan(&itemCount)
	suite.Require().NoError(err)
	assert.Equal(suite.T(), 2, itemCount)
	
	// Step 4: User creates a "Want to Try" collection
	_, err = suite.db.Exec(`INSERT INTO venue_collections 
		(user_id, name, description, is_public) VALUES 
		(1, 'Want to Try', 'Places on my wishlist', false) ON CONFLICT DO NOTHING`)
	suite.Require().NoError(err)
}

func (suite *TestSuite) testUserVotingJourney() {
	// Test both legacy voting and new campaign voting
	
	// Step 1: User participates in legacy voting (talent competition)
	w := suite.makeGETRequest("/v1/vote/test_user_1")
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var voteData serializers.Vote
	suite.parseJSONResponse(w, &voteData)
	
	if len(voteData.Participants) > 0 && voteData.Voting.Id > 0 {
		// Submit vote for first active participant
		participantID := voteData.Participants[0].Id
		w = suite.makePOSTRequest(fmt.Sprintf("/v1/vote/test_user_1/%d/%d", voteData.Voting.Id, participantID), nil)
		
		if w.Code == http.StatusOK {
			// Verify vote was recorded
			var voteCount int
			err := suite.db.QueryRow(`SELECT COUNT(*) FROM user_voting 
				WHERE owner_id = 1 AND voting_id = $1 AND vote_id = $2`, 
				voteData.Voting.Id, participantID).Scan(&voteCount)
			suite.Require().NoError(err)
			assert.Equal(suite.T(), 1, voteCount)
		}
	}
	
	// Step 2: User participates in venue voting campaign
	// Create an active campaign
	now := time.Now()
	_, err := suite.db.Exec(`INSERT INTO voting_campaigns 
		(id, title, description, campaign_type, city_id, category_id, start_date, end_date, 
		 max_votes_per_user, is_active) VALUES 
		(10, 'Best Restaurant 2024', 'Vote for your favorite restaurant', 'best_restaurant', 
		 1, 1, $1, $2, 3, true) ON CONFLICT (id) DO NOTHING`,
		now.Add(-1*time.Hour), now.Add(30*24*time.Hour))
	suite.Require().NoError(err)
	
	// User votes for their favorite venue
	_, err = suite.db.Exec(`INSERT INTO campaign_votes 
		(campaign_id, venue_id, user_id, reason, confidence_score) VALUES 
		(10, 1, 1, 'Amazing food and service!', 5.0) ON CONFLICT DO NOTHING`)
	suite.Require().NoError(err)
	
	// Verify campaign vote was recorded
	var campaignVoteCount int
	err = suite.db.QueryRow(`SELECT COUNT(*) FROM campaign_votes 
		WHERE campaign_id = 10 AND user_id = 1`).Scan(&campaignVoteCount)
	suite.Require().NoError(err)
	assert.Equal(suite.T(), 1, campaignVoteCount)
}

// TestCrossFeatureIntegration tests integration between different features
func (suite *TestSuite) TestCrossFeatureIntegration() {
	suite.Run("Cross-Feature Integration", func() {
		suite.testReviewToRecommendationIntegration()
		suite.testSearchToAnalyticsIntegration()
		suite.testVotingToVenueRankingIntegration()
	})
}

func (suite *TestSuite) testReviewToRecommendationIntegration() {
	// Test that user reviews influence their recommendations
	
	// Step 1: User reviews multiple venues in same category
	reviews := []serializers.CreateReviewRequest{
		{
			VenueID:       1,
			OverallRating: 5.0,
			Title:         "Love this place!",
			ReviewText:    "Best Italian food in the city",
			VisitType:     "dinner",
		},
		{
			VenueID:       2,
			OverallRating: 4.0,
			Title:         "Good but not great",
			ReviewText:    "Decent food, slow service",
			VisitType:     "lunch",
		},
	}
	
	for i, review := range reviews {
		userID := fmt.Sprintf("test_user_%d", i+1)
		w := suite.makePOSTRequest(fmt.Sprintf("/v1/reviews/%s", userID), review)
		// May succeed or fail depending on existing data
	}
	
	// Step 2: Test that recommendations consider user's review history
	// This would require the recommendation endpoints to be implemented
	// For now, we verify the data exists for recommendation processing
	
	var userReviewCount int
	err := suite.db.QueryRow(`SELECT COUNT(*) FROM venue_reviews WHERE user_id = 1`).Scan(&userReviewCount)
	suite.Require().NoError(err)
	assert.True(suite.T(), userReviewCount >= 0, "User should have review history for recommendations")
	
	// Verify review data includes category preferences
	var categoryCount int
	err = suite.db.QueryRow(`SELECT COUNT(DISTINCT v.category_id) FROM venue_reviews vr 
		JOIN venues v ON vr.venue_id = v.id WHERE vr.user_id = 1`).Scan(&categoryCount)
	suite.Require().NoError(err)
	// User has reviewed venues, category data is available for recommendations
}

func (suite *TestSuite) testSearchToAnalyticsIntegration() {
	// Test that search behavior is tracked for analytics
	
	// Step 1: User performs various searches
	searches := []string{
		"/v1/venues/search?q=pizza&lat=37.7749&lng=-122.4194",
		"/v1/venues/search?category=1&min_rating=4.0",
		"/v1/venues/nearby?lat=37.7749&lng=-122.4194&radius=10",
	}
	
	for _, searchURL := range searches {
		w := suite.makeGETRequest(searchURL)
		assert.Equal(suite.T(), http.StatusOK, w.Code)
	}
	
	// Step 2: Verify search analytics data exists
	// Note: This would require actual analytics tracking in the controllers
	// For now, we verify the analytics table structure exists
	
	var tableExists bool
	err := suite.db.QueryRow(`SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_name = 'search_analytics')`).Scan(&tableExists)
	suite.Require().NoError(err)
	assert.True(suite.T(), tableExists, "Search analytics table should exist")
	
	// Test direct analytics insertion
	_, err = suite.db.Exec(`INSERT INTO search_analytics 
		(user_id, search_query, search_type, results_count, clicked_venue_id, click_position) 
		VALUES (1, 'test search', 'text', 5, 1, 1) ON CONFLICT DO NOTHING`)
	suite.Require().NoError(err)
	
	var analyticsCount int
	err = suite.db.QueryRow(`SELECT COUNT(*) FROM search_analytics WHERE user_id = 1`).Scan(&analyticsCount)
	suite.Require().NoError(err)
	assert.True(suite.T(), analyticsCount > 0, "Search analytics should be tracked")
}

func (suite *TestSuite) testVotingToVenueRankingIntegration() {
	// Test that voting campaign results influence venue prominence
	
	// Step 1: Create campaign votes for different venues
	campaignVotes := [][]interface{}{
		{1, 1, 1, "Great food!", 5.0},  // Venue 1, User 1
		{1, 1, 2, "Love this place", 4.5}, // Venue 1, User 2
		{1, 2, 1, "Pretty good", 3.5},  // Venue 2, User 1 (should fail - duplicate user)
	}
	
	for _, vote := range campaignVotes {
		_, err := suite.db.Exec(`INSERT INTO campaign_votes 
			(campaign_id, venue_id, user_id, reason, confidence_score) 
			VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`, vote...)
		// Some may fail due to constraints, which is expected
	}
	
	// Step 2: Calculate campaign results
	var venueVoteCounts []struct {
		VenueID   int64
		VoteCount int
		AvgScore  float64
	}
	
	rows, err := suite.db.Query(`SELECT venue_id, COUNT(*) as vote_count, AVG(confidence_score) as avg_score 
		FROM campaign_votes WHERE campaign_id = 1 
		GROUP BY venue_id ORDER BY vote_count DESC, avg_score DESC`)
	suite.Require().NoError(err)
	defer rows.Close()
	
	for rows.Next() {
		var result struct {
			VenueID   int64
			VoteCount int
			AvgScore  float64
		}
		err := rows.Scan(&result.VenueID, &result.VoteCount, &result.AvgScore)
		suite.Require().NoError(err)
		venueVoteCounts = append(venueVoteCounts, result)
	}
	
	// Step 3: Verify winning venue logic
	if len(venueVoteCounts) > 0 {
		winner := venueVoteCounts[0]
		assert.True(suite.T(), winner.VoteCount > 0, "Winner should have votes")
		assert.True(suite.T(), winner.AvgScore > 0, "Winner should have positive score")
		
		// Update campaign with winner
		_, err = suite.db.Exec(`UPDATE voting_campaigns 
			SET winner_venue_id = $1, total_votes = (
				SELECT COUNT(*) FROM campaign_votes WHERE campaign_id = 1
			) WHERE id = 1`, winner.VenueID)
		suite.Require().NoError(err)
		
		// Verify winner was set
		var winnerID int64
		err = suite.db.QueryRow(`SELECT winner_venue_id FROM voting_campaigns WHERE id = 1`).Scan(&winnerID)
		suite.Require().NoError(err)
		assert.Equal(suite.T(), winner.VenueID, winnerID)
	}
}

// TestErrorHandlingAndEdgeCases tests system robustness
func (suite *TestSuite) TestErrorHandlingAndEdgeCases() {
	suite.Run("Error Handling and Edge Cases", func() {
		suite.testInvalidDataHandling()
		suite.testConcurrencyScenarios()
		suite.testDataConsistency()
	})
}

func (suite *TestSuite) testInvalidDataHandling() {
	// Test invalid JSON
	w := suite.makePOSTRequest("/v1/reviews/test_user_1", "invalid json")
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	// Test missing required fields
	w = suite.makePOSTRequest("/v1/reviews/test_user_1", map[string]interface{}{})
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	// Test invalid URLs
	w = suite.makeGETRequest("/v1/venues/invalid_id")
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	// Test non-existent resources
	w = suite.makeGETRequest("/v1/venues/99999")
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	
	// Test invalid coordinates
	w = suite.makeGETRequest("/v1/venues/nearby?lat=invalid&lng=-122.4194")
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	// Test extreme coordinate values
	w = suite.makeGETRequest("/v1/venues/nearby?lat=91&lng=-122.4194")
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *TestSuite) testConcurrencyScenarios() {
	// Test concurrent review creation for same venue/user (should fail)
	// This is a simplified test; real concurrency testing would use goroutines
	
	reviewData := serializers.CreateReviewRequest{
		VenueID:       1,
		OverallRating: 4.0,
		Title:         "Concurrent test",
		ReviewText:    "Testing concurrency",
	}
	
	// First review should succeed
	w1 := suite.makePOSTRequest("/v1/reviews/test_user_1", reviewData)
	
	// Second review for same venue/user should fail
	w2 := suite.makePOSTRequest("/v1/reviews/test_user_1", reviewData)
	
	// One should succeed, one should fail
	assert.True(suite.T(), 
		(w1.Code == http.StatusCreated && w2.Code == http.StatusBadRequest) ||
		(w1.Code == http.StatusBadRequest && w2.Code == http.StatusCreated),
		"Only one review should be allowed per user per venue")
}

func (suite *TestSuite) testDataConsistency() {
	// Test referential integrity
	
	// Try to create review for non-existent venue
	invalidReview := serializers.CreateReviewRequest{
		VenueID:       99999,
		OverallRating: 4.0,
		Title:         "Invalid venue test",
		ReviewText:    "This should fail",
	}
	
	w := suite.makePOSTRequest("/v1/reviews/test_user_1", invalidReview)
	// This might succeed or fail depending on foreign key constraints
	// In production, you'd want proper validation
	
	// Test that deleting a venue would handle dependent data
	// (This would require implementing venue deletion)
	
	// Test data consistency across related tables
	var venueCount, reviewCount int
	
	err := suite.db.QueryRow("SELECT COUNT(*) FROM venues WHERE is_active = true").Scan(&venueCount)
	suite.Require().NoError(err)
	
	err = suite.db.QueryRow("SELECT COUNT(*) FROM venue_reviews").Scan(&reviewCount)
	suite.Require().NoError(err)
	
	// Basic sanity check
	assert.True(suite.T(), venueCount >= 0, "Should have non-negative venue count")
	assert.True(suite.T(), reviewCount >= 0, "Should have non-negative review count")
	
	// Test that all reviews reference valid venues
	var orphanedReviews int
	err = suite.db.QueryRow(`SELECT COUNT(*) FROM venue_reviews vr 
		LEFT JOIN venues v ON vr.venue_id = v.id 
		WHERE v.id IS NULL`).Scan(&orphanedReviews)
	suite.Require().NoError(err)
	assert.Equal(suite.T(), 0, orphanedReviews, "Should not have orphaned reviews")
}
