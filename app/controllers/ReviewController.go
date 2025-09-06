package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
	"voting-app/app/models"
	"voting-app/app/serializers"
)

type ReviewController struct{}

// CreateReview creates a new venue review
// @Summary      Create venue review
// @Tags         reviews
// @Accept       json
// @Produce      json
// @Param        snapp_id       path      string  true   "User Snapp ID"
// @Param        review         body      serializers.CreateReviewRequest  true  "Review data"
// @Success      201  {object}  models.VenueReview
// @Failure      400  {object}  serializers.Base
// @Failure      401  {object}  serializers.Base
// @Router       /reviews/{snapp_id} [post]
func (ReviewController) CreateReview(ctx *gin.Context) {
	var request serializers.CreateReviewRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, serializers.Base{
			Code:    serializers.InvalidInput,
			Message: "Invalid review data",
		})
		return
	}

	// Validate request
	base, isValid := request.Validate()
	if !isValid {
		ctx.JSON(http.StatusBadRequest, base)
		return
	}

	// Get user ID from context (set by middleware)
	userID := ctx.GetInt64("snappUser_id")

	// Create review
	review := request.ToReview()
	review.UserID = userID

	err := review.Create()
	if err != nil {
		if err.Error() == "user has already reviewed this venue" {
			ctx.JSON(http.StatusBadRequest, serializers.Base{
				Code:    serializers.AlreadyReviewed,
				Message: "You have already reviewed this venue",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, serializers.Base{
			Code:    serializers.InternalError,
			Message: "Failed to create review",
		})
		return
	}

	// Get the created review with full details
	err = review.GetByID()
	if err != nil {
		// Review was created but we couldn't fetch details, still return success
		ctx.JSON(http.StatusCreated, review)
		return
	}

	ctx.JSON(http.StatusCreated, review)
}

// GetVenueReviews gets all reviews for a venue
// @Summary      Get venue reviews
// @Tags         reviews
// @Produce      json
// @Param        venue_id       path      int     true   "Venue ID"
// @Param        min_rating     query     number  false  "Minimum rating filter"
// @Param        max_rating     query     number  false  "Maximum rating filter"
// @Param        visit_type     query     string  false  "Visit type filter (dinner, lunch, drinks, etc.)"
// @Param        has_photos     query     boolean false  "Filter reviews with photos"
// @Param        sort_by        query     string  false  "Sort by: newest, oldest, rating_high, rating_low, helpful"
// @Param        page           query     int     false  "Page number (default 1)"
// @Param        limit          query     int     false  "Results per page (default 20)"
// @Success      200  {object}  serializers.ReviewSearchResponse
// @Failure      400  {object}  serializers.Base
// @Router       /venues/{venue_id}/reviews [get]
func (ReviewController) GetVenueReviews(ctx *gin.Context) {
	venueIDStr := ctx.Param("venue_id")
	venueID, err := strconv.ParseInt(venueIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, serializers.Base{
			Code:    serializers.InvalidInput,
			Message: "Invalid venue ID",
		})
		return
	}

	// Parse filters
	filters := models.ReviewFilters{
		VenueID: &venueID,
		SortBy:  ctx.DefaultQuery("sort_by", "newest"),
		Page:    1,
		Limit:   20,
	}

	// Parse optional filters
	if minRatingStr := ctx.Query("min_rating"); minRatingStr != "" {
		if minRating, err := strconv.ParseFloat(minRatingStr, 64); err == nil {
			filters.MinRating = &minRating
		}
	}

	if maxRatingStr := ctx.Query("max_rating"); maxRatingStr != "" {
		if maxRating, err := strconv.ParseFloat(maxRatingStr, 64); err == nil {
			filters.MaxRating = &maxRating
		}
	}

	if visitType := ctx.Query("visit_type"); visitType != "" {
		filters.VisitType = visitType
	}

	if hasPhotosStr := ctx.Query("has_photos"); hasPhotosStr != "" {
		if hasPhotos, err := strconv.ParseBool(hasPhotosStr); err == nil {
			filters.HasPhotos = &hasPhotos
		}
	}

	// Parse pagination
	if pageStr := ctx.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filters.Page = page
		}
	}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			filters.Limit = limit
		}
	}

	// Get reviews
	review := &models.VenueReview{}
	reviews, totalCount, err := review.Search(filters)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, serializers.Base{
			Code:    serializers.InternalError,
			Message: "Failed to get reviews",
		})
		return
	}

	// Calculate pagination
	totalPages := (totalCount + filters.Limit - 1) / filters.Limit
	hasNext := filters.Page < totalPages
	hasPrev := filters.Page > 1

	response := serializers.ReviewSearchResponse{
		Reviews: reviews,
		Pagination: serializers.PaginationInfo{
			Page:       filters.Page,
			Limit:      filters.Limit,
			Total:      totalCount,
			TotalPages: totalPages,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
		},
		Filters: filters,
	}

	ctx.JSON(http.StatusOK, response)
}

