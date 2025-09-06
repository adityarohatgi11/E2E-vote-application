package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	databases "voting-app/app"
	"voting-app/app/models"

	"github.com/getsentry/sentry-go"
)

// AnalyticsService provides comprehensive analytics and insights
type AnalyticsService struct{}

// VenueAnalytics represents comprehensive venue performance metrics
type VenueAnalytics struct {
	VenueID   int64  `json:"venueId"`
	VenueName string `json:"venueName"`
	TimeRange string `json:"timeRange"`

	// Engagement Metrics
	ProfileViews      int `json:"profileViews"`
	PhotoViews        int `json:"photoViews"`
	PhoneClicks       int `json:"phoneClicks"`
	WebsiteClicks     int `json:"websiteClicks"`
	DirectionRequests int `json:"directionRequests"`

	// Social Metrics
	CheckinsCount int `json:"checkinsCount"`
	ReviewsCount  int `json:"reviewsCount"`
	SharesCount   int `json:"sharesCount"`

	// Rating Trends
	AverageRating      float64        `json:"averageRating"`
	RatingTrend        []DailyRating  `json:"ratingTrend"`
	RatingDistribution map[string]int `json:"ratingDistribution"`

	// Popular Times
	PopularHours map[int]int    `json:"popularHours"` // Hour -> Visit count
	PopularDays  map[string]int `json:"popularDays"`  // Day -> Visit count

	// Search Performance
	SearchImpressions int     `json:"searchImpressions"`
	SearchClicks      int     `json:"searchClicks"`
	ClickThroughRate  float64 `json:"clickThroughRate"`
	AveragePosition   float64 `json:"averagePosition"`

	// Competitive Analysis
	CategoryRank int `json:"categoryRank"`
	LocalRank    int `json:"localRank"`

	// User Demographics
	Demographics UserDemographics `json:"demographics"`

	// Growth Metrics
	GrowthMetrics GrowthData `json:"growthMetrics"`
}

type DailyRating struct {
	Date   string  `json:"date"`
	Rating float64 `json:"rating"`
	Count  int     `json:"count"`
}

type UserDemographics struct {
	AgeGroups      map[string]int `json:"ageGroups"`
	TopCities      []CityMetric   `json:"topCities"`
	ReturnVisitors float64        `json:"returnVisitorRate"`
}

type CityMetric struct {
	CityName string `json:"cityName"`
	Count    int    `json:"count"`
}

type GrowthData struct {
	ReviewsGrowth     float64 `json:"reviewsGrowth"` // % change
	RatingGrowth      float64 `json:"ratingGrowth"`
	ProfileViewGrowth float64 `json:"profileViewGrowth"`
	PeriodComparison  string  `json:"periodComparison"`
}

// PlatformAnalytics represents overall platform metrics
type PlatformAnalytics struct {
	TimeRange string `json:"timeRange"`

	// Overall Metrics
	TotalVenues   int `json:"totalVenues"`
	TotalUsers    int `json:"totalUsers"`
	TotalReviews  int `json:"totalReviews"`
	TotalCheckins int `json:"totalCheckins"`

	// Activity Metrics
	DailyActiveUsers   int `json:"dailyActiveUsers"`
	WeeklyActiveUsers  int `json:"weeklyActiveUsers"`
	MonthlyActiveUsers int `json:"monthlyActiveUsers"`

	// Content Metrics
	ReviewsPerDay   []DateValue `json:"reviewsPerDay"`
	CheckinsPerDay  []DateValue `json:"checkinsPerDay"`
	NewVenuesPerDay []DateValue `json:"newVenuesPerDay"`

	// Quality Metrics
	AverageRating      float64 `json:"averageRating"`
	ReviewQualityScore float64 `json:"reviewQualityScore"`

	// Popular Categories
	TopCategories []CategoryMetric `json:"topCategories"`
	TopCities     []CityMetric     `json:"topCities"`

	// Search Analytics
	TopSearchQueries []SearchQueryMetric `json:"topSearchQueries"`
	SearchTrends     []SearchTrendMetric `json:"searchTrends"`

	// User Engagement
	EngagementMetrics UserEngagementMetrics `json:"engagementMetrics"`
}

type DateValue struct {
	Date  string `json:"date"`
	Value int    `json:"value"`
}

