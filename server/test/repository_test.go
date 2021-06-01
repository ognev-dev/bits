package test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/refto/server/service/repository"

	"github.com/refto/server/database/mock"

	"github.com/refto/server/server/response"

	"github.com/brianvoe/gofakeit"
	"github.com/refto/server/database/model"
	"github.com/refto/server/server/request"
	. "github.com/refto/server/test/apitest"
	"github.com/refto/server/test/assert"
	"github.com/refto/server/util"
)

func TestCreateRepository(t *testing.T) {
	Authorise(t)

	req := request.CreateRepository{
		Path:        util.RandomString() + "/" + util.RandomString(),
		Name:        gofakeit.Name(),
		Description: gofakeit.Name(),
		Type:        model.RepoTypePublic,
	}

	var resp response.CreateRepository
	TestCreate(t, "repositories", req, &resp)

	assert.True(t, resp.Secret != "")
	assert.DatabaseHas(t, "repositories", util.M{
		"path":        req.Path,
		"user_id":     AuthUser.ID,
		"name":        req.Name,
		"description": req.Description,
		"type":        req.Type,
		"confirmed":   false,
	})
}

// Since repo path is unique, adding repo with existing path should raise error
// TODO add repo transfer to another user using secret
func TestCreateRepository_Existing(t *testing.T) {
	Authorise(t)

	m, err := mock.InsertRepository()
	assert.NotError(t, err)

	req := request.CreateRepository{
		Path:        m.Path,
		Name:        gofakeit.Name(),
		Description: gofakeit.Name(),
		Type:        model.RepoTypeGlobal,
	}
	resp, _ := TestCreate422(t, "repositories", req)
	assert.Equals(t, resp.Error, repository.ErrRepoAlreadyClaimed.Error())
}

func TestRepositoryGetNewSecret(t *testing.T) {
	Authorise(t)

	m, err := mock.InsertRepository(model.Repository{
		UserID: AuthUser.ID,
	})
	FailOnError(t, err)

	var resp response.CreateRepository
	POST(t, Request{
		Path:         fmt.Sprintf("repositories/%d/secret/", m.ID),
		Body:         nil,
		BindResponse: &resp,
		AssertStatus: http.StatusOK,
	})
	assert.True(t, resp.Secret != "")
}

func TestGetUserRepositories(t *testing.T) {
	Authorise(t)

	m1, err := mock.InsertRepository(model.Repository{UserID: AuthUser.ID})
	assert.NotError(t, err)
	m2, err := mock.InsertRepository(model.Repository{UserID: AuthUser.ID})
	assert.NotError(t, err)
	_, err = mock.InsertRepository() // not user's
	assert.NotError(t, err)

	var req request.FilterRepositories
	var resp response.FilterRepositories
	TestFilter(t, "user/repositories/", req, &resp)
	assert.Equals(t, 2, resp.Count)

	for _, el := range resp.Data {
		if el.ID == m1.ID {
			continue
		}
		if el.ID == m2.ID {
			continue
		}

		t.Fatalf("invalid element in response: %v", el)
	}
}

func TestGetPublicRepositories(t *testing.T) {
	Authorise(t)

	// hidden repo should not be in response
	_, err := mock.InsertRepository(model.Repository{
		Type:      model.RepoTypeHidden,
		Confirmed: true,
	})
	assert.NotError(t, err)

	// private repo should not be in response
	_, err = mock.InsertRepository(model.Repository{
		Type:      model.RepoTypePrivate,
		Confirmed: true,
	})
	assert.NotError(t, err)

	// public but not confirmed should not be in response
	_, err = mock.InsertRepository(model.Repository{
		Type:      model.RepoTypePublic,
		Confirmed: false,
	})
	assert.NotError(t, err)

	m1, err := mock.InsertRepository(model.Repository{
		Type:      model.RepoTypeGlobal,
		Confirmed: true,
	})
	assert.NotError(t, err)
	m2, err := mock.InsertRepository(model.Repository{
		Type:      model.RepoTypePublic,
		Confirmed: true,
	})
	assert.NotError(t, err)

	var req request.FilterRepositories
	var resp response.FilterRepositories
	TestFilter(t, "/repositories/", req, &resp)
	assert.Equals(t, 2, resp.Count)

	for _, el := range resp.Data {
		if el.ID == m1.ID {
			continue
		}
		if el.ID == m2.ID {
			continue
		}

		t.Fatalf("invalid element in response: %v", el)
	}
}

