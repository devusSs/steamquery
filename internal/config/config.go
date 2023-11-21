package config

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type Config struct {
	SteamUserID64     uint64 `json:"steam_user_id_64"    required:"true"  print:"true"`
	SteamAPIKey       string `json:"steam_api_key"       required:"true"  print:"false"`
	MedianPriceDays   uint   `json:"median_price_days"   required:"false" print:"true"  default:"7"`
	Currency          string `json:"currency"            required:"false" print:"true"  default:"EUR"`
	DecimalSeparator  string `json:"decimal_separator"   required:"false" print:"true"  default:","`
	SpreadSheetID     string `json:"spreadsheet_id"      required:"true"  print:"false"`
	LastUpdatedCell   string `json:"last_updated_cell"   required:"false" print:"true"  default:"G2"`
	ErrorCell         string `json:"error_cell"          required:"false" print:"true"  default:"M2"`
	TotalValueCell    string `json:"total_value_cell"    required:"false" print:"true"  default:"M4"`
	DifferenceCell    string `json:"difference_cell"     required:"false" print:"true"  default:"M5"`
	StartingRow       uint   `json:"starting_row"        required:"false" print:"true"  default:"9"`
	ItemColumn        string `json:"item_column"         required:"false" print:"true"  default:"B"`
	AmountColumn      string `json:"amount_column"       required:"false" print:"true"  default:"F"`
	SinglePriceColumn string `json:"single_price_column" required:"false" print:"true"  default:"H"`
	TotalPriceColumn  string `json:"total_price_column"  required:"false" print:"true"  default:"J"`
}

func (c *Config) String() string {
	v := reflect.ValueOf(c).Elem()
	t := v.Type()

	m := make(map[string]interface{})
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		tag := t.Field(i).Tag.Get("print")

		if tag == "true" {
			m[t.Field(i).Tag.Get("json")] = field.Interface()
		}
	}

	content, err := json.Marshal(m)
	if err != nil {
		return fmt.Sprintf("error marshalling config: %v", err)
	}

	return string(content)
}

func (c *Config) validate() error {
	v := reflect.ValueOf(c).Elem()

	var validationErrors []string

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := v.Type().Field(i)

		requiredTag := fieldType.Tag.Get("required")
		defaultTag := fieldType.Tag.Get("default")

		if requiredTag == "true" && isZeroValue(field) {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("field \"%s\" is required but empty", fieldType.Tag.Get("json")),
			)
		}

		if defaultTag != "" && isZeroValue(field) {
			switch field.Kind() {
			case reflect.String:
				field.SetString(defaultTag)
			case reflect.Int:
				defaultValue, err := strconv.ParseInt(defaultTag, 10, 64)
				if err != nil {
					return fmt.Errorf(
						"error parsing default value for field \"%s\": %v",
						fieldType.Tag.Get("json"),
						err,
					)
				}
				field.SetInt(defaultValue)
			case reflect.Uint:
				defaultValue, err := strconv.ParseUint(defaultTag, 10, 64)
				if err != nil {
					return fmt.Errorf(
						"error parsing default value for field \"%s\": %v",
						fieldType.Tag.Get("json"),
						err,
					)
				}
				field.SetUint(defaultValue)
			}
		}
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening config file: %v", err)
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decoding config file: %v", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation: %v", err)
	}

	return &cfg, nil
}

func isZeroValue(field reflect.Value) bool {
	return reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()) ||
		(field.Kind() == reflect.Ptr && field.IsNil())
}
