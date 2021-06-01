package test

import (
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/refto/server/database/mock"
	"github.com/refto/server/database/model"
	"github.com/refto/server/server/request"
	"github.com/refto/server/server/response"
	. "github.com/refto/server/test/apitest"
	"github.com/refto/server/test/assert"
)

func TestFilterEntities(t *testing.T) {
	// Create test data
	repo, err := mock.InsertRepository(model.Repository{
		Type: model.RepoTypeGlobal,
	})
	assert.NotError(t, err)
	topic1, err := mock.InsertTopic(model.Topic{
		Name:   "topic1",
		RepoID: repo.ID,
	})
	assert.NotError(t, err)
	topic2, err := mock.InsertTopic(model.Topic{
		Name:   "topic2",
		RepoID: repo.ID,
	})
	assert.NotError(t, err)
	topic3, err := mock.InsertTopic(model.Topic{
		Name:   "topic3",
		RepoID: repo.ID,
	})
	assert.NotError(t, err)

	// entity with 1 topic
	_, err = mock.InsertEntity(model.Entity{
		Title:  "ent1",
		Topics: []model.Topic{{ID: topic1.ID}},
		RepoID: repo.ID,
	})
	assert.NotError(t, err)

	// entity with 2 topics
	ent2, err := mock.InsertEntity(model.Entity{
		Title: "ent2",
		Topics: []model.Topic{
			{ID: topic1.ID},
			{ID: topic2.ID},
		},
		RepoID: repo.ID,
	})
	assert.NotError(t, err)

	// entity with 3 topics
	ent3, err := mock.InsertEntity(model.Entity{
		Title: "ent3",
		Topics: []model.Topic{
			{ID: topic1.ID},
			{ID: topic2.ID},
			{ID: topic3.ID},
		},
		RepoID: repo.ID,
	})
	assert.NotError(t, err)

	var req request.FilterEntities
	var resp response.FilterEntities

	// should get ent2 and ent3
	// as they both have t1 and t2 (and ent1 missing t2)
	req.Topics = []string{topic1.Name, topic2.Name}
	TestFilter(t, "entities", req, &resp)
	assert.Equals(t, 2, resp.EntitiesCount)
	for _, v := range resp.Entities {
		if v.ID == ent2.ID || v.ID == ent3.ID {
			continue
		}
		t.Fatal("invalid entity in response")
	}
	// should get t3 as common topic
	assert.Equals(t, 1, len(resp.Topics))
	assert.Equals(t, topic3.Name, resp.Topics[0])

	// should get only ent3
	// (ent2 missing t3 and ent1 have none of them)
	req.Topics = []string{topic2.Name, topic3.Name}
	TestFilter(t, "entities", req, &resp)
	assert.Equals(t, 1, resp.EntitiesCount)
	assert.Equals(t, ent3.ID, resp.Entities[0].ID)
	// should get t1 as common topic
	assert.Equals(t, 1, len(resp.Topics))
	assert.Equals(t, topic1.Name, resp.Topics[0])

	// should get ent1, ent2, ent3
	req.Topics = []string{topic1.Name}
	TestFilter(t, "entities", req, &resp)
	assert.Equals(t, 3, resp.EntitiesCount)
	// should get t2 and t3 as common topics
	assert.Equals(t, 2, len(resp.Topics))
	assert.Equals(t, topic2.Name, resp.Topics[0])
	assert.Equals(t, topic3.Name, resp.Topics[1])

	// should get only ent3
	req.Topics = []string{topic1.Name, topic2.Name, topic3.Name}
	TestFilter(t, "entities", req, &resp)
	assert.Equals(t, 1, resp.EntitiesCount)
	assert.Equals(t, ent3.ID, resp.Entities[0].ID)
	// should not get common topics
	assert.Equals(t, 0, len(resp.Topics))
}

