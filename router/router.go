package router

import (
	"github.com/kiriksik/GeoTrecker/auth"
	"github.com/kiriksik/GeoTrecker/handlers"
	"github.com/labstack/echo/v4"
)

func InitRoutes(e *echo.Echo) {

	e.POST("/api/login", handlers.Login)
	e.POST("/api/register", handlers.Register)

	e.GET("/ws", handlers.WsHandler)

	gr := e.Group("/api")
	gr.Use(auth.Middleware)
	{
		gr.POST("/location", handlers.PostLocation)
		gr.GET("/location/:user_id", handlers.GetLocation)
		gr.GET("/history/:user_id", handlers.GetLocationHistory)
		gr.GET("/active", handlers.GetActiveUsers)
		gr.GET("/nearby", handlers.GetNearbyUsers)
		gr.GET("/geojson/:user_id", handlers.GetGeoJSONHistory)
		gr.GET("/movement/:user_id", handlers.GetMovementInfo)
		gr.GET("/me", handlers.Me)
	}

	adminGroup := gr.Group("/admin")
	adminGroup.Use(auth.Middleware, auth.AdminMiddleware)
	{
		adminGroup.GET("/users", handlers.AdminGetUsers)
		adminGroup.DELETE("/users/:user_id", handlers.AdminDeleteUser)
		adminGroup.GET("/users/:user_id/history", handlers.AdminGetUserHistory)

	}
}
