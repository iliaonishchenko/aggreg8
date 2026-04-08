package osexit_test

import (
	"testing"

	"github.com/iliaonishchenko/aggreg8/cmd/staticlint/osexit"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestOSExitAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), osexit.Analyzer, "a", "b", "c")
}
