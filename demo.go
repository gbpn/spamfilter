package main

import (
	"fmt"

	"github.com/nnqq/bayesian"
)

const (
	good bayesian.Class = "Good"
	bad  bayesian.Class = "Bad"
)

func show(classifier *bayesian.Classifier, list []string) {
	scores, inx, strict, _ := classifier.SafeProbScores(list)
	fmt.Printf("s:%1.4f inx:%v strict: %v, words:%v\n", scores, inx, strict, list)
}

func demomain() {
	// Create a classifier with TF-IDF support.
	classifier := bayesian.NewClassifier("Good", "Bad")

	goodStuff := []string{"tall", "rich", "handsome", "the"}
	badStuff := []string{"poor", "smelly", "ugly", "the"}
	lessgoodStuff := []string{"employed"}

	classifier.Learn(goodStuff, "Good")
	classifier.Learn(badStuff, "Bad")

	show(classifier, []string{"the", "tall"})
	show(classifier, []string{"tall", "rich"})
	show(classifier, []string{"rich", "tall"})
	show(classifier, []string{"tall", "poor"})
	show(classifier, []string{"tall", "poor", "handsome"})
	show(classifier, []string{"tall", "poor", "ugly"})
	show(classifier, []string{"ugly", "poor"})
	show(classifier, []string{"silly", "boy"})
	show(classifier, []string{"employed", "ugly"})

	classifier.Learn(lessgoodStuff, "Good")
	fmt.Println()

	show(classifier, []string{"the", "tall"})
	show(classifier, []string{"tall", "rich"})
	show(classifier, []string{"rich", "tall"})
	show(classifier, []string{"tall", "poor"})
	show(classifier, []string{"tall", "poor", "handsome"})
	show(classifier, []string{"tall", "poor", "ugly"})
	show(classifier, []string{"ugly", "poor"})
	show(classifier, []string{"silly", "boy"})
	show(classifier, []string{"employed", "ugly"})

}
