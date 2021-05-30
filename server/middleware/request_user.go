package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/refto/server/server/handler"
	"github.com/refto/server/server/request"
	authtoken "github.com/refto/server/service/auth_token"
	"github.com/refto/server/service/user"
)

const authSuccessKey = "auth_success"
const authErrorKey = "auth_error"

// RequestUser Middleware
// Attempt to resolve user by given auth token
// Because some public routes response depends on current user
// Authorization also happens here, because it is required to check that auth token is valid,
// but auth errors not thrown here, because this middleware should only resolve that user is authorized
// Instead auth errors placed in request context for later use by auth middleware if needed
func RequestUser(c *gin.Context) {
	sig := c.Request.Header.Get("Authorization")
	if sig == "" {
		c.Next()
		return
	}

	sig = strings.TrimPrefix(sig, "Bearer ")
	if len(sig) < 128 {
		c.Set(authErrorKey, "invalid auth token")
		c.Next()
		return
	}

	token := sig[:64]
	tokenEl, err := authtoken.FindByToken(token)
	if err != nil {
		c.Set(authErrorKey, "unknown auth token")
		c.Next()
		return
	}

	tokenEl.ClientName = request.Client(c)
	tokenEl.ClientIP = c.Request.RemoteAddr
	tokenEl.UserAgent = c.Request.UserAgent()

	if sig != authtoken.Sign(&tokenEl) {
		c.Set(authErrorKey, "invalid auth token signature")
		c.Next()
		return
	}

	if tokenEl.ExpiresAt.Before(time.Now()) {
		c.Set(authErrorKey, "auth token expired")
		c.Next()
		return
	}

	userEl := tokenEl.User
	err = user.SetActiveAt(userEl.ID)
	if err != nil {
		handler.Abort(c, err)
		return
	}

	err = authtoken.Prolong(tokenEl.ID)
	if err != nil {
		handler.Abort(c, err)
		return
	}

	c.Set(authSuccessKey, true)
	request.SetAuthUser(c, *userEl)
	c.Next()
}
