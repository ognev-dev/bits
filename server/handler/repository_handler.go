package handler

import (
	"net/http"

	dataimport "github.com/refto/server/service/data_import"

	"github.com/refto/server/util"

	"github.com/refto/server/database/model"

	"github.com/refto/server/service/repository"

	"github.com/gin-gonic/gin"
	"github.com/refto/server/server/request"
	"github.com/refto/server/server/response"
)

func GetPublicRepositories(c *gin.Context) {
	var req request.FilterRepositories
	if !bindQuery(c, &req) {
		return
	}

	req.Types = []model.RepoType{
		model.RepoTypeGlobal,
		model.RepoTypePublic,
	}
	req.Confirmed = util.NewBool(true)
	data, count, err := repository.Filter(req)
	if err != nil {
		Abort(c, err)
		return
	}

	c.JSON(http.StatusOK, response.FilterRepositories{
		Data:  data,
		Count: count,
	})
}

func GetUserRepositories(c *gin.Context) {
	var req request.FilterRepositories
	if !bindQuery(c, &req) {
		return
	}

	req.UserID = request.AuthUser(c).ID
	data, count, err := repository.Filter(req)
	if err != nil {
		Abort(c, err)
		return
	}

	c.JSON(http.StatusOK, response.FilterRepositories{
		Data:  data,
		Count: count,
	})
}

func GetRepositoryByPath(c *gin.Context) {
	repo := request.Repository(c)

	if repo.Type == model.RepoTypePrivate && !request.AuthUserOf(c, repo.UserID) {
		Abort(c, repository.ErrRepoNotFoundByPath)
		return
	}

	c.JSON(http.StatusOK, repo)
}

func CreateRepository(c *gin.Context) {
	var req request.CreateRepository
	if !bindJSON(c, &req) {
		return
	}

	elem := req.ToModel(c)
	secret, err := repository.Create(&elem)
	if err != nil {
		Abort(c, err)
		return
	}

	resp := response.CreateRepository{
		Secret: secret,
	}

	c.JSON(http.StatusCreated, resp)
}

func UpdateRepository(c *gin.Context) {
	var req request.UpdateRepository
	if !bindJSON(c, &req) {
		return
	}

	m := request.Repository(c)
	req.ToModel(&m)
	err := repository.Update(&m)
	if err != nil {
		Abort(c, err)
		return
	}

	c.JSON(http.StatusOK, m)
}

func GetNewRepositorySecret(c *gin.Context) {
	repo := request.Repository(c)
	if request.InvalidUser(c, repo.ID) {
		return
	}

	secret, err := repository.NewSecret(repo.ID)
	if err != nil {
		Abort(c, err)
		return
	}

	resp := response.CreateRepository{
		Secret: secret,
	}

	c.JSON(http.StatusOK, resp)
}

func ImportRepository(c *gin.Context) {
	repo := request.Repository(c)
	if request.InvalidUser(c, repo.ID) {
		return
	}

	// only confirmed repositories can be imported on demand
	if !repo.Confirmed {
		err := repository.ErrNotConfirmed
		Abort(c, err)
		return
	}

	err := dataimport.FromGitHub(repo)
	if err != nil {
		Abort(c, err)
	}

	c.JSON(http.StatusOK, response.OK("Data from repository successfully imported"))
}

func DeleteRepository(c *gin.Context) {
	if !validRequest(c, request.DeleteRepository{}) {
		return
	}

	err := repository.Delete(request.Repository(c).ID)
	if err != nil {
		Abort(c, err)
		return
	}

	c.JSON(http.StatusOK, response.OK("Repository deleted"))
}
