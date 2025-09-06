package serializers

import (
	"encoding/json"
	"strings"
	"time"
	"voting-app/app/models"
)

// VenueSearchResponse for venue search API
type VenueSearchResponse struct {
	Venues       []models.Venue           `json:"venues"`
	Pagination   PaginationInfo           `json:"pagination"`
	SearchParams models.VenueSearchParams `json:"searchParams"`
	Suggestions  []string                 `json:"suggestions,omitempty"` // Search suggestions
	Filters      VenueFilterOptions       `json:"filters"`               // Available filter options
}

// VenueDetailResponse for detailed venue information
type VenueDetailResponse struct {
	Venue         models.Venue          `json:"venue"`
	ReviewSummary *models.ReviewSummary `json:"reviewSummary"`
	// Events        []models.VenueEvent   `json:"events"`
	SimilarVenues []models.Venue `json:"similarVenues,omitempty"`
	CheckinCount  int            `json:"checkinCount,omitempty"`
}

// VenueFilterOptions provides available filter options for search
type VenueFilterOptions struct {
	Categories    []models.VenueCategory    `json:"categories"`
	Subcategories []models.VenueSubcategory `json:"subcategories"`
	PriceRanges   []string                  `json:"priceRanges"`
	Amenities     []string                  `json:"amenities"`
	Cities        []models.City             `json:"cities"`
}

// CreateVenueRequest for creating new venues
type CreateVenueRequest struct {
	Name             string          `json:"name" binding:"required,min=1,max=255"`
	Description      string          `json:"description,omitempty"`
	ShortDescription string          `json:"shortDescription,omitempty"`
	Address          string          `json:"address" binding:"required"`
	CityID           int64           `json:"cityId" binding:"required"`
	Latitude         float64         `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude        float64         `json:"longitude" binding:"required,min=-180,max=180"`
	PostalCode       string          `json:"postalCode,omitempty"`
	CategoryID       int64           `json:"categoryId" binding:"required"`
	SubcategoryID    *int64          `json:"subcategoryId,omitempty"`
	Phone            string          `json:"phone,omitempty"`
	Email            string          `json:"email,omitempty"`
	Website          string          `json:"website,omitempty"`
	OpeningHours     json.RawMessage `json:"openingHours,omitempty"`
	PriceRange       string          `json:"priceRange,omitempty"`
	AvgCostPerPerson float64         `json:"averageCostPerPerson,omitempty"`
	CoverImage       string          `json:"coverImage,omitempty"`
	Logo             string          `json:"logo,omitempty"`
	Amenities        []string        `json:"amenities,omitempty"`
}

// UpdateVenueRequest for updating venues
type UpdateVenueRequest struct {
	Name             *string         `json:"name,omitempty"`
	Description      *string         `json:"description,omitempty"`
	ShortDescription *string         `json:"shortDescription,omitempty"`
	Address          *string         `json:"address,omitempty"`
	Phone            *string         `json:"phone,omitempty"`
	Email            *string         `json:"email,omitempty"`
	Website          *string         `json:"website,omitempty"`
	OpeningHours     json.RawMessage `json:"openingHours,omitempty"`
	PriceRange       *string         `json:"priceRange,omitempty"`
	AvgCostPerPerson *float64        `json:"averageCostPerPerson,omitempty"`
	CoverImage       *string         `json:"coverImage,omitempty"`
	Logo             *string         `json:"logo,omitempty"`
	Amenities        []string        `json:"amenities,omitempty"`
}

// Validate validates the CreateVenueRequest
func (r *CreateVenueRequest) Validate() (Base, bool) {
	if strings.TrimSpace(r.Name) == "" {
		return Base{
			Code:    InvalidInput,
			Message: "Venue name is required",
		}, false
	}

	if strings.TrimSpace(r.Address) == "" {
		return Base{
			Code:    InvalidInput,
			Message: "Address is required",
		}, false
	}

	if r.CityID <= 0 {
		return Base{
			Code:    InvalidInput,
			Message: "Valid city ID is required",
		}, false
	}

	if r.CategoryID <= 0 {
		return Base{
			Code:    InvalidInput,
			Message: "Valid category ID is required",
		}, false
	}

	if r.Latitude < -90 || r.Latitude > 90 {
		return Base{
			Code:    InvalidInput,
			Message: "Latitude must be between -90 and 90",
		}, false
	}

	if r.Longitude < -180 || r.Longitude > 180 {
		return Base{
			Code:    InvalidInput,
			Message: "Longitude must be between -180 and 180",
		}, false
	}

	// Validate price range if provided
	if r.PriceRange != "" {
		validPriceRanges := []string{"$", "$$", "$$$", "$$$$"}
		isValid := false
		for _, valid := range validPriceRanges {
			if r.PriceRange == valid {
				isValid = true
				break
			}
		}
		if !isValid {
			return Base{
				Code:    InvalidInput,
				Message: "Price range must be one of: $, $$, $$$, $$$$",
			}, false
		}
	}

	return Base{}, true
}

// ToVenue converts CreateVenueRequest to Venue model
func (r *CreateVenueRequest) ToVenue() *models.Venue {
	venue := &models.Venue{
		Name:             r.Name,
		Slug:             generateSlug(r.Name), // You'd implement this function
		Description:      r.Description,
		ShortDesc:        r.ShortDescription,
		Address:          r.Address,
		CityID:           r.CityID,
		Latitude:         r.Latitude,
		Longitude:        r.Longitude,
		PostalCode:       r.PostalCode,
		CategoryID:       r.CategoryID,
		SubcategoryID:    r.SubcategoryID,
		Phone:            r.Phone,
		Email:            r.Email,
		Website:          r.Website,
		OpeningHours:     r.OpeningHours,
		PriceRange:       r.PriceRange,
		AvgCostPerPerson: r.AvgCostPerPerson,
		CoverImage:       r.CoverImage,
		Logo:             r.Logo,
		IsActive:         true,
		IsVerified:       false,
		IsFeatured:       false,
	}

	// Convert amenities slice to JSON
	if len(r.Amenities) > 0 {
		amenitiesJSON, _ := json.Marshal(r.Amenities)
		venue.Amenities = amenitiesJSON
	}

	return venue
}

// PaginationInfo for paginated responses
type PaginationInfo struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"totalPages"`
	HasNext    bool `json:"hasNext"`
	HasPrev    bool `json:"hasPrev"`
}

