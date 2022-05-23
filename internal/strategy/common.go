package strategy

import (
	"github.com/iamjinlei/go-tachart/tachart"
	"github.com/sdcoffey/techan"
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"tinkoff-invest-bot/internal/config"
	"tinkoff-invest-bot/internal/rule-strategy"
	"tinkoff-invest-bot/pkg/sdk"
)

type Operation int

const (
	Buy Operation = iota
	Sell
	Hold
)

// FromConfig создаёт CandlesStrategyProcessor по трейдинг конфигу
func FromConfig(tradingConfig *config.TradingConfig, s *sdk.SDK, logger *zap.Logger) (*CandlesStrategyProcessor, error) {
	f := rule_strategy.List[tradingConfig.StrategyConfig.Name]
	if f == nil {
		return nil, xerrors.Errorf("no ruleStrategy with name %s", tradingConfig.StrategyConfig.Name)
	}

	tradingRecord := techan.NewTradingRecord() // создание структуры стратегии и истории трейдинга
	ruleStrategy, timeSeries := f(*tradingConfig)

	tradingStrategy := CandlesStrategyProcessor{
		tradingConfig: tradingConfig,
		sdk:           s,
		logger:        logger,
		timeSeries:    timeSeries,
		TradingRecord: tradingRecord,
		ruleStrategy:  &ruleStrategy,
		candles:       []tachart.Candle{},
		events:        []tachart.Event{},
	}

	return &tradingStrategy, nil
}
