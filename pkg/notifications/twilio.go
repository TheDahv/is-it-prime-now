package notifications

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const twilioBase = "https://api.twilio.com"
const twilioVersion = "2010-04-01"

// Twilio allows a program to send a notification via Twilio
type Twilio struct {
	AccountSID string
	AuthToken  string
	From       string
}

// NewTwilio creates a new Twilio notification client. It will error if the
// provided auth credentials are empty
func NewTwilio(accountSID string, authToken string, from string) (Twilio, error) {
	client := Twilio{AccountSID: accountSID, AuthToken: authToken, From: from}

	if accountSID == "" {
		return client, fmt.Errorf("Twilio accountSID cannot be empty")
	}
	if authToken == "" {
		return client, fmt.Errorf("Twilio authToken cannot be empty")
	}
	if from == "" {
		return client, fmt.Errorf("Twilio from cannot be empty")
	}

	return client, nil
}

// SendMessage sends a text message notification to the target number
func (t Twilio) SendMessage(to string, message string) (bool, error) {
	u, err := url.ParseRequestURI(twilioBase)
	if err != nil {
		return false, fmt.Errorf("could not create Twilio URL: %v", err)
	}
	u.Path = fmt.Sprintf("/%s/Accounts/%s/Messages.json", twilioVersion, t.AccountSID)

	v := url.Values{}
	v.Set("To", to)
	v.Set("From", t.From)
	v.Set("Body", message)
	rb := *strings.NewReader(v.Encode())

	fmt.Println("request URL:", u.String())
	req, err := http.NewRequest("POST", u.String(), &rb)
	if err != nil {
		return false, fmt.Errorf("could not prepare request: %v", err)
	}
	req.SetBasicAuth(t.AccountSID, t.AuthToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("could not make request: %v", err)
	}

	// TODO set up a debug mode and print the response
	//defer resp.Body.Close()
	//data, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println(data)

	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	return statusOK, nil
}
