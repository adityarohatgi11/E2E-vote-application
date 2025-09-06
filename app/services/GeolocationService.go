package services

import (
	"database/sql"
	"fmt"
	"math"
	"time"
	databases "voting-app/app"
	"voting-app/app/models"

	"github.com/getsentry/sentry-go"
)

// GeolocationService handles all location-based operations
type GeolocationService struct {
	MapboxToken string // You'd get this from environment
	GoogleToken string // Alternative geocoding service
}

// LocationResult represents a geocoding result
type LocationResult struct {
	Address    string  `json:"address"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	City       string  `json:"city"`
	State      string  `json:"state"`
	Country    string  `json:"country"`
	PostalCode string  `json:"postalCode"`
	Confidence float64 `json:"confidence"` // 0-1 confidence score
}

// NearbyResult represents venues near a location
type NearbyResult struct {
	Venues     []models.Venue `json:"venues"`
	UserLat    float64        `json:"userLatitude"`
	UserLng    float64        `json:"userLongitude"`
	Radius     float64        `json:"radiusKm"`
	TotalFound int            `json:"totalFound"`
}

// LocationBounds represents a geographical bounding box
type LocationBounds struct {
	NorthEast LatLng `json:"northEast"`
	SouthWest LatLng `json:"southWest"`
}

type LatLng struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Distance represents distance between two points
type Distance struct {
	Meters     float64 `json:"meters"`
	Kilometers float64 `json:"kilometers"`
	Miles      float64 `json:"miles"`
}

// Geocode converts an address to coordinates
func (gs *GeolocationService) Geocode(address string) (*LocationResult, error) {
	// First try with our local database
	result := gs.searchLocalLocations(address)
	if result != nil {
		return result, nil
	}

	// Fallback to external geocoding service
	return gs.externalGeocode(address)
}

// ReverseGeocode converts coordinates to address
func (gs *GeolocationService) ReverseGeocode(lat, lng float64) (*LocationResult, error) {
	// Validate coordinates
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return nil, fmt.Errorf("invalid coordinates")
	}

	// Try local database first
	result := gs.reverseGeocodeLocal(lat, lng)
	if result != nil {
		return result, nil
	}

	// Fallback to external service
	return gs.externalReverseGeocode(lat, lng)
}

// GetNearbyVenues finds venues within a radius
func (gs *GeolocationService) GetNearbyVenues(lat, lng, radiusKm float64, filters map[string]interface{}) (*NearbyResult, error) {
	// Validate inputs
	if radiusKm <= 0 || radiusKm > 100 {
		radiusKm = 10 // Default to 10km
	}

	query := `
		SELECT v.id, v.name, v.slug, v.short_description, v.address,
			   v.latitude, v.longitude, v.category_id, v.subcategory_id,
			   v.phone, v.website, v.price_range, v.average_rating, v.total_ratings,
			   v.cover_image, v.is_featured,
			   ST_Distance(
				   ST_Point(v.longitude, v.latitude)::geography,
				   ST_Point($1, $2)::geography
			   ) / 1000 AS distance_km,
			   c.name as category_name,
			   city.name as city_name
		FROM venues v
		LEFT JOIN venue_categories c ON v.category_id = c.id
		LEFT JOIN cities city ON v.city_id = city.id
		WHERE v.is_active = true
		  AND ST_DWithin(
			  ST_Point(v.longitude, v.latitude)::geography,
			  ST_Point($1, $2)::geography,
			  $3 * 1000
		  )`

	args := []interface{}{lng, lat, radiusKm}
	argCount := 3

	// Add filters
	if categoryID, exists := filters["category_id"]; exists {
		argCount++
		query += fmt.Sprintf(" AND v.category_id = $%d", argCount)
		args = append(args, categoryID)
	}

	if minRating, exists := filters["min_rating"]; exists {
		argCount++
		query += fmt.Sprintf(" AND v.average_rating >= $%d", argCount)
		args = append(args, minRating)
	}

	if priceRange, exists := filters["price_range"]; exists {
		argCount++
		query += fmt.Sprintf(" AND v.price_range = $%d", argCount)
		args = append(args, priceRange)
	}

	if isOpen, exists := filters["is_open"]; exists && isOpen.(bool) {
		// Add opening hours check (simplified)
		currentHour := time.Now().Hour()
		query += fmt.Sprintf(" AND (v.opening_hours IS NULL OR jsonb_path_exists(v.opening_hours, '$.*.open ? (@ <= \"%d:00\")'))", currentHour)
	}

	query += " ORDER BY distance_km ASC, v.average_rating DESC LIMIT 50"

	rows, err := databases.PostgresDB.Query(query, args...)
	if err != nil {
		sentry.CaptureException(err)
		return nil, err
	}
	defer rows.Close()

	var venues []models.Venue
	for rows.Next() {
		var venue models.Venue
		var distanceKm float64
		var categoryName, cityName sql.NullString
		var subcategoryID sql.NullInt64

		err := rows.Scan(
			&venue.ID, &venue.Name, &venue.Slug, &venue.ShortDesc, &venue.Address,
			&venue.Latitude, &venue.Longitude, &venue.CategoryID, &subcategoryID,
			&venue.Phone, &venue.Website, &venue.PriceRange, &venue.AverageRating, &venue.TotalRatings,
			&venue.CoverImage, &venue.IsFeatured, &distanceKm,
			&categoryName, &cityName,
		)

		if err != nil {
			continue
		}

		venue.Distance = &distanceKm
		if subcategoryID.Valid {
			venue.SubcategoryID = &subcategoryID.Int64
		}

		// Set category and city info
		if categoryName.Valid {
			venue.Category = &models.VenueCategory{
				ID:   venue.CategoryID,
				Name: categoryName.String,
			}
		}
		if cityName.Valid {
			venue.City = &models.City{
				ID:   venue.CityID,
				Name: cityName.String,
			}
		}

		venues = append(venues, venue)
	}

	return &NearbyResult{
		Venues:     venues,
		UserLat:    lat,
		UserLng:    lng,
		Radius:     radiusKm,
		TotalFound: len(venues),
	}, nil
}

// CalculateDistance calculates distance between two points
func (gs *GeolocationService) CalculateDistance(lat1, lng1, lat2, lng2 float64) Distance {
	const R = 6371000 // Earth radius in meters

	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δφ := (lat2 - lat1) * math.Pi / 180
	Δλ := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*
			math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	meters := R * c

	return Distance{
		Meters:     meters,
		Kilometers: meters / 1000,
		Miles:      meters / 1609.344,
	}
}

// GetBounds calculates bounding box for a center point and radius
func (gs *GeolocationService) GetBounds(centerLat, centerLng, radiusKm float64) LocationBounds {
	// Approximate degrees per kilometer
	latDegPerKm := 1.0 / 111.0
	lngDegPerKm := 1.0 / (111.0 * math.Cos(centerLat*math.Pi/180))

	latOffset := radiusKm * latDegPerKm
	lngOffset := radiusKm * lngDegPerKm

	return LocationBounds{
		NorthEast: LatLng{
			Latitude:  centerLat + latOffset,
			Longitude: centerLng + lngOffset,
		},
		SouthWest: LatLng{
			Latitude:  centerLat - latOffset,
			Longitude: centerLng - lngOffset,
		},
	}
}

// FindOptimalMeetingPoint finds the best meeting point for multiple locations
func (gs *GeolocationService) FindOptimalMeetingPoint(locations []LatLng, preferences map[string]interface{}) (*LocationResult, []models.Venue, error) {
	if len(locations) == 0 {
		return nil, nil, fmt.Errorf("no locations provided")
	}

	// Calculate centroid
	var sumLat, sumLng float64
	for _, loc := range locations {
		sumLat += loc.Latitude
		sumLng += loc.Longitude
	}

	centerLat := sumLat / float64(len(locations))
	centerLng := sumLng / float64(len(locations))

	// Calculate maximum distance from center to any point
	maxDistance := 0.0
	for _, loc := range locations {
		distance := gs.CalculateDistance(centerLat, centerLng, loc.Latitude, loc.Longitude)
		if distance.Kilometers > maxDistance {
			maxDistance = distance.Kilometers
		}
	}

	// Add buffer to ensure all points are included
	searchRadius := maxDistance * 1.2
	if searchRadius < 5.0 {
		searchRadius = 5.0 // Minimum 5km radius
	}

	// Get reverse geocoding for the center point
	centerLocation, err := gs.ReverseGeocode(centerLat, centerLng)
	if err != nil {
		centerLocation = &LocationResult{
			Latitude:  centerLat,
			Longitude: centerLng,
			Address:   fmt.Sprintf("%.6f, %.6f", centerLat, centerLng),
		}
	}

	// Find venues near the optimal meeting point
	filters := make(map[string]interface{})
	if categoryID, exists := preferences["category_id"]; exists {
		filters["category_id"] = categoryID
	}
	if minRating, exists := preferences["min_rating"]; exists {
		filters["min_rating"] = minRating
	}

	nearbyResult, err := gs.GetNearbyVenues(centerLat, centerLng, searchRadius, filters)
	var venues []models.Venue
	if err == nil {
		venues = nearbyResult.Venues
	}

	return centerLocation, venues, nil
}

// GetVenuesInBounds finds all venues within a bounding box
func (gs *GeolocationService) GetVenuesInBounds(bounds LocationBounds, filters map[string]interface{}) ([]models.Venue, error) {
	query := `
		SELECT v.id, v.name, v.slug, v.address, v.latitude, v.longitude,
			   v.category_id, v.average_rating, v.total_ratings, v.cover_image
		FROM venues v
		WHERE v.is_active = true
		  AND v.latitude BETWEEN $1 AND $2
		  AND v.longitude BETWEEN $3 AND $4`

	args := []interface{}{
		bounds.SouthWest.Latitude, bounds.NorthEast.Latitude,
		bounds.SouthWest.Longitude, bounds.NorthEast.Longitude,
	}
	argCount := 4

	// Add filters
	if categoryID, exists := filters["category_id"]; exists {
		argCount++
		query += fmt.Sprintf(" AND v.category_id = $%d", argCount)
		args = append(args, categoryID)
	}

	if minRating, exists := filters["min_rating"]; exists {
		argCount++
		query += fmt.Sprintf(" AND v.average_rating >= $%d", argCount)
		args = append(args, minRating)
	}

	query += " ORDER BY v.average_rating DESC LIMIT 100"

	rows, err := databases.PostgresDB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var venues []models.Venue
	for rows.Next() {
		var venue models.Venue
		err := rows.Scan(
			&venue.ID, &venue.Name, &venue.Slug, &venue.Address,
			&venue.Latitude, &venue.Longitude, &venue.CategoryID,
			&venue.AverageRating, &venue.TotalRatings, &venue.CoverImage,
		)
		if err == nil {
			venues = append(venues, venue)
		}
	}

	return venues, nil
}

// Helper methods for geocoding

func (gs *GeolocationService) searchLocalLocations(address string) *LocationResult {
	// Search in our cities database first
	query := `
		SELECT name, latitude, longitude, state, country
		FROM cities
		WHERE LOWER(name) LIKE LOWER($1) OR LOWER(state) LIKE LOWER($1)
		ORDER BY 
		  CASE WHEN LOWER(name) = LOWER($1) THEN 1 ELSE 2 END,
		  name
		LIMIT 1`

	row := databases.PostgresDB.QueryRow(query, "%"+address+"%")

	var name, state, country string
	var lat, lng float64

	err := row.Scan(&name, &lat, &lng, &state, &country)
	if err != nil {
		return nil
	}

	return &LocationResult{
		Address:    fmt.Sprintf("%s, %s, %s", name, state, country),
		Latitude:   lat,
		Longitude:  lng,
		City:       name,
		State:      state,
		Country:    country,
		Confidence: 0.8,
	}
}

func (gs *GeolocationService) reverseGeocodeLocal(lat, lng float64) *LocationResult {
	// Find the nearest city
	query := `
		SELECT name, state, country,
			   ST_Distance(
				   ST_Point($1, $2)::geography,
				   ST_Point(longitude, latitude)::geography
			   ) / 1000 as distance_km
		FROM cities
		ORDER BY distance_km
		LIMIT 1`

	row := databases.PostgresDB.QueryRow(query, lng, lat)

	var name, state, country string
	var distance float64

	err := row.Scan(&name, &state, &country, &distance)
	if err != nil || distance > 50 { // Only if within 50km
		return nil
	}

	return &LocationResult{
		Address:    fmt.Sprintf("Near %s, %s, %s", name, state, country),
		Latitude:   lat,
		Longitude:  lng,
		City:       name,
		State:      state,
		Country:    country,
		Confidence: math.Max(0.1, 1.0-(distance/50.0)),
	}
}

func (gs *GeolocationService) externalGeocode(address string) (*LocationResult, error) {
	// This would integrate with Mapbox, Google Maps, or other geocoding service
	// For now, return a mock result
	return &LocationResult{
		Address:    address,
		Latitude:   0,
		Longitude:  0,
		Confidence: 0.1,
	}, fmt.Errorf("external geocoding not implemented")
}

func (gs *GeolocationService) externalReverseGeocode(lat, lng float64) (*LocationResult, error) {
	// This would integrate with external reverse geocoding service
	return &LocationResult{
		Address:    fmt.Sprintf("%.6f, %.6f", lat, lng),
		Latitude:   lat,
		Longitude:  lng,
		Confidence: 0.1,
	}, fmt.Errorf("external reverse geocoding not implemented")
}

// GetLocationSuggestions provides autocomplete suggestions for locations
func (gs *GeolocationService) GetLocationSuggestions(query string, limit int) ([]LocationResult, error) {
	if limit <= 0 || limit > 20 {
		limit = 10
	}

	dbQuery := `
		SELECT name, state, country, latitude, longitude,
			   similarity(name, $1) + similarity(state, $1) as score
		FROM cities
		WHERE name ILIKE $2 OR state ILIKE $2
		ORDER BY score DESC, name
		LIMIT $3`

	rows, err := databases.PostgresDB.Query(dbQuery, query, "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suggestions []LocationResult
	for rows.Next() {
		var result LocationResult
		var score float64

		err := rows.Scan(
			&result.City, &result.State, &result.Country,
			&result.Latitude, &result.Longitude, &score,
		)
		if err != nil {
			continue
		}

		result.Address = fmt.Sprintf("%s, %s, %s", result.City, result.State, result.Country)
		result.Confidence = math.Min(1.0, score)

		suggestions = append(suggestions, result)
	}

	return suggestions, nil
}
