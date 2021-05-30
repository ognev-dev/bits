package request

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/refto/server/database/model"
	"github.com/refto/server/errors"
)

const ctxCollectionKey = "collection_model"

func Collection(c *gin.Context) model.Collection {
	elem := c.MustGet(ctxCollectionKey)
	return elem.(model.Collection)
}

func SetCollection(c *gin.Context, m model.Collection) {
	c.Set(ctxCollectionKey, m)
}

type FilterCollections struct {
	NoValidation
	Pagination
	Name              string `json:"name" form:"name"`
	UserID            int64  `json:"-" form:"-"`
	EntityID          int64  `json:"entity_id" form:"entity_id"`
	Available         bool   `json:"available" form:"available"`
	WithEntitiesCount bool   `json:"wec" form:"wec"`
}

type CreateCollection struct {
	Name    string `json:"name"`
	Private bool   `json:"private"`
}

func (r *CreateCollection) Validate(*gin.Context) (err error) {
	if strings.TrimSpace(r.Name) == "" {
		return errors.Input{
			"name": "Name is required",
		}
	}

	return
}

func (r *CreateCollection) ToModel(c *gin.Context) (m model.Collection) {
	return model.Collection{
		Name:    r.Name,
		Private: r.Private,
		UserID:  AuthUser(c).ID,
	}
}

type UpdateCollection struct {
	Name    string `json:"name"`
	Private bool   `json:"private"`
}

func (r UpdateCollection) Validate(c *gin.Context) (err error) {
	if AuthUser(c).ID != Collection(c).UserID {
		err = errors.Unprocessable("invalid user")
		return
	}

	if strings.TrimSpace(r.Name) == "" {
		return errors.Input{
			"name": "Name is required",
		}
	}

	return
}

func (r UpdateCollection) ToModel(m *model.Collection) {
	m.Name = r.Name
	m.Private = r.Private
}

type DeleteCollection struct{}

func (r DeleteCollection) Validate(c *gin.Context) (err error) {
	if AuthUser(c).ID != Collection(c).UserID {
		err = errors.Unprocessable("you can't simply delete collection that is not created by you")
		return
	}

	return
}

type AddEntityToCollection struct{}

func (r AddEntityToCollection) Validate(c *gin.Context) (err error) {
	if AuthUser(c).ID != Collection(c).UserID {
		err = errors.Unprocessable("unable to add entity to collection: Ownership violation detected. ")
		return
	}

	return
}

type RemoveEntityFromCollection struct{}

func (r RemoveEntityFromCollection) Validate(c *gin.Context) (err error) {
	if AuthUser(c).ID != Collection(c).UserID {
		err = errors.Unprocessable("unable to remove entity from collection: Owner of collection will not be happy")
		return
	}

	return
}