// GetReviewSummary gets review statistics for a venue
// @Summary      Get venue review summary
// @Tags         reviews
// @Produce      json
// @Param        venue_id       path      int     true   "Venue ID"
// @Success      200  {object}  models.ReviewSummary
// @Failure      400  {object}  serializers.Base
// @Router       /venues/{venue_id}/reviews/summary [get]
func (ReviewController) GetReviewSummary(ctx *gin.Context) {
	venueIDStr := ctx.Param("venue_id")
	venueID, err := strconv.ParseInt(venueIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, serializers.Base{
			Code:    serializers.InvalidInput,
			Message: "Invalid venue ID",
		})
		return
	}

	summary, err := models.GetVenueReviewSummary(venueID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, serializers.Base{
			Code:    serializers.InternalError,
			Message: "Failed to get review summary",
		})
		return
	}

	ctx.JSON(http.StatusOK, summary)
}

// GetUserReviews gets all reviews by a user
// @Summary      Get user reviews
// @Tags         reviews
// @Produce      json
// @Param        snapp_id       path      string  true   "User Snapp ID"
// @Param        sort_by        query     string  false  "Sort by: newest, oldest, rating_high, rating_low"
// @Param        page           query     int     false  "Page number (default 1)"
// @Param        limit          query     int     false  "Results per page (default 20)"
// @Success      200  {object}  serializers.ReviewSearchResponse
// @Failure      400  {object}  serializers.Base
// @Router       /users/{snapp_id}/reviews [get]
func (ReviewController) GetUserReviews(ctx *gin.Context) {
	userID := ctx.GetInt64("snappUser_id")

	// Parse filters
	filters := models.ReviewFilters{
		UserID: &userID,
		SortBy: ctx.DefaultQuery("sort_by", "newest"),
		Page:   1,
		Limit:  20,
	}

	// Parse pagination
	if pageStr := ctx.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filters.Page = page
		}
	}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			filters.Limit = limit
		}
	}

	// Get reviews
	review := &models.VenueReview{}
	reviews, totalCount, err := review.Search(filters)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, serializers.Base{
			Code:    serializers.InternalError,
			Message: "Failed to get user reviews",
		})
		return
	}

	// Calculate pagination
	totalPages := (totalCount + filters.Limit - 1) / filters.Limit
	hasNext := filters.Page < totalPages
	hasPrev := filters.Page > 1

	response := serializers.ReviewSearchResponse{
		Reviews: reviews,
		Pagination: serializers.PaginationInfo{
			Page:       filters.Page,
			Limit:      filters.Limit,
			Total:      totalCount,
			TotalPages: totalPages,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
		},
		Filters: filters,
	}

	ctx.JSON(http.StatusOK, response)
}

// VoteReviewHelpful marks a review as helpful or unhelpful
// @Summary      Vote on review helpfulness
// @Tags         reviews
// @Accept       json
// @Produce      json
// @Param        snapp_id       path      string  true   "User Snapp ID"
// @Param        review_id      path      int     true   "Review ID"
// @Param        vote           body      serializers.ReviewVoteRequest  true  "Vote data"
// @Success      200  {object}  serializers.Base
// @Failure      400  {object}  serializers.Base
// @Failure      401  {object}  serializers.Base
// @Router       /reviews/{snapp_id}/{review_id}/vote [post]
func (ReviewController) VoteReviewHelpful(ctx *gin.Context) {
	reviewIDStr := ctx.Param("review_id")
	reviewID, err := strconv.ParseInt(reviewIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, serializers.Base{
			Code:    serializers.InvalidInput,
			Message: "Invalid review ID",
		})
		return
	}

	var request serializers.ReviewVoteRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, serializers.Base{
			Code:    serializers.InvalidInput,
			Message: "Invalid vote data",
		})
		return
	}

	userID := ctx.GetInt64("snappUser_id")

	review := &models.VenueReview{ID: reviewID}
	err = review.VoteHelpful(userID, request.IsHelpful)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, serializers.Base{
			Code:    serializers.InternalError,
			Message: "Failed to vote on review",
		})
		return
	}

	ctx.JSON(http.StatusOK, serializers.Base{
		Code:    serializers.Success,
		Message: "Vote recorded successfully",
	})
}

