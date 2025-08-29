package main

import (
	"voting-app/app/controllers"
	"voting-app/app/middlewares"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"log"
	"os"
)

func initSentry() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
}

func apiHandler() {
	routes := gin.Default()
	
	// Global middleware
	routes.Use(middlewares.Api())
	routes.Use(middlewares.CORS()) // You'd need to implement this
	
	{
		v1Routes := routes.Group("v1")
		
		// =====================================
		// VENUE DISCOVERY & MANAGEMENT ROUTES
		// =====================================
		
		// Public venue routes (no auth required)
		venueRoutes := v1Routes.Group("/venues")
		{
			venueController := new(controllers.VenueController)
			
			// Search and discovery
			venueRoutes.GET("/search", venueController.Search)
			venueRoutes.GET("/nearby", venueController.GetNearby)
			venueRoutes.GET("/featured", venueController.GetFeatured)
			venueRoutes.GET("/trending", venueController.GetTrending)
			venueRoutes.GET("/categories", venueController.GetCategories)
			
			// Individual venue details
			venueRoutes.GET("/:id", venueController.GetByID)
			venueRoutes.GET("/:id/similar", venueController.GetSimilar)
			venueRoutes.GET("/:id/events", venueController.GetVenueEvents)
			
			// Venue management (requires authentication)
			venueRoutes.Use(middlewares.AuthorizeJWT())
			venueRoutes.POST("/", venueController.CreateVenue)
			venueRoutes.PUT("/:id", venueController.UpdateVenue)
			venueRoutes.DELETE("/:id", venueController.DeleteVenue)
			venueRoutes.POST("/:id/claim", venueController.ClaimVenue)
		}
		
		// =====================================
		// REVIEW & RATING SYSTEM ROUTES
		// =====================================
		
		// Public review routes
		reviewRoutes := v1Routes.Group("/reviews")
		{
			reviewController := new(controllers.ReviewController)
			
			// Get reviews
			reviewRoutes.GET("/trending", reviewController.GetTrendingReviews)
			reviewRoutes.GET("/featured", reviewController.GetFeaturedReviews)
		}
		
		// Venue-specific review routes
		v1Routes.GET("/venues/:venue_id/reviews", controllers.ReviewController{}.GetVenueReviews)
		v1Routes.GET("/venues/:venue_id/reviews/summary", controllers.ReviewController{}.GetReviewSummary)
		
		// User-specific review routes (requires auth)
		userReviewRoutes := v1Routes.Group("/reviews/:snapp_id")
		{
			userReviewRoutes.Use(middlewares.AuthSnappUser())
			reviewController := new(controllers.ReviewController)
			
			// CRUD operations
			userReviewRoutes.POST("/", reviewController.CreateReview)
			userReviewRoutes.GET("/", reviewController.GetUserReviews)
			userReviewRoutes.PUT("/:review_id", reviewController.UpdateReview)
			userReviewRoutes.DELETE("/:review_id", reviewController.DeleteReview)
			
			// Review interactions
			userReviewRoutes.POST("/:review_id/vote", reviewController.VoteReviewHelpful)
			userReviewRoutes.POST("/:review_id/report", reviewController.ReportReview)
		}
		
		// =====================================
		// ENHANCED VOTING CAMPAIGNS
		// =====================================
		
		campaignRoutes := v1Routes.Group("/campaigns")
		{
			campaignController := new(controllers.CampaignController)
			
			// Public campaign routes
			campaignRoutes.GET("/", campaignController.GetActiveCampaigns)
			campaignRoutes.GET("/past", campaignController.GetPastCampaigns)
			campaignRoutes.GET("/:id", campaignController.GetCampaign)
			campaignRoutes.GET("/:id/results", campaignController.GetCampaignResults)
			campaignRoutes.GET("/:id/leaderboard", campaignController.GetLeaderboard)
			
			// Admin routes (requires admin auth)
			campaignRoutes.Use(middlewares.AuthorizeJWT())
			campaignRoutes.POST("/", campaignController.CreateCampaign)
			campaignRoutes.PUT("/:id", campaignController.UpdateCampaign)
			campaignRoutes.DELETE("/:id", campaignController.DeleteCampaign)
		}
		
		// User voting in campaigns
		userCampaignRoutes := v1Routes.Group("/campaigns/:campaign_id/:snapp_id")
		{
			userCampaignRoutes.Use(middlewares.AuthSnappUser())
			campaignController := new(controllers.CampaignController)
			
			userCampaignRoutes.GET("/", campaignController.GetUserCampaignData)
			userCampaignRoutes.POST("/vote/:venue_id", campaignController.SubmitCampaignVote)
			userCampaignRoutes.GET("/votes", campaignController.GetUserVotes)
		}
		
		// =====================================
		// USER COLLECTIONS & LISTS
		// =====================================
		
		collectionRoutes := v1Routes.Group("/collections/:snapp_id")
		{
			collectionRoutes.Use(middlewares.AuthSnappUser())
			collectionController := new(controllers.CollectionController)
			
			// Collection management
			collectionRoutes.GET("/", collectionController.GetUserCollections)
			collectionRoutes.POST("/", collectionController.CreateCollection)
			collectionRoutes.PUT("/:collection_id", collectionController.UpdateCollection)
			collectionRoutes.DELETE("/:collection_id", collectionController.DeleteCollection)
			
			// Collection items
			collectionRoutes.GET("/:collection_id/venues", collectionController.GetCollectionVenues)
			collectionRoutes.POST("/:collection_id/venues", collectionController.AddVenueToCollection)
			collectionRoutes.DELETE("/:collection_id/venues/:venue_id", collectionController.RemoveVenueFromCollection)
		}
		
		// Public collection routes
		v1Routes.GET("/collections/public", controllers.CollectionController{}.GetPublicCollections)
		v1Routes.GET("/collections/:collection_id/public", controllers.CollectionController{}.GetPublicCollection)
		
		// =====================================
		// SOCIAL FEATURES
		// =====================================
		
		socialRoutes := v1Routes.Group("/social/:snapp_id")
		{
			socialRoutes.Use(middlewares.AuthSnappUser())
			socialController := new(controllers.SocialController)
			
			// Check-ins
			socialRoutes.POST("/checkin", socialController.CreateCheckin)
			socialRoutes.GET("/checkins", socialController.GetUserCheckins)
			socialRoutes.GET("/feed", socialController.GetSocialFeed)
			
			// Following
			socialRoutes.POST("/follow/:target_user_id", socialController.FollowUser)
			socialRoutes.DELETE("/follow/:target_user_id", socialController.UnfollowUser)
			socialRoutes.GET("/followers", socialController.GetFollowers)
			socialRoutes.GET("/following", socialController.GetFollowing)
			
			// Recommendations
			socialRoutes.GET("/recommendations", socialController.GetPersonalizedRecommendations)
			socialRoutes.GET("/recommendations/similar-users", socialController.GetSimilarUsers)
		}
		
		// =====================================
		// DISCOVERY & RECOMMENDATIONS
		// =====================================
		
		discoveryRoutes := v1Routes.Group("/discover")
		{
			discoveryController := new(controllers.DiscoveryController)
			
			// Public discovery
			discoveryRoutes.GET("/trending", discoveryController.GetTrending)
			discoveryRoutes.GET("/new", discoveryController.GetNewVenues)
			discoveryRoutes.GET("/popular", discoveryController.GetPopularVenues)
			discoveryRoutes.GET("/by-city/:city_id", discoveryController.GetVenuesByCity)
			discoveryRoutes.GET("/by-category/:category_id", discoveryController.GetVenuesByCategory)
			
			// Personalized discovery (requires auth)
			personalRoutes := discoveryRoutes.Group("/:snapp_id")
			personalRoutes.Use(middlewares.AuthSnappUser())
			personalRoutes.GET("/recommendations", discoveryController.GetPersonalizedRecommendations)
			personalRoutes.GET("/for-you", discoveryController.GetForYouVenues)
			personalRoutes.GET("/based-on-location", discoveryController.GetLocationBasedRecommendations)
			personalRoutes.GET("/similar-to/:venue_id", discoveryController.GetSimilarVenues)
		}
		
		// =====================================
		// EVENTS & PROMOTIONS
		// =====================================
		
		eventRoutes := v1Routes.Group("/events")
		{
			eventController := new(controllers.EventController)
			
			// Public event routes
			eventRoutes.GET("/", eventController.GetEvents)
			eventRoutes.GET("/happening-now", eventController.GetHappeningNow)
			eventRoutes.GET("/today", eventController.GetTodayEvents)
			eventRoutes.GET("/this-week", eventController.GetThisWeekEvents)
			eventRoutes.GET("/:id", eventController.GetEvent)
			
			// Location-based events
			eventRoutes.GET("/nearby", eventController.GetNearbyEvents)
			
			// Venue owner event management (requires auth)
			eventRoutes.Use(middlewares.AuthorizeJWT())
			eventRoutes.POST("/", eventController.CreateEvent)
			eventRoutes.PUT("/:id", eventController.UpdateEvent)
			eventRoutes.DELETE("/:id", eventController.DeleteEvent)
		}
		
		// =====================================
		// ANALYTICS & INSIGHTS
		// =====================================
		
		analyticsRoutes := v1Routes.Group("/analytics")
		{
			analyticsRoutes.Use(middlewares.AuthorizeJWT()) // Admin only
			analyticsController := new(controllers.AnalyticsController)
			
			// Venue analytics
			analyticsRoutes.GET("/venues/:venue_id", analyticsController.GetVenueAnalytics)
			analyticsRoutes.GET("/venues/:venue_id/performance", analyticsController.GetVenuePerformance)
			analyticsRoutes.GET("/venues/top-performing", analyticsController.GetTopPerformingVenues)
			
			// Search analytics
			analyticsRoutes.GET("/search/trends", analyticsController.GetSearchTrends)
			analyticsRoutes.GET("/search/popular-queries", analyticsController.GetPopularQueries)
			
			// User behavior analytics
			analyticsRoutes.GET("/users/engagement", analyticsController.GetUserEngagement)
			analyticsRoutes.GET("/reviews/sentiment", analyticsController.GetReviewSentiment)
			
			// Platform analytics
			analyticsRoutes.GET("/overview", analyticsController.GetPlatformOverview)
			analyticsRoutes.GET("/growth", analyticsController.GetGrowthMetrics)
		}
		
		// =====================================
		// LEGACY ROUTES (BACKWARDS COMPATIBILITY)
		// =====================================
		
		// Keep original voting routes for backwards compatibility
		voteRoutes := v1Routes.Group("/vote/:snapp_id")
		{
			voteRoutes.Use(middlewares.AuthSnappUser())
			voteController := new(controllers.VoteController)
			voteRoutes.GET("/", voteController.Vote)
			voteRoutes.POST("/:voting_id/:vote_id", voteController.SubmitVote)
		}
		
		// File serving
		fileRoutes := v1Routes.Group("/files")
		{
			var fileController controllers.FileController
			fileRoutes.GET("/*file_name", fileController.Serve)
		}
		
		// User authentication
		authRoutes := v1Routes.Group("/auth")
		{
			authController := new(controllers.User)
			authRoutes.POST("register", authController.Register)
			authRoutes.POST("login", authController.Login)
			authRoutes.Use(middlewares.AuthorizeJWT())
			authRoutes.POST("reset-pass", authController.Reset)
			authRoutes.GET("profile", authController.GetProfile)
			authRoutes.PUT("profile", authController.UpdateProfile)
		}
		
		// =====================================
		// UTILITY ROUTES
		// =====================================
		
		utilityRoutes := v1Routes.Group("/utils")
		{
			utilityController := new(controllers.UtilityController)
			
			// Location utilities
			utilityRoutes.GET("/cities", utilityController.GetCities)
			utilityRoutes.GET("/cities/:city_id/districts", utilityController.GetDistricts)
			utilityRoutes.GET("/geocode", utilityController.Geocode)
			utilityRoutes.GET("/reverse-geocode", utilityController.ReverseGeocode)
			
			// Search suggestions
			utilityRoutes.GET("/suggestions", utilityController.GetSearchSuggestions)
			utilityRoutes.GET("/autocomplete", utilityController.GetAutocomplete)
			
			// Health check
			utilityRoutes.GET("/health", utilityController.HealthCheck)
			utilityRoutes.GET("/version", utilityController.GetVersion)
		}
	}

	err := routes.Run()
	if err != nil {
		sentry.CaptureException(err)
	}
}

func main() {
	initSentry()
	apiHandler()
}
