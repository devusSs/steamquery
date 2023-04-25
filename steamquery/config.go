package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
)

type config struct {
	ItemList        []map[string]string `json:"item_list"`
	LastUpdatedCell string              `json:"last_updated_cell"`
	ErrorCell       string              `json:"error_cell"`
	SpreadSheetID   string              `json:"spreadsheet_id"`
	UpdateInterval  int                 `json:"update_interval"`
	SteamAPIKey     string              `json:"steam_api_key"`
}

func loadConfig(fileName string) (*config, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	input, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var cfg config

	err = json.Unmarshal(input, &cfg)

	return &cfg, err
}

func (c *config) checkConfig() error {
	if len(c.ItemList) == 0 {
		return errors.New("missing item list in config")
	}

	if c.LastUpdatedCell == "" {
		return errors.New("missing last updated cell in config")
	}

	if c.ErrorCell == "" {
		return errors.New("missing error cell in config")
	}

	if c.SpreadSheetID == "" {
		return errors.New("missing spreadsheet id in config")
	}

	if c.UpdateInterval == 0 {
		return errors.New("missing update interval in config")
	}

	if c.SteamAPIKey == "" {
		return errors.New("missing steam api key in config")
	}

	return nil
}
