package models

import (
	databases "voting-app/app"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/getsentry/sentry-go"
	"time"
)

// Venue represents a restaurant, bar, cafe, or other venue
type Venue struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description string  `json:"description,omitempty"`
	ShortDesc   string  `json:"shortDescription,omitempty"`
	
	// Location
	Address    string  `json:"address"`
	CityID     int64   `json:"cityId"`
	City       *City   `json:"city,omitempty"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	PostalCode string  `json:"postalCode,omitempty"`
	
	// Category
	CategoryID    int64             `json:"categoryId"`
	Category      *VenueCategory    `json:"category,omitempty"`
	SubcategoryID *int64            `json:"subcategoryId,omitempty"`
	Subcategory   *VenueSubcategory `json:"subcategory,omitempty"`
	
	// Contact
	Phone   string `json:"phone,omitempty"`
	Email   string `json:"email,omitempty"`
	Website string `json:"website,omitempty"`
	
	// Business Hours
	OpeningHours json.RawMessage `json:"openingHours,omitempty"`
	
	// Pricing
	PriceRange       string  `json:"priceRange,omitempty"` // $, $$, $$$, $$$$
	AvgCostPerPerson float64 `json:"averageCostPerPerson,omitempty"`
	
	// Media
	CoverImage string `json:"coverImage,omitempty"`
	Logo       string `json:"logo,omitempty"`
	
	// Ratings (cached for performance)
	AverageRating float64 `json:"averageRating"`
	TotalRatings  int     `json:"totalRatings"`
	TotalReviews  int     `json:"totalReviews"`
	
	// Features
	Amenities json.RawMessage `json:"amenities,omitempty"`
	
	// Status
	IsActive    bool `json:"isActive"`
	IsVerified  bool `json:"isVerified"`
	IsFeatured  bool `json:"isFeatured"`
	
	// Owner
	OwnerID   *int64     `json:"ownerId,omitempty"`
	ClaimedAt *time.Time `json:"claimedAt,omitempty"`
	
	// Computed fields
	Distance      *float64 `json:"distance,omitempty"`      // Distance from user in km
	IsOpen        *bool    `json:"isOpen,omitempty"`        // Currently open
	NextOpenTime  *string  `json:"nextOpenTime,omitempty"`  // When it opens next
	ReviewSummary *string  `json:"reviewSummary,omitempty"` // AI-generated summary
	
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type VenueCategory struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
	IsActive    bool   `json:"isActive"`
}

type VenueSubcategory struct {
	ID          int64  `json:"id"`
	CategoryID  int64  `json:"categoryId"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IsActive    bool   `json:"isActive"`
}

