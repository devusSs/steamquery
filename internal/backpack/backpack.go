package backpack

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var (
	ZeroPriceError = fmt.Errorf("item has no price")
)

func IsAvailable() bool {
	resp, err := http.Get(checkURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

type RequestOptions struct {
	MedianTime uint
	Currency   string
}

func (r *RequestOptions) checkDefaults() {
	if r.MedianTime == 0 {
		r.MedianTime = defaultMedianTime
	}
	if r.Currency == "" {
		r.Currency = defaultCurrency
	}
}

func GetItemPrice(marketHashName string, options ...*RequestOptions) (float64, error) {
	opt := &RequestOptions{}
	if len(options) > 0 {
		opt = options[0]
	}
	opt.checkDefaults()

	u, err := url.Parse(itemPriceURL)
	if err != nil {
		return 0, fmt.Errorf("parsing url: %w", err)
	}

	v := url.Values{}
	v.Set("id", marketHashName)
	v.Set("time", strconv.Itoa(int(opt.MedianTime)))
	v.Set("extend", "true")
	v.Set("currency", opt.Currency)
	u.RawQuery = v.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("doing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("received %d: %s", resp.StatusCode, resp.Status)
	}

	var res itemPriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return 0, fmt.Errorf("decoding response: %w", err)
	}

	if string(res.Success) != "true" {
		return 0, ZeroPriceError
	}

	median := strings.ReplaceAll(res.MedianPrice, ",", "")
	median = strings.ReplaceAll(median, "-", "0")

	price, err := strconv.ParseFloat(median, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing price: %w", err)
	}

	return price, nil
}

const (
	checkURL     string = "https://csgobackpack.net/"
	itemPriceURL string = "https://csgobackpack.net/api/GetItemPrice/"
)

const (
	defaultMedianTime uint   = 7
	defaultCurrency   string = "EUR"
)

type itemPriceResponse struct {
	Success           json.RawMessage `json:"success"`
	AveragePrice      string          `json:"average_price"`
	MedianPrice       string          `json:"median_price"`
	AmountSold        string          `json:"amount_sold"`
	StandardDeviation string          `json:"standard_deviation"`
	LowestPrice       string          `json:"lowest_price"`
	HighestPrice      string          `json:"highest_price"`
	FirstSaleDate     string          `json:"first_sale_date"`
	Time              string          `json:"time"`
	Icon              string          `json:"icon"`
	Currency          string          `json:"currency"`
}
