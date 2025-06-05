package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kiriksik/GeoTrecker/models"
	"github.com/kiriksik/GeoTrecker/utils"
	redis_pac "github.com/redis/go-redis/v9"

	"github.com/kiriksik/GeoTrecker/redis"
	"github.com/labstack/echo/v4"
)

const locationTTL = 5 * time.Minute

func PostLocation(c echo.Context) error {
	var loc models.Location
	if err := c.Bind(&loc); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
	}

	_, err := redis.RBD.GeoAdd(redis.Ctx, "locations", &redis_pac.GeoLocation{
		Name:      loc.UserID,
		Longitude: loc.Longitude,
		Latitude:  loc.Latitude,
	}).Result()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Redis GEOADD failed"})
	}

	loc.UpdatedAt = time.Now()
	locJSON, _ := json.Marshal(loc)
	redis.RBD.Set(redis.Ctx, "location_data:"+loc.UserID, locJSON, locationTTL)

	historyKey := "location_history:" + loc.UserID
	redis.RBD.RPush(redis.Ctx, historyKey, locJSON)
	redis.RBD.LTrim(redis.Ctx, historyKey, -100, -1)

	broadcast <- locJSON

	return c.JSON(http.StatusOK, map[string]string{"status": "location stored"})
}

func GetLocation(c echo.Context) error {
	userID := c.Param("user_id")
	val, err := redis.RBD.Get(redis.Ctx, "location_data:"+userID).Result()

	if err == redis.NilError {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Location not found"})
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Redis read error"})
	}

	var loc models.Location
	if err := json.Unmarshal([]byte(val), &loc); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Decoding error"})
	}

	return c.JSON(http.StatusOK, loc)
}

func GetLocationHistory(c echo.Context) error {
	userID := c.Param("user_id")
	historyKey := "location_history:" + userID

	values, err := redis.RBD.LRange(redis.Ctx, historyKey, 0, -1).Result()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Redis LRange failed"})
	}

	var locations []models.Location
	for _, val := range values {
		var loc models.Location
		if err := json.Unmarshal([]byte(val), &loc); err == nil {
			locations = append(locations, loc)
		}
	}

	return c.JSON(http.StatusOK, locations)
}

func GetActiveUsers(c echo.Context) error {
	var activeUsers []string

	iter := redis.RBD.Scan(redis.Ctx, 0, "location_data:*", 100).Iterator()
	for iter.Next(redis.Ctx) {
		key := iter.Val()

		ttl, err := redis.RBD.TTL(redis.Ctx, key).Result()
		if err != nil {
			continue
		}
		if ttl <= 0 {
			continue
		}

		userID := strings.TrimPrefix(key, "location_data:")
		if userID != "" {
			activeUsers = append(activeUsers, userID)
		}
	}

	if err := iter.Err(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Redis iteration failed"})
	}

	return c.JSON(http.StatusOK, map[string][]string{"users": activeUsers})
}

func GetNearbyUsers(c echo.Context) error {
	latStr := c.QueryParam("lat")
	lonStr := c.QueryParam("lon")
	radiusStr := c.QueryParam("radius")

	var lat, lon, radiusKm float64
	lat, err1 := strconv.ParseFloat(latStr, 64)
	lon, err2 := strconv.ParseFloat(lonStr, 64)
	radiusKm, err3 := strconv.ParseFloat(radiusStr, 64)

	if err1 != nil || err2 != nil || err3 != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid query parameters"})
	}

	results, err := redis.RBD.GeoSearchLocation(redis.Ctx, "locations", &redis_pac.GeoSearchLocationQuery{
		GeoSearchQuery: redis_pac.GeoSearchQuery{
			Longitude:  lon,
			Latitude:   lat,
			Radius:     radiusKm,
			RadiusUnit: "km",
		},
		WithCoord: true,
		WithDist:  true,
	}).Result()

	if err != nil {
		log.Printf("GeoSearchLocation error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Redis GEOSEARCH failed"})
	}

	var response []map[string]interface{}

	for _, res := range results {
		response = append(response, map[string]interface{}{
			"user_id":     res.Name,
			"lat":         res.Latitude,
			"lon":         res.Longitude,
			"distance_km": res.Dist,
		})
	}

	return c.JSON(http.StatusOK, response)
}

func GetGeoJSONHistory(c echo.Context) error {
	userID := c.Param("user_id")
	historyKey := "location_history:" + userID

	values, err := redis.RBD.LRange(redis.Ctx, historyKey, 0, -1).Result()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get history"})
	}

	var features []models.Feature

	for _, item := range values {
		var loc models.Location
		if err := json.Unmarshal([]byte(item), &loc); err != nil {
			continue
		}

		f := models.Feature{
			Type: "Feature",
		}
		f.Properties.Timestamp = loc.UpdatedAt.Format(time.RFC3339)
		f.Geometry.Type = "Point"
		f.Geometry.Coordinates = []float64{loc.Longitude, loc.Latitude}

		features = append(features, f)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"type":     "FeatureCollection",
		"features": features,
	})
}

func GetMovementInfo(c echo.Context) error {
	userID := c.Param("user_id")
	historyKey := "location_history:" + userID

	values, err := redis.RBD.LRange(redis.Ctx, historyKey, -2, -1).Result()
	if err != nil || len(values) < 2 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Not enough data"})
	}

	var prev, curr models.Location
	if err := json.Unmarshal([]byte(values[0]), &prev); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Invalid previous point"})
	}
	if err := json.Unmarshal([]byte(values[1]), &curr); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Invalid current point"})
	}

	distanceKm := utils.Haversine(prev.Latitude, prev.Longitude, curr.Latitude, curr.Longitude)
	direction := utils.Bearing(prev.Latitude, prev.Longitude, curr.Latitude, curr.Longitude)
	timeDiff := curr.UpdatedAt.Sub(prev.UpdatedAt).Seconds()

	var speedKph float64
	if timeDiff > 0 {
		speedKph = (distanceKm / timeDiff) * 3600
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user_id":     userID,
		"from":        prev,
		"to":          curr,
		"distance_km": distanceKm,
		"direction":   direction,
		"speed_kph":   speedKph,
	})
}
