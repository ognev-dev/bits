package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/refto/server/server/request"
	"github.com/refto/server/server/response"
	"github.com/refto/server/service/topic"
)

func GetDataImportLogs(c *gin.Context) {
	var req request.FilterTopics
	if !bindQuery(c, &req) {
		return
	}

	data, count, err := topic.Filter(req)
	if err != nil {
		Abort(c, err)
		return
	}

	resp := response.FilterTopics{
		Data:  data,
		Count: count,
	}

	c.JSON(http.StatusOK, resp)
}
