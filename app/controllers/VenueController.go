package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"
	"voting-app/app/models"
	"voting-app/app/serializers"

	"github.com/gin-gonic/gin"
)

type VenueController struct{}

// Search performs advanced venue search
// @Summary      Search venues with filters
// @Tags         venues
// @Accept       json
// @Produce      json
// @Param        q              query     string  false  "Search query"
// @Param        category       query     int     false  "Category ID"
// @Param        subcategory    query     int     false  "Subcategory ID"
// @Param        city           query     int     false  "City ID"
// @Param        lat            query     number  false  "Latitude for location search"
// @Param        lng            query     number  false  "Longitude for location search"
// @Param        radius         query     number  false  "Search radius in km (default 10)"
// @Param        price_range    query     string  false  "Price ranges (comma separated: $,$$,$$$,$$$$)"
// @Param        min_rating     query     number  false  "Minimum rating (1-5)"
// @Param        amenities      query     string  false  "Required amenities (comma separated)"
// @Param        is_open        query     boolean false  "Currently open venues only"
// @Param        is_featured    query     boolean false  "Featured venues only"
// @Param        sort_by        query     string  false  "Sort by: rating, distance, popularity, newest"
// @Param        page           query     int     false  "Page number (default 1)"
// @Param        limit          query     int     false  "Results per page (default 20, max 100)"
// @Success      200  {object}  serializers.VenueSearchResponse
// @Failure      400  {object}  serializers.Base
// @Router       /venues/search [get]
func (VenueController) Search(ctx *gin.Context) {
	// Parse search parameters
	params := models.VenueSearchParams{
		Query:  ctx.Query("q"),
		SortBy: ctx.DefaultQuery("sort_by", "rating"),
		Page:   1,
		Limit:  20,
	}

	// Parse numeric parameters
	if categoryStr := ctx.Query("category"); categoryStr != "" {
		if categoryID, err := strconv.ParseInt(categoryStr, 10, 64); err == nil {
			params.CategoryID = &categoryID
		}
	}

	if subcategoryStr := ctx.Query("subcategory"); subcategoryStr != "" {
		if subcategoryID, err := strconv.ParseInt(subcategoryStr, 10, 64); err == nil {
			params.SubcategoryID = &subcategoryID
		}
	}

	if cityStr := ctx.Query("city"); cityStr != "" {
		if cityID, err := strconv.ParseInt(cityStr, 10, 64); err == nil {
			params.CityID = &cityID
		}
	}

	// Parse location parameters
	if latStr := ctx.Query("lat"); latStr != "" {
		if lat, err := strconv.ParseFloat(latStr, 64); err == nil {
			params.Latitude = &lat
		}
	}

	if lngStr := ctx.Query("lng"); lngStr != "" {
		if lng, err := strconv.ParseFloat(lngStr, 64); err == nil {
			params.Longitude = &lng
		}
	}

	if radiusStr := ctx.Query("radius"); radiusStr != "" {
		if radius, err := strconv.ParseFloat(radiusStr, 64); err == nil {
			params.Radius = &radius
		}
	} else if params.Latitude != nil && params.Longitude != nil {
		// Default radius of 10km for location searches
		defaultRadius := 10.0
		params.Radius = &defaultRadius
	}

	// Parse price range
	if priceRangeStr := ctx.Query("price_range"); priceRangeStr != "" {
		params.PriceRange = strings.Split(priceRangeStr, ",")
	}

	// Parse minimum rating
	if minRatingStr := ctx.Query("min_rating"); minRatingStr != "" {
		if minRating, err := strconv.ParseFloat(minRatingStr, 64); err == nil {
			params.MinRating = &minRating
		}
	}

	// Parse amenities
	if amenitiesStr := ctx.Query("amenities"); amenitiesStr != "" {
		params.Amenities = strings.Split(amenitiesStr, ",")
	}

	// Parse boolean flags
	if isOpenStr := ctx.Query("is_open"); isOpenStr != "" {
		if isOpen, err := strconv.ParseBool(isOpenStr); err == nil {
			params.IsOpen = &isOpen
		}
	}

	if isFeaturedStr := ctx.Query("is_featured"); isFeaturedStr != "" {
		if isFeatured, err := strconv.ParseBool(isFeaturedStr); err == nil {
			params.IsFeatured = &isFeatured
		}
	}

	// Parse pagination
	if pageStr := ctx.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			params.Page = page
		}
	}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			params.Limit = limit
		}
	}

	// Perform search
	venue := &models.Venue{}
	venues, totalCount, err := venue.Search(params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, serializers.Base{
			Code:    serializers.InternalError,
			Message: "Failed to search venues",
		})
		return
	}

	// Calculate pagination info
	totalPages := (totalCount + params.Limit - 1) / params.Limit
	hasNext := params.Page < totalPages
	hasPrev := params.Page > 1

	response := serializers.VenueSearchResponse{
		Venues: venues,
		Pagination: serializers.PaginationInfo{
			Page:       params.Page,
			Limit:      params.Limit,
			Total:      totalCount,
			TotalPages: totalPages,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
		},
		SearchParams: params,
	}

	ctx.JSON(http.StatusOK, response)
}

