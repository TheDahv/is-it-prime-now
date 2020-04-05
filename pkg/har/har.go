package har

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

type pair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// RequestHAR are the elements we want to extract out of the Prime Now checkout
// page request saved in HAR
type RequestHAR struct {
	Log struct {
		Entries []struct {
			Request Request `json:"request"`
		} `json:"entries"`
	} `json:"log"`
}

// Request is an HAR request
type Request struct {
	URL         string `json:"url"`
	Headers     []pair `json:"headers"`
	Cookies     []pair `json:"cookies"`
	QueryString []pair `json:"queryString"`
}

// FindRequest finds the first request that causes the matches function to
// return true when passed the request URL
func (h RequestHAR) FindRequest(matches func(string) bool) (Request, bool) {
	var r Request
	var ok bool

	for _, entry := range h.Log.Entries {
		if matches(entry.Request.URL) {
			return entry.Request, true
		}
	}

	return r, ok
}

// Parse reads a HAR file and returns its parsed representation
func Parse(input io.Reader) (RequestHAR, error) {
	var har RequestHAR

	data, err := ioutil.ReadAll(input)
	if err != nil {
		return har, fmt.Errorf("could not read input: %v", err)
	}

	if err := json.Unmarshal(data, &har); err != nil {
		return har, fmt.Errorf("could not parse HAR input: %v", err)
	}

	return har, nil
}
