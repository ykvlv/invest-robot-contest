package sdk

import (
	api "tinkoff-invest-bot/investapi"
)

// MarketDataConsumer интерфейс получателя (потребителя) информации о MarketData
type MarketDataConsumer interface {
	// Consume будет вызываться для каждого нового сообщения из стрима MarketDataStream
	Consume(data *api.MarketDataResponse)
}
