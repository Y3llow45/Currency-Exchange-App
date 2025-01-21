package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Currency Exchange")

	baseCurrency := widget.NewEntry()
	baseCurrency.SetPlaceHolder("Enter base currency (e.g., USD)")

	targetCurrency := widget.NewEntry()
	targetCurrency.SetPlaceHolder("Enter target currency (e.g., EUR)")

	amountEntry := widget.NewEntry()
	amountEntry.SetPlaceHolder("Enter amount")

	resultLabel := widget.NewLabel("Result will appear here")

	convertButton := widget.NewButton("Convert", func() {
		resultLabel.SetText("20 bucks")
	})

	content := container.NewVBox(
		widget.NewLabel("Currency Exchange App"),
		baseCurrency,
		targetCurrency,
		amountEntry,
		convertButton,
		resultLabel,
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(400, 300))
	w.ShowAndRun()
}
