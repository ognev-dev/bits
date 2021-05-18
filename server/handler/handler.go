package handler

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v9"
	"github.com/refto/server/config"
	se "github.com/refto/server/server/error"
	"github.com/refto/server/server/response"
	log "github.com/sirupsen/logrus"
)

type Validatable interface {
	Validate(c *gin.Context) error
}

func abort400(c *gin.Context, err error) {
	Abort(c, se.New400(err.Error()))
}

// Comment due to linting (unused)
//func abort422(c *gin.Context, err error) {
//	Abort(c, se.New422(err.Error()))
//}

func Abort(c *gin.Context, err error) {
	resp := response.Error{}
	code := http.StatusInternalServerError

	switch e := err.(type) {
	case se.Error:
		code = e.Code
		resp.Error = e.Error()
	case se.List:
		resp.Errors = err.(se.List)
	case se.Input:
		resp.InputErrors = err.(se.Input)
		code = http.StatusUnprocessableEntity
	default:
		resp.Error = err.Error()
		switch err {
		case pg.ErrNoRows:
			code = http.StatusNotFound
		}
	}

	log.Println(err.Error())
	if !config.IsReleaseEnv() {
		log.Println(string(debug.Stack()))
	}

	c.AbortWithStatusJSON(code, resp)
}

func bindJSON(c *gin.Context, req Validatable) (ok bool) {
	if err := c.ShouldBindJSON(req); err != nil {
		abort400(c, err)
		return
	}

	return validRequest(c, req)
}

func bindQuery(c *gin.Context, req Validatable) (ok bool) {
	if err := c.ShouldBindQuery(req); err != nil {
		abort400(c, err)
		return
	}

	return validRequest(c, req)
}

func validRequest(c *gin.Context, req Validatable) (ok bool) {
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
		abort400(c, errors.New("ID param missing from request"))
		return
	}

	intVal, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		abort400(c, errors.New("ID param must be int64"))
		return
	}

	if intVal < 1 {
		abort400(c, errors.New("invalid ID"))
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