type CategoryMetric struct {
	CategoryName  string  `json:"categoryName"`
	VenueCount    int     `json:"venueCount"`
	ReviewCount   int     `json:"reviewCount"`
	AverageRating float64 `json:"averageRating"`
}

type SearchQueryMetric struct {
	Query        string  `json:"query"`
	Count        int     `json:"count"`
	ResultsCount int     `json:"resultsCount"`
	ClickThrough float64 `json:"clickThrough"`
}

type SearchTrendMetric struct {
	Query  string      `json:"query"`
	Trend  []DateValue `json:"trend"`
	Growth float64     `json:"growth"`
}

type UserEngagementMetrics struct {
	AverageSessionDuration time.Duration `json:"averageSessionDuration"`
	BounceRate             float64       `json:"bounceRate"`
	PagesPerSession        float64       `json:"pagesPerSession"`
	ReturnUserRate         float64       `json:"returnUserRate"`
}

// GetVenueAnalytics returns comprehensive analytics for a specific venue
func (as *AnalyticsService) GetVenueAnalytics(venueID int64, timeRange string) (*VenueAnalytics, error) {
	// Parse time range
	startDate, endDate, err := as.parseTimeRange(timeRange)
	if err != nil {
		return nil, err
	}

	analytics := &VenueAnalytics{
		VenueID:            venueID,
		TimeRange:          timeRange,
		RatingDistribution: make(map[string]int),
		PopularHours:       make(map[int]int),
		PopularDays:        make(map[string]int),
	}

	// Get venue name
	var venueName string
	err = databases.PostgresDB.QueryRow("SELECT name FROM venues WHERE id = $1", venueID).Scan(&venueName)
	if err != nil {
		return nil, err
	}
	analytics.VenueName = venueName

	// Get engagement metrics
	err = as.getVenueEngagementMetrics(venueID, startDate, endDate, analytics)
	if err != nil {
		sentry.CaptureException(err)
	}

	// Get rating analytics
	err = as.getVenueRatingAnalytics(venueID, startDate, endDate, analytics)
	if err != nil {
		sentry.CaptureException(err)
	}

	// Get popular times
	err = as.getVenuePopularTimes(venueID, startDate, endDate, analytics)
	if err != nil {
		sentry.CaptureException(err)
	}

	// Get search performance
	err = as.getVenueSearchPerformance(venueID, startDate, endDate, analytics)
	if err != nil {
		sentry.CaptureException(err)
	}

	// Get competitive ranking
	err = as.getVenueRanking(venueID, analytics)
	if err != nil {
		sentry.CaptureException(err)
	}

	// Get user demographics
	err = as.getVenueDemographics(venueID, startDate, endDate, analytics)
	if err != nil {
		sentry.CaptureException(err)
	}

	// Calculate growth metrics
	err = as.calculateVenueGrowth(venueID, startDate, endDate, analytics)
	if err != nil {
		sentry.CaptureException(err)
	}

	return analytics, nil
}

// GetPlatformAnalytics returns overall platform performance metrics
func (as *AnalyticsService) GetPlatformAnalytics(timeRange string) (*PlatformAnalytics, error) {
	startDate, endDate, err := as.parseTimeRange(timeRange)
	if err != nil {
		return nil, err
	}

	analytics := &PlatformAnalytics{
		TimeRange: timeRange,
	}

	// Get overall counts
	err = as.getPlatformOverviewMetrics(analytics)
	if err != nil {
		return nil, err
	}

	// Get activity metrics
	err = as.getPlatformActivityMetrics(startDate, endDate, analytics)
	if err != nil {
		sentry.CaptureException(err)
	}

	// Get content metrics
	err = as.getPlatformContentMetrics(startDate, endDate, analytics)
	if err != nil {
		sentry.CaptureException(err)
	}

	// Get top categories and cities
	err = as.getPlatformTopMetrics(startDate, endDate, analytics)
	if err != nil {
		sentry.CaptureException(err)
	}

	// Get search analytics
	err = as.getPlatformSearchAnalytics(startDate, endDate, analytics)
	if err != nil {
		sentry.CaptureException(err)
	}

	return analytics, nil
}

