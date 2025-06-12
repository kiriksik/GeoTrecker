package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/kiriksik/GeoTrecker/models"
	"github.com/kiriksik/GeoTrecker/redis"
	"github.com/labstack/echo/v4"
)

func AdminGetUsers(c echo.Context) error {
	keys, err := redis.RBD.Keys(redis.Ctx, "user:*").Result()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Redis error"})
	}

	users := []map[string]string{}
	for _, key := range keys {
		val, err := redis.RBD.Get(redis.Ctx, key).Result()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Redis error"})
		}
		var user models.User
		json.Unmarshal([]byte(val), &user)
		users = append(users, map[string]string{
			"username": user.Username,
			"role":     user.Role,
		})

	}

	return c.JSON(http.StatusOK, users)
}

func AdminDeleteUser(c echo.Context) error {
	userID := c.Param("user_id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User ID required"})
	}

	err := redis.RBD.Del(redis.Ctx, "user:"+userID).Err()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Redis delete error"})
	}
	redis.RBD.Del(redis.Ctx, "history:"+userID)
	return c.JSON(http.StatusOK, map[string]string{"status": "User deleted"})
}

func AdminGetUserHistory(c echo.Context) error {
	userID := c.Param("user_id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User ID required"})
	}

	data, err := redis.RBD.LRange(redis.Ctx, "history:"+userID, 0, -1).Result()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Redis error"})
	}

	return c.JSON(http.StatusOK, data)
}
