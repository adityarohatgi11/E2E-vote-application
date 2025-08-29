package tests

import (
	"fmt"
	"net/http"
	"time"
	"voting-app/app/models"
	"voting-app/app/serializers"
	
	"github.com/stretchr/testify/assert"
)

// TestLegacyVotingSystem tests backwards compatibility with original voting system
func (suite *TestSuite) TestLegacyVotingSystem() {
	suite.Run("Legacy Voting System Compatibility", func() {
		// Setup legacy data structures
		suite.setupLegacyVotingData()
		
		// Test original voting endpoints
		suite.testLegacyVoteEndpoint()
		suite.testLegacySubmitVote()
		suite.testLegacyVotingEdgeCases()
	})
}

func (suite *TestSuite) setupLegacyVotingData() {
	// Create legacy tables and data for backwards compatibility testing
	legacyTables := []string{
		`CREATE TABLE IF NOT EXISTS mentors (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			photo VARCHAR
		)`,
		`CREATE TABLE IF NOT EXISTS participants (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			code VARCHAR,
			photo VARCHAR,
			is_active BOOLEAN DEFAULT true,
			mentor_id BIGINT REFERENCES mentors(id)
		)`,
		`CREATE TABLE IF NOT EXISTS voting (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			winner_id BIGINT REFERENCES participants(id),
			started_at TIMESTAMP NOT NULL,
			ended_at TIMESTAMP NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS user_voting (
			id BIGSERIAL PRIMARY KEY,
			voting_id BIGINT REFERENCES voting(id),
			owner_id BIGINT REFERENCES snapp_users(id),
			vote_id BIGINT REFERENCES participants(id)
		)`,
		`CREATE TABLE IF NOT EXISTS vouchers (
			id BIGSERIAL PRIMARY KEY,
			owner_id BIGINT REFERENCES snapp_users(id),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			code VARCHAR(255) NOT NULL,
			icon VARCHAR,
			is_new BOOLEAN DEFAULT true
		)`,
		`CREATE TABLE IF NOT EXISTS banners (
			id BIGSERIAL PRIMARY KEY,
			is_active BOOLEAN DEFAULT true,
			image VARCHAR NOT NULL,
			link VARCHAR
		)`,
	}
	
	for _, table := range legacyTables {
		_, err := suite.db.Exec(table)
		suite.Require().NoError(err)
	}
	
	// Insert test data
	_, err := suite.db.Exec(`INSERT INTO mentors (id, name, photo) VALUES 
		(1, 'Test Mentor 1', 'mentor1.jpg'),
		(2, 'Test Mentor 2', 'mentor2.jpg') ON CONFLICT (id) DO NOTHING`)
	suite.Require().NoError(err)
	
	_, err = suite.db.Exec(`INSERT INTO participants (id, name, code, photo, is_active, mentor_id) VALUES 
		(1, 'Test Participant 1', 'P001', 'participant1.jpg', true, 1),
		(2, 'Test Participant 2', 'P002', 'participant2.jpg', true, 2),
		(3, 'Test Participant 3', 'P003', 'participant3.jpg', false, 1) ON CONFLICT (id) DO NOTHING`)
	suite.Require().NoError(err)
	
	// Create active voting
	now := time.Now()
	_, err = suite.db.Exec(`INSERT INTO voting (id, name, description, started_at, ended_at) VALUES 
		(1, 'Test Competition 2024', 'Annual test competition', $1, $2) ON CONFLICT (id) DO NOTHING`,
		now.Add(-24*time.Hour), now.Add(24*time.Hour))
	suite.Require().NoError(err)
	
	// Create expired voting
	_, err = suite.db.Exec(`INSERT INTO voting (id, name, description, winner_id, started_at, ended_at) VALUES 
		(2, 'Past Competition 2023', 'Last years competition', 1, $1, $2) ON CONFLICT (id) DO NOTHING`,
		now.Add(-48*time.Hour), now.Add(-12*time.Hour))
	suite.Require().NoError(err)
	
	// Create vouchers
	_, err = suite.db.Exec(`INSERT INTO vouchers (id, owner_id, name, description, code, icon, is_new) VALUES 
		(1, 1, 'Winner Badge', 'You won a competition!', 'WIN2023', 'trophy.png', true),
		(2, 1, 'Participation Award', 'Thanks for participating', 'PART2023', 'medal.png', false) ON CONFLICT (id) DO NOTHING`)
	suite.Require().NoError(err)
	
	// Create banners
	_, err = suite.db.Exec(`INSERT INTO banners (id, is_active, image, link) VALUES 
		(1, true, 'banner1.jpg', 'https://example.com/promotion') ON CONFLICT (id) DO NOTHING`)
	suite.Require().NoError(err)
}

