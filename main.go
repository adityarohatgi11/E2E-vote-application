package main

import (
	"log"
	"os"
	"voting-app/app/controllers"
	"voting-app/app/middlewares"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
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
	routes.Use(middlewares.Api())

	{
		v1Routes := routes.Group("v1")
		{
			voteRoutes := v1Routes.Group("/vote/:snapp_id")
			{
				voteRoutes.Use(middlewares.AuthSnappUser())
				voteController := new(controllers.VoteController)
				voteRoutes.GET("/", voteController.Vote)
				voteRoutes.POST("/:voting_id/:vote_id", voteController.SubmitVote)
			}
			fileRoutes := v1Routes.Group("/files")
			{
				var fileController controllers.FileController
				fileRoutes.GET("/*file_name", fileController.Serve)
			}
			authRoutes := v1Routes.Group("/auth")
			{
				authController := new(controllers.User)
				authRoutes.POST("register", authController.Register)
				authRoutes.POST("login", authController.Login)
				authRoutes.Use(middlewares.AuthorizeJWT())
				authRoutes.POST("reset-pass", authController.Reset)
			}

		}

	}

	log.Println("Starting HTTP server...")
	err := routes.Run()
	if err != nil {
		log.Printf("Server error: %v", err)
		sentry.CaptureException(err)
		log.Fatal("Failed to start server")
	}
}

func main() {
	log.Println("Starting VoteEngine application...")
	initSentry()
	log.Println("Sentry initialized successfully")
	apiHandler()
}
