package sdk

import (
	"tinkoff-invest-bot/investapi"
)

func QuotationToFloat(q *investapi.Quotation) float64 {
	return float64(q.Units) + float64(q.Nano)/1000000000
}

func MoneyValueToFloat(q *investapi.MoneyValue) float64 {
	return float64(q.Units) + float64(q.Nano)/1000000000
}