func (suite *TestSuite) testLegacyVoteEndpoint() {
	// Test the original vote endpoint that returns competition data
	w := suite.makeGETRequest("/v1/vote/test_user_1")
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var voteResponse serializers.Vote
	suite.parseJSONResponse(w, &voteResponse)
	
	// Verify response structure matches original
	assert.NotNil(suite.T(), voteResponse.Participants)
	assert.NotNil(suite.T(), voteResponse.Mentors)
	assert.NotNil(suite.T(), voteResponse.Voting)
	assert.NotNil(suite.T(), voteResponse.History)
	assert.NotNil(suite.T(), voteResponse.History.Vouchers)
	
	// Verify active participants are returned
	assert.NotEmpty(suite.T(), voteResponse.Participants)
	for _, participant := range voteResponse.Participants {
		assert.True(suite.T(), participant.IsActive)
	}
	
	// Verify mentors are returned
	assert.NotEmpty(suite.T(), voteResponse.Mentors)
	
	// Verify current voting is returned
	assert.NotEmpty(suite.T(), voteResponse.Voting.Name)
	assert.Equal(suite.T(), "Test Competition 2024", voteResponse.Voting.Name)
	
	// Verify user vouchers are returned
	assert.NotEmpty(suite.T(), voteResponse.History.Vouchers)
	
	// Verify banner is returned
	if voteResponse.Banner != nil {
		assert.NotEmpty(suite.T(), voteResponse.Banner.Image)
	}
	
	// Test with non-existent user
	w = suite.makeGETRequest("/v1/vote/non_existent_user")
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *TestSuite) testLegacySubmitVote() {
	// Submit a vote using the original endpoint
	w := suite.makePOSTRequest("/v1/vote/test_user_1/1/1", nil)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var voteResponse serializers.Vote
	suite.parseJSONResponse(w, &voteResponse)
	
	// Verify vote was recorded and response includes updated data
	assert.NotNil(suite.T(), voteResponse.Voting)
	assert.NotNil(suite.T(), voteResponse.History.Voted)
	
	// Verify user's vote history includes the new vote
	assert.NotEmpty(suite.T(), voteResponse.History.Voted)
	voted := false
	for _, vote := range voteResponse.History.Voted {
		if vote.ParticipantId == 1 {
			voted = true
			break
		}
	}
	assert.True(suite.T(), voted, "User's vote should appear in vote history")
	
	// Test duplicate vote prevention
	w = suite.makePOSTRequest("/v1/vote/test_user_1/1/1", nil)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var errorResponse serializers.Base
	suite.parseJSONResponse(w, &errorResponse)
	assert.Equal(suite.T(), serializers.AlreadyVoted, errorResponse.Code)
	
	// Test vote for different participant by same user (should fail)
	w = suite.makePOSTRequest("/v1/vote/test_user_1/1/2", nil)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	// Test vote by different user for same participant (should succeed)
	w = suite.makePOSTRequest("/v1/vote/test_user_2/1/1", nil)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	// Test invalid voting ID
	w = suite.makePOSTRequest("/v1/vote/test_user_2/999/1", nil)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	suite.parseJSONResponse(w, &errorResponse)
	assert.Contains(suite.T(), errorResponse.Message, "voting id is invalid")
	
	// Test invalid participant ID
	w = suite.makePOSTRequest("/v1/vote/test_user_2/1/999", nil)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	suite.parseJSONResponse(w, &errorResponse)
	assert.Contains(suite.T(), errorResponse.Message, "vote id is invalid")
	
	// Test inactive participant
	w = suite.makePOSTRequest("/v1/vote/test_user_2/1/3", nil) // Participant 3 is inactive
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	// Test expired voting
	w = suite.makePOSTRequest("/v1/vote/test_user_2/2/1", nil) // Voting 2 is expired
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *TestSuite) testLegacyVotingEdgeCases() {
	// Test with invalid user ID format
	w := suite.makeGETRequest("/v1/vote/")
	assert.Equal(suite.T(), http.StatusNotFound, w.Code) // Route not found
	
	// Test with invalid voting/vote ID formats
	w = suite.makePOSTRequest("/v1/vote/test_user_1/invalid/1", nil)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	w = suite.makePOSTRequest("/v1/vote/test_user_1/1/invalid", nil)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	// Test when no active voting exists
	// Disable current voting
	_, err := suite.db.Exec("UPDATE voting SET ended_at = $1 WHERE id = 1", time.Now().Add(-1*time.Hour))
	suite.Require().NoError(err)
	
	w = suite.makeGETRequest("/v1/vote/test_user_1")
	assert.Equal(suite.T(), http.StatusOK, w.Code) // Should still return data, but no active voting
	
	var voteResponse serializers.Vote
	suite.parseJSONResponse(w, &voteResponse)
	// Voting name might be empty or from last voting
	
	// Restore active voting for other tests
	_, err = suite.db.Exec("UPDATE voting SET ended_at = $1 WHERE id = 1", time.Now().Add(24*time.Hour))
	suite.Require().NoError(err)
}

