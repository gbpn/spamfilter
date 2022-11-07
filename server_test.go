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

func DoReq(api server, rtype string, path string, body string) (*http.Request, string, error) {

	req, _ := http.NewRequest(rtype, path, strings.NewReader(body))
	w := httptest.NewRecorder()
	api.router.ServeHTTP(w,req)
	rbody,err := ioutil.ReadAll(w.Body)
	return req,string(rbody),err
}

func TestNewClassifier(t *testing.T) {
	var api server
	api.setupRouter()
	_,body,err := DoReq(api, "PUT", "/classifier/test1", "")
	assert.NoError(t, err)
	assert.Contains(t, body, "EOF")

	_,body,err = DoReq(api, "PUT", "/classifier/test1", "{}")
	assert.NoError(t, err)
	assert.Contains(t, body, "2 classes")

	_,body,err = DoReq(api, "PUT", "/classifier/test1", "{\"classes\":[\"good\",\"bad\"]}")
	assert.NoError(t, err)
	println("===")
	fmt.Println(body)
	println("===")

}


