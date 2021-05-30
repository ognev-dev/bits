package route

import (
	"github.com/gin-gonic/gin"
	"github.com/refto/server/server/handler"
)

func publicTopicRoutes(r *gin.RouterGroup) {
	r.GET("topics/", handler.GetTopics)
}
