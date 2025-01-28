package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	apiURL      = "https://v6.exchangerate-api.com/v6/6d7dcd183cb12ac6ee96ac6e/latest/USD"
	dataFile    = "exchange_rates.json"
	dateFormat  = "2006-01-02"
)

type ExchangeRates struct {
	Result           string             `json:"result"`
	TimeLastUpdateUnix int64            `json:"time_last_update_unix"`
	BaseCode         string             `json:"base_code"`
	ConversionRates  map[string]float64 `json:"conversion_rates"`
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
		return nil, fmt.Errorf("API returned non-success result")
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
	w := a.NewWindow("Currency Exchange")

	data, err := CheckAndFetchData()
	if err != nil {
		fmt.Println("Error fetching exchange rates:", err)
		return
	}

	currencies := []string{}
	for currency := range data.ConversionRates {
		currencies = append(currencies, currency)
	}

	baseCurrency := widget.NewSelect(currencies, func(value string) {})
	baseCurrency.PlaceHolder = "Select base currency"

	targetCurrency := widget.NewSelect(currencies, func(value string) {})
	targetCurrency.PlaceHolder = "Select target currency"

	amountEntry := widget.NewEntry()
	amountEntry.SetPlaceHolder("Enter amount")

	resultLabel := widget.NewLabel("Result will appear here")

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
		resultLabel.SetText(fmt.Sprintf("Converted Amount: %.2f %s", convertedValue, target))
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
