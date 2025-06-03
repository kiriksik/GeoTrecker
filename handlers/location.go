package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"fmt"

	"github.com/kiriksik/GeoTrecker/models"
	"github.com/kiriksik/GeoTrecker/redis"
	"github.com/labstack/echo/v4"
)

const locationTTL = 5 * time.Minute

func PostLocation(c echo.Context) error {
	var loc models.Location
	if err := c.Bind(&loc); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	loc.UpdatedAt = time.Now().UTC()

	jsonData, err := json.Marshal(loc)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Encoding failed"})
	}

	key := fmt.Sprintf("location:%s", loc.UserID)
	err = redis.RBD.Set(redis.Ctx, key, jsonData, locationTTL).Err()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Redis write failed"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
