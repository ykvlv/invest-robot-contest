package strategy

import (
	"time"

	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"

	"tinkoff-invest-bot/investapi"
	"tinkoff-invest-bot/pkg/sdk"
)

func HistoricCandleToTechanCandle(c *investapi.HistoricCandle, period time.Duration) *techan.Candle {
	timePeriod := techan.NewTimePeriod(c.Time.AsTime(), period)
	candle := techan.NewCandle(timePeriod)

	candle.OpenPrice = big.NewDecimal(sdk.QuotationToFloat(c.Open))
	candle.ClosePrice = big.NewDecimal(sdk.QuotationToFloat(c.Close))
	candle.MaxPrice = big.NewDecimal(sdk.QuotationToFloat(c.High))
	candle.MinPrice = big.NewDecimal(sdk.QuotationToFloat(c.Low))
	candle.Volume = big.NewFromInt(int(c.Volume))
	return candle
}

func HistoricCandlesToTechanCandles(c []*investapi.HistoricCandle, period time.Duration) []*techan.Candle {
	techanCandles := make([]*techan.Candle, len(c))
	for i, candle := range c {
		techanCandles[i] = HistoricCandleToTechanCandle(candle, period)
	}
	return techanCandles
}

func CandleToTechanCandle(c *investapi.Candle, period time.Duration) *techan.Candle {
	timePeriod := techan.NewTimePeriod(c.Time.AsTime(), period)
	candle := techan.NewCandle(timePeriod)

	candle.OpenPrice = big.NewDecimal(sdk.QuotationToFloat(c.Open))
	candle.ClosePrice = big.NewDecimal(sdk.QuotationToFloat(c.Close))
	candle.MaxPrice = big.NewDecimal(sdk.QuotationToFloat(c.High))
	candle.MinPrice = big.NewDecimal(sdk.QuotationToFloat(c.Low))
	candle.Volume = big.NewFromInt(int(c.Volume))
	return candle
}