// ReviewSearchResponse for review search results
type ReviewSearchResponse struct {
	Reviews    []models.VenueReview `json:"reviews"`
	Pagination PaginationInfo       `json:"pagination"`
	Filters    models.ReviewFilters `json:"filters"`
}

// CreateReviewRequest for creating new reviews
type CreateReviewRequest struct {
	VenueID         int64           `json:"venueId" binding:"required"`
	OverallRating   float64         `json:"overallRating" binding:"required,min=1,max=5"`
	DetailedRatings json.RawMessage `json:"detailedRatings,omitempty"`
	Title           string          `json:"title,omitempty"`
	ReviewText      string          `json:"reviewText,omitempty"`
	VisitDate       *time.Time      `json:"visitDate,omitempty"`
	VisitType       string          `json:"visitType,omitempty"`
	PartySize       int             `json:"partySize,omitempty"`
	Photos          []string        `json:"photos,omitempty"`
}

// UpdateReviewRequest for updating reviews
type UpdateReviewRequest struct {
	OverallRating   *float64        `json:"overallRating,omitempty"`
	DetailedRatings json.RawMessage `json:"detailedRatings,omitempty"`
	Title           *string         `json:"title,omitempty"`
	ReviewText      *string         `json:"reviewText,omitempty"`
	VisitDate       *time.Time      `json:"visitDate,omitempty"`
	VisitType       *string         `json:"visitType,omitempty"`
	PartySize       *int            `json:"partySize,omitempty"`
	Photos          []string        `json:"photos,omitempty"`
}

// ReviewVoteRequest for voting on review helpfulness
type ReviewVoteRequest struct {
	IsHelpful bool `json:"isHelpful" binding:"required"`
}

