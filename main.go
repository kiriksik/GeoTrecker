package main

import (
	"fmt"

	"github.com/kiriksik/GeoTrecker/config"
	"github.com/kiriksik/GeoTrecker/handlers"
	"github.com/kiriksik/GeoTrecker/redis"
	"github.com/kiriksik/GeoTrecker/router"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	config.LoadConfig()
	redis.InitRedis()

	e := echo.New()
	e.Use(middleware.CORS())

	e.GET("/ping", func(c echo.Context) error {
		err := redis.RBD.Set(redis.Ctx, "test-key", "hello", 0).Err()
		if err != nil {
			return err
		}

		val, err := redis.RBD.Get(redis.Ctx, "test-key").Result()
		if err != nil {
			return err
		}

		return c.String(200, val)
	})

	addr := fmt.Sprintf(":%s", config.Cfg.AppPort)
	router.InitRoutes(e)
	go handlers.BroadcastLocations()
	e.Logger.Fatal(e.Start(addr))
}
