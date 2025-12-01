package exitcheck_test

import (
	"testing"

	"github.com/koyif/metrics/pkg/exitcheck"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, exitcheck.Analyzer, "a", "main", "alias")
}
