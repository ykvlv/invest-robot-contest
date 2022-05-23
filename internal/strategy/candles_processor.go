package strategy

import (
	"fmt"

	"github.com/iamjinlei/go-tachart/tachart"
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"go.uber.org/zap"

	"tinkoff-invest-bot/internal/config"
	"tinkoff-invest-bot/investapi"
	"tinkoff-invest-bot/pkg/sdk"
)

const (
	graphDirName string = "./graphs/"
)

type FinishEvent struct{}

// CandlesStrategyProcessor запускалка всех стратегий, работа которых основана на свечках
type CandlesStrategyProcessor struct {
	tradingConfig *config.TradingConfig
	sdk           *sdk.SDK
	logger        *zap.Logger

	timeSeries    *techan.TimeSeries
	TradingRecord *techan.TradingRecord
	ruleStrategy  *techan.RuleStrategy

	candles []tachart.Candle
	events  []tachart.Event

	blockChannel chan FinishEvent
}

func (w *CandlesStrategyProcessor) Init(candles []*techan.Candle) {
	for _, candle := range candles {
		if w.timeSeries.AddCandle(candle) {
			w.candles = append(w.candles, tachart.Candle{
				Label: candle.Period.Start.Format("02.01/15:04"),
				O:     candle.OpenPrice.Float(),
				H:     candle.MaxPrice.Float(),
				L:     candle.MinPrice.Float(),
				C:     candle.ClosePrice.Float(),
				V:     candle.Volume.Float(),
			})
		}
	}
}

// GenGraph генерирует график в .html и ложит его в директорию с графиками
func (w CandlesStrategyProcessor) GenGraph(dirname string, filename string) {
	err := config.CreateDirIfNotExist(dirname)
	if err != nil {
		w.logger.Info("Can't create dir")
	}
	cfg := tachart.NewConfig().
		SetChartWidth(1400).
		SetChartHeight(800).AddOverlay(tachart.NewEMA(100))

	c := tachart.New(*cfg)
	err = c.GenStatic(w.candles, w.events, dirname+filename)
	if err != nil {
		w.logger.Info("Can't gen graph")
	}
}

func (w *CandlesStrategyProcessor) AddEvent(op Operation, orderId string, executedPrice float64, totalAmount float64) {
	var eventType tachart.EventType
	switch op {
	case Buy:
		eventType = tachart.Open
	case Sell:
		eventType = tachart.Close
	case Hold:
	default:
	}
	w.events = append(w.events, tachart.Event{
		Type:  eventType,
		Label: w.candles[len(w.candles)-1].Label,
	})
	w.TradingRecord.Operate(techan.Order{
		Side:          techan.OrderSide(op),
		Security:      orderId,
		Price:         big.NewDecimal(executedPrice),
		Amount:        big.NewDecimal(totalAmount),
		ExecutionTime: w.timeSeries.LastCandle().Period.End,
	})
}

func (w *CandlesStrategyProcessor) Step(candle *techan.Candle, drawGraph bool) Operation {
	if w.timeSeries.AddCandle(candle) {
		w.candles = append(w.candles, tachart.Candle{
			Label: candle.Period.Start.Format("02.01/15:04"),
			O:     candle.OpenPrice.Float(),
			H:     candle.MaxPrice.Float(),
			L:     candle.MinPrice.Float(),
			C:     candle.ClosePrice.Float(),
			V:     candle.Volume.Float(),
		})
		fmt.Printf("Added candle %v for %s: %f\n", w.timeSeries.LastIndex(), w.tradingConfig.Ticker, candle.ClosePrice.Float())
		if drawGraph {
			go w.GenGraph(graphDirName, w.tradingConfig.Ticker+"_"+w.tradingConfig.AccountId+".html")
		}
	} // добавляем пришедшую свечу (неважно откуда)

	if w.ruleStrategy.ShouldEnter(w.timeSeries.LastIndex(), w.TradingRecord) {
		return Buy
	} else if w.ruleStrategy.ShouldExit(w.timeSeries.LastIndex(), w.TradingRecord) {
		return Sell
	} else {
		return Hold
	}
}

