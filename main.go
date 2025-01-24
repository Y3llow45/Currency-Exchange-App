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

	popularCurrencies := []string{
		"USD", "EUR", "GBP", "RUB", "BGN", "JPY", "CNY", "AUD", "CAD",
		"CHF", "SEK", "NZD", "SGD", "ZAR",
	}

	baseCurrency := widget.NewSelect(popularCurrencies, func(value string) {
	})
	baseCurrency.PlaceHolder = "Select base currency"

	targetCurrency := widget.NewSelect(popularCurrencies, func(value string) {
	})
	targetCurrency.PlaceHolder = "Select target currency"

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
