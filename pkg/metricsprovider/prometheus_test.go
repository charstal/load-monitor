package metricsprovider

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	opt = MetricsProviderOpts{
		Name:               PromClientName,
		Address:            "http://10.214.241.226:39090/",
		InsecureSkipVerify: true,
		AuthToken:          "",
	}
)

func TestHealthy(t *testing.T) {

	fmt.Println("hello")
	client, err := NewPromClient(opt)
	assert.Nil(t, err)
	fmt.Println("world")
	code, err := client.Healthy()
	fmt.Print(code)
	assert.Nil(t, err)
	assert.Equal(t, 0, code)
}
