package utils

import (
	"fmt"

	"github.com/shopspring/decimal"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func FormatKES(price decimal.Decimal) string {
	return message.NewPrinter(language.English).Sprintf("%s", price.StringFixed(2))
}

func SeatLabel(row string, number int) string {
	return fmt.Sprintf("%s%d", row, number)
}
