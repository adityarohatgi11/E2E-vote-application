package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	databases "voting-app/app"
	"voting-app/app/models"

	"github.com/getsentry/sentry-go"
)

// RecommendationEngine provides personalized venue recommendations
type RecommendationEngine struct{}

// RecommendationScore represents a venue with its recommendation score
type RecommendationScore struct {
	Venue   models.Venue `json:"venue"`
	Score   float64      `json:"score"`
	Reasons []string     `json:"reasons"` // Why this venue was recommended
}

// UserPreferences represents user's preferences extracted from their behavior
type UserPreferences struct {
	UserID              int64             `json:"userId"`
	PreferredCategories map[int64]float64 `json:"preferredCategories"` // CategoryID -> Weight
	PreferredPriceRange []string          `json:"preferredPriceRange"`
	PreferredAmenities  []string          `json:"preferredAmenities"`
	AverageRatingGiven  float64           `json:"averageRatingGiven"`
	PreferredLocations  []LocationPref    `json:"preferredLocations"`
	ActivityHours       map[string]int    `json:"activityHours"`   // Hour -> Frequency
	SocialInfluence     []int64           `json:"socialInfluence"` // Followed users
}

type LocationPref struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Weight    float64 `json:"weight"`
	Radius    float64 `json:"radius"` // in km
}

// RecommendationContext provides context for generating recommendations
type RecommendationContext struct {
	UserID      int64    `json:"userId"`
	UserLat     *float64 `json:"userLat,omitempty"`
	UserLng     *float64 `json:"userLng,omitempty"`
	TimeOfDay   string   `json:"timeOfDay,omitempty"` // morning, afternoon, evening, night
	Occasion    string   `json:"occasion,omitempty"`  // casual, date, business, celebration
	GroupSize   int      `json:"groupSize,omitempty"`
	MaxDistance float64  `json:"maxDistance"` // in km
	Limit       int      `json:"limit"`
}

// GetPersonalizedRecommendations generates personalized venue recommendations
func (re *RecommendationEngine) GetPersonalizedRecommendations(ctx RecommendationContext) ([]RecommendationScore, error) {
	// Step 1: Extract user preferences
	preferences, err := re.extractUserPreferences(ctx.UserID)
	if err != nil {
		return nil, err
	}

	// Step 2: Get candidate venues
	candidates, err := re.getCandidateVenues(ctx, preferences)
	if err != nil {
		return nil, err
	}

	// Step 3: Score each venue
	scores := make([]RecommendationScore, 0, len(candidates))
	for _, venue := range candidates {
		score := re.calculateRecommendationScore(venue, preferences, ctx)
		if score.Score > 0 {
			scores = append(scores, score)
		}
	}

	// Step 4: Sort by score and return top results
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	if len(scores) > ctx.Limit {
		scores = scores[:ctx.Limit]
	}

	return scores, nil
}

// extractUserPreferences analyzes user's past behavior to extract preferences
func (re *RecommendationEngine) extractUserPreferences(userID int64) (*UserPreferences, error) {
	prefs := &UserPreferences{
		UserID:              userID,
		PreferredCategories: make(map[int64]float64),
		ActivityHours:       make(map[string]int),
	}

	// Analyze reviews to extract category preferences
	err := re.analyzeReviewPreferences(userID, prefs)
	if err != nil {
		sentry.CaptureException(err)
	}

	// Analyze check-ins for location and time preferences
	err = re.analyzeCheckinPreferences(userID, prefs)
	if err != nil {
		sentry.CaptureException(err)
	}

	// Analyze social connections
	err = re.analyzeSocialPreferences(userID, prefs)
	if err != nil {
		sentry.CaptureException(err)
	}

	return prefs, nil
}

