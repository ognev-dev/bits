package test

import (
	"net/http"
	"testing"

	"github.com/refto/server/config"
	"github.com/refto/server/server/request"
	githubwebhook "github.com/refto/server/service/github_webhook"
	. "github.com/refto/server/test/apitest"
	"github.com/refto/server/test/assert"
)

func TestWebhookDataPushed(t *testing.T) {
	// TODO
	// To run this test you'll need real repo at conf.GitHub.DataRepo
	// And also because repo cloning, data validation and import is executed in separate routine
	// this test will succeed anyway
	// This test might helpful for debug
	// Skipping until proper way of testing this route will be implemented
	t.Skip()

	conf := config.Get()
	req := request.GitHubRepoPushed{
		Repo: request.GitHubRepo{
			CloneURL: conf.GitHub.DataRepo,
		},
	}

	sig, err := githubwebhook.MakeHMAC(
		`{"repository":{"clone_url":"`+conf.GitHub.DataRepo+`"}}`,
		conf.GitHub.DataPushedHookSecret,
	)
	assert.NotError(t, err)

	POST(t, Request{
		Path:         "hooks/data-pushed",
		Body:         req,
		BindResponse: nil,
		AssertStatus: http.StatusOK,
		Headers: Headers{
			"X-GitHub-Event":  "push",
			"X-Hub-Signature": sig,
		},
	})
}

func TestWebhookPullRequest(t *testing.T) {
	t.Skip()

	conf := config.Get()
	req := request.GitHubRepoPushed{
		Repo: request.GitHubRepo{
			CloneURL: conf.GitHub.DataRepo,
		},
	}

	sig, err := githubwebhook.MakeHMAC(
		`{"repository":{"clone_url":"`+conf.GitHub.DataRepo+`"}}`,
		conf.GitHub.DataPushedHookSecret,
	)
	assert.NotError(t, err)

	POST(t, Request{
		Path:         "hooks/data-pushed",
		Body:         req,
		BindResponse: nil,
		AssertStatus: http.StatusOK,
		Headers: Headers{
			"X-GitHub-Event":  "push",
			"X-Hub-Signature": sig,
		},
	})
}
