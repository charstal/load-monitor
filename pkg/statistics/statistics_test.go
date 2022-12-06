package statistics

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	c, err := NewOfflineReader()
	assert.Nil(t, err)

	fmt.Printf("%v", c)
}