type City struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	State     string  `json:"state,omitempty"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	Timezone  string  `json:"timezone,omitempty"`
}

// VenueSearchParams for advanced venue discovery
type VenueSearchParams struct {
	Query         string    `json:"query,omitempty"`
	CategoryID    *int64    `json:"categoryId,omitempty"`
	SubcategoryID *int64    `json:"subcategoryId,omitempty"`
	CityID        *int64    `json:"cityId,omitempty"`
	Latitude      *float64  `json:"latitude,omitempty"`
	Longitude     *float64  `json:"longitude,omitempty"`
	Radius        *float64  `json:"radius,omitempty"` // in km
	PriceRange    []string  `json:"priceRange,omitempty"`
	MinRating     *float64  `json:"minRating,omitempty"`
	Amenities     []string  `json:"amenities,omitempty"`
	IsOpen        *bool     `json:"isOpen,omitempty"`
	IsFeatured    *bool     `json:"isFeatured,omitempty"`
	SortBy        string    `json:"sortBy,omitempty"` // rating, distance, popularity, newest
	Page          int       `json:"page"`
	Limit         int       `json:"limit"`
}

func (v *Venue) TableName() string {
	return "venues"
}

// GetByID retrieves a venue by ID with all related data
func (v *Venue) GetByID() error {
	query := `
		SELECT v.id, v.name, v.slug, v.description, v.short_description,
			   v.address, v.city_id, v.latitude, v.longitude, v.postal_code,
			   v.category_id, v.subcategory_id, v.phone, v.email, v.website,
			   v.opening_hours, v.price_range, v.average_cost_per_person,
			   v.cover_image, v.logo, v.average_rating, v.total_ratings, v.total_reviews,
			   v.amenities, v.is_active, v.is_verified, v.is_featured,
			   v.owner_id, v.claimed_at, v.created_at, v.updated_at,
			   c.name as city_name, c.state, c.country,
			   cat.name as category_name, cat.icon as category_icon,
			   sub.name as subcategory_name
		FROM venues v
		LEFT JOIN cities c ON v.city_id = c.id
		LEFT JOIN venue_categories cat ON v.category_id = cat.id
		LEFT JOIN venue_subcategories sub ON v.subcategory_id = sub.id
		WHERE v.id = $1 AND v.is_active = true`
	
	row := databases.PostgresDB.QueryRow(query, v.ID)
	
	var subcategoryID sql.NullInt64
	var ownerID sql.NullInt64
	var claimedAt sql.NullTime
	var cityName, state, country, categoryName, categoryIcon, subcategoryName sql.NullString
	
	err := row.Scan(
		&v.ID, &v.Name, &v.Slug, &v.Description, &v.ShortDesc,
		&v.Address, &v.CityID, &v.Latitude, &v.Longitude, &v.PostalCode,
		&v.CategoryID, &subcategoryID, &v.Phone, &v.Email, &v.Website,
		&v.OpeningHours, &v.PriceRange, &v.AvgCostPerPerson,
		&v.CoverImage, &v.Logo, &v.AverageRating, &v.TotalRatings, &v.TotalReviews,
		&v.Amenities, &v.IsActive, &v.IsVerified, &v.IsFeatured,
		&ownerID, &claimedAt, &v.CreatedAt, &v.UpdatedAt,
		&cityName, &state, &country,
		&categoryName, &categoryIcon,
		&subcategoryName,
	)
	
	if err != nil {
		if err != sql.ErrNoRows {
			sentry.CaptureException(err)
		}
		return err
	}
	
	// Set optional fields
	if subcategoryID.Valid {
		v.SubcategoryID = &subcategoryID.Int64
	}
	if ownerID.Valid {
		v.OwnerID = &ownerID.Int64
	}
	if claimedAt.Valid {
		v.ClaimedAt = &claimedAt.Time
	}
	
	// Set related objects
	if cityName.Valid {
		v.City = &City{
			ID:      v.CityID,
			Name:    cityName.String,
			State:   state.String,
			Country: country.String,
		}
	}
	
	if categoryName.Valid {
		v.Category = &VenueCategory{
			ID:   v.CategoryID,
			Name: categoryName.String,
			Icon: categoryIcon.String,
		}
	}
	
	if subcategoryName.Valid && v.SubcategoryID != nil {
		v.Subcategory = &VenueSubcategory{
			ID:         *v.SubcategoryID,
			CategoryID: v.CategoryID,
			Name:       subcategoryName.String,
		}
	}
	
	return nil
}

// Search performs advanced venue search with filters and location
func (v *Venue) Search(params VenueSearchParams) ([]Venue, int, error) {
	// Build dynamic query based on search parameters
	baseQuery := `
		SELECT v.id, v.name, v.slug, v.short_description,
			   v.address, v.latitude, v.longitude,
			   v.category_id, v.phone, v.website,
			   v.price_range, v.average_rating, v.total_ratings,
			   v.cover_image, v.is_featured,
			   c.name as city_name,
			   cat.name as category_name, cat.icon as category_icon`
	
	var distanceSelect string
	if params.Latitude != nil && params.Longitude != nil {
		distanceSelect = fmt.Sprintf(`,
			ST_Distance(
				ST_Point(v.longitude, v.latitude)::geography,
				ST_Point(%f, %f)::geography
			) / 1000 as distance`, *params.Longitude, *params.Latitude)
	}
	
	fromClause := `
		FROM venues v
		LEFT JOIN cities c ON v.city_id = c.id
		LEFT JOIN venue_categories cat ON v.category_id = cat.id`
	
	whereClause := "WHERE v.is_active = true"
	var args []interface{}
	argCount := 0
	
	// Add search filters
	if params.Query != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND (v.name ILIKE $%d OR v.description ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+params.Query+"%")
	}
	
	if params.CategoryID != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND v.category_id = $%d", argCount)
		args = append(args, *params.CategoryID)
	}
	
	if params.CityID != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND v.city_id = $%d", argCount)
		args = append(args, *params.CityID)
	}
	
	if params.MinRating != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND v.average_rating >= $%d", argCount)
		args = append(args, *params.MinRating)
	}
	
	if len(params.PriceRange) > 0 {
		argCount++
		whereClause += fmt.Sprintf(" AND v.price_range = ANY($%d)", argCount)
		args = append(args, params.PriceRange)
	}
	
	// Location radius filter
	if params.Latitude != nil && params.Longitude != nil && params.Radius != nil {
		argCount += 3
		whereClause += fmt.Sprintf(` AND ST_DWithin(
			ST_Point(v.longitude, v.latitude)::geography,
			ST_Point($%d, $%d)::geography,
			$%d * 1000)`, argCount-2, argCount-1, argCount)
		args = append(args, *params.Longitude, *params.Latitude, *params.Radius)
	}
	
	if params.IsFeatured != nil && *params.IsFeatured {
		whereClause += " AND v.is_featured = true"
	}
	
	// Sorting
	var orderBy string
	switch params.SortBy {
	case "rating":
		orderBy = "ORDER BY v.average_rating DESC, v.total_ratings DESC"
	case "distance":
		if params.Latitude != nil && params.Longitude != nil {
			orderBy = "ORDER BY distance ASC"
		} else {
			orderBy = "ORDER BY v.average_rating DESC"
		}
	case "newest":
		orderBy = "ORDER BY v.created_at DESC"
	default:
		orderBy = "ORDER BY v.is_featured DESC, v.average_rating DESC, v.total_ratings DESC"
	}
	
	// Pagination
	if params.Limit == 0 {
		params.Limit = 20
	}
	if params.Page < 1 {
		params.Page = 1
	}
	offset := (params.Page - 1) * params.Limit
	
	limitClause := fmt.Sprintf(" LIMIT %d OFFSET %d", params.Limit, offset)
	
	// Execute query
	fullQuery := baseQuery + distanceSelect + fromClause + whereClause + orderBy + limitClause
	
	rows, err := databases.PostgresDB.Query(fullQuery, args...)
	if err != nil {
		sentry.CaptureException(err)
		return nil, 0, err
	}
	defer rows.Close()
	
	var venues []Venue
	for rows.Next() {
		var venue Venue
		var cityName, categoryName, categoryIcon sql.NullString
		var distance sql.NullFloat64
		
		scanArgs := []interface{}{
			&venue.ID, &venue.Name, &venue.Slug, &venue.ShortDesc,
			&venue.Address, &venue.Latitude, &venue.Longitude,
			&venue.CategoryID, &venue.Phone, &venue.Website,
			&venue.PriceRange, &venue.AverageRating, &venue.TotalRatings,
			&venue.CoverImage, &venue.IsFeatured,
			&cityName, &categoryName, &categoryIcon,
		}
		
		if distanceSelect != "" {
			scanArgs = append(scanArgs, &distance)
		}
		
		err := rows.Scan(scanArgs...)
		if err != nil {
			sentry.CaptureException(err)
			continue
		}
		
		// Set related data
		if cityName.Valid {
			venue.City = &City{Name: cityName.String}
		}
		if categoryName.Valid {
			venue.Category = &VenueCategory{
				ID:   venue.CategoryID,
				Name: categoryName.String,
				Icon: categoryIcon.String,
			}
		}
		if distance.Valid {
			venue.Distance = &distance.Float64
		}
		
		venues = append(venues, venue)
	}
	
	// Get total count for pagination
	countQuery := "SELECT COUNT(*) FROM venues v" + fromClause + whereClause
	var totalCount int
	err = databases.PostgresDB.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		sentry.CaptureException(err)
		return venues, 0, err
	}
	
	return venues, totalCount, nil
}

// GetNearby finds venues near a location
func (v *Venue) GetNearby(lat, lng, radius float64, limit int) ([]Venue, error) {
	params := VenueSearchParams{
		Latitude:  &lat,
		Longitude: &lng,
		Radius:    &radius,
		SortBy:    "distance",
		Limit:     limit,
		Page:      1,
	}
	
	venues, _, err := v.Search(params)
	return venues, err
}

// GetFeatured returns featured venues
func (v *Venue) GetFeatured(limit int) ([]Venue, error) {
	featured := true
	params := VenueSearchParams{
		IsFeatured: &featured,
		SortBy:     "rating",
		Limit:      limit,
		Page:       1,
	}
	
	venues, _, err := v.Search(params)
	return venues, err
}

// UpdateRatingCache updates the cached rating statistics
func (v *Venue) UpdateRatingCache() error {
	query := `
		UPDATE venues 
		SET average_rating = (
			SELECT COALESCE(AVG(overall_rating), 0) 
			FROM venue_reviews 
			WHERE venue_id = $1 AND moderation_status = 'approved'
		),
		total_ratings = (
			SELECT COUNT(*) 
			FROM venue_reviews 
			WHERE venue_id = $1 AND moderation_status = 'approved'
		),
		total_reviews = (
			SELECT COUNT(*) 
			FROM venue_reviews 
			WHERE venue_id = $1 AND moderation_status = 'approved' AND review_text IS NOT NULL
		),
		updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`
	
	_, err := databases.PostgresDB.Exec(query, v.ID)
	if err != nil {
		sentry.CaptureException(err)
	}
	return err
}

// Create creates a new venue
func (v *Venue) Create() error {
	query := `
		INSERT INTO venues (
			name, slug, description, short_description, address, city_id,
			latitude, longitude, postal_code, category_id, subcategory_id,
			phone, email, website, opening_hours, price_range, average_cost_per_person,
			cover_image, logo, amenities, owner_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21
		) RETURNING id, created_at, updated_at`
	
	err := databases.PostgresDB.QueryRow(
		query,
		v.Name, v.Slug, v.Description, v.ShortDesc, v.Address, v.CityID,
		v.Latitude, v.Longitude, v.PostalCode, v.CategoryID, v.SubcategoryID,
		v.Phone, v.Email, v.Website, v.OpeningHours, v.PriceRange, v.AvgCostPerPerson,
		v.CoverImage, v.Logo, v.Amenities, v.OwnerID,
	).Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt)
	
	if err != nil {
		sentry.CaptureException(err)
	}
	return err
}

// GetCategories returns all venue categories
func GetVenueCategories() ([]VenueCategory, error) {
	query := "SELECT id, name, description, icon, is_active FROM venue_categories WHERE is_active = true ORDER BY name"
	
	rows, err := databases.PostgresDB.Query(query)
	if err != nil {
		sentry.CaptureException(err)
		return nil, err
	}
	defer rows.Close()
	
	var categories []VenueCategory
	for rows.Next() {
		var cat VenueCategory
		var description, icon sql.NullString
		
		err := rows.Scan(&cat.ID, &cat.Name, &description, &icon, &cat.IsActive)
		if err != nil {
			sentry.CaptureException(err)
			continue
		}
		
		cat.Description = description.String
		cat.Icon = icon.String
		categories = append(categories, cat)
	}
	
	return categories, nil
}
