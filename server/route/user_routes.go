package route

import (
	"github.com/gin-gonic/gin"
	"github.com/refto/server/server/handler"
)

func publicUserRoutes(r *gin.RouterGroup) {
	r.POST("user/login/", handler.LoginWithGithub)
}

func userRoutes(r *gin.RouterGroup) {
	r.GET("user/repositories/", handler.GetUserRepositories)
}
