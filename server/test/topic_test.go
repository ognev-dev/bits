package test

import (
	"testing"

	"github.com/go-pg/pg/v9"
	"github.com/refto/server/database"
	"github.com/refto/server/database/mock"
	"github.com/refto/server/database/model"
	"github.com/refto/server/server/request"
	"github.com/refto/server/server/response"
	. "github.com/refto/server/test/apitest"
	"github.com/refto/server/test/assert"
)

func TestFilterTopic(t *testing.T) {
	// Create test data
	topics := []string{"test_this", "test_that", "test_else"}
	for _, name := range topics {
		_, err := mock.InsertTopic(model.Topic{Name: name})
		assert.NotError(t, err)
	}

	var req request.FilterTopics
	var resp response.FilterTopics

	topicsCount := 0
	_, err := database.ORM().Query(pg.Scan(&topicsCount), "SELECT COUNT(id) FROM topics")
	assert.NotError(t, err)

	cases := map[string]int{
		"":          topicsCount, // all
		"test_th":   2,           // partial match of: this, that
		"test_else": 1,           // exact match
	}
	for query, count := range cases {
		req.Name = query
		TestFilter(t, "topics", req, &resp)
		assert.Equals(t, count, resp.Count)
	}
}