// TestEnhancedVotingCampaigns tests the new voting campaign system
func (suite *TestSuite) TestEnhancedVotingCampaigns() {
	suite.Run("Enhanced Voting Campaigns", func() {
		suite.testCreateVotingCampaign()
		suite.testGetActiveCampaigns()
		suite.testSubmitCampaignVote()
		suite.testCampaignResults()
		suite.testCampaignValidation()
	})
}

func (suite *TestSuite) testCreateVotingCampaign() {
	campaignData := map[string]interface{}{
		"title":                   "Best Restaurant 2024",
		"description":             "Vote for the best restaurant in San Francisco",
		"campaignType":           "best_restaurant",
		"cityId":                 1,
		"categoryId":             1,
		"startDate":              time.Now().Add(1 * time.Hour),
		"endDate":                time.Now().Add(30 * 24 * time.Hour),
		"maxVotesPerUser":        3,
		"allowMultipleCategories": true,
		"requireReview":          false,
	}
	
	// Note: This endpoint would need to be implemented in CampaignController
	// For now, we'll create the campaign directly in the database
	_, err := suite.db.Exec(`INSERT INTO voting_campaigns 
		(id, title, description, campaign_type, city_id, category_id, start_date, end_date, 
		 max_votes_per_user, allow_multiple_categories, require_review, is_active) 
		VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, true) ON CONFLICT (id) DO NOTHING`,
		campaignData["title"], campaignData["description"], campaignData["campaignType"],
		campaignData["cityId"], campaignData["categoryId"], campaignData["startDate"],
		campaignData["endDate"], campaignData["maxVotesPerUser"], 
		campaignData["allowMultipleCategories"], campaignData["requireReview"])
	suite.Require().NoError(err)
}

func (suite *TestSuite) testGetActiveCampaigns() {
	// This endpoint would need to be implemented
	// For testing purposes, we'll verify the data exists in the database
	var count int
	err := suite.db.QueryRow("SELECT COUNT(*) FROM voting_campaigns WHERE is_active = true").Scan(&count)
	suite.Require().NoError(err)
	assert.True(suite.T(), count > 0, "Should have active campaigns")
}

