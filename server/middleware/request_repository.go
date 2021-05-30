package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/refto/server/database/model"
	"github.com/refto/server/server/handler"
	"github.com/refto/server/server/request"
	"github.com/refto/server/service/repository"
)

func RequestRepository(idParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var elem model.Repository
		var err error
		if idParam == "path" {
			// path param is composite of account and name, so little trick here
			path := c.Param("account") + "/" + c.Param("name")
			elem, err = repository.FindByPath(path)
		} else {
			var id int64
			if !handler.BindID(c, &id, idParam) {
				return
			}
			elem, err = repository.FindByID(id)
		}
		if err != nil {
			handler.Abort(c, err)
			return
		}

		request.SetRepository(c, elem)
		c.Next()
	}
}
