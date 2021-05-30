package route

import (
	"github.com/gin-gonic/gin"
	"github.com/refto/server/server/handler"
	"github.com/refto/server/server/middleware"
)

func publicRepositoryRoutes(r *gin.RouterGroup) {
	r.Group("repositories/").
		GET("/", handler.GetPublicRepositories)

	r.Group("repositories/:account/:name").
		Use(middleware.RequestRepository("path")).
		GET("/", handler.GetRepositoryByPath)
}

func repositoryRoutes(r *gin.RouterGroup) {
	r.Group("repositories/").
		POST("/", handler.CreateRepository)

	r.Group("repositories/:id/").
		Use(middleware.RequestRepository("id")).
		PUT("/", handler.UpdateRepository).
		DELETE("/", handler.DeleteRepository).
		POST("/secret/", handler.GetNewRepositorySecret).
		POST("/import/", handler.ImportRepository)

}
