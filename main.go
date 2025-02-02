package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var apiURL = "https://v6.exchangerate-api.com/v6/6d7dCd184cb12ac6ee96ac63/latest/USD"
const (
	dataFile = "exchange_rates.json"
	dateFormat = "2006-01-02"
)

type ExchangeRates struct {
	Result           string             `json:"result"`
	TimeLastUpdateUnix int64            `json:"time_last_update_unix"`
	BaseCode         string             `json:"base_code"`
	ConversionRates  map[string]float64 `json:"conversion_rates"`
}

var history []string

func addToHistory(from string, amount string, to string, converted float64) {
	amountValue, _ := strconv.ParseFloat(amount, 64)
	entry := fmt.Sprintf("%s: %.2f - %s: %.2f", from, amountValue, to, converted)
	history = append(history, entry)

	if len(history) > 100 { //keep last 100
		history = history[1:]
	}
}

func FetchData() (*ExchangeRates, error) {
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	var data ExchangeRates
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if data.Result != "success" {
		return nil, fmt.Errorf("api request failed")
	}
	return &data, nil
}

func SaveData(data *ExchangeRates) error {
	fileData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}
	return ioutil.WriteFile(dataFile, fileData, 0644)
}

func LoadData() (*ExchangeRates, error) {
	fileData, err := ioutil.ReadFile(dataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read data file: %w", err)
	}
	var data ExchangeRates
	if err := json.Unmarshal(fileData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}
	return &data, nil
}

func CheckAndFetchData() (*ExchangeRates, error) {
	if _, err := os.Stat(dataFile); err == nil {
		data, err := LoadData()
		if err == nil {
			lastUpdate := time.Unix(data.TimeLastUpdateUnix, 0)
			if lastUpdate.Format(dateFormat) == time.Now().Format(dateFormat) {
				return data, nil
			}
		}
	}

	data, err := FetchData()
	if err != nil {
		return nil, err
	}

	if err := SaveData(data); err != nil {
		return nil, err
	}
	return data, nil
}

func main() {
	a := app.New()
	w := a.NewWindow("Currency Exchange App")

  apiURLEntry := widget.NewEntry()
	apiURLEntry.SetPlaceHolder("Enter API URL")
	apiURLEntry.Hide()

  toggleButton := widget.NewButton("Set API URL", func() {
		if apiURLEntry.Visible() {
			apiURLEntry.Hide()
      newURL := apiURLEntry.Text
      if newURL != "" {
				apiURL = newURL
			}
		} else {
			apiURLEntry.Show()
		}
	})

	data, err := CheckAndFetchData()
	if err != nil {
		fmt.Println("Error fetching exchange rates:", err)
		return
	}

	currencies := []string{}
	for currency := range data.ConversionRates {
		currencies = append(currencies, currency)
	}
	sort.Strings(currencies)

	baseCurrency := widget.NewSelect(currencies, func(value string) {})
	baseCurrency.PlaceHolder = "Select base currency"

	targetCurrency := widget.NewSelect(currencies, func(value string) {})
	targetCurrency.PlaceHolder = "Select target currency"

	spacer := widget.NewLabel("")
	spacer.Hide()

	historyContainer := container.NewVBox()
	for _, entry := range history {
		historyContainer.Add(widget.NewLabel(entry))
	}

	amountEntry := widget.NewEntry()
	amountEntry.SetPlaceHolder("Enter amount")

	resultLabel := widget.NewLabel("Result will appear here")

	scrollableHistory := container.NewVScroll(historyContainer)
	scrollableHistory.SetMinSize(fyne.NewSize(400, 200))

	convertButton := widget.NewButton("Convert", func() {
		base := baseCurrency.Selected
		target := targetCurrency.Selected
		amount := amountEntry.Text

		if base == "" || target == "" || amount == "" {
			resultLabel.SetText("Please fill all fields")
			return
		}

		baseRate, baseOk := data.ConversionRates[base]
		targetRate, targetOk := data.ConversionRates[target]
		if !baseOk || !targetOk {
			resultLabel.SetText("Invalid currency selected")
			return
		}

		amountValue, err := strconv.ParseFloat(amount, 64)
		if err != nil {
			resultLabel.SetText("Invalid amount")
			return
		}

		convertedValue := (amountValue / baseRate) * targetRate
		addToHistory(base, amount, target, convertedValue)
		resultLabel.SetText(fmt.Sprintf("Converted Amount: %.2f %s", convertedValue, target))
		historyContainer.Add(widget.NewLabel(fmt.Sprintf("%s: %.2f - %s: %.2f", base, amountValue, target, convertedValue)))
		scrollableHistory.ScrollToBottom()
	})

	content := container.NewVBox(
		widget.NewLabel("Currency Exchange App"),
    toggleButton,
		apiURLEntry,
		baseCurrency,
		targetCurrency, 
		amountEntry,
		convertButton,
		resultLabel,
		widget.NewLabel("Conversion History:"),
		scrollableHistory,
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(400, 500))
	w.SetFixedSize(false)
	w.ShowAndRun()
}