// Consume будет вызван для каждой новой свечки, которая соответствует figi в трейдинг конфиге
func (w CandlesStrategyProcessor) Consume(data *investapi.MarketDataResponse) {
	op := w.Step(
		CandleToTechanCandle(
			data.GetCandle(),
			sdk.IntervalToDuration(w.tradingConfig.StrategyConfig.Interval),
		),
		true,
	)

	switch op {
	case Buy:
		isEnough, trackingId, err := w.sdk.IsEnoughMoneyToBuy(
			w.tradingConfig.AccountId,
			w.tradingConfig.IsSandbox,
			w.tradingConfig.Figi,
			w.tradingConfig.Currency,
			w.tradingConfig.StrategyConfig.Quantity,
		)
		if err != nil {
			w.logger.Info(
				"Can't check available to buy share",
				zap.String("accountId", w.tradingConfig.AccountId),
				zap.String("figi", w.tradingConfig.Figi),
				zap.String("ticker", w.tradingConfig.Ticker),
				zap.String("ruleStrategy", w.tradingConfig.StrategyConfig.Name),
				zap.String("trackingId", trackingId),
				zap.Error(err),
			)
		}

		if isEnough {
			w.buy()
		} else {
			w.logger.Info(
				"Can't buy share because not enough money",
				zap.String("accountId", w.tradingConfig.AccountId),
				zap.String("figi", w.tradingConfig.Figi),
				zap.String("ticker", w.tradingConfig.Ticker),
				zap.String("ruleStrategy", w.tradingConfig.StrategyConfig.Name),
				zap.String("trackingId", trackingId),
			)
		}

	case Sell:
		isAvailable, trackingId, err := w.sdk.IsEnoughMoneyToBuy(
			w.tradingConfig.AccountId,
			w.tradingConfig.IsSandbox,
			w.tradingConfig.Figi,
			w.tradingConfig.Currency,
			w.tradingConfig.StrategyConfig.Quantity,
		)
		if err != nil {
			w.logger.Info(
				"Can't check available to sell share",
				zap.String("accountId", w.tradingConfig.AccountId),
				zap.String("figi", w.tradingConfig.Figi),
				zap.String("ticker", w.tradingConfig.Ticker),
				zap.String("ruleStrategy", w.tradingConfig.StrategyConfig.Name),
				zap.String("trackingId", trackingId),
				zap.Error(err),
			)
		}

		if isAvailable {
			w.sell()
		} else {
			w.logger.Info(
				"Can't sell share because not enough quantity of shares",
				zap.String("accountId", w.tradingConfig.AccountId),
				zap.String("figi", w.tradingConfig.Figi),
				zap.String("ticker", w.tradingConfig.Ticker),
				zap.String("ruleStrategy", w.tradingConfig.StrategyConfig.Name),
				zap.String("trackingId", trackingId),
			)
		}
	case Hold:
	default:
	}
}

func (w CandlesStrategyProcessor) buy() {
	orderId := sdk.GenerateOrderId()

	var resp *investapi.PostOrderResponse
	var trackingId string
	var err error

	if w.tradingConfig.IsSandbox {
		resp, trackingId, err = w.sdk.SandboxMarketBuy(
			w.tradingConfig.Figi,
			w.tradingConfig.StrategyConfig.Quantity,
			w.tradingConfig.AccountId,
			orderId,
		)
	} else {
		resp, trackingId, err = w.sdk.RealMarketBuy(
			w.tradingConfig.Figi,
			w.tradingConfig.StrategyConfig.Quantity,
			w.tradingConfig.AccountId,
			orderId,
		)
	}

	if err != nil {
		w.logger.Info(
			"Can't Buy share",
			zap.String("accountId", w.tradingConfig.AccountId),
			zap.String("figi", w.tradingConfig.Figi),
			zap.String("ticker", w.tradingConfig.Ticker),
			zap.Float64("price", sdk.MoneyValueToFloat(resp.GetExecutedOrderPrice())),
			zap.String("ruleStrategy", w.tradingConfig.StrategyConfig.Name),
			zap.String("orderId", orderId),
			zap.String("trackingId", trackingId),
			zap.Error(err),
		)
	} else {
		w.AddEvent(Buy, orderId, sdk.MoneyValueToFloat(resp.GetExecutedOrderPrice()), sdk.MoneyValueToFloat(resp.GetTotalOrderAmount()))

		w.logger.Info(
			"Buy new share",
			zap.String("accountId", w.tradingConfig.AccountId),
			zap.String("figi", w.tradingConfig.Figi),
			zap.String("ticker", w.tradingConfig.Ticker),
			zap.Float64("price", sdk.MoneyValueToFloat(resp.GetExecutedOrderPrice())),
			zap.String("ruleStrategy", w.tradingConfig.StrategyConfig.Name),
			zap.String("orderId", orderId),
			zap.String("trackingId", trackingId),
		)
	}
}

