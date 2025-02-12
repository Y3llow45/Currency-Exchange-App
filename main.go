package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var apiURL string

const (
	dataFile   = "exchange_rates.json"
	dateFormat = "2006-01-02"
  configFile = "config.env"
)

type ExchangeRates struct {
	Result            string             `json:"result"`
	TimeLastUpdateUnix int64             `json:"time_last_update_unix"`
	BaseCode          string             `json:"base_code"`
	ConversionRates   map[string]float64 `json:"conversion_rates"`
  ApiURL            string             `json:"api_url"`
}

var history []string
var (
	errorLabel    *widget.Label
	baseCurrency  *widget.Select
	targetCurrency *widget.Select
  data          *ExchangeRates
)

func LoadConfig() error {
	fileData, err := ioutil.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
      apiURL = ""
			return ioutil.WriteFile(configFile, []byte("API_URL="), 0644)
		}
		return err
	}
	lines := strings.Split(string(fileData), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "API_URL=") {
			apiURL = strings.TrimPrefix(line, "API_URL=")
			break
		}
	}
	return nil
}

func SaveConfig() error {
	return ioutil.WriteFile(configFile, []byte(fmt.Sprintf("API_URL=%s", apiURL)), 0644)
}

func addToHistory(from string, amount string, to string, converted float64) {
	amountValue, _ := strconv.ParseFloat(amount, 64)
	entry := fmt.Sprintf("%s: %.2f - %s: %.2f", from, amountValue, to, converted)
	history = append(history, entry)

	if len(history) > 100 { // keep the last 100
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
		return nil, fmt.Errorf("API request failed")
	}
  data.ApiURL = apiURL
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
  apiURL = data.ApiURL
	return &data, nil
}

func CheckData() (*ExchangeRates, error) {
	if _, err := os.Stat(dataFile); err == nil {
		data, err := LoadData()
		if err == nil {
			lastUpdate := time.Unix(data.TimeLastUpdateUnix, 0)
			if time.Since(lastUpdate) <= 4*time.Second {
				return data, nil
			}
		}
	}
	return nil, fmt.Errorf("no valid data found")
}

func FetchAndSaveData() (*ExchangeRates, error) {
	data, err := FetchData()
	if err != nil {
		return nil, err
	}
	if err := SaveData(data); err != nil {
		return nil, err
	}
	return data, nil
}

func SetData() *ExchangeRates {
  LoadConfig()
  if apiURL == "" {
		errorLabel.SetText("Set API URL first")
		errorLabel.Show()
		return &ExchangeRates{ConversionRates: make(map[string]float64)}
	}

	data, err := CheckData()
	if err != nil {
		data, err = FetchAndSaveData()
		if err != nil {
			errorLabel.SetText(fmt.Sprintf("Error: %s", err.Error()))
			errorLabel.Show()
			return &ExchangeRates{ConversionRates: make(map[string]float64)}
		}
	}
	errorLabel.Hide()

	currencies := []string{}
	for currency := range data.ConversionRates {
		currencies = append(currencies, currency)
	}
	sort.Strings(currencies)

	baseCurrency.Options = currencies
	targetCurrency.Options = currencies
	baseCurrency.Refresh()
	targetCurrency.Refresh()

	return data
}

func main() {
	a := app.New()
	w := a.NewWindow("Currency Exchange App")
  LoadConfig()

	apiURLEntry := widget.NewEntry()
	apiURLEntry.SetPlaceHolder("Enter API URL")
	apiURLEntry.Hide()

	errorLabel = widget.NewLabel("")
	errorLabel.Hide()

	baseCurrency = widget.NewSelect([]string{}, func(value string) {})
	baseCurrency.PlaceHolder = "Select base currency"

	targetCurrency = widget.NewSelect([]string{}, func(value string) {})
	targetCurrency.PlaceHolder = "Select target currency"

  data = SetData()

	toggleButton := widget.NewButton("Set API URL", func() {
		if apiURLEntry.Visible() {
			apiURLEntry.Hide()
			newURL := apiURLEntry.Text
			if newURL != "" {
				apiURL = newURL
        SaveConfig()
        LoadConfig()
				SetData()
			}
		} else {
			apiURLEntry.Show()
		}
	})

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
		lastUpdate := time.Unix(data.TimeLastUpdateUnix, 0)
		if time.Since(lastUpdate) > 10*time.Minute {
			newData, err := FetchAndSaveData()
			if err == nil {
				data = newData
			}
		}

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
		errorLabel,
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
