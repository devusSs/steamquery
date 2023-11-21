package format

import (
	"fmt"
	"strings"
)

func FormatPricePrintable(price float64, separator string, currencySign string) string {
	return addCurrencySign(replaceDecimalSeparator(price, separator), currencySign)
}

func FormatPriceCalculatable(price string, separator string, currencySign string) string {
	replacedDot := strings.ReplaceAll(price, ".", "")
	replacedSep := strings.ReplaceAll(replacedDot, separator, ".")
	return strings.ReplaceAll(replacedSep, currencySign, "")
}

func replaceDecimalSeparator(price float64, separator string) string {
	return strings.ReplaceAll(fmt.Sprintf("%.2f", price), ".", ",")
}

func addCurrencySign(price string, sign string) string {
	return fmt.Sprintf("%s%s", price, sign)
}
