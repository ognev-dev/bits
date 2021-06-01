package mock

import (
	fake "github.com/brianvoe/gofakeit"
	"github.com/refto/server/database"
	"github.com/refto/server/database/model"
)

func Topic(opt ...model.Topic) (m model.Topic, err error) {
	if len(opt) == 1 {
		m = opt[0]
	}

	if m.Name == "" {
		m.Name = fake.Name()
	}
	if m.RepoID == 0 {
		repo, err := InsertRepository()
		if err != nil {
			return m, err
		}
		m.RepoID = repo.ID
	}

	return
}

func InsertTopic(opt ...model.Topic) (m model.Topic, err error) {
	m, err = Topic(opt...)
	if err != nil {
		return
	}

	err = database.ORM().Insert(&m)
	return
}
