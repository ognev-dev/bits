package handler

import (
	"bytes"
	"io/ioutil"
	"runtime/debug"
	"strconv"

	"github.com/refto/server/errors"

	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v9"
	"github.com/refto/server/server/response"
	log "github.com/sirupsen/logrus"
)

var (
	ErrIDParamMissingFromRequest = errors.BadRequest("ID param missing from request")
	ErrIDParamMustBeInt64        = errors.BadRequest("ID param must be positive int64")
)

type Validator interface {
	Validate(c *gin.Context) error
}

func Abort(c *gin.Context, err error) {
	resp := response.Error{
		Code: errors.CodeInternal,
	}

	switch e := err.(type) {
	case errors.Error:
		resp.Code = e.Code
		resp.Error = e.Error()
	case errors.Input:
		resp.Code = errors.CodeUnprocessable
		resp.InputErrors = e
	default:
		resp.Error = err.Error()
		switch err {
		case pg.ErrNoRows:
			resp.Code = errors.CodeNotFound
		}
	}

	if resp.Code >= errors.CodeInternal {
		log.Println(string(debug.Stack()))
	}

	c.AbortWithStatusJSON(resp.Code, resp)
}

func bindJSON(c *gin.Context, req Validator) (ok bool) {
	if err := c.ShouldBindJSON(req); err != nil {
		Abort(c, errors.BadRequest(err.Error()))
		return
	}

	return validRequest(c, req)
}

func bindQuery(c *gin.Context, req Validator) (ok bool) {
	if err := c.ShouldBindQuery(req); err != nil {
		Abort(c, errors.BadRequest(err.Error()))
		return
	}

	return validRequest(c, req)
}

func validRequest(c *gin.Context, req Validator) (ok bool) {
	if err := req.Validate(c); err != nil {
		Abort(c, err)
		return
	}

	return true
}

func BindID(c *gin.Context, id *int64, paramNameOpt ...string) (ok bool) {
	name := "id"
	if len(paramNameOpt) == 1 {
		name = paramNameOpt[0]
	}

	val := c.Param(name)
	if val == "" {
		Abort(c, ErrIDParamMissingFromRequest)
		return
	}

	intVal, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		Abort(c, ErrIDParamMustBeInt64)
		return
	}

	if intVal < 1 {
		Abort(c, ErrIDParamMustBeInt64)
		return
	}

	*id = intVal
	return true
}

func copyRequestBody(c *gin.Context) (body []byte, err error) {
	body, err = ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return
	}
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body)) // Write body back

	return
}