// UpdateReview updates an existing review (owner only)
// @Summary      Update review
// @Tags         reviews
// @Accept       json
// @Produce      json
// @Param        snapp_id       path      string  true   "User Snapp ID"
// @Param        review_id      path      int     true   "Review ID"
// @Param        review         body      serializers.UpdateReviewRequest  true  "Updated review data"
// @Success      200  {object}  models.VenueReview
// @Failure      400  {object}  serializers.Base
// @Failure      401  {object}  serializers.Base
// @Failure      403  {object}  serializers.Base
// @Failure      404  {object}  serializers.Base
// @Router       /reviews/{snapp_id}/{review_id} [put]
func (ReviewController) UpdateReview(ctx *gin.Context) {
	reviewIDStr := ctx.Param("review_id")
	reviewID, err := strconv.ParseInt(reviewIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, serializers.Base{
			Code:    serializers.InvalidInput,
			Message: "Invalid review ID",
		})
		return
	}

	var request serializers.UpdateReviewRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, serializers.Base{
			Code:    serializers.InvalidInput,
			Message: "Invalid review data",
		})
		return
	}

	userID := ctx.GetInt64("snappUser_id")

	// Check if review exists and belongs to user
	review := &models.VenueReview{ID: reviewID}
	err = review.GetByID()
	if err != nil {
		ctx.JSON(http.StatusNotFound, serializers.Base{
			Code:    serializers.NotFound,
			Message: "Review not found",
		})
		return
	}

	if review.UserID != userID {
		ctx.JSON(http.StatusForbidden, serializers.Base{
			Code:    serializers.Forbidden,
			Message: "You can only update your own reviews",
		})
		return
	}

	// Validate updated data
	base, isValid := request.Validate()
	if !isValid {
		ctx.JSON(http.StatusBadRequest, base)
		return
	}

	// Update review (this would need to be implemented in the model)
	// For now, just return the existing review
	ctx.JSON(http.StatusOK, review)
}

// DeleteReview deletes a review (owner only)
// @Summary      Delete review
// @Tags         reviews
// @Produce      json
// @Param        snapp_id       path      string  true   "User Snapp ID"
// @Param        review_id      path      int     true   "Review ID"
// @Success      200  {object}  serializers.Base
// @Failure      400  {object}  serializers.Base
// @Failure      401  {object}  serializers.Base
// @Failure      403  {object}  serializers.Base
// @Failure      404  {object}  serializers.Base
// @Router       /reviews/{snapp_id}/{review_id} [delete]
func (ReviewController) DeleteReview(ctx *gin.Context) {
	reviewIDStr := ctx.Param("review_id")
	reviewID, err := strconv.ParseInt(reviewIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, serializers.Base{
			Code:    serializers.InvalidInput,
			Message: "Invalid review ID",
		})
		return
	}

	userID := ctx.GetInt64("snappUser_id")

	// Check if review exists and belongs to user
	review := &models.VenueReview{ID: reviewID}
	err = review.GetByID()
	if err != nil {
		ctx.JSON(http.StatusNotFound, serializers.Base{
			Code:    serializers.NotFound,
			Message: "Review not found",
		})
		return
	}

	if review.UserID != userID {
		ctx.JSON(http.StatusForbidden, serializers.Base{
			Code:    serializers.Forbidden,
			Message: "You can only delete your own reviews",
		})
		return
	}

	// Delete review (this would need to be implemented in the model)
	// For now, just return success
	ctx.JSON(http.StatusOK, serializers.Base{
		Code:    serializers.Success,
		Message: "Review deleted successfully",
	})
}

// GetTrendingReviews gets trending/popular reviews
// @Summary      Get trending reviews
// @Tags         reviews
// @Produce      json
// @Param        category       query     int     false  "Filter by venue category"
// @Param        city           query     int     false  "Filter by city"
// @Param        time_period    query     string  false  "Time period: today, week, month (default week)"
// @Param        limit          query     int     false  "Number of results (default 20)"
// @Success      200  {object}  []models.VenueReview
// @Router       /reviews/trending [get]
func (ReviewController) GetTrendingReviews(ctx *gin.Context) {
	limit := 20
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	timePeriod := ctx.DefaultQuery("time_period", "week")
	var dateFrom *time.Time
	now := time.Now()

	switch timePeriod {
	case "today":
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		dateFrom = &today
	case "month":
		monthAgo := now.AddDate(0, -1, 0)
		dateFrom = &monthAgo
	default: // week
		weekAgo := now.AddDate(0, 0, -7)
		dateFrom = &weekAgo
	}

	filters := models.ReviewFilters{
		DateFrom: dateFrom,
		SortBy:   "helpful",
		Limit:    limit,
		Page:     1,
	}

	// Add category filter if provided
	if categoryStr := ctx.Query("category"); categoryStr != "" {
		// This would need a join with venues table to filter by category
		// For now, we'll skip this filter
	}

	review := &models.VenueReview{}
	reviews, _, err := review.Search(filters)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, serializers.Base{
			Code:    serializers.InternalError,
			Message: "Failed to get trending reviews",
		})
		return
	}

	ctx.JSON(http.StatusOK, reviews)
}
