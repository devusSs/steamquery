package main

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	totalValuePreRun  float64
	totalValuePostRun float64
)

func (s *spreadsheetService) getTotalValueCell(cfg *config) (float64, error) {
	res, err := s.service.Spreadsheets.Values.Get(spreadsheetID, cfg.TotalValueCell).Do()
	if err != nil {
		return 0, err
	}

	if len(res.Values) == 0 {
		return 0, nil
	}

	value := res.Values[0]
	valueString := fmt.Sprintf("%v", value)
	valueWithoutCurrency := strings.ReplaceAll(valueString, "€", "")
	valueWithoutDot := strings.ReplaceAll(valueWithoutCurrency, ".", "")
	valueProperComma := strings.ReplaceAll(valueWithoutDot, ",", ".")
	valueFinal := strings.ReplaceAll(valueProperComma, "[", "")
	valueFinal = strings.ReplaceAll(valueFinal, "]", "")

	return strconv.ParseFloat(valueFinal, 64)
}

func calculateTotalValueDifference() string {
	result := totalValuePostRun - totalValuePreRun
	return fmt.Sprintf("%.2f", result)
}
