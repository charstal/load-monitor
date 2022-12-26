package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const RLClientUrl = "http://localhost:8000"

func TestRLClient(t *testing.T) {
	rlclient, err := NewRLClient(RLClientUrl)

	assert.Nil(t, err)

	err = rlclient.Healthy()

	assert.Nil(t, err)
}

func TestRLClientPredict(t *testing.T) {
	rlclient, err := NewRLClient(RLClientUrl)

	assert.Nil(t, err)
	podName := "123456"
	podLabel := "a"
	nodes := []string{"a", "b"}
	request, _ := MakePredictRequest(podName, podLabel, nodes)
	responce, err := rlclient.Predict(request)
	assert.Nil(t, err)
	fmt.Printf("%v", responce)
}
