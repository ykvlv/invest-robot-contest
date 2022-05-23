package engine

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"tinkoff-invest-bot/internal/config"
	"tinkoff-invest-bot/internal/strategy"
	"tinkoff-invest-bot/pkg/sdk"
)

// investRobot Микро-робот, который после запуска начинает торговать укзанным инструментом согласно переданному трейдинг конфигу
type investRobot struct {
	robotConfig     *config.RobotConfig
	tradingConfig   *config.TradingConfig
	tradingStrategy *strategy.CandlesStrategyProcessor
	logger          *zap.Logger
	sdk             *sdk.SDK

	restartDelay time.Duration
}

// New создать новый инстанс микро-робота
func New(conf *config.RobotConfig, tradingConfig *config.TradingConfig, s *sdk.SDK, logger *zap.Logger) (*investRobot, error) {
	tradingStrategy, err := strategy.FromConfig(tradingConfig, s, logger)
	if err != nil {
		return nil, err
	}

	// При старте микро-робота он сразу же загружает предыдущие свечки,
	// чтобы моментально начать торговать
	c, _, err := s.GetCandles(
		tradingConfig.Figi,
		time.Now().AddDate(0, 0, -1),
		time.Now(),
		sdk.IntervalToCandleInterval(tradingConfig.StrategyConfig.Interval),
	)
	if err != nil {
		return nil, err
	}

	tradingStrategy.Init(strategy.HistoricCandlesToTechanCandles(c, sdk.IntervalToDuration(tradingConfig.StrategyConfig.Interval)))
	logger.Info(fmt.Sprintf("Initialization %s with %v candles", tradingConfig.Ticker, len(c)))

	return &investRobot{
		robotConfig:     conf,
		tradingConfig:   tradingConfig,
		tradingStrategy: tradingStrategy,
		logger:          logger,
		sdk:             s,

		restartDelay: 10 * time.Second,
	}, nil
}

// Run запускает микро-робота,
// микро-робот будет автоматически перезапускаться в случае ошибки
func (r *investRobot) Run() {
	for {
		r.logger.Info(
			"Micro-robot started",
			zap.String("ticker", r.tradingConfig.Figi),
		)

		if err := r.run(); err != nil {
			r.logger.Info(
				"Micro-robot finished with error",
				zap.String("ticker", r.tradingConfig.Figi),
				zap.Error(err),
			)
		} else {
			r.logger.Info(
				"Micro-robot finished successfully",
				zap.String("ticker", r.tradingConfig.Figi),
			)
		}

		time.Sleep(r.restartDelay)
	}
}

func (r *investRobot) run() error {
	canTrade, _, err := r.sdk.CanTradeNow(r.tradingConfig.Exchange)
	if err != nil {
		return xerrors.Errorf("can't receive trading schedules: %w", err)
	}
	if !canTrade {
		return xerrors.Errorf("instrument %s is not available, exchange is closed", r.tradingConfig.Ticker)
	}

	err = (*r.tradingStrategy).Start()
	if err != nil {
		return xerrors.Errorf("can't start robot trading strategy, %v", err)
	}

	(*r.tradingStrategy).BlockUntilEnd()

	err = (*r.tradingStrategy).Stop()
	if err != nil {
		return xerrors.Errorf("can't stop robot trading strategy, %v", err)
	}
	return nil
}
