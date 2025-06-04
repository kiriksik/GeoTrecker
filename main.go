package main

import (
	"fmt"

	"github.com/kiriksik/GeoTrecker/config"
	"github.com/kiriksik/GeoTrecker/redis"
	"github.com/kiriksik/GeoTrecker/router"
	"github.com/labstack/echo/v4"
)

func main() {
	config.LoadConfig()
	redis.InitRedis()

	e := echo.New()

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
	e.Logger.Fatal(e.Start(addr))
}