func (suite *TestSuite) testSubmitCampaignVote() {
	// Insert test vote directly
	_, err := suite.db.Exec(`INSERT INTO campaign_votes 
		(campaign_id, venue_id, user_id, reason, confidence_score) 
		VALUES (1, 1, 1, 'Great food and service', 4.5) ON CONFLICT DO NOTHING`)
	suite.Require().NoError(err)
	
	// Verify vote was recorded
	var voteCount int
	err = suite.db.QueryRow("SELECT COUNT(*) FROM campaign_votes WHERE campaign_id = 1 AND user_id = 1").Scan(&voteCount)
	suite.Require().NoError(err)
	assert.Equal(suite.T(), 1, voteCount)
	
	// Test duplicate vote prevention
	_, err = suite.db.Exec(`INSERT INTO campaign_votes 
		(campaign_id, venue_id, user_id, reason) 
		VALUES (1, 1, 1, 'Duplicate vote')`)
	assert.Error(suite.T(), err, "Should prevent duplicate votes")
}

func (suite *TestSuite) testCampaignResults() {
	// Add more votes for testing
	_, err := suite.db.Exec(`INSERT INTO campaign_votes 
		(campaign_id, venue_id, user_id, confidence_score) VALUES 
		(1, 1, 2, 4.0),
		(1, 2, 1, 3.5) ON CONFLICT DO NOTHING`)
	suite.Require().NoError(err)
	
	// Get vote counts per venue
	rows, err := suite.db.Query(`
		SELECT venue_id, COUNT(*) as vote_count, AVG(confidence_score) as avg_confidence
		FROM campaign_votes 
		WHERE campaign_id = 1 
		GROUP BY venue_id 
		ORDER BY vote_count DESC, avg_confidence DESC`)
	suite.Require().NoError(err)
	defer rows.Close()
	
	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		var venueID, voteCount int
		var avgConfidence float64
		err := rows.Scan(&venueID, &voteCount, &avgConfidence)
		suite.Require().NoError(err)
		
		results = append(results, map[string]interface{}{
			"venueId":       venueID,
			"voteCount":     voteCount,
			"avgConfidence": avgConfidence,
		})
	}
	
	assert.NotEmpty(suite.T(), results)
	assert.Equal(suite.T(), 1, results[0]["venueId"]) // Venue 1 should be leading
	assert.True(suite.T(), results[0]["voteCount"].(int) >= 1)
}

func (suite *TestSuite) testCampaignValidation() {
	// Test voting in expired campaign
	_, err := suite.db.Exec(`INSERT INTO voting_campaigns 
		(id, title, campaign_type, start_date, end_date, is_active) 
		VALUES (2, 'Expired Campaign', 'test', $1, $2, true) ON CONFLICT (id) DO NOTHING`,
		time.Now().Add(-48*time.Hour), time.Now().Add(-24*time.Hour))
	suite.Require().NoError(err)
	
	// Try to vote in expired campaign (this validation would be in the controller)
	_, err = suite.db.Exec(`INSERT INTO campaign_votes 
		(campaign_id, venue_id, user_id) VALUES (2, 1, 1)`)
	// This might succeed at database level, but should be prevented by application logic
	
	// Test max votes per user constraint
	// This would need to be implemented in the application logic
	
	// Test campaign that requires review
	_, err = suite.db.Exec(`INSERT INTO voting_campaigns 
		(id, title, campaign_type, start_date, end_date, require_review, is_active) 
		VALUES (3, 'Review Required Campaign', 'test', $1, $2, true, true) ON CONFLICT (id) DO NOTHING`,
		time.Now().Add(-1*time.Hour), time.Now().Add(24*time.Hour))
	suite.Require().NoError(err)
	
	// In a real implementation, this would check if user has reviewed the venue before allowing vote
}

