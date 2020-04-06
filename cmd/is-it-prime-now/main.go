package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"

	"github.com/thedahv/prime-now-checker/pkg/checker"
	"github.com/thedahv/prime-now-checker/pkg/har"
)

const interval = 5 * time.Minute

func main() {
	a := app.New()
	state := appState{}

	w := a.NewWindow("Is it Prime Now?")

	output := widget.NewMultiLineEntry()
	output.Disable()

	var startButton *widget.Button
	var stopButton *widget.Button

	setStartButtonEnabled := func() {
		if state.request == nil {
			startButton.Enable()
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

				log.Println("waiting for windows")
				windows := <-windowChan
				log.Println("got windows")
				log.Println(windows)
				out := ""
				if len(windows) == 0 {
					out = "No checkout windows"
				} else {
					for _, window := range windows {
						out += string(window) + "\n"
					}
				}
				output.SetText(out)
			}
		}()
	})

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
		widget.NewEntry(),
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
	ctx         context.Context
	cancel      context.CancelFunc
	isChecking  bool
	phoneNumber string
	request     *http.Request
	windows     []checker.DeliveryWindow
}
