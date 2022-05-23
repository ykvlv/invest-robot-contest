package sdk

import (
	"golang.org/x/xerrors"

	api "tinkoff-invest-bot/investapi"
)

// SubscribeCandles Подписать консьюмера на информацию о новых свечах
func (s *SDK) SubscribeCandles(figi string, interval api.SubscriptionInterval, consumer *MarketDataConsumer) error {
	consumers, contains := s.candlesConsumers[figi]
	if !contains {
		subscribeRequest := api.MarketDataRequest{
			Payload: &api.MarketDataRequest_SubscribeCandlesRequest{
				SubscribeCandlesRequest: &api.SubscribeCandlesRequest{
					SubscriptionAction: api.SubscriptionAction_SUBSCRIPTION_ACTION_SUBSCRIBE,
					Instruments: []*api.CandleInstrument{
						{
							Figi:     figi,
							Interval: interval,
						},
					},
				},
			},
		}
		if err := s.marketDataStreamClient.Send(&subscribeRequest); err != nil {
			return err
		}
		s.candlesConsumers[figi] = make([]*MarketDataConsumer, 0)
	}

	s.candlesConsumers[figi] = append(consumers, consumer)
	return nil
}

// UnsubscribeCandles Отписать консьюмера от информацию о новых свечах
func (s *SDK) UnsubscribeCandles(figi string, consumer *MarketDataConsumer) error {
	consumers, contains := s.candlesConsumers[figi]
	if !contains {
		return xerrors.Errorf("no such consumer subscribed on figi %s", figi)
	}

	for i, c := range consumers {
		if c == consumer {
			s.candlesConsumers[figi] = append(consumers[:i], consumers[i+1:]...)
			break
		}
	}

	if len(consumers) == 0 {
		unsubscribeRequest := api.MarketDataRequest{
			Payload: &api.MarketDataRequest_SubscribeCandlesRequest{
				SubscribeCandlesRequest: &api.SubscribeCandlesRequest{
					SubscriptionAction: api.SubscriptionAction_SUBSCRIPTION_ACTION_UNSUBSCRIBE,
					Instruments: []*api.CandleInstrument{
						{Figi: figi},
					},
				},
			},
		}
		err := s.marketDataStreamClient.Send(&unsubscribeRequest)
		if err != nil {
			return err
		}
		delete(s.candlesConsumers, figi)
	}
	return nil
}