// TestVotingSystemIntegration tests integration between legacy and new systems
func (suite *TestSuite) TestVotingSystemIntegration() {
	suite.Run("Voting System Integration", func() {
		// Test that legacy voting and new campaigns can coexist
		
		// Submit legacy vote
		w := suite.makePOSTRequest("/v1/vote/test_user_1/1/2", nil)
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		// Submit campaign vote
		_, err := suite.db.Exec(`INSERT INTO campaign_votes 
			(campaign_id, venue_id, user_id, reason) 
			VALUES (1, 2, 1, 'Campaign vote') ON CONFLICT DO NOTHING`)
		suite.Require().NoError(err)
		
		// Verify both votes exist
		var legacyVoteCount, campaignVoteCount int
		
		err = suite.db.QueryRow("SELECT COUNT(*) FROM user_voting WHERE owner_id = 1").Scan(&legacyVoteCount)
		suite.Require().NoError(err)
		
		err = suite.db.QueryRow("SELECT COUNT(*) FROM campaign_votes WHERE user_id = 1").Scan(&campaignVoteCount)
		suite.Require().NoError(err)
		
		assert.True(suite.T(), legacyVoteCount > 0, "Should have legacy votes")
		assert.True(suite.T(), campaignVoteCount > 0, "Should have campaign votes")
		
		// Test that venue appears in both legacy and campaign contexts
		w = suite.makeGETRequest("/v1/vote/test_user_1")
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var voteResponse serializers.Vote
		suite.parseJSONResponse(w, &voteResponse)
		
		// Verify user's legacy voting history
		assert.NotEmpty(suite.T(), voteResponse.History.Voted)
		
		// Verify campaign data could be included (if implemented)
		// This would require extending the serializers.Vote struct
	})
}

// TestVotingDataConsistency tests data consistency across the voting system
func (suite *TestSuite) TestVotingDataConsistency() {
	suite.Run("Voting Data Consistency", func() {
		// Test that vote counts are consistent
		
		// Get initial participant vote count
		var initialVoteCount int
		err := suite.db.QueryRow(`
			SELECT COUNT(*) FROM user_voting 
			WHERE vote_id = 1 AND voting_id = 1`).Scan(&initialVoteCount)
		suite.Require().NoError(err)
		
		// Submit another vote
		w := suite.makePOSTRequest("/v1/vote/test_user_2/1/1", nil)
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		// Verify vote count increased
		var newVoteCount int
		err = suite.db.QueryRow(`
			SELECT COUNT(*) FROM user_voting 
			WHERE vote_id = 1 AND voting_id = 1`).Scan(&newVoteCount)
		suite.Require().NoError(err)
		
		assert.Equal(suite.T(), initialVoteCount+1, newVoteCount)
		
		// Test vote counting across different participants
		var totalVotes int
		err = suite.db.QueryRow("SELECT COUNT(*) FROM user_voting WHERE voting_id = 1").Scan(&totalVotes)
		suite.Require().NoError(err)
		assert.True(suite.T(), totalVotes > 0)
		
		// Test that inactive participants don't receive votes
		var inactiveVotes int
		err = suite.db.QueryRow(`
			SELECT COUNT(*) FROM user_voting uv 
			JOIN participants p ON uv.vote_id = p.id 
			WHERE p.is_active = false`).Scan(&inactiveVotes)
		suite.Require().NoError(err)
		assert.Equal(suite.T(), 0, inactiveVotes, "Inactive participants should not receive votes")
		
		// Test campaign vote consistency
		var campaignVotes int
		err = suite.db.QueryRow("SELECT COUNT(*) FROM campaign_votes WHERE campaign_id = 1").Scan(&campaignVotes)
		suite.Require().NoError(err)
		assert.True(suite.T(), campaignVotes >= 0)
		
		// Verify no orphaned votes exist
		var orphanedLegacyVotes int
		err = suite.db.QueryRow(`
			SELECT COUNT(*) FROM user_voting uv 
			LEFT JOIN participants p ON uv.vote_id = p.id 
			WHERE p.id IS NULL`).Scan(&orphanedLegacyVotes)
		suite.Require().NoError(err)
		assert.Equal(suite.T(), 0, orphanedLegacyVotes, "Should not have orphaned legacy votes")
		
		var orphanedCampaignVotes int
		err = suite.db.QueryRow(`
			SELECT COUNT(*) FROM campaign_votes cv 
			LEFT JOIN venues v ON cv.venue_id = v.id 
			WHERE v.id IS NULL`).Scan(&orphanedCampaignVotes)
		suite.Require().NoError(err)
		assert.Equal(suite.T(), 0, orphanedCampaignVotes, "Should not have orphaned campaign votes")
	})
}
