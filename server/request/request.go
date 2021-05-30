package request

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/refto/server/database/model"
	"github.com/refto/server/errors"
)

var ErrInvalidOwnerDefault = errors.Forbidden("Hey, that's not belongs to you")

type Pagination struct {
	Page  int `json:"page,omitempty" form:"page"`
	Limit int `json:"limit,omitempty" form:"limit"`
}

type NoValidation struct{}

func (r *NoValidation) Validate(*gin.Context) (err error) {
	return
}

const ctxUserKey = "request_user"
const ctxClientKey = "request_client"

func HasAuthUser(c *gin.Context) (ok bool) {
	_, ok = c.Get(ctxUserKey)
	return
}

func AuthUserOf(c *gin.Context, userID int64) bool {
	if !HasAuthUser(c) {
		return false
	}
	if AuthUser(c).ID == userID {
		return true
	}
	return false
}

func InvalidUser(c *gin.Context, userID int64, errOpt ...error) bool {
	if !AuthUserOf(c, userID) {
		err := ErrInvalidOwnerDefault
		if len(errOpt) == 1 {
			err = errOpt[0]
		}

		c.JSON(http.StatusForbidden, err.Error())
		c.Abort()
		return true
	}

	return false
}

func AuthUser(c *gin.Context) model.User {
	u := c.MustGet(ctxUserKey)
	return u.(model.User)
}

func SetAuthUser(c *gin.Context, u model.User) {
	c.Set(ctxUserKey, u)
}

func Client(c *gin.Context) string {
	return c.GetString(ctxClientKey)
}

func SetClient(c *gin.Context) {
	name := c.Request.Header.Get("X-Client")
	c.Set(ctxClientKey, name)
}