// analyzeReviewPreferences extracts preferences from user's reviews
func (re *RecommendationEngine) analyzeReviewPreferences(userID int64, prefs *UserPreferences) error {
	query := `
		SELECT v.category_id, v.price_range, v.amenities, r.overall_rating, r.visit_type,
			   COUNT(*) as frequency
		FROM venue_reviews r
		JOIN venues v ON r.venue_id = v.id
		WHERE r.user_id = $1 AND r.moderation_status = 'approved'
		GROUP BY v.category_id, v.price_range, v.amenities, r.overall_rating, r.visit_type
		ORDER BY frequency DESC`

	rows, err := databases.PostgresDB.Query(query, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var totalRating float64
	var ratingCount int
	priceRangeCount := make(map[string]int)
	amenitiesCount := make(map[string]int)

	for rows.Next() {
		var categoryID int64
		var priceRange, visitType string
		var amenitiesJSON []byte
		var rating float64
		var frequency int

		err := rows.Scan(&categoryID, &priceRange, &amenitiesJSON, &rating, &visitType, &frequency)
		if err != nil {
			continue
		}

		// Weight preferences by rating and frequency
		weight := rating * float64(frequency) / 5.0 // Normalize by max rating
		prefs.PreferredCategories[categoryID] += weight

		// Accumulate price range preferences
		if priceRange != "" {
			priceRangeCount[priceRange] += frequency
		}

		// Accumulate amenities preferences
		if amenitiesJSON != nil {
			var amenities []string
			if json.Unmarshal(amenitiesJSON, &amenities) == nil {
				for _, amenity := range amenities {
					amenitiesCount[amenity] += frequency
				}
			}
		}

		totalRating += rating * float64(frequency)
		ratingCount += frequency
	}

	// Calculate average rating given
	if ratingCount > 0 {
		prefs.AverageRatingGiven = totalRating / float64(ratingCount)
	}

	// Extract top price ranges
	for priceRange, count := range priceRangeCount {
		if count >= 2 { // Minimum threshold
			prefs.PreferredPriceRange = append(prefs.PreferredPriceRange, priceRange)
		}
	}

	// Extract top amenities
	for amenity, count := range amenitiesCount {
		if count >= 2 { // Minimum threshold
			prefs.PreferredAmenities = append(prefs.PreferredAmenities, amenity)
		}
	}

	return nil
}

// analyzeCheckinPreferences extracts location and timing preferences from check-ins
func (re *RecommendationEngine) analyzeCheckinPreferences(userID int64, prefs *UserPreferences) error {
	query := `
		SELECT v.latitude, v.longitude, EXTRACT(hour FROM c.created_at) as hour,
			   COUNT(*) as frequency
		FROM venue_checkins c
		JOIN venues v ON c.venue_id = v.id
		WHERE c.user_id = $1
		GROUP BY v.latitude, v.longitude, hour
		ORDER BY frequency DESC`

	rows, err := databases.PostgresDB.Query(query, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	locationClusters := make(map[string]*LocationPref)

	for rows.Next() {
		var lat, lng float64
		var hour int
		var frequency int

		err := rows.Scan(&lat, &lng, &hour, &frequency)
		if err != nil {
			continue
		}

		// Cluster nearby locations
		locationKey := re.getLocationCluster(lat, lng)
		if cluster, exists := locationClusters[locationKey]; exists {
			cluster.Weight += float64(frequency)
		} else {
			locationClusters[locationKey] = &LocationPref{
				Latitude:  lat,
				Longitude: lng,
				Weight:    float64(frequency),
				Radius:    2.0, // 2km radius
			}
		}

		// Track activity hours
		hourStr := formatHour(hour)
		prefs.ActivityHours[hourStr] += frequency
	}

	// Convert location clusters to slice
	for _, cluster := range locationClusters {
		if cluster.Weight >= 2 { // Minimum threshold
			prefs.PreferredLocations = append(prefs.PreferredLocations, *cluster)
		}
	}

	return nil
}

// analyzeSocialPreferences extracts preferences from social connections
func (re *RecommendationEngine) analyzeSocialPreferences(userID int64, prefs *UserPreferences) error {
	query := `
		SELECT following_id FROM user_follows WHERE follower_id = $1`

	rows, err := databases.PostgresDB.Query(query, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var followingID int64
		if rows.Scan(&followingID) == nil {
			prefs.SocialInfluence = append(prefs.SocialInfluence, followingID)
		}
	}

	return nil
}

// getCandidateVenues retrieves potential venues for recommendation
func (re *RecommendationEngine) getCandidateVenues(ctx RecommendationContext, prefs *UserPreferences) ([]models.Venue, error) {
	// Build query based on context and preferences
	baseQuery := `
		SELECT DISTINCT v.id, v.name, v.slug, v.description, v.short_description,
			   v.address, v.latitude, v.longitude, v.category_id, v.subcategory_id,
			   v.phone, v.website, v.price_range, v.average_rating, v.total_ratings,
			   v.cover_image, v.amenities, v.is_featured
		FROM venues v
		WHERE v.is_active = true AND v.average_rating >= 3.0`

	var conditions []string
	var args []interface{}
	argCount := 0

	// Location filter
	if ctx.UserLat != nil && ctx.UserLng != nil {
		argCount += 3
		conditions = append(conditions, `ST_DWithin(
			ST_Point(v.longitude, v.latitude)::geography,
			ST_Point($`+fmt.Sprintf("%d", argCount-2)+`, $`+fmt.Sprintf("%d", argCount-1)+`)::geography,
			$`+fmt.Sprintf("%d", argCount)+` * 1000)`)
		args = append(args, *ctx.UserLng, *ctx.UserLat, ctx.MaxDistance)
	}

	// Exclude venues user has already reviewed
	argCount++
	conditions = append(conditions, `v.id NOT IN (
		SELECT venue_id FROM venue_reviews WHERE user_id = $`+fmt.Sprintf("%d", argCount)+`)`)
	args = append(args, ctx.UserID)

	// Add preferred categories if available
	if len(prefs.PreferredCategories) > 0 {
		var categoryIDs []int64
		for categoryID := range prefs.PreferredCategories {
			categoryIDs = append(categoryIDs, categoryID)
		}
		if len(categoryIDs) > 0 {
			argCount++
			conditions = append(conditions, `(v.category_id = ANY($`+fmt.Sprintf("%d", argCount)+`) OR v.is_featured = true)`)
			args = append(args, categoryIDs)
		}
	}

	// Build final query
	finalQuery := baseQuery
	if len(conditions) > 0 {
		finalQuery += " AND " + strings.Join(conditions, " AND ")
	}
	finalQuery += " LIMIT 200" // Limit candidates for performance

	rows, err := databases.PostgresDB.Query(finalQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var venues []models.Venue
	for rows.Next() {
		var venue models.Venue
		var subcategoryID sql.NullInt64

		err := rows.Scan(
			&venue.ID, &venue.Name, &venue.Slug, &venue.Description, &venue.ShortDesc,
			&venue.Address, &venue.Latitude, &venue.Longitude, &venue.CategoryID, &subcategoryID,
			&venue.Phone, &venue.Website, &venue.PriceRange, &venue.AverageRating, &venue.TotalRatings,
			&venue.CoverImage, &venue.Amenities, &venue.IsFeatured,
		)

		if err != nil {
			continue
		}

		if subcategoryID.Valid {
			venue.SubcategoryID = &subcategoryID.Int64
		}

		venues = append(venues, venue)
	}

	return venues, nil
}

// calculateRecommendationScore calculates recommendation score for a venue
func (re *RecommendationEngine) calculateRecommendationScore(venue models.Venue, prefs *UserPreferences, ctx RecommendationContext) RecommendationScore {
	score := RecommendationScore{
		Venue:   venue,
		Reasons: make([]string, 0),
	}

	var totalScore float64 = 0

	// 1. Category preference score (30% weight)
	if categoryWeight, exists := prefs.PreferredCategories[venue.CategoryID]; exists {
		categoryScore := categoryWeight * 0.3
		totalScore += categoryScore
		score.Reasons = append(score.Reasons, "Matches your preferred category")
	}

	// 2. Rating quality score (25% weight)
	ratingScore := (venue.AverageRating / 5.0) * 0.25
	if venue.AverageRating >= 4.0 {
		ratingScore *= 1.2 // Boost highly rated venues
		score.Reasons = append(score.Reasons, "Highly rated venue")
	}
	totalScore += ratingScore

	// 3. Location preference score (20% weight)
	if ctx.UserLat != nil && ctx.UserLng != nil {
		distance := calculateDistance(*ctx.UserLat, *ctx.UserLng, venue.Latitude, venue.Longitude)
		locationScore := math.Max(0, (ctx.MaxDistance-distance)/ctx.MaxDistance) * 0.2
		totalScore += locationScore

		if distance <= 2.0 {
			score.Reasons = append(score.Reasons, "Close to your location")
		}

		// Check preferred locations
		for _, locPref := range prefs.PreferredLocations {
			prefDistance := calculateDistance(locPref.Latitude, locPref.Longitude, venue.Latitude, venue.Longitude)
			if prefDistance <= locPref.Radius {
				totalScore += (locPref.Weight / 10.0) * 0.1
				score.Reasons = append(score.Reasons, "In an area you frequent")
				break
			}
		}
	}

	// 4. Price range preference score (10% weight)
	if venue.PriceRange != "" {
		for _, prefPrice := range prefs.PreferredPriceRange {
			if venue.PriceRange == prefPrice {
				totalScore += 0.1
				score.Reasons = append(score.Reasons, "Matches your price preference")
				break
			}
		}
	}

	// 5. Amenities preference score (10% weight)
	if venue.Amenities != nil {
		var venueAmenities []string
		if json.Unmarshal(venue.Amenities, &venueAmenities) == nil {
			amenityMatches := 0
			for _, prefAmenity := range prefs.PreferredAmenities {
				for _, venueAmenity := range venueAmenities {
					if prefAmenity == venueAmenity {
						amenityMatches++
						break
					}
				}
			}
			if amenityMatches > 0 {
				amenityScore := (float64(amenityMatches) / float64(len(prefs.PreferredAmenities))) * 0.1
				totalScore += amenityScore
				score.Reasons = append(score.Reasons, "Has amenities you prefer")
			}
		}
	}

	// 6. Social influence score (5% weight)
	socialScore := re.calculateSocialScore(venue.ID, prefs.SocialInfluence)
	if socialScore > 0 {
		totalScore += socialScore * 0.05
		score.Reasons = append(score.Reasons, "Popular with people you follow")
	}

	// 7. Freshness and trending bonus
	if venue.IsFeatured {
		totalScore += 0.05
		score.Reasons = append(score.Reasons, "Featured venue")
	}

	// 8. Context-based scoring
	totalScore += re.calculateContextScore(venue, ctx)

	score.Score = math.Max(0, math.Min(1, totalScore)) // Normalize to 0-1
	return score
}

// calculateSocialScore calculates social influence score
func (re *RecommendationEngine) calculateSocialScore(venueID int64, followedUsers []int64) float64 {
	if len(followedUsers) == 0 {
		return 0
	}

	// Count positive reviews from followed users
	query := `
		SELECT COUNT(*) FROM venue_reviews 
		WHERE venue_id = $1 AND user_id = ANY($2) AND overall_rating >= 4.0`

	var positiveReviewCount int
	err := databases.PostgresDB.QueryRow(query, venueID, followedUsers).Scan(&positiveReviewCount)
	if err != nil {
		return 0
	}

	return float64(positiveReviewCount) / float64(len(followedUsers))
}

// calculateContextScore applies context-based scoring
func (re *RecommendationEngine) calculateContextScore(venue models.Venue, ctx RecommendationContext) float64 {
	var contextScore float64 = 0

	// Time-based scoring (simplified)
	if ctx.TimeOfDay == "evening" && venue.CategoryID == 2 { // Assuming 2 is bars/nightlife
		contextScore += 0.1
	} else if ctx.TimeOfDay == "morning" && venue.CategoryID == 3 { // Assuming 3 is cafes
		contextScore += 0.1
	}

	// Group size considerations
	if ctx.GroupSize > 4 && venue.PriceRange != "$$$$" {
		contextScore += 0.05 // Favor more affordable options for larger groups
	}

	return contextScore
}

// Helper functions

func calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	// Haversine formula for distance calculation
	const R = 6371 // Earth radius in kilometers

	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func (re *RecommendationEngine) getLocationCluster(lat, lng float64) string {
	// Simple clustering by rounding to 2 decimal places (~1km precision)
	return fmt.Sprintf("%.2f,%.2f", lat, lng)
}

func formatHour(hour int) string {
	if hour < 6 {
		return "night"
	} else if hour < 12 {
		return "morning"
	} else if hour < 17 {
		return "afternoon"
	} else {
		return "evening"
	}
}

// GetSimilarVenues finds venues similar to a given venue
func (re *RecommendationEngine) GetSimilarVenues(venueID int64, limit int) ([]models.Venue, error) {
	// Get the reference venue
	venue := &models.Venue{ID: venueID}
	err := venue.GetByID()
	if err != nil {
		return nil, err
	}

	// Find similar venues based on category, location, price range, and amenities
	query := `
		SELECT v.id, v.name, v.slug, v.short_description, v.address,
			   v.latitude, v.longitude, v.category_id, v.price_range,
			   v.average_rating, v.total_ratings, v.cover_image,
			   ST_Distance(
				   ST_Point(v.longitude, v.latitude)::geography,
				   ST_Point($2, $3)::geography
			   ) / 1000 as distance
		FROM venues v
		WHERE v.is_active = true 
		  AND v.id != $1
		  AND (v.category_id = $4 OR v.subcategory_id = $5)
		  AND ST_DWithin(
			  ST_Point(v.longitude, v.latitude)::geography,
			  ST_Point($2, $3)::geography,
			  50000
		  )
		ORDER BY 
		  CASE WHEN v.category_id = $4 THEN 1 ELSE 2 END,
		  CASE WHEN v.price_range = $6 THEN 1 ELSE 2 END,
		  v.average_rating DESC,
		  distance ASC
		LIMIT $7`

	rows, err := databases.PostgresDB.Query(
		query, venue.ID, venue.Longitude, venue.Latitude, venue.CategoryID,
		venue.SubcategoryID, venue.PriceRange, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var venues []models.Venue
	for rows.Next() {
		var v models.Venue
		var distance float64

		err := rows.Scan(
			&v.ID, &v.Name, &v.Slug, &v.ShortDesc, &v.Address,
			&v.Latitude, &v.Longitude, &v.CategoryID, &v.PriceRange,
			&v.AverageRating, &v.TotalRatings, &v.CoverImage, &distance,
		)

		if err != nil {
			continue
		}

		v.Distance = &distance
		venues = append(venues, v)
	}

	return venues, nil
}
