package apitest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/refto/server/server/response"

	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v9/orm"
	"github.com/refto/server/config"
	"github.com/refto/server/database"
	"github.com/refto/server/database/mock"
	"github.com/refto/server/database/model"
	"github.com/refto/server/logger"
	"github.com/refto/server/server/route"
	authtoken "github.com/refto/server/service/auth_token"
	"github.com/refto/server/test/assert"
	"github.com/refto/server/util"
)

var (
	DB        orm.DB
	AuthUser  *model.User
	authToken string
	router    *gin.Engine
	Conf      *config.Config
)

type Headers map[string]string

type Request struct {
	method       string
	Path         string
	Body         interface{}
	bodyReader   io.Reader
	Headers      Headers
	BindResponse interface{}
	AssertStatus int
}

const ContentTypeJSON = "application/json"

func init() {
	// changing working dir to project root level
	// I want to use main .config.yaml, thats why
	err := os.Chdir("../..")
	if err != nil {
		panic(err)
	}
	Conf = config.Get()

	// override base config
	Conf.Dir.Logs = "tests.log"
	Conf.AppEnv = config.TestEnv

	logger.Setup()

	db := database.Conn()
	_, _ = db.Exec("SET AUTOCOMMIT TO OFF")
	tx, err := db.Conn().Begin()
	if err != nil {
		panic(err)
	}
	database.SetDB(tx)
	DB = database.ORM()

	router = gin.New()
	route.Register(router)
}

