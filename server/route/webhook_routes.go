package route

import (
	"github.com/gin-gonic/gin"
	"github.com/refto/server/server/handler"
)

func publicWebHookRoutes(r *gin.RouterGroup) {
	r.POST("hooks/data-pushed/", handler.ImportDataFromRepoByGitHubWebHook)
	r.POST("hooks/pull-request/", handler.ProcessPullRequestActions)
}