// GetByID retrieves a venue by ID with all details
// @Summary      Get venue details
// @Tags         venues
// @Produce      json
// @Param        id             path      int     true   "Venue ID"
// @Param        user_lat       query     number  false  "User latitude for distance calculation"
// @Param        user_lng       query     number  false  "User longitude for distance calculation"
// @Success      200  {object}  serializers.VenueDetailResponse
// @Failure      404  {object}  serializers.Base
// @Router       /venues/{id} [get]
func (VenueController) GetByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	venueID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, serializers.Base{
			Code:    serializers.InvalidInput,
			Message: "Invalid venue ID",
		})
		return
	}

	venue := &models.Venue{ID: venueID}
	err = venue.GetByID()
	if err != nil {
		ctx.JSON(http.StatusNotFound, serializers.Base{
			Code:    serializers.NotFound,
			Message: "Venue not found",
		})
		return
	}

	// Calculate distance if user location provided
	if latStr := ctx.Query("user_lat"); latStr != "" {
		if lngStr := ctx.Query("user_lng"); lngStr != "" {
			if userLat, err1 := strconv.ParseFloat(latStr, 64); err1 == nil {
				if userLng, err2 := strconv.ParseFloat(lngStr, 64); err2 == nil {
					// Calculate distance using Haversine formula (simplified)
					distance := calculateDistance(userLat, userLng, venue.Latitude, venue.Longitude)
					venue.Distance = &distance
				}
			}
		}
	}

	// Get review summary
	reviewSummary, err := models.GetVenueReviewSummary(venueID)
	if err != nil {
		reviewSummary = &models.ReviewSummary{VenueID: venueID}
	}

	// Get recent events (commented out for now)
	// events := getVenueEvents(venueID, 5)

	response := serializers.VenueDetailResponse{
		Venue:         *venue,
		ReviewSummary: reviewSummary,
		// Events:        events,
	}

	ctx.JSON(http.StatusOK, response)
}

