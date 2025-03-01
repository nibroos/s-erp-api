package auth

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

// GetAuthUser extracts and returns the authenticated user data from the JWT token
func GetAuthUser(ctx *fiber.Ctx) (jwt.MapClaims, error) {
	authHeader := ctx.Get("Authorization")
	if authHeader == "" {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Missing or malformed JWT")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.NewError(fiber.StatusUnauthorized, "Unexpected signing method")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid or expired JWT")
	}

	claims := token.Claims.(jwt.MapClaims)

	return claims, nil
}
