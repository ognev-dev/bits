package handler

import (
	"errors"
	"net/http"

	"github.com/refto/server/service/repository"

	"github.com/gin-gonic/gin"
	"github.com/refto/server/database/model"
	"github.com/refto/server/server/request"
	"github.com/refto/server/server/response"
	"github.com/refto/server/service/collection"
	"github.com/refto/server/service/entity"
	"github.com/refto/server/service/topic"
)

func GetEntities(c *gin.Context) {
	var req request.FilterEntities
	if !bindQuery(c, &req) {
		return
	}

	var (
		// if only one topic selected, there can be a definition for this topic
		// if so return it as first element in response
		definition *model.Entity
		err        error
	)

	if len(req.Topics) == 1 && req.Page < 2 {
		definition, err = entity.Definition(req.Topics[0])
		if err != nil {
			Abort(c, err)
			return
		}
	}

	if req.Collection != 0 {
		col, err := collection.FindByID(req.Collection)
		if err != nil {
			Abort(c, err)
			return
		}

		if col.Private && col.UserID != request.AuthUser(c).ID {
			err = errors.New("unable to display entities from private collection")
			Abort(c, err)
			return
		}
	}

	if req.Repo != 0 {
		repo, err := repository.FindByID(req.Repo)
		if err != nil {
			Abort(c, err)
			return
		}

		if repo.IsPrivate() && repo.UserID != request.AuthUser(c).ID {
			err = errors.New("unable to display entities from private repository")
			Abort(c, err)
			return
		}
	}

	if request.HasAuthUser(c) {
		req.User = request.AuthUser(c).ID
	}

	data, count, err := entity.Filter(req)
	if err != nil {
		Abort(c, err)
		return
	}

	topics, err := topic.Common(topic.CommonTopicsParams{
		Topics:       req.Topics,
		CollectionID: req.Collection,
		RepoID:       req.Repo,
	})
	if err != nil {
		Abort(c, err)
		return
	}

	resp := response.FilterEntities{
		Definition:    definition,
		Entities:      data,
		EntitiesCount: count,
		Topics:        topics,
	}

	c.JSON(http.StatusOK, resp)
}

func GetEntityByID(c *gin.Context) {
	var id int64
	if !BindID(c, &id) {
		return
	}

	e, err := entity.FindByID(id)
	if err != nil {
		Abort(c, err)
		return
	}

	if request.HasAuthUser(c) {
		e.Collections, _, err = collection.Filter(request.FilterCollections{
			UserID:   request.AuthUser(c).ID,
			EntityID: e.ID,
		})
		if err != nil {
			Abort(c, err)
			return
		}
	}

	c.JSON(http.StatusOK, e)
}
