package mock

import (
	"github.com/refto/server/database"
	"github.com/refto/server/database/model"
)

func CollectionEntity(opt ...model.CollectionEntity) (m model.CollectionEntity, err error) {
	if len(opt) == 1 {
		m = opt[0]
	}

	if m.CollectionID == 0 {
		if m.Collection == nil {
			m.Collection = &model.Collection{}
			*m.Collection, err = InsertCollection()
			if err != nil {
				return m, err
			}
		}
		m.CollectionID = m.Collection.ID
	}
	if m.EntityID == 0 {
		if m.Entity == nil {
			m.Entity = &model.Entity{}
			*m.Entity, err = InsertEntity()
			if err != nil {
				return m, err
			}
		}

		m.EntityID = m.Entity.ID
	}

	return
}

func InsertCollectionEntity(opt ...model.CollectionEntity) (m model.CollectionEntity, err error) {
	m, err = CollectionEntity(opt...)
	if err != nil {
		return
	}

	err = database.ORM().Insert(&m)
	return
}
