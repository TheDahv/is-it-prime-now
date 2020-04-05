package har

import (
	"os"
	"testing"
)

func TestParserFromFile(t *testing.T) {
	f, err := os.Open("./test-data/prime-now-request-with-options.har")
	if err != nil {
		t.Logf("could not open test file: %v", err)
		t.FailNow()
	}

	har, err := Parse(f)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if l := len(har.Log.Entries); l != 28 {
		t.Errorf("expected 28 entries, got %d", l)
	}
}
