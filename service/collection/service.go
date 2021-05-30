package collection

import (
	"errors"
	"math/rand"

	"github.com/go-pg/pg/v9"
	"github.com/refto/server/database"
	"github.com/refto/server/database/filter"
	"github.com/refto/server/database/model"
	"github.com/refto/server/server/request"
)

func Filter(req request.FilterCollections) (data []model.Collection, count int, err error) {
	q := database.ORM().
		Model(&data).
		Apply(filter.PageFilter(req.Page, req.Limit)).
		Apply(filter.UserFilter(req.UserID)).
		Order("created_at DESC")

	if req.EntityID != 0 {
		if req.Available {
			q.Where("collection.id NOT IN (SELECT collection_id FROM collection_entities WHERE entity_id = ?)", req.EntityID)
		} else {
			q.Join("JOIN collection_entities ce ON ce.collection_id=collection.id").
				Where("ce.entity_id = ?", req.EntityID)
		}
	}

	if req.Name != "" {
		q.Where("collection.name ILIKE ?", "%"+req.Name+"%")
	}

	if req.WithEntitiesCount {
		q.Join("LEFT JOIN collection_entities ce2 ON ce2.collection_id=collection.id").
			ColumnExpr("collection.*, COUNT(ce2.collection_id) AS entities_count").
			Group("collection.id")

	}

	count, err = q.SelectAndCount()
	return
}

func Create(elem *model.Collection) (err error) {
	elem.Token, err = NewToken()
	if err != nil {
		return
	}

	err = database.ORM().Insert(elem)
	return
}

func Update(elem *model.Collection) (err error) {
	err = database.ORM().Update(elem)
	return
}

func Delete(id int64) (err error) {
	_, err = database.ORM().
		Model(&model.CollectionEntity{}).
		Where("collection_id = ?", id).
		Delete()
	if err != nil {
		return
	}

	_, err = database.ORM().
		Model(&model.Collection{}).
		Where("id = ?", id).
		Delete()

	return
}

func FindByID(id int64) (m model.Collection, err error) {
	err = database.ORM().
		Model(&m).
		Where("id = ?", id).
		First()

	if err == pg.ErrNoRows {
		err = errors.New("collection not found")
	}

	return
}

func FindByToken(token string) (m model.Collection, err error) {
	err = database.ORM().
		Model(&m).
		Where("token = ?", token).
		First()

	if err == pg.ErrNoRows {
		err = errors.New("collection not found")
	}

	return
}

func AddEntity(collectionID, entityID int64) (err error) {
	err = database.ORM().Insert(&model.CollectionEntity{
		CollectionID: collectionID,
		EntityID:     entityID,
	})

	return
}

func RemoveEntity(collectionID, entityID int64) (err error) {
	_, err = database.ORM().
		Model(&model.CollectionEntity{}).
		Where("collection_id = ?", collectionID).
		Where("entity_id = ?", entityID).
		Delete()

	return
}

// NewToken creates token of 1 lowercase char
// Increases it's length by one char if such token already exists
// and do so until unique token is created
func NewToken() (token string, err error) {
	chars := []byte("-.abcdefghijklmnopqrstuvwxyz1234567890")
	for length := 1; ; length++ {
		t := make([]byte, length)
		for char := 0; char < length; char++ {
			t[char] = chars[rand.Intn(len(chars))]
		}
		token = string(t)
		err = database.ORM().
			Model(&model.Collection{}).
			Where("token = ?", token).
			First()

		if err == pg.ErrNoRows {
			return token, nil
		}

		if err != nil {
			return "", err
		}
	}
}
