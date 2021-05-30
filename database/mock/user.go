package mock

import (
	"math"
	"math/rand"
	"time"

	fake "github.com/brianvoe/gofakeit"
	"github.com/refto/server/database"
	"github.com/refto/server/database/model"
	"github.com/refto/server/util"
)

func User(opt ...model.User) (m model.User, err error) {
	if len(opt) == 1 {
		m = opt[0]
	}

	if m.Name == "" {
		m.Name = fake.Name()
	}
	if m.Login == "" {
		m.Login = fake.Email()
	}
	if m.AvatarURL == "" {
		m.AvatarURL = fake.URL()
	}
	if m.TelegramID == 0 {
		m.TelegramID = 1
	}
	if m.GithubID == 0 {
		m.GithubID = int64(rand.Intn(math.MaxInt32))
	}
	if m.GithubToken == "" {
		m.GithubToken = util.RandomString()
	}
	if m.Email == "" {
		m.Email = fake.Email()
	}
	if m.ActiveAt.IsZero() {
		m.ActiveAt = time.Now()
	}

	return
}

func InsertUser(opt ...model.User) (m model.User, err error) {
	m, err = User(opt...)
	if err != nil {
		return
	}

	err = database.ORM().Insert(&m)
	return
}
