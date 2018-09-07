package defercheck

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestBasic(t *testing.T) {
	testdata := analysistest.TestData()
	results := analysistest.Run(t, testdata, Analyzer, "basic")
	assert.Equal(t, 1, len(results))
	assert.NoError(t, results[0].Err)
}
