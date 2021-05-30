package request

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/refto/server/database/model"
	"github.com/refto/server/errors"
	"github.com/refto/server/util"
)

const ctxRepositoryKey = "repository_model"

func Repository(c *gin.Context) model.Repository {
	elem := c.MustGet(ctxRepositoryKey)
	return elem.(model.Repository)
}

func SetRepository(c *gin.Context, m model.Repository) {
	c.Set(ctxRepositoryKey, m)
}

type FilterRepositories struct {
	NoValidation
	Pagination
	Path string `json:"path" form:"path"`
	Name string `json:"name" form:"name"`

	// internal only
	Types     []model.RepoType `json:"-" form:"-"`
	UserID    int64            `json:"-" form:"-"`
	Confirmed *bool            `json:"-" form:"-"`
}

type CreateRepository struct {
	Path        string         `json:"path"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Type        model.RepoType `json:"type"`
}

func (r *CreateRepository) Validate(*gin.Context) (err error) {
	errs := errors.NewInput()
	errs.AddIf(util.IsEmptyString(r.Name), "name", "Name is required")
	errs.AddIf(util.IsEmptyString(r.Description), "description", "Description is required")
	errs.AddIf(util.IsEmptyString(r.Path), "path", "Path is required")
	errs.AddIf(util.IsEmptyString(r.Type.String()), "type", "Type is required")
	errs.AddIf(!r.Type.IsValid(), "type", fmt.Sprintf("Invalid type: '%s'", r.Type))

	if errs.Has() {
		return errs
	}

	return nil
}

func (r *CreateRepository) ToModel(c *gin.Context) (m model.Repository) {
	// cannot create repo with type "global"
	if r.Type == model.RepoTypeGlobal {
		r.Type = model.RepoTypePublic
	}

	return model.Repository{
		UserID:      AuthUser(c).ID,
		Path:        r.Path,
		Name:        r.Name,
		Description: r.Description,
		Type:        r.Type,
	}
}

type UpdateRepository struct {
	Name        *string         `json:"name"`
	Description *string         `json:"description"`
	Type        *model.RepoType `json:"type"`
}

func (r *UpdateRepository) Validate(c *gin.Context) (err error) {
	if InvalidUser(c, Repository(c).UserID) {
		return
	}

	errs := errors.NewInput()
	if r.Name != nil {
		errs.AddIf(util.IsEmptyString(*r.Name), "name", "Name cannot be empty")
	}
	if r.Description != nil {
		errs.AddIf(util.IsEmptyString(*r.Description), "description", "Description cannot be empty")
	}
	if r.Type != nil {
		errs.AddIf(util.IsEmptyString(r.Type.String()), "type", "Type cannot be empty")
		errs.AddIf(!r.Type.IsValid(), "type", fmt.Sprintf("Invalid type: '%s'", r.Type))
	}

	if errs.Has() {
		return errs
	}
	return nil
}

func (r UpdateRepository) ToModel(m *model.Repository) {

	if r.Name != nil {
		m.Name = *r.Name
	}
	if r.Description != nil {
		m.Description = *r.Description
	}
	// cannot update repo to type "global"
	if r.Type != nil && *r.Type != model.RepoTypeGlobal {
		m.Type = *r.Type
	}
}

type DeleteRepository struct{}

func (r DeleteRepository) Validate(c *gin.Context) (err error) {
	if InvalidUser(c, Repository(c).UserID) {
		return
	}

	return nil
}