func (w CandlesStrategyProcessor) sell() {
	orderId := sdk.GenerateOrderId()

	var resp *investapi.PostOrderResponse
	var trackingId string
	var err error

	if w.tradingConfig.IsSandbox {
		resp, trackingId, err = w.sdk.SandboxMarketSell(
			w.tradingConfig.Figi,
			w.tradingConfig.StrategyConfig.Quantity,
			w.tradingConfig.AccountId,
			orderId,
		)
	} else {
		resp, trackingId, err = w.sdk.RealMarketSell(
			w.tradingConfig.Figi,
			w.tradingConfig.StrategyConfig.Quantity,
			w.tradingConfig.AccountId,
			orderId,
		)
	}

	if err != nil {
		w.logger.Info(
			"Can't sell new share",
			zap.String("accountId", w.tradingConfig.AccountId),
			zap.String("figi", w.tradingConfig.Figi),
			zap.String("ticker", w.tradingConfig.Ticker),
			zap.Float64("price", sdk.MoneyValueToFloat(resp.GetExecutedOrderPrice())),
			zap.String("ruleStrategy", w.tradingConfig.StrategyConfig.Name),
			zap.String("orderId", orderId),
			zap.String("trackingId", trackingId),
			zap.Error(err),
		)
	} else {
		w.AddEvent(Sell, orderId, sdk.MoneyValueToFloat(resp.GetExecutedOrderPrice()), sdk.MoneyValueToFloat(resp.GetTotalOrderAmount()))

		w.logger.Info(
			"Sell share",
			zap.String("accountId", w.tradingConfig.AccountId),
			zap.String("figi", w.tradingConfig.Figi),
			zap.String("ticker", w.tradingConfig.Ticker),
			zap.Float64("price", sdk.MoneyValueToFloat(resp.GetExecutedOrderPrice())),
			zap.Float64("income", w.TradingRecord.LastTrade().ExitOrder().Amount.Sub(w.TradingRecord.LastTrade().EntranceOrder().Amount).Float()),
			zap.String("ruleStrategy", w.tradingConfig.StrategyConfig.Name),
			zap.String("orderId", orderId),
			zap.String("trackingId", trackingId),
		)
		// w.blockChannel <- FinishEvent{}
	}
}

func (w CandlesStrategyProcessor) Start() error {
	var cons sdk.MarketDataConsumer = w
	err := w.sdk.SubscribeCandles(w.tradingConfig.Figi, sdk.IntervalToSubscriptionInterval(w.tradingConfig.StrategyConfig.Interval), &cons)
	if err != nil {
		return err
	}

	w.logger.Info(
		"Algorithm started",
		zap.String("figi", w.tradingConfig.Figi),
		zap.String("ruleStrategy", w.tradingConfig.StrategyConfig.Name),
	)

	return nil
}

func (w CandlesStrategyProcessor) Stop() error {
	var cons sdk.MarketDataConsumer = w
	if err := w.sdk.UnsubscribeCandles(w.tradingConfig.Figi, &cons); err != nil {
		return err
	}
	w.logger.Info(
		"Algorithm stopped",
		zap.String("figi", w.tradingConfig.Figi),
		zap.String("ruleStrategy", w.tradingConfig.StrategyConfig.Name),
	)
	return nil
}

func (w *CandlesStrategyProcessor) BlockUntilEnd() {
	<-w.blockChannel
}
