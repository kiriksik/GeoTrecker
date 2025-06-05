package handlers

import (
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
	}

	err = redis.RBD.Set(redis.Ctx, "user:"+user.Username, user.Password, 0).Err()
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

	hashedPassword, err := redis.RBD.Get(redis.Ctx, "user:"+req.Username).Result()
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not found"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Wrong password"})
	}

	token, err := auth.GenerateToken(req.Username, 24*time.Hour)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Token generation failed"})
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

func Me(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user_id":  userID,
		"username": userID,
	})
}
