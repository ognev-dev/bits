package handler

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/refto/server/server/request"
	dataimport "github.com/refto/server/service/data_import"
	datawarden "github.com/refto/server/service/data_warden"
	githubpullrequest "github.com/refto/server/service/github_pull_request"
	githubwebhook "github.com/refto/server/service/github_webhook"
	jsonschema "github.com/refto/server/service/json_schema"
	"github.com/refto/server/service/repository"
	log "github.com/sirupsen/logrus"
)

// ImportDataFromRepoByGitHubWebHook is a webhook's handler that is triggered by GitHub
// Once commits pushed to a branch, Github will send request to a route which will call this method
// (Trigger must set manually on Github)
// Here we simply check for valid  signature, then clone repo to have data locally and then import it.
// Note: Payloads are capped at 25 MB. If your event generates a larger payload, a webhook will not be fired. This may happen, for example, on a create event if many branches or tags are pushed at once. We suggest monitoring your payload size to ensure delivery.
// Note: You will not receive a webhook for this event when you push more than three tags at once.
// https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads#push
func ImportDataFromRepoByGitHubWebHook(c *gin.Context) {
	var headers request.GitHubWebHookHeaders
	err := c.ShouldBindHeader(&headers)
	if err != nil {
		Abort(c, err)
		return
	}

	// event must be a push event
	if headers.EventName != githubwebhook.PushEvent {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// copy request body now, it'll need soon to check signature
	body, err := copyRequestBody(c)
	if err != nil {
		log.Error(err)
		Abort(c, err)
		return
	}

	var req request.GitHubRepoPushed
	err = c.ShouldBindJSON(&req)
	if err != nil {
		Abort(c, err)
		return
	}

	repo, err := repository.FindByPath(req.Repo.Path)
	if err != nil {
		Abort(c, err)
		return
	}

	validSig, err := githubwebhook.IsValidHMAC(body, headers.EventSig, repo.Secret)
	if err != nil {
		log.Error("[ERROR] " + err.Error())
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if !validSig {
		log.Error("[ERROR] github's webhook invalid signature")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// import data on goroutine, because repo is authorized, but might take some time to import
	// TODO: make selective import using diff ?
	// TODO import using queue
	go func() {
		// repoURL initially unknown
		// TODO when adding new repo get repo from GH and if user that adds the repo is owner of the repo
		//  then mark repo as confirmed and also set repoURL
		repo.CloneURL = req.Repo.CloneURL

		err = dataimport.FromGitHub(repo)
		if err != nil {
			log.Error("[ERROR] import from GH: " + err.Error())
		}
	}()

	c.JSON(http.StatusOK, "ok")
}

func ProcessPullRequestActions(c *gin.Context) {
	var headers request.GitHubWebHookHeaders
	err := c.ShouldBindHeader(&headers)
	if err != nil {
		Abort(c, err)
		return
	}

	// event must be a pull_request
	if headers.EventName != githubwebhook.PullRequestEvent {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var req request.GitHubPullRequestEvent
	err = c.ShouldBindJSON(&req)
	if err != nil {
		Abort(c, err)
		return
	}

	repo, err := repository.FindByPath(req.Repo.Path)
	if err != nil {
		Abort(c, err)
		return
	}

	// check signature
	body, err := copyRequestBody(c)
	if err != nil {
		log.Error(err)
		Abort(c, err)
		return
	}

	validSig, err := githubwebhook.IsValidHMAC(body, headers.EventSig, repo.Secret)
	if err != nil {
		log.Error("[ERROR] " + err.Error())
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if !validSig {
		log.Error("[ERROR] github's webhook (pull_request) invalid signature")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if !req.Action.ShouldValidate() {
		c.JSON(http.StatusOK, "ok")
		return
	}

	// TODO good to check for merge conflicts here
	// but I don't know how to, seems like github API didnt have info on this

	// because data check might take some time, mark HEAD as status "pending"
	// (in fact it is "in-progress", but some feedback better than nothing)
	dw := datawarden.New(repo.Path)
	_, _, err = dw.Status(
		req.PR.Head.SHA,
		githubpullrequest.StatusPending,
		"Checking data...",
		nil,
	)
	if err != nil {
		log.Error("[ERROR] data warden set commit status: " + err.Error())
		return
	}

	go func() {
		var err error

		// if any errors happened, i'd mark the HEAD as failure and comment error to pull request
		// the error might be not related to pull request checks (like internal error or whatever)
		defer func() {
			if err != nil {
				comment, _, err2 := dw.Comment(req.Number, err.Error())
				if err2 != nil {
					log.Error("[ERROR] data warden comment: " + err2.Error())
					return
				}

				var commentURL *string
				if comment.ID != nil {
					url := dw.PRCommentLink(req.Number, *comment.ID)
					commentURL = &url
				}

				_, _, err = dw.Status(
					req.PR.Head.SHA,
					githubpullrequest.StatusFailure,
					err.Error(),
					commentURL,
				)
				if err != nil {
					log.Error("[ERROR] data warden set commit status: " + err.Error())
					return
				}
			}
		}()

		// to make sure that data checks will not go into conflicts
		// i'd clone each PR's HEAD in " {repoPath}/pr_{PR_ID}_{HEAD_SHA}" directory
		cloneDir := path.Join(repo.Path, fmt.Sprintf("pr_%d_%s", req.Number, req.PR.Head.SHA))
		_ = os.MkdirAll(path.Join("pr-checks", cloneDir), 0755)

		_, err = git.PlainClone(cloneDir, false, &git.CloneOptions{
			URL:               req.PR.Head.Repo.CloneURL,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})
		if err != nil {
			err = fmt.Errorf("[ERROR] git clone: " + err.Error())
			log.Error(err)
			return
		}

		defer func() {
			err := os.RemoveAll(cloneDir)
			if err != nil {
				err = fmt.Errorf("[ERROR] remove cloned repo (%s) %s: ", cloneDir, err.Error())
				log.Error(err)
			}
		}()

		// TODO validate not only json schema but everything that should be validated?
		//  like URLs must be valid URL, dates must be valid dates, etc
		//   and so on (probably can be done with json schema custom validators)
		_, err = jsonschema.Validate(cloneDir)
		if err != nil {
			err = fmt.Errorf("[ERROR] data validate: %s", err.Error())
			log.Error("[ERROR] data validate: " + err.Error())
			return
		}

		// TODO make test run on database
		//  1. create new db
		//  2. insert repo
		//  3. run validation
		//  4. drop db

		// TODO set reviewers according to topics
		//  _, _, err = client.PullRequests.RequestReviewers(context.Background(), "refto", "data", 1, github.ReviewersRequest{
		//  Reviewers: []string{
		//		"data-warden",
		//	 },
		//  })

		// all good, mark commit as success
		_, _, err = dw.Status(
			req.PR.Head.SHA,
			githubpullrequest.StatusSuccess,
			dw.DataCheckSuccessMessage(),
			nil,
		)
		if err != nil {
			log.Error("[ERROR] data warden set commit status: " + err.Error())
			return
		}
	}()

	c.JSON(http.StatusOK, "ok")
}
