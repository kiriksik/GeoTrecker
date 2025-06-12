package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/kiriksik/GeoTrecker/auth"
	"github.com/kiriksik/GeoTrecker/models"
	"github.com/kiriksik/GeoTrecker/redis"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func Register(c echo.Context) error {
	var req models.AuthRequest
	log.Println("Try to register")
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	exists, err := redis.RBD.Exists(redis.Ctx, "user:"+req.Username).Result()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Redis error"})
	}
	if exists == 1 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User already exists"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error hashing password"})
	}

	user := models.User{
		ID:       req.Username,
		Username: req.Username,
		Password: string(hashedPassword),
		Role:     "user",
	}

	userJson, err := json.Marshal(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "JSON marshal error"})
	}

	err = redis.RBD.Set(redis.Ctx, "user:"+user.Username, userJson, 0).Err()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Redis save error"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "registered"})
}

func Login(c echo.Context) error {
	var req models.AuthRequest
	log.Println("Try to login")
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	userJson, err := redis.RBD.Get(redis.Ctx, "user:"+req.Username).Result()
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not found"})
	}

	var user models.User
	err = json.Unmarshal([]byte(userJson), &user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "JSON unmarshal error"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Wrong password"})
	}

	token, err := auth.GenerateToken(user.ID, user.Role, 24*time.Hour)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Token generation failed"})
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

func Me(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	role, okRole := c.Get("role").(string)
	if !ok || !okRole || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user_id":  userID,
		"username": userID,
		"role":     role,
	})
}
