package sdk

import (
	"fmt"
	"time"
	"tinkoff-invest-bot/investapi"
)

var Intervals = []string{"1_MIN", "5_MIN"}

func IntervalToSubscriptionInterval(s string) investapi.SubscriptionInterval {
	switch s {
	case "1_MIN":
		return investapi.SubscriptionInterval_SUBSCRIPTION_INTERVAL_ONE_MINUTE
	case "5_MIN":
		return investapi.SubscriptionInterval_SUBSCRIPTION_INTERVAL_FIVE_MINUTES
	default:
		panic(fmt.Sprintf("Значение \"%s\" для интервала свечи не определено, есть только %s", s, Intervals))
	}
}

func IntervalToCandleInterval(s string) investapi.CandleInterval {
	switch s {
	case "1_MIN":
		return investapi.CandleInterval_CANDLE_INTERVAL_1_MIN
	case "5_MIN":
		return investapi.CandleInterval_CANDLE_INTERVAL_5_MIN
	default:
		panic(fmt.Sprintf("Значение \"%s\" для интервала свечи не определено, есть только %s", s, Intervals))
	}
}

func IntervalToDuration(s string) time.Duration {
	switch s {
	case "1_MIN":
		return time.Minute
	case "5_MIN":
		return time.Minute * 5
	default:
		panic(fmt.Sprintf("Значение \"%s\" для интервала свечи не определено, есть только %s", s, Intervals))
	}
}
