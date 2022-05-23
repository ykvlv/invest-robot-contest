package rule_strategy

import (
	"fmt"

	"github.com/sdcoffey/techan"

	"tinkoff-invest-bot/internal/config"
)

func simpleEMA(tradingConfig config.TradingConfig) (techan.RuleStrategy, *techan.TimeSeries) {
	var w = tradingConfig.StrategyConfig.Other[window]
	if w == 0 {
		panic(fmt.Sprintf("Значение %s в конфигурации %s_%s не обнаружено", window, tradingConfig.Ticker, tradingConfig.AccountId))
	}

	series := techan.NewTimeSeries()                        // история всех свечей
	closePrices := techan.NewClosePriceIndicator(series)    // отсеивает High, Low, Open, на выходе только Close
	movingAverage := techan.NewEMAIndicator(closePrices, w) // Создает экспоненциальное среднее с окном в n свечей

	entryRule := techan.And( // правило входа
		techan.NewCrossUpIndicatorRule(movingAverage, closePrices), // когда свеча закрытия пересечет EMA (станет выше EMA)
		techan.PositionNewRule{})                                   // и сделок не открыто — мы покупаем
	exitRule := techan.And( // правило выхода
		techan.NewCrossDownIndicatorRule(closePrices, movingAverage), // когда свеча закроется ниже EMA
		techan.PositionOpenRule{})                                    // и сделка открыта — продаем
	ruleStrategy := techan.RuleStrategy{
		UnstablePeriod: w,
		EntryRule:      entryRule,
		ExitRule:       exitRule,
	}
	return ruleStrategy, series
}

func doubleEMA(tradingConfig config.TradingConfig) (techan.RuleStrategy, *techan.TimeSeries) {
	var sw = tradingConfig.StrategyConfig.Other[shortWindow]
	var lw = tradingConfig.StrategyConfig.Other[longWindow]
	if sw == 0 || lw == 0 {
		panic(fmt.Sprintf("Значения %s или %s в конфигурации %s_%s не обнаружены", shortWindow, longWindow, tradingConfig.Ticker, tradingConfig.AccountId))
	}

	series := techan.NewTimeSeries()                             // история всех свечей
	closePrices := techan.NewClosePriceIndicator(series)         // отсеивает High, Low, Open, на выходе только Close
	shortEMAIndicator := techan.NewEMAIndicator(closePrices, sw) // Создает экспоненциальное средне с окном в n свечей
	longEMAIndicator := techan.NewEMAIndicator(closePrices, lw)

	entryRule := techan.And( // правило входа
		techan.NewCrossUpIndicatorRule(longEMAIndicator, shortEMAIndicator), // когда короткая EMA пересечет (станет ВЫШЕ) длинную EMA
		techan.PositionNewRule{}) // и сделок не открыто
	exitRule := techan.And( // правило выхода
		techan.NewCrossDownIndicatorRule(shortEMAIndicator, longEMAIndicator), // когда короткая EMA пересечет (станет НИЖЕ) длинную EMA
		techan.PositionOpenRule{}) // и сделка открыта — продаем
	ruleStrategy := techan.RuleStrategy{
		UnstablePeriod: lw, // период когда стратегия нестабильна
		EntryRule:      entryRule,
		ExitRule:       exitRule,
	}
	return ruleStrategy, series
}

func tripleEMA(tradingConfig config.TradingConfig) (techan.RuleStrategy, *techan.TimeSeries) {
	var sw = tradingConfig.StrategyConfig.Other[shortWindow]
	var mw = tradingConfig.StrategyConfig.Other[middleWindow]
	var lw = tradingConfig.StrategyConfig.Other[longWindow]
	if sw == 0 || lw == 0 || mw == 0 {
		panic(fmt.Sprintf("Значения %s или %s или %s в конфигурации %s_%s не обнаружены", shortWindow, middleWindow, longWindow, tradingConfig.Ticker, tradingConfig.AccountId))
	}

	series := techan.NewTimeSeries()                             // история всех свечей
	closePrices := techan.NewClosePriceIndicator(series)         // отсеивает High, Low, Open, на выходе только Close
	shortEMAIndicator := techan.NewEMAIndicator(closePrices, sw) // Создает экспоненциальное средне с окном в n свечей
	middleEMAIndicator := techan.NewEMAIndicator(closePrices, mw)
	longEMAIndicator := techan.NewEMAIndicator(closePrices, lw)

	entryRule := techan.And( // правило входа
		techan.And(
			techan.NewCrossUpIndicatorRule(middleEMAIndicator, shortEMAIndicator), // когда короткая EMA пересечет (станет ВЫШЕ) среднюю EMA
			techan.NewCrossUpIndicatorRule(longEMAIndicator, middleEMAIndicator),  // и средняя EMA пересечет (станет ВЫШЕ) длинную EMA
		),
		techan.PositionNewRule{}) // и сделок не открыто
	exitRule := techan.And( // правило выхода
		techan.NewCrossDownIndicatorRule(shortEMAIndicator, middleEMAIndicator), // когда короткая EMA пересечет (станет НИЖЕ) среднюю EMA
		techan.PositionOpenRule{}) // и сделка открыта — продаем
	ruleStrategy := techan.RuleStrategy{
		UnstablePeriod: lw, // период когда стратегия нестабильна
		EntryRule:      entryRule,
		ExitRule:       exitRule,
	}
	return ruleStrategy, series
}
