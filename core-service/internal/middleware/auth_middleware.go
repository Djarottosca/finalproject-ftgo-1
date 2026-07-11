package middleware

import (
	"net/http"

	"github.com/Djarottosca/finalproject-ftgo-1/pkg/jwt"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/response"
	"github.com/labstack/echo/v4"
)

func AuthMiddleware(authManager *jwt.AuthManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenString := c.Request().Header.Get("Authorization")
			if tokenString == "" {
				return response.ErrorResponse(c, http.StatusUnauthorized, "token tidak ditemukan")
			}

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