func TestFilterEntitiesByCollection(t *testing.T) {
	Authorise(t)

	// Create test data
	repo, err := mock.InsertRepository(model.Repository{
		Type: model.RepoTypeGlobal,
	})
	FailOnError(t, err)
	e1, err := mock.InsertEntity(model.Entity{RepoID: repo.ID})
	FailOnError(t, err)
	e2, err := mock.InsertEntity(model.Entity{RepoID: repo.ID})
	FailOnError(t, err)
	e3, err := mock.InsertEntity(model.Entity{RepoID: repo.ID})
	FailOnError(t, err)

	col, err := mock.InsertCollection(model.Collection{
		User: AuthUser,
	})
	FailOnError(t, err)
	ce1, err := mock.InsertCollectionEntity(model.CollectionEntity{
		CollectionID: col.ID,
		Entity:       &e1,
	})
	assert.NotError(t, err)
	ce2, err := mock.InsertCollectionEntity(model.CollectionEntity{
		CollectionID: col.ID,
		Entity:       &e2,
	})
	assert.NotError(t, err)
	ce3, err := mock.InsertCollectionEntity(model.CollectionEntity{
		CollectionID: col.ID,
		Entity:       &e3,
	})
	assert.NotError(t, err)

	_, err = mock.InsertCollectionEntity()
	assert.NotError(t, err)

	req := request.FilterEntities{
		Collection: col.ID,
	}
	var resp response.FilterEntities

	TestFilter(t, "entities", req, &resp)
	assert.Equals(t, 3, resp.EntitiesCount)
	assert.Equals(t, 3, len(resp.Entities))
	for _, v := range resp.Entities {
		if v.ID == ce1.EntityID || v.ID == ce2.EntityID || v.ID == ce3.EntityID {
			continue
		}

		t.Errorf("unexpected entity in response (%+v)", v)
		break
	}
}

func TestGetEntityByID(t *testing.T) {
	Logout()

	e, err := mock.InsertEntity()
	assert.NotError(t, err)

	var resp model.Entity
	TestGet(t, fmt.Sprintf("entities/%d/", e.ID), &resp)

	assert.Equals(t, e.ID, resp.ID)
}

func TestGetEntityByID_ShouldGetCollections(t *testing.T) {
	Authorise(t)

	e, err := mock.InsertEntity()
	assert.NotError(t, err)

	col1, err := mock.InsertCollection(model.Collection{User: AuthUser})
	assert.NotError(t, err)
	col2, err := mock.InsertCollection(model.Collection{User: AuthUser})
	assert.NotError(t, err)
	col3, err := mock.InsertCollection(model.Collection{User: AuthUser})
	assert.NotError(t, err)
	colX, err := mock.InsertCollection()
	assert.NotError(t, err)
	_, err = mock.InsertCollectionEntity(model.CollectionEntity{
		EntityID:     e.ID,
		CollectionID: col1.ID,
	})
	assert.NotError(t, err)
	_, err = mock.InsertCollectionEntity(model.CollectionEntity{
		EntityID:     e.ID,
		CollectionID: col2.ID,
	})
	assert.NotError(t, err)
	_, err = mock.InsertCollectionEntity(model.CollectionEntity{
		EntityID:     e.ID,
		CollectionID: col3.ID,
	})

	// collection that is not create by current user
	// should not be in response
	assert.NotError(t, err)
	_, err = mock.InsertCollectionEntity(model.CollectionEntity{
		EntityID:     e.ID,
		CollectionID: colX.ID,
	})
	assert.NotError(t, err)

	var resp model.Entity
	TestGet(t, fmt.Sprintf("entities/%d/", e.ID), &resp)

	spew.Dump(resp)

	assert.Equals(t, e.ID, resp.ID)
	assert.True(t, 3 == len(resp.Collections))

	for _, v := range resp.Collections {
		if v.ID == col1.ID || v.ID == col2.ID || v.ID == col3.ID {
			continue
		}

		t.Errorf("unexpected collection in response (%+v)", v)
		break
	}
}
