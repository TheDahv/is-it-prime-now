package checker

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/thedahv/prime-now-checker/pkg/har"
)

const primeCheckoutURL = "https://primenow.amazon.com/checkout/enter-checkout"

// FindCheckoutRequest finds the originating GET request to the checkout page
func FindCheckoutRequest(har har.RequestHAR) (har.Request, bool) {
	return har.FindRequest(func(url string) bool {
		return strings.Contains(url, primeCheckoutURL)
	})
}

// BuildCheckoutRequestOptions builds a Go HTTP request that a client can use to
// request the checkout page from Amazon
func BuildCheckoutRequestOptions(archiveRequest har.Request) (*http.Request, error) {
	req, err := http.NewRequest("GET", archiveRequest.URL, nil)
	if err != nil {
		return req, fmt.Errorf("unable to create request: %v", err)
	}

	req.Form = make(url.Values)
	for _, q := range archiveRequest.QueryString {
		req.Form.Add(q.Name, q.Value)
	}
	for _, h := range archiveRequest.Headers {
		req.Header.Add(h.Name, h.Value)
	}
	/*
		for _, c := range archiveRequest.Cookies {
			req.AddCookie(&http.Cookie{
				Name:  c.Name,
				Value: url.QueryEscape(c.Value),
			})
		}
	*/

	return req, nil
}

// DeliveryWindow is a delivery option on the checkout page
type DeliveryWindow string

// Check calls Amazon Prime now at the given Request and attempts to return
// DeliveryWindows from the parsed page response
func Check(client http.Client, req *http.Request) ([]DeliveryWindow, error) {
	resp, err := client.Do(req)

	if err != nil {
		fmt.Printf("request error: %v\n", err)
		return []DeliveryWindow{}, fmt.Errorf("cannot make request to Amazon: %v", err)
	}
	defer resp.Body.Close()

	rdr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return []DeliveryWindow{}, fmt.Errorf("could not decompress response: %v", err)
	}

	return ParseWindowsFromCheckoutPage(rdr)
}

// CheckEvery will check the given request at the specified interval until the
// checker is canceled.
func CheckEvery(ctx context.Context, interval time.Duration, req *http.Request) (chan []DeliveryWindow, error) {
	out := make(chan []DeliveryWindow)
	client := http.Client{}

	go func() {
		// Run the first check
		windows, err := Check(client, req)
		if err != nil {
			log.Fatalf("cannot parse the checkout page: %v", err)
		}

		out <- windows

		// Set up a timer for every time thereafter
		timer := time.Tick(interval)
		select {
		case <-ctx.Done():
			{
				close(out)
				return
			}
		case <-timer:
			{
				windows, err := Check(client, req)
				if err != nil {
					log.Fatalf("cannot parse the checkout page: %v", err)
				}

				out <- windows
			}
		}
	}()

	return out, nil
}

const optionPath = "div.delivery-window-radio-button-section" +
	" span.a-radio-label" +
	" > span[data-testid].a-color-base"

// ParseWindowsFromCheckoutPage reads a checkout page and tries to parse out
// delivery window options from it
func ParseWindowsFromCheckoutPage(page io.Reader) ([]DeliveryWindow, error) {
	var results []DeliveryWindow

	document, err := goquery.NewDocumentFromReader(page)
	if err != nil {
		return results, fmt.Errorf("unable to load page: %v", err)
	}

	document.Find(optionPath).Each(func(index int, element *goquery.Selection) {
		window := strings.TrimSpace(element.Text())
		results = append(results, DeliveryWindow(window))
	})

	// TODO wrap in a debug flag
	debugFile, err := os.OpenFile("./debug.html", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("could not open debug file: %v", err)
	}
	raw, _ := goquery.OuterHtml(document.First())
	fmt.Fprint(debugFile, raw)
	debugFile.Close()

	return results, nil
}
