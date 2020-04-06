package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"

	"github.com/thedahv/prime-now-checker/pkg/checker"
	"github.com/thedahv/prime-now-checker/pkg/har"
	"github.com/thedahv/prime-now-checker/pkg/notifications"
)

// TwilioFromNumber is the number notifications will come from. Injected at
// build time.
var TwilioFromNumber string

// TwilioSID is the Twilio account ID. Injected at build time.
var TwilioSID string

// TwilioAuthToken is the Twilio auth token. Injected at build time.
var TwilioAuthToken string

const notificationsMax = 5

const interval = 5 * time.Minute

func main() {
	a := app.New()
	state := appState{}

	notifier, err := notifications.NewTwilio(TwilioSID, TwilioAuthToken, TwilioFromNumber)
	if err != nil {
		log.Fatalf("unable to start Twilio client: %v", err)
	}

	w := a.NewWindow("Is it Prime Now?")

	output := widget.NewMultiLineEntry()
	output.Disable()

	var startButton *widget.Button
	var stopButton *widget.Button

	setStartButtonEnabled := func() {
		if state.request != nil && len(state.phoneNumber) > 0 {
			startButton.Enable()
		} else {
			startButton.Disable()
		}
	}

	toggleButtons := func() {
		if state.isChecking {
			startButton.Hide()
			stopButton.Show()
		} else {
			startButton.Show()
			stopButton.Hide()
		}
	}

	cancelChecking := func() {
		state.isChecking = false
		state.ctx = nil
		state.notificationsSent = 0
		state.snoozeMsgSent = false

		if state.cancel != nil {
			state.cancel()
		}
		state.cancel = nil
	}

	startButton = widget.NewButton("Start Checking", func() {
		state.isChecking = true
		defer toggleButtons()

		state.ctx, state.cancel = context.WithCancel(context.Background())

		windowChan, err := checker.CheckEvery(state.ctx, interval, state.request)
		if err != nil {
			output.SetText(err.Error())
			return
		}

		go func() {
			// TODO add a "next check at: ..." label
			for {
				if !state.isChecking {
					break
				}

				windows := <-windowChan
				out := ""
				if len(windows) == 0 {
					out = "No checkout windows"
				} else {
					for _, window := range windows {
						out += string(window) + "\n"
					}

					// Send notification
					if state.notificationsSent < notificationsMax {
						var msg string
						if len(windows) == 1 {
							msg = "There is 1 delivery window available for your Prime cart"
						} else {
							msg = fmt.Sprintf(
								"There are %d delivery windows available for your Prime cart",
								len(windows),
							)
						}
						notifier.SendMessage(state.phoneNumber, msg)
					} else if !state.snoozeMsgSent {
						notifier.SendMessage(
							state.phoneNumber,
							"I'm going to snooze for now. Restart me to resume notifications",
						)
					}
				}
				output.SetText(out)
			}
		}()
	})
	startButton.Disable()

	stopButton = widget.NewButton("Stop Checking", func() {
		cancelChecking()
		toggleButtons()
	})
	toggleButtons()

	phoneEntry := widget.NewEntry()
	phoneEntry.OnChanged = func(number string) {
		state.phoneNumber = number
		setStartButtonEnabled()
	}

	harEntryPath := widget.NewEntry()
	harEntryPath.OnChanged = func(path string) {
		defer setStartButtonEnabled()

		if len(path) == 0 {
			state.request = nil
			return
		}

		f, err := os.Open(path)
		if err != nil {
			output.SetText(err.Error())
			return
		}
		defer f.Close()

		parsed, err := har.Parse(f)
		if err != nil {
			output.SetText(err.Error())
			return
		}

		req, ok := checker.FindCheckoutRequest(parsed)
		if !ok {
			output.SetText("could not find a Checkout request")
			return
		}

		state.request, err = checker.BuildCheckoutRequestOptions(req)
		if err != nil {
			output.SetText(err.Error())
			return
		}
		// Clear any errors or output if there were any
		output.SetText("")
	}

	w.SetContent(widget.NewVBox(
		widget.NewLabel("Is it Prime Now?"),
		widget.NewLabel("Phone Number"),
		phoneEntry,
		widget.NewLabel("Checkout Page HAR"),
		harEntryPath,
		startButton,
		stopButton,
		output,
		widget.NewButton("Quit", func() {
			a.Quit()
		}),
	))

	w.ShowAndRun()
}

type appState struct {
	ctx               context.Context
	cancel            context.CancelFunc
	isChecking        bool
	notificationsSent int
	snoozeMsgSent     bool
	phoneNumber       string
	request           *http.Request
	windows           []checker.DeliveryWindow
}
