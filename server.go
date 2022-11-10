package main

import (
	"bytes"
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

type server struct {
	classes []bayesian.Class
	classifiers map[string]*bayesian.Classifier
	router *gin.Engine
}

func (s *server) setupRouter() {
	s.classifiers = make(map[string]*bayesian.Classifier)
	s.router = gin.New()

	s.router.Use(gin.Recovery())

	s.router.PUT("/classifier/:name", s.addClassifier)
	s.router.DELETE("/classifier/:name", s.deleteClassifier)
	s.router.GET("/classifier/:name", s.getClassifier)
	s.router.GET("/classifier/:name/raw", s.exportClassifier)
	s.router.POST("/classifier/:name/train", s.train)
	s.router.GET("/classifier/:name/predict", s.predict)
}

func (s *server) predict(c *gin.Context) {
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

	scores, inx, strict, _ := model.SafeProbScores([]string{body.Phrase})

	winner := "false"
	if strict {
		winner = "true"
	}

	c.JSON(http.StatusOK, gin.H{"scores": scores, "inx": s.classes[inx], "winner": winner})
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
	//c.JSONP(http.StatusOK, output)
}

func (s *server) exportClassifier(c *gin.Context) {
	name := c.Param("name")
	model, ok := s.classifiers[name]
	if !ok {
		c.AbortWithError(http.StatusNotFound, errors.New("Not found"))
		return
	}
	serialized := new(bytes.Buffer)
	err := model.WriteTo(serialized)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, serialized.String())
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