// TrackVenueView records a venue profile view
func (as *AnalyticsService) TrackVenueView(venueID, userID int64, viewType string) error {
	// Insert or update daily analytics
	query := `
		INSERT INTO venue_analytics (venue_id, date, profile_views, photo_views, phone_clicks, website_clicks, direction_requests)
		VALUES ($1, CURRENT_DATE, 
			CASE WHEN $2 = 'profile' THEN 1 ELSE 0 END,
			CASE WHEN $2 = 'photo' THEN 1 ELSE 0 END,
			CASE WHEN $2 = 'phone' THEN 1 ELSE 0 END,
			CASE WHEN $2 = 'website' THEN 1 ELSE 0 END,
			CASE WHEN $2 = 'directions' THEN 1 ELSE 0 END
		)
		ON CONFLICT (venue_id, date) DO UPDATE SET
			profile_views = venue_analytics.profile_views + CASE WHEN $2 = 'profile' THEN 1 ELSE 0 END,
			photo_views = venue_analytics.photo_views + CASE WHEN $2 = 'photo' THEN 1 ELSE 0 END,
			phone_clicks = venue_analytics.phone_clicks + CASE WHEN $2 = 'phone' THEN 1 ELSE 0 END,
			website_clicks = venue_analytics.website_clicks + CASE WHEN $2 = 'website' THEN 1 ELSE 0 END,
			direction_requests = venue_analytics.direction_requests + CASE WHEN $2 = 'directions' THEN 1 ELSE 0 END`

	_, err := databases.PostgresDB.Exec(query, venueID, viewType)
	if err != nil {
		sentry.CaptureException(err)
	}

	return err
}

