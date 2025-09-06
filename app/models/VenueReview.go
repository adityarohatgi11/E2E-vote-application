package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	databases "voting-app/app"

	"github.com/getsentry/sentry-go"
)

// VenueReview represents a detailed review of a venue
type VenueReview struct {
	ID      int64 `json:"id"`
	VenueID int64 `json:"venueId"`
	UserID  int64 `json:"userId"`

	// Rating Information
	OverallRating   float64         `json:"overallRating"`
	DetailedRatings json.RawMessage `json:"detailedRatings,omitempty"` // {"food": 4.5, "service": 4.0, "ambiance": 5.0}

	// Review Content
	Title      string `json:"title,omitempty"`
	ReviewText string `json:"reviewText,omitempty"`

	// Visit Information
	VisitDate *time.Time `json:"visitDate,omitempty"`
	VisitType string     `json:"visitType,omitempty"` // dinner, lunch, drinks, event
	PartySize int        `json:"partySize,omitempty"`

	// Media
	Photos json.RawMessage `json:"photos,omitempty"` // Array of photo URLs

	// Moderation
	IsVerified       bool   `json:"isVerified"`
	IsFeatured       bool   `json:"isFeatured"`
	IsFlagged        bool   `json:"isFlagged"`
	ModerationStatus string `json:"moderationStatus"` // pending, approved, rejected

	// Engagement
	HelpfulVotes   int `json:"helpfulVotes"`
	UnhelpfulVotes int `json:"unhelpfulVotes"`

	// User Information (joined)
	User     *SnappUser `json:"user,omitempty"`
	UserName string     `json:"userName,omitempty"`

	// Venue Information (joined)
	Venue     *Venue `json:"venue,omitempty"`
	VenueName string `json:"venueName,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ReviewSummary for analytics and display
type ReviewSummary struct {
	VenueID         int64              `json:"venueId"`
	AverageRating   float64            `json:"averageRating"`
	TotalReviews    int                `json:"totalReviews"`
	RatingBreakdown map[string]int     `json:"ratingBreakdown"` // {"5": 10, "4": 5, "3": 2, "2": 1, "1": 0}
	DetailedAverage map[string]float64 `json:"detailedAverage"` // {"food": 4.2, "service": 4.1, ...}
	RecentReviews   []VenueReview      `json:"recentReviews"`
	TopReviews      []VenueReview      `json:"topReviews"` // Most helpful reviews
}

// ReviewFilters for searching and filtering reviews
type ReviewFilters struct {
	VenueID    *int64     `json:"venueId,omitempty"`
	UserID     *int64     `json:"userId,omitempty"`
	MinRating  *float64   `json:"minRating,omitempty"`
	MaxRating  *float64   `json:"maxRating,omitempty"`
	VisitType  string     `json:"visitType,omitempty"`
	HasPhotos  *bool      `json:"hasPhotos,omitempty"`
	IsFeatured *bool      `json:"isFeatured,omitempty"`
	DateFrom   *time.Time `json:"dateFrom,omitempty"`
	DateTo     *time.Time `json:"dateTo,omitempty"`
	SortBy     string     `json:"sortBy,omitempty"` // newest, oldest, rating_high, rating_low, helpful
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
}

func (r *VenueReview) TableName() string {
	return "venue_reviews"
}

// Create creates a new review
func (r *VenueReview) Create() error {
	// Check if user has already reviewed this venue
	var existingCount int
	err := databases.PostgresDB.QueryRow(
		"SELECT COUNT(*) FROM venue_reviews WHERE venue_id = $1 AND user_id = $2",
		r.VenueID, r.UserID,
	).Scan(&existingCount)

	if err != nil {
		sentry.CaptureException(err)
		return err
	}

	if existingCount > 0 {
		return fmt.Errorf("user has already reviewed this venue")
	}

	// Validate rating range
	if r.OverallRating < 1.0 || r.OverallRating > 5.0 {
		return fmt.Errorf("rating must be between 1.0 and 5.0")
	}

	// Insert new review
	query := `
		INSERT INTO venue_reviews (
			venue_id, user_id, overall_rating, detailed_ratings,
			title, review_text, visit_date, visit_type, party_size,
			photos, moderation_status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at`

	err = databases.PostgresDB.QueryRow(
		query,
		r.VenueID, r.UserID, r.OverallRating, r.DetailedRatings,
		r.Title, r.ReviewText, r.VisitDate, r.VisitType, r.PartySize,
		r.Photos, "pending",
	).Scan(&r.ID, &r.CreatedAt, &r.UpdatedAt)

	if err != nil {
		sentry.CaptureException(err)
		return err
	}

	// Update venue rating cache
	venue := &Venue{ID: r.VenueID}
	go venue.UpdateRatingCache() // Update in background

	return nil
}

// GetByID retrieves a review by ID with user and venue info
func (r *VenueReview) GetByID() error {
	query := `
		SELECT r.id, r.venue_id, r.user_id, r.overall_rating, r.detailed_ratings,
			   r.title, r.review_text, r.visit_date, r.visit_type, r.party_size,
			   r.photos, r.is_verified, r.is_featured, r.is_flagged, r.moderation_status,
			   r.helpful_votes, r.unhelpful_votes, r.created_at, r.updated_at,
			   v.name as venue_name,
			   u.snapp_id as user_snapp_id
		FROM venue_reviews r
		LEFT JOIN venues v ON r.venue_id = v.id
		LEFT JOIN snapp_users u ON r.user_id = u.id
		WHERE r.id = $1`

	row := databases.PostgresDB.QueryRow(query, r.ID)

	var visitDate sql.NullTime
	var userSnapID sql.NullString

	err := row.Scan(
		&r.ID, &r.VenueID, &r.UserID, &r.OverallRating, &r.DetailedRatings,
		&r.Title, &r.ReviewText, &visitDate, &r.VisitType, &r.PartySize,
		&r.Photos, &r.IsVerified, &r.IsFeatured, &r.IsFlagged, &r.ModerationStatus,
		&r.HelpfulVotes, &r.UnhelpfulVotes, &r.CreatedAt, &r.UpdatedAt,
		&r.VenueName, &userSnapID,
	)

	if err != nil {
		if err != sql.ErrNoRows {
			sentry.CaptureException(err)
		}
		return err
	}

	if visitDate.Valid {
		r.VisitDate = &visitDate.Time
	}

	// Set user name (you might want to get from user profile instead)
	if userSnapID.Valid {
		r.UserName = userSnapID.String // Or get display name from user service
	}

	return nil
}

// Search finds reviews based on filters
func (r *VenueReview) Search(filters ReviewFilters) ([]VenueReview, int, error) {
	baseQuery := `
		SELECT r.id, r.venue_id, r.user_id, r.overall_rating, r.detailed_ratings,
			   r.title, r.review_text, r.visit_date, r.visit_type, r.party_size,
			   r.photos, r.is_verified, r.is_featured, r.helpful_votes, r.unhelpful_votes,
			   r.created_at, r.updated_at,
			   v.name as venue_name,
			   u.snapp_id as user_snapp_id
		FROM venue_reviews r
		LEFT JOIN venues v ON r.venue_id = v.id
		LEFT JOIN snapp_users u ON r.user_id = u.id`

	whereClause := "WHERE r.moderation_status = 'approved'"
	var args []interface{}
	argCount := 0

	// Add filters
	if filters.VenueID != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND r.venue_id = $%d", argCount)
		args = append(args, *filters.VenueID)
	}

	if filters.UserID != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND r.user_id = $%d", argCount)
		args = append(args, *filters.UserID)
	}

	if filters.MinRating != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND r.overall_rating >= $%d", argCount)
		args = append(args, *filters.MinRating)
	}

	if filters.MaxRating != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND r.overall_rating <= $%d", argCount)
		args = append(args, *filters.MaxRating)
	}

	if filters.VisitType != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND r.visit_type = $%d", argCount)
		args = append(args, filters.VisitType)
	}

	if filters.HasPhotos != nil && *filters.HasPhotos {
		whereClause += " AND r.photos IS NOT NULL AND jsonb_array_length(r.photos) > 0"
	}

	if filters.IsFeatured != nil {
		if *filters.IsFeatured {
			whereClause += " AND r.is_featured = true"
		} else {
			whereClause += " AND r.is_featured = false"
		}
	}

	if filters.DateFrom != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND r.created_at >= $%d", argCount)
		args = append(args, *filters.DateFrom)
	}

	if filters.DateTo != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND r.created_at <= $%d", argCount)
		args = append(args, *filters.DateTo)
	}

	// Sorting
	var orderBy string
	switch filters.SortBy {
	case "oldest":
		orderBy = "ORDER BY r.created_at ASC"
	case "rating_high":
		orderBy = "ORDER BY r.overall_rating DESC, r.created_at DESC"
	case "rating_low":
		orderBy = "ORDER BY r.overall_rating ASC, r.created_at DESC"
	case "helpful":
		orderBy = "ORDER BY r.helpful_votes DESC, r.created_at DESC"
	default: // newest
		orderBy = "ORDER BY r.created_at DESC"
	}

	// Pagination
	if filters.Limit == 0 {
		filters.Limit = 20
	}
	if filters.Page < 1 {
		filters.Page = 1
	}
	offset := (filters.Page - 1) * filters.Limit

	limitClause := fmt.Sprintf(" LIMIT %d OFFSET %d", filters.Limit, offset)

	// Execute query
	fullQuery := baseQuery + " " + whereClause + " " + orderBy + limitClause

	rows, err := databases.PostgresDB.Query(fullQuery, args...)
	if err != nil {
		sentry.CaptureException(err)
		return nil, 0, err
	}
	defer rows.Close()

	var reviews []VenueReview
	for rows.Next() {
		var review VenueReview
		var visitDate sql.NullTime
		var userSnapID sql.NullString

		err := rows.Scan(
			&review.ID, &review.VenueID, &review.UserID, &review.OverallRating, &review.DetailedRatings,
			&review.Title, &review.ReviewText, &visitDate, &review.VisitType, &review.PartySize,
			&review.Photos, &review.IsVerified, &review.IsFeatured, &review.HelpfulVotes, &review.UnhelpfulVotes,
			&review.CreatedAt, &review.UpdatedAt,
			&review.VenueName, &userSnapID,
		)

		if err != nil {
			sentry.CaptureException(err)
			continue
		}

		if visitDate.Valid {
			review.VisitDate = &visitDate.Time
		}
		if userSnapID.Valid {
			review.UserName = userSnapID.String
		}

		reviews = append(reviews, review)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM venue_reviews r " + whereClause
	var totalCount int
	err = databases.PostgresDB.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		sentry.CaptureException(err)
		return reviews, 0, err
	}

	return reviews, totalCount, nil
}

// GetVenueReviewSummary returns comprehensive review statistics for a venue
func GetVenueReviewSummary(venueID int64) (*ReviewSummary, error) {
	summary := &ReviewSummary{
		VenueID:         venueID,
		RatingBreakdown: make(map[string]int),
		DetailedAverage: make(map[string]float64),
	}

	// Get basic statistics
	basicQuery := `
		SELECT 
			COALESCE(AVG(overall_rating), 0) as avg_rating,
			COUNT(*) as total_reviews,
			COUNT(CASE WHEN overall_rating >= 4.5 THEN 1 END) as rating_5,
			COUNT(CASE WHEN overall_rating >= 3.5 AND overall_rating < 4.5 THEN 1 END) as rating_4,
			COUNT(CASE WHEN overall_rating >= 2.5 AND overall_rating < 3.5 THEN 1 END) as rating_3,
			COUNT(CASE WHEN overall_rating >= 1.5 AND overall_rating < 2.5 THEN 1 END) as rating_2,
			COUNT(CASE WHEN overall_rating < 1.5 THEN 1 END) as rating_1
		FROM venue_reviews 
		WHERE venue_id = $1 AND moderation_status = 'approved'`

	var rating5, rating4, rating3, rating2, rating1 int
	err := databases.PostgresDB.QueryRow(basicQuery, venueID).Scan(
		&summary.AverageRating, &summary.TotalReviews,
		&rating5, &rating4, &rating3, &rating2, &rating1,
	)

	if err != nil {
		sentry.CaptureException(err)
		return nil, err
	}

	summary.RatingBreakdown["5"] = rating5
	summary.RatingBreakdown["4"] = rating4
	summary.RatingBreakdown["3"] = rating3
	summary.RatingBreakdown["2"] = rating2
	summary.RatingBreakdown["1"] = rating1

	// Get recent reviews
	recentFilters := ReviewFilters{
		VenueID: &venueID,
		SortBy:  "newest",
		Limit:   5,
		Page:    1,
	}

	review := &VenueReview{}
	recentReviews, _, err := review.Search(recentFilters)
	if err == nil {
		summary.RecentReviews = recentReviews
	}

	// Get top helpful reviews
	topFilters := ReviewFilters{
		VenueID: &venueID,
		SortBy:  "helpful",
		Limit:   3,
		Page:    1,
	}

	topReviews, _, err := review.Search(topFilters)
	if err == nil {
		summary.TopReviews = topReviews
	}

	return summary, nil
}

// VoteHelpful marks a review as helpful/unhelpful
func (r *VenueReview) VoteHelpful(userID int64, isHelpful bool) error {
	// Check if user has already voted
	var existingVote bool
	err := databases.PostgresDB.QueryRow(
		"SELECT is_helpful FROM review_votes WHERE review_id = $1 AND user_id = $2",
		r.ID, userID,
	).Scan(&existingVote)

	if err == nil {
		// Update existing vote
		if existingVote != isHelpful {
			_, err = databases.PostgresDB.Exec(
				"UPDATE review_votes SET is_helpful = $1 WHERE review_id = $2 AND user_id = $3",
				isHelpful, r.ID, userID,
			)
		}
	} else if err == sql.ErrNoRows {
		// Create new vote
		_, err = databases.PostgresDB.Exec(
			"INSERT INTO review_votes (review_id, user_id, is_helpful) VALUES ($1, $2, $3)",
			r.ID, userID, isHelpful,
		)
	} else {
		sentry.CaptureException(err)
		return err
	}

	if err != nil {
		sentry.CaptureException(err)
		return err
	}

	// Update review vote counts
	_, err = databases.PostgresDB.Exec(`
		UPDATE venue_reviews SET 
			helpful_votes = (SELECT COUNT(*) FROM review_votes WHERE review_id = $1 AND is_helpful = true),
			unhelpful_votes = (SELECT COUNT(*) FROM review_votes WHERE review_id = $1 AND is_helpful = false)
		WHERE id = $1`,
		r.ID,
	)

	if err != nil {
		sentry.CaptureException(err)
	}

	return err
}

// ApproveReview approves a review for display
func (r *VenueReview) ApproveReview() error {
	_, err := databases.PostgresDB.Exec(
		"UPDATE venue_reviews SET moderation_status = 'approved', updated_at = CURRENT_TIMESTAMP WHERE id = $1",
		r.ID,
	)

	if err != nil {
		sentry.CaptureException(err)
		return err
	}

	// Update venue rating cache
	venue := &Venue{ID: r.VenueID}
	go venue.UpdateRatingCache()

	return nil
}