// Validate validates the CreateReviewRequest
func (r *CreateReviewRequest) Validate() (Base, bool) {
	if r.VenueID <= 0 {
		return Base{
			Code:    InvalidInput,
			Message: "Valid venue ID is required",
		}, false
	}

	if r.OverallRating < 1.0 || r.OverallRating > 5.0 {
		return Base{
			Code:    InvalidInput,
			Message: "Rating must be between 1.0 and 5.0",
		}, false
	}

	// Validate visit type if provided
	if r.VisitType != "" {
		validVisitTypes := []string{"breakfast", "lunch", "dinner", "drinks", "coffee", "event", "takeout"}
		isValid := false
		for _, valid := range validVisitTypes {
			if r.VisitType == valid {
				isValid = true
				break
			}
		}
		if !isValid {
			return Base{
				Code:    InvalidInput,
				Message: "Invalid visit type",
			}, false
		}
	}

	if r.PartySize < 0 || r.PartySize > 50 {
		return Base{
			Code:    InvalidInput,
			Message: "Party size must be between 0 and 50",
		}, false
	}

	return Base{}, true
}

// Validate validates the UpdateReviewRequest
func (r *UpdateReviewRequest) Validate() (Base, bool) {
	if r.OverallRating != nil {
		if *r.OverallRating < 1.0 || *r.OverallRating > 5.0 {
			return Base{
				Code:    InvalidInput,
				Message: "Rating must be between 1.0 and 5.0",
			}, false
		}
	}

	if r.PartySize != nil {
		if *r.PartySize < 0 || *r.PartySize > 50 {
			return Base{
				Code:    InvalidInput,
				Message: "Party size must be between 0 and 50",
			}, false
		}
	}

	return Base{}, true
}

// ToReview converts CreateReviewRequest to VenueReview model
func (r *CreateReviewRequest) ToReview() *models.VenueReview {
	review := &models.VenueReview{
		VenueID:          r.VenueID,
		OverallRating:    r.OverallRating,
		DetailedRatings:  r.DetailedRatings,
		Title:            r.Title,
		ReviewText:       r.ReviewText,
		VisitDate:        r.VisitDate,
		VisitType:        r.VisitType,
		PartySize:        r.PartySize,
		ModerationStatus: "pending",
	}

	// Convert photos slice to JSON
	if len(r.Photos) > 0 {
		photosJSON, _ := json.Marshal(r.Photos)
		review.Photos = photosJSON
	}

	return review
}

// VenueCollectionResponse for venue collections/lists
type VenueCollectionResponse struct {
	// Collections []models.VenueCollection `json:"collections"`
	Pagination PaginationInfo `json:"pagination"`
}

// CreateCollectionRequest for creating venue collections
type CreateCollectionRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description,omitempty"`
	IsPublic    bool   `json:"isPublic"`
	CoverImage  string `json:"coverImage,omitempty"`
}

// AddVenueToCollectionRequest for adding venues to collections
type AddVenueToCollectionRequest struct {
	VenueID int64  `json:"venueId" binding:"required"`
	Note    string `json:"note,omitempty"`
}

// VotingCampaignResponse for voting campaigns
type VotingCampaignResponse struct {
	// Campaigns  []models.VotingCampaign `json:"campaigns"`
	Pagination PaginationInfo `json:"pagination"`
}

// CreateCampaignRequest for creating voting campaigns
type CreateCampaignRequest struct {
	Title                   string    `json:"title" binding:"required,min=1,max=255"`
	Description             string    `json:"description,omitempty"`
	CampaignType            string    `json:"campaignType" binding:"required"`
	CityID                  *int64    `json:"cityId,omitempty"`
	CategoryID              *int64    `json:"categoryId,omitempty"`
	StartDate               time.Time `json:"startDate" binding:"required"`
	EndDate                 time.Time `json:"endDate" binding:"required"`
	MaxVotesPerUser         int       `json:"maxVotesPerUser"`
	AllowMultipleCategories bool      `json:"allowMultipleCategories"`
	RequireReview           bool      `json:"requireReview"`
}

// SubmitCampaignVoteRequest for voting in campaigns
type SubmitCampaignVoteRequest struct {
	VenueID         int64   `json:"venueId" binding:"required"`
	Reason          string  `json:"reason,omitempty"`
	ConfidenceScore float64 `json:"confidenceScore,omitempty"`
}

// Helper function to generate URL-friendly slugs
func generateSlug(name string) string {
	// Simple slug generation - in production, use a proper library
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "&", "and")
	// Remove special characters (basic implementation)
	allowedChars := "abcdefghijklmnopqrstuvwxyz0123456789-"
	var result strings.Builder
	for _, char := range slug {
		if strings.ContainsRune(allowedChars, char) {
			result.WriteRune(char)
		}
	}
	return result.String()
}
