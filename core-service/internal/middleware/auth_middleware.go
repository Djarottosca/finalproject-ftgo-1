package middleware

import (
	"context"
	"net/http"
	"strings"

	echo "github.com/labstack/echo/v4"

	"github.com/Djarottosca/finalproject-ftgo-1/pkg/jwt"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/response"
)

// blacklist is the subset of cache.TokenBlacklist this middleware needs —
// kept as an interface so middleware doesn't import the cache package
// directly and stays easy to test.
type blacklist interface {
	Contains(ctx context.Context, token string) (bool, error)
}

func AuthMiddleware(authManager *jwt.AuthManager, bl blacklist) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenString := c.Request().Header.Get("Authorization")
			if tokenString == "" {
				return response.ErrorResponse(c, http.StatusUnauthorized, "token tidak ditemukan")
			}

			// Check token that have 'Bearer' prefix
			if !strings.HasPrefix(tokenString, "Bearer ") {
				return response.ErrorResponse(c, http.StatusUnauthorized, "token tidak valid")
			}
			tokenString = strings.TrimPrefix(tokenString, "Bearer ")

			claims, err := authManager.VerifyToken(tokenString)
			if err != nil {
				return response.ErrorResponse(c, http.StatusUnauthorized, "token tidak valid")
			}

			if revoked, err := bl.Contains(c.Request().Context(), tokenString); err == nil && revoked {
				return response.ErrorResponse(c, http.StatusUnauthorized, "token sudah logout")
			}

			c.Set("user_id", claims.UserID)
			c.Set("role", claims.Role)
			c.Set("raw_token", tokenString)

			return next(c)
		}
	}
}

// scope per-user cart, order
func UserIDFromContext(c echo.Context) (int, bool) {
	userID, ok := c.Get("user_id").(int)
	return userID, ok
}
