package checker

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/thedahv/prime-now-checker/pkg/har"
)

func TestParserFromFile(t *testing.T) {
	t.SkipNow()

	f, err := os.Open("./test-data/prime-now-request-with-options.har")
	if err != nil {
		t.Logf("could not open test file: %v", err)
		t.FailNow()
	}

	archive, err := har.Parse(f)
	checkoutRequest, ok := FindCheckoutRequest(archive)
	if !ok {
		t.Error("expected to find a checkout URL, but didn't find one")
	}
	t.Log(checkoutRequest)
}

func TestRequestOptionsFile(t *testing.T) {
	f, err := os.Open("./test-data/prime-now-request-with-options.har")
	if err != nil {
		t.Logf("could not open test file: %v", err)
		t.FailNow()
	}

	archive, err := har.Parse(f)
	checkoutRequest, ok := FindCheckoutRequest(archive)
	if !ok {
		t.Error("expected to find a checkout URL, but didn't find one")
	}

	req, err := BuildCheckoutRequestOptions(checkoutRequest)
	if err != nil {
		t.Errorf("expected no error building request options, got: %v", err)
	}

	if !strings.Contains(req.URL.String(), primeCheckoutURL) {
		t.Errorf("expected request to have checkout URL base, got: %s", req.URL)
	}
}

// TODO rewrite as something we can inject a reader instead of an HTTP response
func TestCheckoutRequest(t *testing.T) {
	t.SkipNow()

	f, err := os.Open("./test-data/prime-now-request-with-options.har")
	if err != nil {
		t.Logf("could not open test file: %v", err)
		t.FailNow()
	}

	archive, err := har.Parse(f)
	checkoutRequest, ok := FindCheckoutRequest(archive)
	if !ok {
		t.Error("expected to find a checkout URL, but didn't find one")
	}

	req, err := BuildCheckoutRequestOptions(checkoutRequest)
	if err != nil {
		t.Errorf("expected no error building request options, got: %v", err)
	}

	client := http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	defer resp.Body.Close()

	rdr, err := gzip.NewReader(resp.Body)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	body, err := ioutil.ReadAll(rdr)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	fmt.Println(string(body))
}

func TestParsePage(t *testing.T) {
	f, err := os.Open("./test-data/prime-now-request-with-options.html")
	if err != nil {
		t.Logf("unable to open test file: %v", err)
		t.FailNow()
	}

	windows, err := ParseWindowsFromCheckoutPage(f)
	if err != nil {
		t.Errorf("expected no parse error, got: %v", err)
	}

	if l := len(windows); l < 4 {
		t.Errorf("expected 4 windows, got %d", l)
	}
}