// TrackSearch records search analytics
func (as *AnalyticsService) TrackSearch(userID int64, query string, filters map[string]interface{}, results []models.Venue, clickedVenueID *int64, clickPosition *int) error {
	// Get user location if available
	var userLat, userLng *float64
	if lat, exists := filters["latitude"]; exists {
		if latFloat, ok := lat.(float64); ok {
			userLat = &latFloat
		}
	}
	if lng, exists := filters["longitude"]; exists {
		if lngFloat, ok := lng.(float64); ok {
			userLng = &lngFloat
		}
	}

	// Get search radius
	var searchRadius *int
	if radius, exists := filters["radius"]; exists {
		if radiusFloat, ok := radius.(float64); ok {
			radiusInt := int(radiusFloat)
			searchRadius = &radiusInt
		}
	}

	// Serialize filters
	filtersJSON, _ := json.Marshal(filters)

	// Insert search record
	insertQuery := `
		INSERT INTO search_analytics (
			user_id, search_query, search_type, filters_used,
			user_latitude, user_longitude, search_radius,
			results_count, clicked_venue_id, click_position
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	searchType := "text"
	if query == "" {
		searchType = "filter"
	}

	_, err := databases.PostgresDB.Exec(
		insertQuery,
		userID, query, searchType, filtersJSON,
		userLat, userLng, searchRadius,
		len(results), clickedVenueID, clickPosition,
	)

	if err != nil {
		sentry.CaptureException(err)
	}

	return err
}

// Helper methods for analytics calculation

func (as *AnalyticsService) parseTimeRange(timeRange string) (time.Time, time.Time, error) {
	now := time.Now()
	var startDate time.Time

	switch timeRange {
	case "today":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "yesterday":
		yesterday := now.AddDate(0, 0, -1)
		startDate = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, now.Location())
		now = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 0, now.Location())
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	case "quarter":
		startDate = now.AddDate(0, -3, 0)
	case "year":
		startDate = now.AddDate(-1, 0, 0)
	default:
		startDate = now.AddDate(0, 0, -7) // Default to last week
	}

	return startDate, now, nil
}

func (as *AnalyticsService) getVenueEngagementMetrics(venueID int64, startDate, endDate time.Time, analytics *VenueAnalytics) error {
	query := `
		SELECT 
			COALESCE(SUM(profile_views), 0) as profile_views,
			COALESCE(SUM(photo_views), 0) as photo_views,
			COALESCE(SUM(phone_clicks), 0) as phone_clicks,
			COALESCE(SUM(website_clicks), 0) as website_clicks,
			COALESCE(SUM(direction_requests), 0) as direction_requests,
			COALESCE(SUM(checkins), 0) as checkins,
			COALESCE(SUM(reviews_count), 0) as reviews_count,
			COALESCE(SUM(shares), 0) as shares
		FROM venue_analytics
		WHERE venue_id = $1 AND date BETWEEN $2 AND $3`

	err := databases.PostgresDB.QueryRow(query, venueID, startDate, endDate).Scan(
		&analytics.ProfileViews,
		&analytics.PhotoViews,
		&analytics.PhoneClicks,
		&analytics.WebsiteClicks,
		&analytics.DirectionRequests,
		&analytics.CheckinsCount,
		&analytics.ReviewsCount,
		&analytics.SharesCount,
	)

	return err
}

func (as *AnalyticsService) getVenueRatingAnalytics(venueID int64, startDate, endDate time.Time, analytics *VenueAnalytics) error {
	// Get current average rating
	err := databases.PostgresDB.QueryRow(
		"SELECT average_rating FROM venues WHERE id = $1", venueID,
	).Scan(&analytics.AverageRating)
	if err != nil {
		return err
	}

	// Get rating distribution
	distQuery := `
		SELECT 
			CASE 
				WHEN overall_rating >= 4.5 THEN '5'
				WHEN overall_rating >= 3.5 THEN '4'
				WHEN overall_rating >= 2.5 THEN '3'
				WHEN overall_rating >= 1.5 THEN '2'
				ELSE '1'
			END as rating_bucket,
			COUNT(*) as count
		FROM venue_reviews
		WHERE venue_id = $1 AND created_at BETWEEN $2 AND $3
		GROUP BY rating_bucket`

	rows, err := databases.PostgresDB.Query(distQuery, venueID, startDate, endDate)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var bucket string
		var count int
		if rows.Scan(&bucket, &count) == nil {
			analytics.RatingDistribution[bucket] = count
		}
	}

	// Get daily rating trend
	trendQuery := `
		SELECT 
			DATE(created_at) as date,
			AVG(overall_rating) as avg_rating,
			COUNT(*) as review_count
		FROM venue_reviews
		WHERE venue_id = $1 AND created_at BETWEEN $2 AND $3
		GROUP BY DATE(created_at)
		ORDER BY date`

	rows, err = databases.PostgresDB.Query(trendQuery, venueID, startDate, endDate)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var daily DailyRating
		var date time.Time
		if rows.Scan(&date, &daily.Rating, &daily.Count) == nil {
			daily.Date = date.Format("2006-01-02")
			analytics.RatingTrend = append(analytics.RatingTrend, daily)
		}
	}

	return nil
}

func (as *AnalyticsService) getVenuePopularTimes(venueID int64, startDate, endDate time.Time, analytics *VenueAnalytics) error {
	// Get popular hours from check-ins
	hourQuery := `
		SELECT 
			EXTRACT(hour FROM created_at) as hour,
			COUNT(*) as count
		FROM venue_checkins
		WHERE venue_id = $1 AND created_at BETWEEN $2 AND $3
		GROUP BY hour
		ORDER BY hour`

	rows, err := databases.PostgresDB.Query(hourQuery, venueID, startDate, endDate)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var hour, count int
		if rows.Scan(&hour, &count) == nil {
			analytics.PopularHours[hour] = count
		}
	}

	// Get popular days
	dayQuery := `
		SELECT 
			TO_CHAR(created_at, 'Day') as day_name,
			COUNT(*) as count
		FROM venue_checkins
		WHERE venue_id = $1 AND created_at BETWEEN $2 AND $3
		GROUP BY day_name, EXTRACT(dow FROM created_at)
		ORDER BY EXTRACT(dow FROM created_at)`

	rows, err = databases.PostgresDB.Query(dayQuery, venueID, startDate, endDate)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var day string
		var count int
		if rows.Scan(&day, &count) == nil {
			analytics.PopularDays[strings.TrimSpace(day)] = count
		}
	}

	return nil
}

func (as *AnalyticsService) getVenueSearchPerformance(venueID int64, startDate, endDate time.Time, analytics *VenueAnalytics) error {
	// Get search impressions and clicks
	query := `
		SELECT 
			COUNT(*) as impressions,
			COUNT(CASE WHEN clicked_venue_id = $1 THEN 1 END) as clicks,
			AVG(CASE WHEN clicked_venue_id = $1 THEN click_position END) as avg_position
		FROM search_analytics sa
		WHERE sa.created_at BETWEEN $2 AND $3
		  AND EXISTS (
			  SELECT 1 FROM venues v 
			  WHERE v.id = $1 
			  AND (
				  LOWER(v.name) LIKE LOWER('%' || sa.search_query || '%')
				  OR v.category_id::text = ANY(string_to_array(sa.filters_used->>'category_id', ','))
			  )
		  )`

	var avgPosition sql.NullFloat64
	err := databases.PostgresDB.QueryRow(query, venueID, startDate, endDate).Scan(
		&analytics.SearchImpressions,
		&analytics.SearchClicks,
		&avgPosition,
	)

	if err != nil {
		return err
	}

	if avgPosition.Valid {
		analytics.AveragePosition = avgPosition.Float64
	}

	if analytics.SearchImpressions > 0 {
		analytics.ClickThroughRate = float64(analytics.SearchClicks) / float64(analytics.SearchImpressions)
	}

	return nil
}

func (as *AnalyticsService) getVenueRanking(venueID int64, analytics *VenueAnalytics) error {
	// Get category ranking
	var categoryID int64
	err := databases.PostgresDB.QueryRow("SELECT category_id FROM venues WHERE id = $1", venueID).Scan(&categoryID)
	if err != nil {
		return err
	}

	// Category rank
	categoryRankQuery := `
		SELECT COUNT(*) + 1 as rank
		FROM venues
		WHERE category_id = $1 AND average_rating > (SELECT average_rating FROM venues WHERE id = $2)`

	err = databases.PostgresDB.QueryRow(categoryRankQuery, categoryID, venueID).Scan(&analytics.CategoryRank)
	if err != nil {
		analytics.CategoryRank = 0
	}

	// Local rank (within same city)
	localRankQuery := `
		SELECT COUNT(*) + 1 as rank
		FROM venues v1
		JOIN venues v2 ON v1.city_id = v2.city_id
		WHERE v2.id = $1 AND v1.average_rating > v2.average_rating`

	err = databases.PostgresDB.QueryRow(localRankQuery, venueID).Scan(&analytics.LocalRank)
	if err != nil {
		analytics.LocalRank = 0
	}

	return nil
}

func (as *AnalyticsService) getVenueDemographics(venueID int64, startDate, endDate time.Time, analytics *VenueAnalytics) error {
	analytics.Demographics = UserDemographics{
		AgeGroups: make(map[string]int),
	}

	// Get top cities of visitors (based on check-ins)
	cityQuery := `
		SELECT c.name, COUNT(*) as count
		FROM venue_checkins vc
		JOIN snapp_users su ON vc.user_id = su.id
		JOIN cities c ON ST_DWithin(
			ST_Point(vc.user_latitude, vc.user_longitude)::geography,
			ST_Point(c.longitude, c.latitude)::geography,
			50000
		)
		WHERE vc.venue_id = $1 AND vc.created_at BETWEEN $2 AND $3
		GROUP BY c.name
		ORDER BY count DESC
		LIMIT 5`

	rows, err := databases.PostgresDB.Query(cityQuery, venueID, startDate, endDate)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var city CityMetric
			if rows.Scan(&city.CityName, &city.Count) == nil {
				analytics.Demographics.TopCities = append(analytics.Demographics.TopCities, city)
			}
		}
	}

	// Calculate return visitor rate
	returnQuery := `
		SELECT 
			COUNT(DISTINCT user_id) as total_users,
			COUNT(DISTINCT CASE WHEN visit_count > 1 THEN user_id END) as return_users
		FROM (
			SELECT user_id, COUNT(*) as visit_count
			FROM venue_checkins
			WHERE venue_id = $1 AND created_at BETWEEN $2 AND $3
			GROUP BY user_id
		) user_visits`

	var totalUsers, returnUsers int
	err = databases.PostgresDB.QueryRow(returnQuery, venueID, startDate, endDate).Scan(&totalUsers, &returnUsers)
	if err == nil && totalUsers > 0 {
		analytics.Demographics.ReturnVisitors = float64(returnUsers) / float64(totalUsers)
	}

	return nil
}

func (as *AnalyticsService) calculateVenueGrowth(venueID int64, startDate, endDate time.Time, analytics *VenueAnalytics) error {
	// Calculate growth compared to previous period
	duration := endDate.Sub(startDate)
	prevStartDate := startDate.Add(-duration)
	prevEndDate := startDate

	// Reviews growth
	var currentReviews, prevReviews int

	databases.PostgresDB.QueryRow(
		"SELECT COUNT(*) FROM venue_reviews WHERE venue_id = $1 AND created_at BETWEEN $2 AND $3",
		venueID, startDate, endDate,
	).Scan(&currentReviews)

	databases.PostgresDB.QueryRow(
		"SELECT COUNT(*) FROM venue_reviews WHERE venue_id = $1 AND created_at BETWEEN $2 AND $3",
		venueID, prevStartDate, prevEndDate,
	).Scan(&prevReviews)

	if prevReviews > 0 {
		analytics.GrowthMetrics.ReviewsGrowth = (float64(currentReviews-prevReviews) / float64(prevReviews)) * 100
	}

	// Profile views growth
	var currentViews, prevViews int

	databases.PostgresDB.QueryRow(
		"SELECT COALESCE(SUM(profile_views), 0) FROM venue_analytics WHERE venue_id = $1 AND date BETWEEN $2 AND $3",
		venueID, startDate, endDate,
	).Scan(&currentViews)

	databases.PostgresDB.QueryRow(
		"SELECT COALESCE(SUM(profile_views), 0) FROM venue_analytics WHERE venue_id = $1 AND date BETWEEN $2 AND $3",
		venueID, prevStartDate, prevEndDate,
	).Scan(&prevViews)

	if prevViews > 0 {
		analytics.GrowthMetrics.ProfileViewGrowth = (float64(currentViews-prevViews) / float64(prevViews)) * 100
	}

	analytics.GrowthMetrics.PeriodComparison = fmt.Sprintf("vs previous %s", analytics.TimeRange)

	return nil
}

// Platform-wide analytics helper methods would follow similar patterns...
func (as *AnalyticsService) getPlatformOverviewMetrics(analytics *PlatformAnalytics) error {
	query := `
		SELECT 
			(SELECT COUNT(*) FROM venues WHERE is_active = true),
			(SELECT COUNT(*) FROM snapp_users),
			(SELECT COUNT(*) FROM venue_reviews),
			(SELECT COUNT(*) FROM venue_checkins),
			(SELECT AVG(average_rating) FROM venues WHERE total_ratings > 0)`

	err := databases.PostgresDB.QueryRow(query).Scan(
		&analytics.TotalVenues,
		&analytics.TotalUsers,
		&analytics.TotalReviews,
		&analytics.TotalCheckins,
		&analytics.AverageRating,
	)

	return err
}

func (as *AnalyticsService) getPlatformActivityMetrics(startDate, endDate time.Time, analytics *PlatformAnalytics) error {
	// Daily active users
	err := databases.PostgresDB.QueryRow(
		`SELECT COUNT(DISTINCT user_id) FROM venue_checkins WHERE created_at >= CURRENT_DATE`,
	).Scan(&analytics.DailyActiveUsers)

	if err != nil {
		analytics.DailyActiveUsers = 0
	}

	// Weekly active users
	err = databases.PostgresDB.QueryRow(
		`SELECT COUNT(DISTINCT user_id) FROM venue_checkins WHERE created_at >= CURRENT_DATE - INTERVAL '7 days'`,
	).Scan(&analytics.WeeklyActiveUsers)

	if err != nil {
		analytics.WeeklyActiveUsers = 0
	}

	// Monthly active users
	err = databases.PostgresDB.QueryRow(
		`SELECT COUNT(DISTINCT user_id) FROM venue_checkins WHERE created_at >= CURRENT_DATE - INTERVAL '30 days'`,
	).Scan(&analytics.MonthlyActiveUsers)

	if err != nil {
		analytics.MonthlyActiveUsers = 0
	}

	return nil
}

func (as *AnalyticsService) getPlatformContentMetrics(startDate, endDate time.Time, analytics *PlatformAnalytics) error {
	// Reviews per day
	reviewQuery := `
		SELECT DATE(created_at), COUNT(*) 
		FROM venue_reviews 
		WHERE created_at BETWEEN $1 AND $2 
		GROUP BY DATE(created_at) 
		ORDER BY DATE(created_at)`

	rows, err := databases.PostgresDB.Query(reviewQuery, startDate, endDate)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var date time.Time
			var count int
			if rows.Scan(&date, &count) == nil {
				analytics.ReviewsPerDay = append(analytics.ReviewsPerDay, DateValue{
					Date:  date.Format("2006-01-02"),
					Value: count,
				})
			}
		}
	}

	// Similar queries for checkins and new venues...

	return nil
}

func (as *AnalyticsService) getPlatformTopMetrics(startDate, endDate time.Time, analytics *PlatformAnalytics) error {
	// Top categories
	categoryQuery := `
		SELECT vc.name, COUNT(v.id) as venue_count, COUNT(vr.id) as review_count, AVG(v.average_rating) as avg_rating
		FROM venue_categories vc
		LEFT JOIN venues v ON vc.id = v.category_id AND v.is_active = true
		LEFT JOIN venue_reviews vr ON v.id = vr.venue_id AND vr.created_at BETWEEN $1 AND $2
		GROUP BY vc.id, vc.name
		ORDER BY venue_count DESC, review_count DESC
		LIMIT 10`

	rows, err := databases.PostgresDB.Query(categoryQuery, startDate, endDate)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var category CategoryMetric
			var avgRating sql.NullFloat64
			if rows.Scan(&category.CategoryName, &category.VenueCount, &category.ReviewCount, &avgRating) == nil {
				if avgRating.Valid {
					category.AverageRating = avgRating.Float64
				}
				analytics.TopCategories = append(analytics.TopCategories, category)
			}
		}
	}

	return nil
}

func (as *AnalyticsService) getPlatformSearchAnalytics(startDate, endDate time.Time, analytics *PlatformAnalytics) error {
	// Top search queries
	queryQuery := `
		SELECT search_query, COUNT(*) as search_count, 
			   AVG(results_count) as avg_results,
			   COUNT(clicked_venue_id)::float / COUNT(*)::float as ctr
		FROM search_analytics
		WHERE created_at BETWEEN $1 AND $2 AND search_query != ''
		GROUP BY search_query
		ORDER BY search_count DESC
		LIMIT 20`

	rows, err := databases.PostgresDB.Query(queryQuery, startDate, endDate)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var query SearchQueryMetric
			var avgResults sql.NullFloat64
			if rows.Scan(&query.Query, &query.Count, &avgResults, &query.ClickThrough) == nil {
				if avgResults.Valid {
					query.ResultsCount = int(avgResults.Float64)
				}
				analytics.TopSearchQueries = append(analytics.TopSearchQueries, query)
			}
		}
	}

	return nil
}

// GetTopPerformingVenues returns the best performing venues
func (as *AnalyticsService) GetTopPerformingVenues(timeRange string, category *int64, city *int64, limit int) ([]VenueAnalytics, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	startDate, endDate, err := as.parseTimeRange(timeRange)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT v.id, v.name, v.average_rating, v.total_ratings,
			   COALESCE(va.total_views, 0) as total_views,
			   COALESCE(va.total_checkins, 0) as total_checkins
		FROM venues v
		LEFT JOIN (
			SELECT venue_id,
				   SUM(profile_views + photo_views) as total_views,
				   SUM(checkins) as total_checkins
			FROM venue_analytics
			WHERE date BETWEEN $1 AND $2
			GROUP BY venue_id
		) va ON v.id = va.venue_id
		WHERE v.is_active = true AND v.total_ratings >= 5`

	args := []interface{}{startDate, endDate}
	argCount := 2

	if category != nil {
		argCount++
		query += fmt.Sprintf(" AND v.category_id = $%d", argCount)
		args = append(args, *category)
	}

	if city != nil {
		argCount++
		query += fmt.Sprintf(" AND v.city_id = $%d", argCount)
		args = append(args, *city)
	}

	query += ` ORDER BY 
		(v.average_rating * 0.4 + 
		 LEAST(COALESCE(va.total_views, 0) / 100.0, 5.0) * 0.3 + 
		 LEAST(COALESCE(va.total_checkins, 0) / 10.0, 5.0) * 0.3) DESC
		LIMIT $` + fmt.Sprintf("%d", argCount+1)
	args = append(args, limit)

	rows, err := databases.PostgresDB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []VenueAnalytics
	for rows.Next() {
		var va VenueAnalytics
		var totalViews, totalCheckins int

		err := rows.Scan(
			&va.VenueID, &va.VenueName, &va.AverageRating,
			&va.ReviewsCount, &totalViews, &totalCheckins,
		)
		if err != nil {
			continue
		}

		va.ProfileViews = totalViews
		va.CheckinsCount = totalCheckins
		va.TimeRange = timeRange

		results = append(results, va)
	}

	return results, nil
}

// Additional analytics methods for sentiment analysis, growth metrics, etc. would be implemented similarly...
