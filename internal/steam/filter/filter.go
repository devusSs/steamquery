package filter

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/devusSs/steamquery/pkg/steam"
)

func LoadFilterOptions(filePath string) error {
	if filePath == "" {
		filter = filterOptions{
			Tradable:   false,
			Marketable: true,
		}
		return nil
	}

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening filter file: %w", err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&filter); err != nil {
		return fmt.Errorf("error decoding filter file: %w", err)
	}

	return nil
}

func FilterInventoryResponse(
	inv steam.SteamInventoryResponse,
) steam.SteamInventoryResponse {
	r := steam.SteamInventoryResponse{}
	for _, item := range inv.Descriptions {
		if filter.Tradable && item.Tradable == 0 {
			continue
		}
		if filter.Marketable && item.Marketable == 0 {
			continue
		}
		r.Descriptions = append(r.Descriptions, item)
	}
	return r
}

func GetFilterSettings() filterOptions {
	return filter
}

func GetItemAmountMap(inv steam.SteamInventoryResponse) map[string]int {
	m := make(map[string]int)
	for _, item := range inv.Descriptions {
		m[item.MarketHashName]++
	}
	return m
}

func AddItems(amountMap map[string]int, itemsFile string) (map[string]int, error) {
	if itemsFile == "" {
		return amountMap, nil
	}

	f, err := os.Open(itemsFile)
	if err != nil {
		return nil, fmt.Errorf("item file not found: %w", err)
	}
	defer f.Close()

	var items itemFileStructure
	if err := json.NewDecoder(f).Decode(&items); err != nil {
		return nil, fmt.Errorf("could not unmarshal items: %w", err)
	}

	for _, item := range items.Items {
		if _, ok := amountMap[item.MarketHashName]; ok {
			amountMap[item.MarketHashName] += item.Amount
			continue
		}
		amountMap[item.MarketHashName] = item.Amount
	}

	return amountMap, nil
}

type filterOptions struct {
	Tradable   bool `json:"tradable"`
	Marketable bool `json:"marketable"`
}

func (f filterOptions) String() string {
	return fmt.Sprintf("tradable: %t, marketable: %t", f.Tradable, f.Marketable)
}

var (
	filter filterOptions
)

type itemFileStructure struct {
	Items []item `json:"items"`
}

type item struct {
	MarketHashName string `json:"market_hash_name"`
	Amount         int    `json:"amount"`
}
