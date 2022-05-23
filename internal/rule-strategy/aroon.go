package rule_strategy

import (
	"fmt"

	"github.com/sdcoffey/techan"

	"tinkoff-invest-bot/internal/config"
)

func simpleAroon(tradingConfig config.TradingConfig) (techan.RuleStrategy, *techan.TimeSeries) {
	var w = tradingConfig.StrategyConfig.Other[window]
	if w == 0 {
		panic(fmt.Sprintf("Значение %s в конфигурации %s_%s не обнаружено", window, tradingConfig.Ticker, tradingConfig.AccountId))
	}

	series := techan.NewTimeSeries()                   // история всех свечей
	lowPrices := techan.NewLowPriceIndicator(series)   // отсеивает High, Close, Open, на выходе только Low
	highPrices := techan.NewHighPriceIndicator(series) // на выходе только High

	aroonDownIndicator := techan.NewAroonDownIndicator(lowPrices, w)
	aroonUpIndicator := techan.NewAroonUpIndicator(highPrices, w)

	entryRule := techan.And( // правило входа
		techan.NewCrossUpIndicatorRule(aroonDownIndicator, aroonUpIndicator), // когда aroonUpIndicator пересечет (станет ВЫШЕ) aroonDownIndicator
		techan.PositionNewRule{}) // и сделок не открыто
	exitRule := techan.And( // правило выхода
		techan.NewCrossDownIndicatorRule(aroonUpIndicator, aroonDownIndicator), // когда aroonUpIndicator пересечет (станет НИЖЕ) aroonDownIndicator
		techan.PositionOpenRule{}) // и сделка открыта — продаем
	ruleStrategy := techan.RuleStrategy{
		UnstablePeriod: w, // период когда стратегия нестабильна
		EntryRule:      entryRule,
		ExitRule:       exitRule,
	}
	return ruleStrategy, series
}
