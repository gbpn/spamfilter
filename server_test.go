package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type testHTTPRequest struct{
	body string
	err error
	response *httptest.ResponseRecorder
	request *http.Request
}

func doReq(api server, rtype string, path string, body string) (testHTTPRequest) {

	res := testHTTPRequest{}

	res.request, _ = http.NewRequest(rtype, path, strings.NewReader(body))
	res.response = httptest.NewRecorder()
	api.router.ServeHTTP(res.response,res.request)
	rbody,err := ioutil.ReadAll(res.response.Body)
	res.err = err
	res.body = string(rbody)
	return res
}

func TestCreate(t *testing.T) {
	var api server
	gin.SetMode("test")
	api.setupRouter()
	// Classifier must have a json body
	res := doReq(api, "PUT", "/classifier/test1", "")
	assert.NoError(t, res.err)
	assert.Contains(t, res.body, "EOF")

	// Classifier json must have actual arguments
	res = doReq(api, "PUT", "/classifier/test1", "{}")
	assert.NoError(t, res.err)
	assert.Equal(t,400, res.response.Code)
	assert.Contains(t, res.body, "2 classes")

	// And a good one should work
	res = doReq(api, "PUT", "/classifier/test1", "{\"classes\":[\"good\",\"bad\"]}")
	assert.NoError(t, res.err)
	assert.Contains(t, res.body, "ok")

	// But the same classifier twice should be a 409 conflict
	res = doReq(api, "PUT", "/classifier/test1", "{\"classes\":[\"good\",\"bad\"]}")
	assert.NoError(t, res.err)
	assert.Equal(t,409, res.response.Code)
	assert.Contains(t, res.body, "already exists")

}
func TestDelete(t *testing.T) {
	var api server
	gin.SetMode("test")
	api.setupRouter()

	res := doReq(api, "DELETE", "/classifier/test1","")
	assert.NoError(t, res.err)
	assert.Contains(t, res.body, "not found")
	assert.Equal(t,404, res.response.Code)

	// Make one to delete
	res = doReq(api, "PUT", "/classifier/test1", "{\"classes\":[\"good\",\"bad\"]}")
	assert.NoError(t, res.err)
	assert.Equal(t,200, res.response.Code)
	assert.Contains(t, res.body, "ok")

	// Make sure we can delete it
	res = doReq(api, "DELETE", "/classifier/test1","")
	assert.NoError(t, res.err)
	assert.Equal(t,200, res.response.Code)
	assert.Contains(t, res.body, "ok")

	// But not twice
	res = doReq(api, "DELETE", "/classifier/test1","")
	assert.Contains(t, res.body, "not found")
	assert.Equal(t,404, res.response.Code)
}


