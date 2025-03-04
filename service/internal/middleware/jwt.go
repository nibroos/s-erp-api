package middleware

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

// JWTMiddleware is a middleware for JWT authentication
func JWTMiddleware() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		authHeader := ctx.Get("Authorization")
		if authHeader == "" {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Missing or malformed JWT"})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Unexpected signing method")
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid or expired JWT"})
		}

		claims := token.Claims.(jwt.MapClaims)
		ctx.Locals("user", claims)

		return ctx.Next()
	}
}

// GenerateJWT generates a new JWT toke
func GenerateJWT(userID uint, roles []string, permissions []string, bid *uint) (string, error) {
	expiresAt := os.Getenv("JWT_EXPIRES_MINUTE_AT")
	expiredAtInt, err := strconv.Atoi(expiresAt)
	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"user_id":     userID,
		"bid":         bid,
		"roles":       roles,
		"permissions": permissions,
		"exp":         time.Now().Add(time.Minute * time.Duration(expiredAtInt)).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// VerifyJWT verifies a JWT token
func VerifyJWT(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.NewError(fiber.StatusUnauthorized, "Unexpected signing method")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	return token.Claims.(jwt.MapClaims), nil
}

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
