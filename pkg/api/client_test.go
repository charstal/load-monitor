package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceClient(t *testing.T) {
	client, err := NewServiceClient("http://10.214.241.226:32020")
	assert.Nil(t, err)

	m, err := client.GetLatestWatcherMetrics()
	assert.Nil(t, err)
	fmt.Printf("%v", m)

	err = client.Healthy()
	assert.Nil(t, err)
}
