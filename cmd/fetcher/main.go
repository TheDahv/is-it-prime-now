package main

import (
	"net/http"
	"os"

	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"

	"github.com/thedahv/prime-now-checker/pkg/checker"
	"github.com/thedahv/prime-now-checker/pkg/har"
)

func main() {
	a := app.New()
	state := appState{}

	w := a.NewWindow("Is it Prime Now?")

	output := widget.NewMultiLineEntry()
	output.Disable()

	client := http.Client{}

	submitButton := widget.NewButton("Check for Delivery Windows", func() {
		windows, err := checker.Check(client, state.request)
		if err != nil {
			output.SetText(err.Error())
			return
		}

		out := ""
		if len(windows) == 0 {
			out = "No checkout windows"
		} else {
			for _, window := range windows {
				out += string(window) + "\n"
			}
		}
		output.SetText(out)
	})

	phoneEntry := widget.NewEntry()
	phoneEntry.OnChanged = func(number string) {
		state.phoneNumber = number
	}

	harEntryPath := widget.NewEntry()
	harEntryPath.OnChanged = func(path string) {
		if len(path) == 0 {
			submitButton.Disable()
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

		submitButton.Enable()
	}

	w.SetContent(widget.NewVBox(
		widget.NewLabel("Is it Prime Now?"),
		widget.NewLabel("Phone Number"),
		widget.NewEntry(),
		widget.NewLabel("Checkout Page HAR"),
		harEntryPath,
		submitButton,
		output,
		widget.NewButton("Quit", func() {
			a.Quit()
		}),
	))

	w.ShowAndRun()
}

type appState struct {
	phoneNumber string
	request     *http.Request
	windows     []checker.DeliveryWindow
}