// makeRequest makes request (Duh...)
func makeRequest(t *testing.T, r Request) *httptest.ResponseRecorder {
	req, err := http.NewRequest(
		r.method,
		r.Path,
		r.bodyReader,
	)
	assert.NotError(t, err)
	t.Log(fmt.Sprintf("> (%d) %s %s", r.AssertStatus, r.method, r.Path))
	if r.method == http.MethodPost || r.method == http.MethodPut {
		req.Header.Set("Content-Type", ContentTypeJSON)
	}
	req.Header.Set("Accept", ContentTypeJSON)
	req.Header.Set("X-Client", "test")

	for k, v := range r.Headers {
		req.Header.Set(k, v)
	}

	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

// TestRequest sends given request,
// asserts that status is correct
// binds response and returns response
func TestRequest(t *testing.T, req Request) *httptest.ResponseRecorder {
	var body []byte
	var err error

	req.Path = path.Join("/", config.Get().Server.ApiBasePath, req.Path)

	if !strings.HasSuffix(req.Path, "/") && !strings.Contains(req.Path, "/?") {
		req.Path += "/"
	}

	if len(req.Headers) == 0 {
		req.Headers = Headers{}
	}

	if req.Body != nil {
		body, err = json.Marshal(req.Body)
		assert.NotError(t, err)
	}

	req.bodyReader = strings.NewReader(string(body))
	resp := makeRequest(t, req)
	if resp.Code != req.AssertStatus {
		t.Errorf(`%s "%s" responds with %d, expecting %d`, req.method, req.Path, resp.Code, req.AssertStatus)
		t.Errorf("Response:\n%s", resp.Body.String())
		if req.Body != nil {
			t.Errorf("Request:\n%+v", req.bodyReader)
		}
		t.FailNow()
	}

	if responseContentType(resp) == ContentTypeJSON {
		err = json.Unmarshal(resp.Body.Bytes(), &req.BindResponse)
		if err != nil {
			t.Log(resp.Body.String())
			t.Fatal(err)
		}
	}

	return resp
}

func responseContentType(resp *httptest.ResponseRecorder) string {
	ct, ok := resp.Header()["Content-Type"]
	if !ok || len(ct) == 0 {
		return ""
	}

	frags := strings.Split(ct[0], ";")
	return frags[0]
}

// POST is a wrapper for TestRequest
func POST(t *testing.T, req Request) *httptest.ResponseRecorder {
	req.method = "POST"
	return TestRequest(t, req)
}

// PUT is a wrapper for TestRequest
func PUT(t *testing.T, req Request) *httptest.ResponseRecorder {
	req.method = "PUT"
	return TestRequest(t, req)
}

// PATCH is a wrapper for TestRequest
func PATCH(t *testing.T, req Request) *httptest.ResponseRecorder {
	req.method = "PATCH"
	return TestRequest(t, req)
}

// DELETE is a wrapper for TestRequest
func DELETE(t *testing.T, req Request) *httptest.ResponseRecorder {
	req.method = "DELETE"
	return TestRequest(t, req)
}

// GET is a wrapper for TestRequest
func GET(t *testing.T, req Request) *httptest.ResponseRecorder {
	req.method = "GET"
	return TestRequest(t, req)
}

// TestCreate makes "create" POST request that expects response type of response arg and 201 status code
func TestCreate(t *testing.T, path string, body, response interface{}) *httptest.ResponseRecorder {
	req := Request{
		Path:         path,
		Body:         body,
		BindResponse: response,
		AssertStatus: http.StatusCreated,
	}

	return POST(t, req)
}

// TestCreate422 makes "create" POST request that expects 422 status code
func TestCreate422(t *testing.T, path string, body interface{}) (resp response.Error, rec *httptest.ResponseRecorder) {
	req := Request{
		Path:         path,
		Body:         body,
		BindResponse: &resp,
		AssertStatus: http.StatusUnprocessableEntity,
	}

	rec = POST(t, req)
	return
}

// TestUpdate makes "update" request
func TestUpdate(t *testing.T, path string, body, response interface{}) *httptest.ResponseRecorder {
	req := Request{
		Path:         path,
		Body:         body,
		BindResponse: response,
		AssertStatus: http.StatusOK,
	}

	return PUT(t, req)
}

// Patch makes "patch" request
func Patch(t *testing.T, path string, body, response interface{}) *httptest.ResponseRecorder {
	req := Request{
		Path:         path,
		Body:         body,
		BindResponse: response,
		AssertStatus: http.StatusOK,
	}

	return PATCH(t, req)
}

// TestDelete makes "delete" request
func TestDelete(t *testing.T, path string, response interface{}) *httptest.ResponseRecorder {
	req := Request{
		Path:         path,
		BindResponse: response,
		AssertStatus: http.StatusOK,
	}

	return DELETE(t, req)
}

// TestGet makes simple "get" request
func TestGet(t *testing.T, path string, response interface{}) *httptest.ResponseRecorder {
	req := Request{
		Path:         path,
		BindResponse: response,
		AssertStatus: http.StatusOK,
	}

	return GET(t, req)
}

// TestGet404 makes a "get" request that expects 404 status code
func TestGet404(t *testing.T, path string) (resp response.Error, rec *httptest.ResponseRecorder) {
	req := Request{
		Path:         path,
		BindResponse: &resp,
		AssertStatus: http.StatusNotFound,
	}

	rec = GET(t, req)
	return
}

// TestFilter makes "get" request with query params
func TestFilter(t *testing.T, path string, request, response interface{}) *httptest.ResponseRecorder {
	query, err := util.StructToQueryString(request)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	if query != "" {
		path += "?" + query
	}

	req := Request{
		Path:         path,
		BindResponse: response,
		AssertStatus: http.StatusOK,
	}

	return GET(t, req)
}

func Authorise(t *testing.T) *model.User {
	Logout()
	user, err := mock.InsertUser()
	assert.NotError(t, err)
	AuthoriseAs(t, &user)

	return &user
}

func AuthoriseAs(t *testing.T, user *model.User) {
	token := &model.AuthToken{
		UserID:     user.ID,
		ClientName: "test",
	}
	err := authtoken.Create(token)
	assert.NotError(t, err)

	AuthUser = user
	authToken = authtoken.Sign(token)
}

func Logout() {
	AuthUser = nil
	authToken = ""
}
