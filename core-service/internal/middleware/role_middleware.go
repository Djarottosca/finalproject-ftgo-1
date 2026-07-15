package middleware

import (
	"net/http"

	echo "github.com/labstack/echo/v4"

	"github.com/Djarottosca/finalproject-ftgo-1/pkg/response"
)

// RequireRole restricts a route to the given roles. Must run after
// AuthMiddleware, which sets "role" in the echo context from the JWT claims.
func RequireRole(roles ...string) echo.MiddlewareFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, _ := c.Get("role").(string)
			if _, ok := allowed[role]; !ok {
				return response.ErrorResponse(c, http.StatusForbidden, "akses ditolak untuk role ini")
			}
			return next(c)
		}
	}
}
