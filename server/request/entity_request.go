package request

import (
	"github.com/gin-gonic/gin"
	"github.com/refto/server/database/model"
)

const ctxEntityKey = "entity_model"

func Entity(c *gin.Context) model.Entity {
	elem := c.MustGet(ctxEntityKey)
	return elem.(model.Entity)
}

func SetEntity(c *gin.Context, m model.Entity) {
	c.Set(ctxEntityKey, m)
}

type FilterEntities struct {
	NoValidation
	Pagination
	Topics []string `json:"topics,omitempty" form:"topics"`
	Addr   string   `json:"addr" form:"addr"`
	Name   string   `json:"name" form:"name"`
	Query  string   `json:"query" form:"query"`
}
