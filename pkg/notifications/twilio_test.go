package notifications

import (
	"fmt"
	"os"
	"testing"
)

func TestTwilio(t *testing.T) {
	testSID := os.Getenv("TWILIO_ACCOUNT_SID")
	testAuth := os.Getenv("TWILIO_AUTH_TOKEN")
	fmt.Println("testSID", testSID)
	fmt.Println("testAuth", testAuth)

	from := "+15005550006"
	to := "+5571981265131"

	if len(os.Getenv("TWILIO_FROM_NUMBER")) > 0 {
		from = os.Getenv("TWILIO_FROM_NUMBER")
	}

	client, err := NewTwilio(testSID, testAuth, from)
	if err != nil {
		t.Errorf("expected no initialization errors, got %v", err)
	}
	ok, err := client.SendMessage(to, "hello from Twilio")
	if !ok {
		t.Errorf("expected ok, got %v", ok)
	}
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
