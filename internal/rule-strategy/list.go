package rule_strategy

import (
	"github.com/sdcoffey/techan"

	"tinkoff-invest-bot/internal/config"
)

type RuleStrategy func(tradingConfig config.TradingConfig) (techan.RuleStrategy, *techan.TimeSeries)

const (
	shortWindow  = "short_window"
	middleWindow = "middle_window"
	longWindow   = "long_window"
	window       = "window"
)

// Тут объявляется список доступных стратегий, которые можно использовать в своих трейдинг конфигах.
// Чтобы расширить функционал, нужно создать функцию, которая удовлетворяет типу RuleStrategy
var (
	// List это единственное место, где задаются стратегии
	List = map[string]RuleStrategy{
		"simpleEMA":   simpleEMA,
		"doubleEMA":   doubleEMA,
		"tripleEMA":   tripleEMA,
		"simpleAroon": simpleAroon,
	}
	RequiredParameters = map[string][]string{
		"simpleEMA":   {window},
		"doubleEMA":   {shortWindow, longWindow},
		"tripleEMA":   {shortWindow, middleWindow, longWindow},
		"simpleAroon": {window},
	}
)
