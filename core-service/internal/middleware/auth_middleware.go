package middleware

import (
	"net/http"
	"strings"

	echo "github.com/labstack/echo/v4"

	"github.com/Djarottosca/finalproject-ftgo-1/pkg/jwt"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/response"
)

func AuthMiddleware(authManager *jwt.AuthManager) echo.MiddlewareFunc {
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

			c.Set("user_id", claims.UserID)
			c.Set("role", claims.Role)

			return next(c)
		}
	}
}

// scope per-user cart, order
func UserIDFromContext(c echo.Context) (int, bool) {
	userID, ok := c.Get("user_id").(int)
	return userID, ok
}
