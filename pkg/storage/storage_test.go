package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	_, err := NewStorage()
	assert.Nil(t, err)

}

func TestSimpleTest(t *testing.T) {
	client, err := NewStorage()
	assert.Nil(t, err)
	client.test()
}
