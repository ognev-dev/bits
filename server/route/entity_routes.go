package route

import (
	"github.com/gin-gonic/gin"
	"github.com/refto/server/server/handler"
)

func publicEntityRoutes(r *gin.RouterGroup) {
	r.GET("entities/", handler.GetEntities)
	r.GET("entities/:id/", handler.GetEntityByID)
}
