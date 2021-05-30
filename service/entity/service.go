package entity

import (
	"github.com/go-pg/pg/v9"
	"github.com/refto/server/database"
	"github.com/refto/server/database/filter"
	"github.com/refto/server/database/model"
	"github.com/refto/server/server/request"
)

// Definitions is a special kind of data that displayed only if one topic selected
// Should filter out definitions from regular data
const DefinitionType = "definition"

// To get needed definition I need to know its token
// And since token is a path to data, i need to know that path to build valid token
// So all definitions should be stored in one location
// and it must be persistent
const DefinitionTokenPrefix = "definitions/"

func Filter(req request.FilterEntities) (data []model.Entity, count int, err error) {
	q := database.ORM().
		Model(&data).
		Apply(filter.PageFilter(req.Page, req.Limit))

	if req.Name != "" {
		q.Where("title ILIKE ?", "%"+req.Name+"%")
	}
	if req.Addr != "" {
		q.Where("data ->> 'home_addr' IS NOT NULL AND data ->> 'home_addr' ILIKE ?", "%"+req.Addr+"%")
	}
	if req.Query != "" {
		// TODO this query will also match keys, but only values needed
		q.Where("data::text  ILIKE ?", "%"+req.Query+"%")
	}

	if len(req.Topics) > 0 {
		q.Join("JOIN entity_topics et ON et.entity_id=entity.id").
			Join("JOIN topics t ON et.topic_id=t.id").
			Group("entity.id")
	}

	if len(req.Topics) == 1 {
		q.Where("t.name = ?", req.Topics[0]).
			// Add specific order when only one topic is selected
			OrderExpr("array_position(array['person', 'book', 'conference', 'software']::text[], type)")
	}

	if len(req.Topics) > 1 {
		q.WhereIn("t.name IN (?)", req.Topics).
			Having("COUNT(t.id) = ?", len(req.Topics))
	}

	// should not match definitions
	if len(req.Topics) > 0 {
		q.Where("type != ?", DefinitionType)
	}

	if req.Collection != 0 {
		q.Join("JOIN collection_entities ce ON ce.entity_id=entity.id").
			Where("ce.collection_id = ?", req.Collection)
	}

	q.OrderExpr("updated_at DESC, created_at DESC")
	count, err = q.SelectAndCount()

	return
}

func Definition(name string) (def *model.Entity, err error) {
	def = &model.Entity{}
	err = database.ORM().
		Model(def).
		Where("type = ?", DefinitionType).
		Where("token = ?", DefinitionTokenPrefix+name).
		First()

	if err == pg.ErrNoRows {
		def = nil
		err = nil
	}

	return
}

func CreateOrUpdate(elem *model.Entity) (err error) {
	old := model.Entity{}
	err = database.ORM().
		Model(&old).
		Where("token = ?", elem.Token).
		Where("repo_id = ?", elem.RepoID).
		First()
	if err != nil && err != pg.ErrNoRows {
		return
	}

	// insert new
	if err == pg.ErrNoRows {
		err = database.ORM().Insert(elem)
		if err != nil {
			return
		}
	} else {
		old.Token = elem.Token
		old.Title = elem.Title
		old.Type = elem.Type
		old.DeletedAt = nil

		old.Data = elem.Data
		// TODO check if new data is different from new data and set updated_at
		//if old.Data != elem.Data {
		//	now := time.Now()
		//	old.Data = elem.Data
		//	old.UpdatedAt = &now
		//}

		err = database.ORM().Update(&old)
		if err != nil {
			return
		}
		*elem = old
	}

	return nil
}

func FindByID(id int64) (m model.Entity, err error) {
	err = database.ORM().
		Model(&m).
		Where("id = ?", id).
		First()

	return
}
