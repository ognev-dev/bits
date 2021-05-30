package mock

import (
	fake "github.com/brianvoe/gofakeit"
	"github.com/go-pg/pg/v9"
	"github.com/refto/server/database"
	"github.com/refto/server/database/model"
	entitytopic "github.com/refto/server/service/entity_topic"
)

func Entity(opt ...model.Entity) (m model.Entity, err error) {
	if len(opt) == 1 {
		m = opt[0]
	}

	if m.Token == "" {
		m.Token = fake.UUID()
	}
	if m.Title == "" {
		m.Title = fake.Name()
	}
	if m.Type == "" {
		m.Type = "book"
	}
	if len(m.Data) == 0 {
		m.Data = model.EntityData{"key": "val"}
	}
	if len(m.Topics) == 0 {
		m.Topics = make([]model.Topic, 5)
		for i := range m.Topics {
			m.Topics[i], err = Topic()
			if err != nil {
				return
			}
		}
	}
	if m.RepoID == 0 {
		var repo model.Repository
		repo, err = InsertRepository()
		if err != nil {
			return
		}
		m.RepoID = repo.ID
	}

	return
}

func InsertEntity(opt ...model.Entity) (m model.Entity, err error) {
	m, err = Entity(opt...)
	if err != nil {
		return
	}

	err = database.ORM().Insert(&m)
	if err != nil {
		return
	}

	for _, v := range m.Topics {
		if v.ID == 0 {
			err = database.ORM().Model(&v).Where("name=?", v.Name).First()
			if err == pg.ErrNoRows {
				err = database.ORM().Insert(&v)
			}
			if err != nil {
				return
			}
		}
		et := model.EntityTopic{
			EntityID: m.ID,
			TopicID:  v.ID,
		}
		err = entitytopic.Create(et)
		if err != nil {
			return
		}
	}

	return
}
