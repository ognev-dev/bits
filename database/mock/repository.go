package mock

import (
	fake "github.com/brianvoe/gofakeit"
	"github.com/refto/server/database"
	"github.com/refto/server/database/model"
	"github.com/refto/server/util"
)

func Repository(opt ...model.Repository) (m model.Repository, err error) {
	if len(opt) == 1 {
		m = opt[0]
	}

	if m.Name == "" {
		m.Name = fake.Name()
	}
	if m.Description == "" {
		m.Description = fake.Name()
	}
	if m.User != nil {
		m.UserID = m.User.ID
	}
	if m.UserID == 0 {
		var u model.User
		u, err = InsertUser()
		if err != nil {
			return
		}
		m.UserID = u.ID
		m.User = &u
	}
	if m.Path == "" {
		m.Path = util.RandomString() + "/" + util.RandomString()
	}
	if m.Type == "" {
		m.Type = model.RepoType(fake.RandString(model.RepositoryTypesList))
	}
	if m.Secret == "" {
		m.Secret = util.RandomString()
	}

	return
}

func InsertRepository(opt ...model.Repository) (m model.Repository, err error) {
	m, err = Repository(opt...)
	if err != nil {
		return
	}

	err = database.ORM().Insert(&m)
	return
}
