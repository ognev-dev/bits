package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/refto/server/server/request"
	"github.com/refto/server/server/response"
	"github.com/refto/server/service/collection"
)

func GetCollections(c *gin.Context) {
	var req request.FilterCollections
	if !bindQuery(c, &req) {
		return
	}

	req.UserID = request.AuthUser(c).ID
	data, count, err := collection.Filter(req)
	if err != nil {
		Abort(c, err)
		return
	}

	c.JSON(http.StatusOK, response.FilterCollections{
		Data:  data,
		Count: count,
	})
}

func GetCollectionByToken(c *gin.Context) {
	c.JSON(http.StatusOK, request.Collection(c))
}

func CreateCollection(c *gin.Context) {
	var req request.CreateCollection
	if !bindJSON(c, &req) {
		return
	}

	elem := req.ToModel(c)
	err := collection.Create(&elem)
	if err != nil {
		Abort(c, err)
		return
	}

	c.JSON(http.StatusCreated, elem)
}

func UpdateCollection(c *gin.Context) {
	var req request.UpdateCollection
	if !bindJSON(c, &req) {
		return
	}

	elem := request.Collection(c)
	req.ToModel(&elem)
	err := collection.Update(&elem)
	if err != nil {
		Abort(c, err)
		return
	}

	c.JSON(http.StatusOK, elem)
}

func DeleteCollection(c *gin.Context) {
	if !validRequest(c, request.DeleteCollection{}) {
		return
	}

	err := collection.Delete(request.Collection(c).ID)
	if err != nil {
		Abort(c, err)
		return
	}

	c.JSON(http.StatusOK, response.OK("Collection deleted"))
}

func AddEntityToCollection(c *gin.Context) {
	if !validRequest(c, request.AddEntityToCollection{}) {
		return
	}

	collectionID := request.Collection(c).ID
	entityID := request.Entity(c).ID

	err := collection.AddEntity(collectionID, entityID)
	if err != nil {
		Abort(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.OK("Entity added to collection"))
}

func RemoveEntityFromCollection(c *gin.Context) {
	if !validRequest(c, request.RemoveEntityFromCollection{}) {
		return
	}

	collectionID := request.Collection(c).ID
	entityID := request.Entity(c).ID

	err := collection.RemoveEntity(collectionID, entityID)
	if err != nil {
		Abort(c, err)
		return
	}

	c.JSON(http.StatusOK, response.OK("Entity removed from collection"))
}
