package main

import (
	"errors"
	"strings"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nnqq/bayesian"
)

type reqClassifier struct {
	Classes []bayesian.Class `json:"types"`
}

type reqPredict struct {
	Phrase string `json:"phrase"`
}

type reqTrainData struct {
	Classes []bayesian.Class `json:"classes"`
	Phrases []string         `json:"phrases"`
}

type server struct {
	classifiers map[string]*bayesian.Classifier
}

func (s server) predict(c *gin.Context) {
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
	c.JSON(http.StatusOK, gin.H{"scores": scores, "inx": inx, "strict": strict})
}

func (s server) train(c *gin.Context) {
	name := c.Param("name")
	model, ok := s.classifiers[name]
	if !ok {
		c.AbortWithError(http.StatusNotFound, errors.New("Not found"))
		return
	}
	body := reqTrainData{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	for _, phrase := range body.Phrases {
		for _, class := range body.Classes {
			model.Learn(strings.Fields(phrase), class)
		}
	}
	c.JSON(http.StatusOK, gin.H{"result": "ok"})
}

func (s server) deleteClassifier(c *gin.Context) {
	name := c.Param("name")
	if _, ok := s.classifiers[name]; ok {
		delete(s.classifiers, name)
		c.JSON(http.StatusOK, gin.H{"result": "ok"})
		return
	}
	c.AbortWithError(http.StatusNotFound, errors.New("Not found"))
}

func (s server) addClassifier(c *gin.Context) {
	name := c.Param("name")
	body := reqClassifier{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if _, set := s.classifiers[name]; set {
		c.AbortWithError(http.StatusConflict, errors.New("Already exists"))
		return
	}
	s.classifiers[name] = bayesian.NewClassifier(body.Classes...)
	c.JSON(http.StatusOK, gin.H{"result": "ok"})
}

func (s server) runAPI() {
	router := gin.Default()
	router.PUT("/classifier/:name", s.addClassifier)
	router.DELETE("/classifier/:name", s.deleteClassifier)
	router.POST("/classifier/train/:name", s.train)
	router.GET("/classifier/predict/:name", s.predict)
	router.Run()
}