func TestGetRepositoryByPath_PublicAndHidden(t *testing.T) {
	m, err := mock.InsertRepository(model.Repository{Type: model.RepoTypePublic})
	assert.NotError(t, err)

	var resp model.Repository
	TestGet(t, "repositories/"+m.Path, &resp)
	assert.Equals(t, m.ID, resp.ID)
	assert.Equals(t, m.Path, resp.Path)
	assert.Equals(t, m.Name, resp.Name)
	assert.Equals(t, m.Description, resp.Description)
	assert.Equals(t, m.Type, resp.Type)
	assert.Equals(t, m.Confirmed, resp.Confirmed)

	m, err = mock.InsertRepository(model.Repository{Type: model.RepoTypeHidden})
	assert.NotError(t, err)

	TestGet(t, "repositories/"+m.Path, &resp)
	assert.Equals(t, m.ID, resp.ID)
	assert.Equals(t, m.Path, resp.Path)
	assert.Equals(t, m.Name, resp.Name)
	assert.Equals(t, m.Description, resp.Description)
	assert.Equals(t, m.Type, resp.Type)
	assert.Equals(t, m.Confirmed, resp.Confirmed)
}

func TestGetRepositoryByPath_Private(t *testing.T) {
	m, err := mock.InsertRepository(model.Repository{Type: model.RepoTypePrivate})
	assert.NotError(t, err)

	TestGet404(t, "repositories/"+m.Path)

	AuthoriseAs(t, m.User)

	var resp model.Repository
	TestGet(t, "repositories/"+m.Path, &resp)

	assert.Equals(t, resp.ID, m.ID)
	assert.Equals(t, resp.Type, model.RepoTypePrivate)

}

func TestUpdateRepository(t *testing.T) {
	Authorise(t)

	m, err := mock.InsertRepository(model.Repository{
		Type:   model.RepoTypePublic,
		UserID: AuthUser.ID,
	})
	assert.NotError(t, err)

	repoType := model.RepoTypeHidden
	req := request.UpdateRepository{
		Name:        util.NewString(util.RandomString(10)),
		Description: util.NewString(util.RandomString(10)),
		Type:        &repoType,
	}

	var resp model.Repository
	TestUpdate(t, "repositories/"+fmt.Sprintf("%d", m.ID), req, &resp)

	assert.Equals(t, resp.ID, m.ID)
	assert.Equals(t, resp.Type, model.RepoTypeHidden)

	assert.DatabaseHas(t, "repositories", util.M{
		"id":          m.ID,
		"path":        m.Path,
		"user_id":     m.UserID,
		"name":        *req.Name,
		"description": *req.Description,
		"type":        *req.Type,
	})
}

func TestDeleteRepository(t *testing.T) {
	Authorise(t)

	repo, err := mock.InsertRepository(model.Repository{UserID: AuthUser.ID})
	assert.NotError(t, err)
	entity, err := mock.InsertEntity(model.Entity{RepoID: repo.ID})
	assert.NotError(t, err)
	_, err = mock.InsertCollectionEntity(model.CollectionEntity{
		EntityID: entity.ID,
	})
	assert.NotError(t, err)

	var resp response.Success
	TestDelete(t, "repositories/"+fmt.Sprintf("%d", repo.ID), &resp)

	assert.DatabaseMissing(t, "collection_entities", util.M{
		"entity_id": entity.ID,
	})
	assert.DatabaseMissing(t, "entities", util.M{
		"id": entity.ID,
	})
	assert.DatabaseMissing(t, "repositories", util.M{
		"id": repo.ID,
	})
}
