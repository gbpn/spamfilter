package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nnqq/bayesian"
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

func TestModelCreate(t *testing.T) {
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

	res = doReq(api, "GET", "/classifier/test1", "{}")
	assert.NoError(t, res.err)
	assert.Equal(t,200, res.response.Code)

}
func TestModelDelete(t *testing.T) {
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

func TestInvalidClass(t *testing.T) {
	var api server
	gin.SetMode("test")
	api.setupRouter()
	res := doReq(api, "PUT", "/classifier/missing", "{\"classes\":[\"good\",\"bad\"]}")
	invalidData := reqTrainData{
		Classes: []bayesian.Class{"nope"},
		Phrases: []string{ "No such class", },
	}
	invalidDatabody,_ := json.Marshal(invalidData)
	res = doReq(api, "POST", "/classifier/missing/train", string(invalidDatabody))
	assert.NoError(t, res.err)
	assert.Equal(t,404, res.response.Code)
	assert.Contains(t, res.body, "model missing does not have class nope")
}

func TestTrain(t *testing.T) {
	var api server
	gin.SetMode("test")
	api.setupRouter()
	res := doReq(api, "PUT", "/classifier/spam","{\"classes\":[\"good\",\"spam\"]}")

	goodData := reqTrainData{
		Classes: []bayesian.Class{"good"},
		Phrases: []string{
			"I love apples",
			"Dogs are the best",
			"Lets hang",
			"What is his \"problem\", dude?",
			"Please buy my product",
			"I need to see the doctor. It burns so bad",
			"Can I borrow 50 bucks?",
		},
	}
	goodDatabody,_ := json.Marshal(goodData)
	res = doReq(api, "POST", "/classifier/spam/train", string(goodDatabody))
	spamData := reqTrainData{
		Classes: []bayesian.Class{"spam"},
		Phrases: []string{
			"Donate to our campaign",
			"I am up for reelection",
			"Really, you should reelect me.",
			"Cody for reelection",
			"Your uncle died and left you a million dollars",
			"Give me your credit card number",
		},
	}
	spambody,_ := json.Marshal(spamData)
	res = doReq(api, "POST", "/classifier/spam/train", string(spambody))
	
	res = doReq(api, "GET", "/classifier/spam", "{}")
	assert.NoError(t, res.err)
	assert.Equal(t,200, res.response.Code)
	println(res.body)
}


