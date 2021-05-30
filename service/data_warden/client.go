package datawarden

import (
	"net/http"
	"sync"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v32/github"
	"github.com/refto/server/config"
)

var (
	clientOnce sync.Once
	clientInst *github.Client
)

func client() *github.Client {
	clientOnce.Do(func() {
		conf := config.Get().GitHub.DataWarden
		tr := http.DefaultTransport
		itr, err := ghinstallation.NewKeyFromFile(tr, conf.AppID, conf.InstallID, conf.PEMPath)
		if err != nil {
			panic("data warden: app auth: " + err.Error())
		}

		clientInst = github.NewClient(&http.Client{
			Transport: itr,
		})
	})

	return clientInst
}
