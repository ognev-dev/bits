package middleware

import (
	"net/http"

	"github.com/refto/server/errors"

	"github.com/gin-gonic/gin"
	"github.com/refto/server/server/handler"
)

// RequestAuth middleware
// See RequestUser middleware for auth details
func RequestAuth(c *gin.Context) {
	if !c.GetBool(authSuccessKey) {
		errText := c.GetString(authErrorKey)
		if errText == "" {
			errText = http.StatusText(http.StatusUnauthorized)
		}

		handler.Abort(c, errors.Unauthorized(errText))
		return
	}

	c.Next()
}
