package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	log := Log{}
	log.init(3)
	log.tostdout()

	assert.Equal(t, 3, log.ring.Len())
}

func TestFive(t *testing.T) {
	log := Log{}
	log.init(3)
	data := []string{"One", "Two", "Three", "Four", "Five","Six"}
	for _, s := range data {
		log.notice(s)
	}

	assert.Equal(t, 3, log.ring.Len())


	results := log.toarray()
	for i := range results {
		assert.Contains(t, results[i], "NOTICE")
	}

	expected := []string{"Six", "Five","Four"}
	for i := range expected {
		assert.Contains(t, results[i], expected[i])
	}

}

