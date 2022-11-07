package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

func TestNewClassifier(t *testing.T) {
	var api server
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
	println("Bad one stats here")
	res = doReq(api, "PUT", "/classifier/test1", "{\"classes\":[\"good\",\"bad\"]}")
	assert.NoError(t, res.err)
	fmt.Printf("%+v\n", res.response)
	assert.Equal(t,409, res.response.Code)
	assert.Contains(t, res.body, "already exists")
	println("Bad one ends here")

}


