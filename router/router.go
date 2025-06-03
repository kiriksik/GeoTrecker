package router

import (
	"github.com/kiriksik/GeoTrecker/handlers"
	"github.com/labstack/echo/v4"
)

func InitRoutes(e *echo.Echo) {
	gr := e.Group("/api")
	{
		gr.POST("/location", handlers.PostLocation)
	}
}
