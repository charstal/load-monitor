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

func TestCopyFile(t *testing.T) {
	c, err := NewOfflineReader()
	c.pullFromEtcd()
	assert.Nil(t, err)
	err = c.fetchStatisticsFile()
	assert.Nil(t, err)

	fmt.Printf("%v", c)
}

func TestCheckFileMD5(t *testing.T) {
	c, err := NewOfflineReader()
	c.pullFromEtcd()
	assert.Nil(t, err)
	err = c.fetchStatisticsFile()
	assert.Nil(t, err)
	bo, err := c.checkFileMd5(c.tmpFilePath, c.tmpFileMD5)
	assert.Nil(t, err)
	assert.Equal(t, bo, true)

}

func TestRenameFile(t *testing.T) {
	c, err := NewOfflineReader()
	c.pullFromEtcd()
	assert.Nil(t, err)
	err = c.fetchStatisticsFile()
	assert.Nil(t, err)
	err = c.transferTmpFile2LocalFile()
	assert.Nil(t, err)
}

func TestLoadCsv(t *testing.T) {
	c, err := NewOfflineReader()
	c.pullFromEtcd()
	assert.Nil(t, err)
	err = c.fetchStatisticsFile()
	assert.Nil(t, err)
	err = c.transferTmpFile2LocalFile()
	assert.Nil(t, err)
	err = c.readFromCsv()
	assert.Nil(t, err)

	t.Logf("%v", c.statisData)
}

func TestUpdate(t *testing.T) {
	c, err := NewOfflineReader()
	assert.Nil(t, err)

	err = c.Update()
	assert.Nil(t, err)

	t.Logf("%v", c)
}

func TestGetMetrics(t *testing.T) {
	c, err := NewOfflineReader()
	assert.Nil(t, err)

	err = c.Update()
	assert.Nil(t, err)

	t.Logf("%v", c.GetMetrics())
}
