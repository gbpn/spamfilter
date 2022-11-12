package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nnqq/bayesian"
)

type reqClassifier struct {
	Classes []bayesian.Class `json:"classes"`
}

type classifierInfo struct {
	Name string
	Learned int
	WordCount []int
	Classes []bayesian.Class 
	Words map[bayesian.Class]map[string]float64
}

type reqPredict struct {
	Phrase string `json:"phrase"`
}

type reqTrainData struct {
	Classes []bayesian.Class `json:"classes"`
	Phrases []string         `json:"phrases"`
}

type respSerialize struct {
	Classes []bayesian.Class
	WordCount []int
	Learned int
	Obj string
}

type server struct {
	classes []bayesian.Class
	classifiers map[string]*bayesian.Classifier
	router *gin.Engine
}

func (s *server) setupRouter() {
	s.classifiers = make(map[string]*bayesian.Classifier)
	s.router = gin.New()

	s.router.Use(gin.Recovery())

	s.router.GET("/", s.info)
	s.router.PUT("/classifier/:name", s.addClassifier)
	s.router.DELETE("/classifier/:name", s.deleteClassifier)
	s.router.GET("/classifier/:name", s.getClassifier)
	s.router.GET("/classifier/:name/export", s.exportClassifier)
	s.router.PUT("/classifier/:name/import", s.importClassifier)
	s.router.POST("/classifier/:name/train", s.train)
	s.router.GET("/classifier/:name/predict", s.predict)
}

// FIXME this does not yet work. json cant marshal a function name
func (s *server) info(c *gin.Context) {
	resp := respInfo{
		Routes : s.router.Routes(),
	}
	obj,_  := json.Marshal(resp)
	//println(err.Error())
	c.Data(http.StatusOK, "application/json", obj)
}


type respInfo struct {
	Routes []gin.RouteInfo `json:"Routes"`
}

func (s *server) predict(c *gin.Context) {
	var total float64

	name := c.Param("name")
	model, ok := s.classifiers[name]
	if !ok {
		c.AbortWithError(http.StatusNotFound, errors.New("Not found"))
		return
	}
	body := reqPredict{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	input := strings.Split(body.Phrase," ")
	scores, inx, strict, _ := model.SafeProbScores(input)
	percents := make([]float64, len(scores))

	for _,v := range scores {
		total += v
	}

	for i,v := range scores {
		percents[i] = float64(int(v*10000)) / 100
	}

	c.JSON(http.StatusOK, gin.H{
		"id": inx,
		"name": s.classes[inx],
		"percent": float64(int(scores[inx]*10000))/100,
		"percents": percents,
		"raw": scores,
		"winner": strict,
	})
}

func (s *server) getClassifier(c *gin.Context) {
	name := c.Param("name")
	model, ok := s.classifiers[name]
	

	if !ok {
		c.AbortWithError(http.StatusNotFound, errors.New("Not found"))
		return
	}
	output := classifierInfo{}
	output.Name = name
	output.Learned = model.Learned()
	output.Classes = model.Classes
	output.WordCount = model.WordCount()
	output.Words=make(map[bayesian.Class]map[string]float64)
	for _,class := range model.Classes {
		output.Words[class] = model.WordsByClass(class)
	}
	out, _ := json.MarshalIndent(output,"","   ")
	c.Data(http.StatusOK, "application/json", out)
}

func (s *server) importClassifier(c *gin.Context) {
	name := c.Param("name")

	data := respSerialize{}

	if err := c.ShouldBindJSON(&data); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	raw,err := b64.StdEncoding.DecodeString(data.Obj)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	obj := strings.NewReader(string(raw))

	model, err :=  bayesian.NewClassifierFromReader(obj)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.classifiers[name] = model

	c.JSON(http.StatusOK, gin.H{"result": "ok"})
}
func (s *server) exportClassifier(c *gin.Context) {
	name := c.Param("name")
	model, ok := s.classifiers[name]
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("Can not find model: %s", name)})
		return
	}

	serialized := new(bytes.Buffer)
	err := model.WriteTo(serialized)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := respSerialize{
		Classes: model.Classes,
	  Learned : model.Learned(),
	  WordCount : model.WordCount(),
		Obj: b64.StdEncoding.EncodeToString(serialized.Bytes()),
	}
	obj,err  := json.Marshal(out)
	c.Data(http.StatusOK, "application/json", obj)
}

func (s *server) modelEnsureClass(name string, class bayesian.Class) (bool) {
	model, ok := s.classifiers[name]
	if !ok {
		return false
	}
	for _,v := range model.Classes {
		if v == class {
			return true
		}
	}
	return false
}


func (s *server) train(c *gin.Context) {
	name := c.Param("name")
	model, ok := s.classifiers[name]
	if !ok {
	  c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	body := reqTrainData{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	for _, v := range body.Classes {
		if s.modelEnsureClass(name, v) != true {
			msg := fmt.Sprintf("model %s does not have class %s", name, v)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": msg})
			return 
		}
	}

	for _, phrase := range body.Phrases {
		for _, class := range body.Classes {
			model.Learn(strings.Fields(phrase), class)
		}
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok"})
}

func (s *server) deleteClassifier(c *gin.Context) {
	name := c.Param("name")
	if _, ok := s.classifiers[name]; ok {
		delete(s.classifiers, name)
		c.JSON(http.StatusOK, gin.H{"result": "ok"})
		return
	}
	c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "not found"})
}

func (s *server) addClassifier(c *gin.Context) {
	name := c.Param("name")
	body := reqClassifier{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError , gin.H{"error": err.Error()})
		return
	}
	if _, set := s.classifiers[name]; set {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error":"already exists"})
		return
	}
	if len(body.Classes) < 2 {
		c.AbortWithStatusJSON(http.StatusBadRequest , gin.H{"error": "At least 2 classes must be provided"})
		return
	}
	s.classifiers[name] = bayesian.NewClassifier(body.Classes...)
	s.classes = body.Classes
	c.JSON(http.StatusOK, gin.H{"result": "ok"})
}


func (s *server) Go() {
	s.setupRouter()
	s.router.Run()
}

