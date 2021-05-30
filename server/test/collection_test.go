package test

import (
	"fmt"
	"testing"

	"github.com/brianvoe/gofakeit"
	"github.com/refto/server/database/mock"
	"github.com/refto/server/database/model"
	"github.com/refto/server/server/request"
	"github.com/refto/server/server/response"
	. "github.com/refto/server/test/apitest"
	"github.com/refto/server/test/assert"
	"github.com/refto/server/util"
)

func TestGetCollections(t *testing.T) {
	Authorise(t)

	c1, err := mock.InsertCollection(model.Collection{UserID: AuthUser.ID, Name: "C1"})
	assert.NotError(t, err)
	c2, err := mock.InsertCollection(model.Collection{UserID: AuthUser.ID, Name: "C2"})
	assert.NotError(t, err)
	c3, err := mock.InsertCollection(model.Collection{UserID: AuthUser.ID, Name: "C3"})
	assert.NotError(t, err)
	_, err = mock.InsertCollection() // should not be in response
	assert.NotError(t, err)

	var req request.FilterCollections
	var resp response.FilterCollections

	TestFilter(t, "collections", req, &resp)

	assert.True(t, 3 == resp.Count)
	assert.True(t, 3 == len(resp.Data))

	for _, v := range resp.Data {
		if v.UserID != AuthUser.ID {
			t.Errorf("collection must be owned by authorized user. expecting %d, got %d", v.UserID, AuthUser.ID)
			break
		}

		if v.ID == c1.ID || v.ID == c2.ID || v.ID == c3.ID {
			continue
		}

		t.Errorf("unexpected collection in response (%+v)", v)
		break
	}

	// Test entities count
	// create 3 entities for c1
	_, err = mock.InsertCollectionEntity(model.CollectionEntity{Collection: &c1})
	assert.NotError(t, err)
	_, err = mock.InsertCollectionEntity(model.CollectionEntity{Collection: &c1})
	assert.NotError(t, err)
	_, err = mock.InsertCollectionEntity(model.CollectionEntity{Collection: &c1})
	assert.NotError(t, err)
	// create 2 entities for c3
	_, err = mock.InsertCollectionEntity(model.CollectionEntity{Collection: &c3})
	assert.NotError(t, err)
	_, err = mock.InsertCollectionEntity(model.CollectionEntity{Collection: &c3})
	assert.NotError(t, err)

	assertCounts := map[int64]int{
		c1.ID: 3,
		c2.ID: 0,
		c3.ID: 2,
	}

	req.WithEntitiesCount = true
	TestFilter(t, "collections", req, &resp)

	for _, v := range resp.Data {
		expectCount, ok := assertCounts[v.ID]
		if !ok {
			t.Error("invalid collection in response!")
			break
		}

		assert.Equals(t, expectCount, v.EntitiesCount, fmt.Sprintf("collection %d: invalid entities count", v.ID))
	}
}

func TestGetCollections_FilterByName(t *testing.T) {
	Authorise(t)

	c1, err := mock.InsertCollection(model.Collection{UserID: AuthUser.ID, Name: "111Name"})
	assert.NotError(t, err)
	c2, err := mock.InsertCollection(model.Collection{UserID: AuthUser.ID, Name: "222Name"})
	assert.NotError(t, err)
	c3, err := mock.InsertCollection(model.Collection{UserID: AuthUser.ID, Name: "333Name"})
	assert.NotError(t, err)
	_, err = mock.InsertCollection(model.Collection{UserID: AuthUser.ID, Name: "xxx"}) // should not be in response
	assert.NotError(t, err)
	_, err = mock.InsertCollection(model.Collection{Name: "xxxName"}) // should not be in response
	assert.NotError(t, err)

	req := request.FilterCollections{
		Name: "name",
	}
	var resp response.FilterCollections

	TestFilter(t, "collections", req, &resp)

	assert.True(t, 3 == resp.Count)
	assert.True(t, 3 == len(resp.Data))

	for _, v := range resp.Data {
		if v.UserID != AuthUser.ID {
			t.Errorf("collection must be owned by authorized user. expecting %d, got %d", v.UserID, AuthUser.ID)
			break
		}

		if v.ID == c1.ID || v.ID == c2.ID || v.ID == c3.ID {
			continue
		}

		t.Errorf("unexpected collection in response (%+v)", v)
		break
	}
}