// GetNearby finds venues near a location
// @Summary      Get nearby venues
// @Tags         venues
// @Produce      json
// @Param        lat            query     number  true   "Latitude"
// @Param        lng            query     number  true   "Longitude"
// @Param        radius         query     number  false  "Search radius in km (default 5)"
// @Param        category       query     int     false  "Filter by category ID"
// @Param        limit          query     int     false  "Number of results (default 20)"
// @Success      200  {object}  []models.Venue
// @Failure      400  {object}  serializers.Base
// @Router       /venues/nearby [get]
func (VenueController) GetNearby(ctx *gin.Context) {
	latStr := ctx.Query("lat")
	lngStr := ctx.Query("lng")

	if latStr == "" || lngStr == "" {
		ctx.JSON(http.StatusBadRequest, serializers.Base{
			Code:    serializers.InvalidInput,
			Message: "Latitude and longitude are required",
		})
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, serializers.Base{
			Code:    serializers.InvalidInput,
			Message: "Invalid latitude",
		})
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, serializers.Base{
			Code:    serializers.InvalidInput,
			Message: "Invalid longitude",
		})
		return
	}

	radius := 5.0 // Default 5km
	if radiusStr := ctx.Query("radius"); radiusStr != "" {
		if r, err := strconv.ParseFloat(radiusStr, 64); err == nil {
			radius = r
		}
	}

	limit := 20
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	venue := &models.Venue{}
	venues, err := venue.GetNearby(lat, lng, radius, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, serializers.Base{
			Code:    serializers.InternalError,
			Message: "Failed to find nearby venues",
		})
		return
	}

	ctx.JSON(http.StatusOK, venues)
}

// GetFeatured returns featured venues
// @Summary      Get featured venues
// @Tags         venues
// @Produce      json
// @Param        category       query     int     false  "Filter by category ID"
// @Param        city           query     int     false  "Filter by city ID"
// @Param        limit          query     int     false  "Number of results (default 10)"
// @Success      200  {object}  []models.Venue
// @Router       /venues/featured [get]
func (VenueController) GetFeatured(ctx *gin.Context) {
	limit := 10
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	venue := &models.Venue{}
	venues, err := venue.GetFeatured(limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, serializers.Base{
			Code:    serializers.InternalError,
			Message: "Failed to get featured venues",
		})
		return
	}

	ctx.JSON(http.StatusOK, venues)
}

// GetCategories returns all venue categories
// @Summary      Get venue categories
// @Tags         venues
// @Produce      json
// @Success      200  {object}  []models.VenueCategory
// @Router       /venues/categories [get]
func (VenueController) GetCategories(ctx *gin.Context) {
	categories, err := models.GetVenueCategories()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, serializers.Base{
			Code:    serializers.InternalError,
			Message: "Failed to get categories",
		})
		return
	}

	ctx.JSON(http.StatusOK, categories)
}

// CreateVenue creates a new venue (admin or owner only)
// @Summary      Create new venue
// @Tags         venues
// @Accept       json
// @Produce      json
// @Param        venue          body      serializers.CreateVenueRequest  true  "Venue data"
// @Success      201  {object}  models.Venue
// @Failure      400  {object}  serializers.Base
// @Failure      401  {object}  serializers.Base
// @Router       /venues [post]
func (VenueController) CreateVenue(ctx *gin.Context) {
	var request serializers.CreateVenueRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, serializers.Base{
			Code:    serializers.InvalidInput,
			Message: "Invalid venue data",
		})
		return
	}

	// Validate required fields
	base, isValid := request.Validate()
	if !isValid {
		ctx.JSON(http.StatusBadRequest, base)
		return
	}

	// Create venue
	venue := request.ToVenue()
	// Set owner from authenticated user
	if userID := ctx.GetInt64("user_id"); userID > 0 {
		venue.OwnerID = &userID
		now := time.Now()
		venue.ClaimedAt = &now
	}

	err := venue.Create()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, serializers.Base{
			Code:    serializers.InternalError,
			Message: "Failed to create venue",
		})
		return
	}

	ctx.JSON(http.StatusCreated, venue)
}

// Helper functions

func calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	// Simple Haversine formula implementation
	// For production, use a proper geospatial library
	const R = 6371 // Earth radius in kilometers

	dLat := (lat2 - lat1) * (3.14159 / 180)
	_ = (lng2 - lng1) * (3.14159 / 180) // dLng not used in simplified calc

	a := 0.5 - (0.5 * (1 + (dLat * dLat / 4)))
	return R * 2 * (1.5708 - a) // Simplified calculation
}

func getVenueEvents(venueID int64, limit int) []interface{} {
	// This would be implemented in a VenueEvent model
	// For now, return empty slice
	return []interface{}{}
}