func TestCreateCollection(t *testing.T) {
	Authorise(t)

	req := request.CreateCollection{
		Name: gofakeit.Name(),
	}
	var resp model.Collection

	TestCreate(t, "collections", req, &resp)

	assert.True(t, resp.Token != "")
	assert.DatabaseHas(t, "collections", util.M{
		"id":      resp.ID,
		"token":   resp.Token,
		"user_id": AuthUser.ID,
		"private": false,
	})
}

func TestUpdateCollection(t *testing.T) {
	Authorise(t)

	elem, err := mock.InsertCollection(model.Collection{UserID: AuthUser.ID})
	assert.NotError(t, err)

	req := request.UpdateCollection{
		Name:    gofakeit.Name(),
		Private: true,
	}
	var resp model.Collection

	TestUpdate(t, fmt.Sprintf("collections/%d", elem.ID), req, &resp)

	assert.DatabaseHas(t, "collections", util.M{
		"id":      elem.ID,
		"token":   elem.Token,
		"user_id": AuthUser.ID,
		"name":    req.Name,
		"private": req.Private,
	})
}

func TestDeleteCollection(t *testing.T) {
	Authorise(t)

	c, err := mock.InsertCollection(model.Collection{UserID: AuthUser.ID})
	assert.NotError(t, err)
	ce, err := mock.InsertCollectionEntity(model.CollectionEntity{CollectionID: c.ID})
	assert.NotError(t, err)

	assert.DatabaseHas(t, "collections", util.M{
		"id": c.ID,
	})
	assert.DatabaseHas(t, "collection_entities", util.M{
		"collection_id": ce.CollectionID,
		"entity_id":     ce.EntityID,
	})

	var resp response.Success

	TestDelete(t, fmt.Sprintf("collections/%d", c.ID), &resp)

	assert.DatabaseMissing(t, "collection_entities", util.M{
		"collection_id": ce.CollectionID,
		"entity_id":     ce.EntityID,
	})
	assert.DatabaseMissing(t, "collections", util.M{
		"id": c.ID,
	})
}

func TestAddEntityToCollection(t *testing.T) {
	Authorise(t)

	c, err := mock.InsertCollection(model.Collection{UserID: AuthUser.ID})
	assert.NotError(t, err)
	e, err := mock.InsertEntity()
	assert.NotError(t, err)

	assert.DatabaseMissing(t, "collection_entities", util.M{
		"collection_id": c.ID,
		"entity_id":     e.ID,
	})

	var resp response.Success

	TestCreate(t, fmt.Sprintf("collections/%d/entities/%d", c.ID, e.ID), nil, &resp)

	assert.DatabaseHas(t, "collection_entities", util.M{
		"collection_id": c.ID,
		"entity_id":     e.ID,
	})
}

func TestRemoveEntityFromCollection(t *testing.T) {
	Authorise(t)

	c, err := mock.InsertCollection(model.Collection{UserID: AuthUser.ID})
	assert.NotError(t, err)
	ce, err := mock.InsertCollectionEntity(model.CollectionEntity{CollectionID: c.ID})
	assert.NotError(t, err)

	assert.DatabaseHas(t, "collections", util.M{
		"id": c.ID,
	})
	assert.DatabaseHas(t, "collection_entities", util.M{
		"collection_id": ce.CollectionID,
		"entity_id":     ce.EntityID,
	})

	var resp response.Success

	TestDelete(t, fmt.Sprintf("collections/%d/entities/%d", c.ID, ce.EntityID), &resp)

	assert.DatabaseMissing(t, "collection_entities", util.M{
		"collection_id": ce.CollectionID,
		"entity_id":     ce.EntityID,
	})
}
