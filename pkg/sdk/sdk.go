package sdk

import (
	"context"
	"crypto/tls"
	"io"
	"log"

	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	api "tinkoff-invest-bot/investapi"
)

type SDK struct {
	ctx  context.Context
	conn *grpc.ClientConn

	instruments      api.InstrumentsServiceClient
	marketData       api.MarketDataServiceClient
	marketDataStream api.MarketDataStreamServiceClient
	operations       api.OperationsServiceClient
	orders           api.OrdersServiceClient
	sandbox          api.SandboxServiceClient
	stopOrders       api.StopOrdersServiceClient
	users            api.UsersServiceClient

	marketDataStreamClient api.MarketDataStreamService_MarketDataStreamClient

	candlesConsumers map[string][]*MarketDataConsumer
}

// New создаёт новый инстанс SDK
func New(address string, token string, appName string, ctx context.Context) (*SDK, error) {
	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(
			credentials.NewTLS(&tls.Config{InsecureSkipVerify: true}),
		),
	)
	if err != nil {
		return nil, xerrors.Errorf("can't connect to gRPC server: %v", err)
	}

	ctx = prepareOutgoingContext(ctx, token, appName)

	marketDataStream := api.NewMarketDataStreamServiceClient(conn)
	stream, err := marketDataStream.MarketDataStream(ctx)
	if err != nil {
		return nil, xerrors.Errorf("can't careate market date stream: %v", err)
	}

	return &SDK{
		ctx:  ctx,
		conn: conn,

		instruments:      api.NewInstrumentsServiceClient(conn),
		marketData:       api.NewMarketDataServiceClient(conn),
		marketDataStream: marketDataStream,
		operations:       api.NewOperationsServiceClient(conn),
		orders:           api.NewOrdersServiceClient(conn),
		sandbox:          api.NewSandboxServiceClient(conn),
		stopOrders:       api.NewStopOrdersServiceClient(conn),
		users:            api.NewUsersServiceClient(conn),

		marketDataStreamClient: stream,

		candlesConsumers: make(map[string][]*MarketDataConsumer, 0),
	}, nil
}

// Run запускает двунаправленный стрим для получения информации
// Новую информацию об акции скармливает нужному консьюмеру
func (s *SDK) Run() {
	go func() {
		for {
			newMessage, err := s.marketDataStreamClient.Recv()
			if err == io.EOF {
				log.Fatalf("Data stream closed. No more data.")
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive new message : %v", err)
			}

			if newMessage != nil && newMessage.GetCandle() != nil { // notify candles subscribers
				figi := newMessage.GetCandle().GetFigi()
				figiConsumers, contains := s.candlesConsumers[figi]
				if contains {
					for _, consumer := range figiConsumers {
						(*consumer).Consume(newMessage)
					}
				}
			}
		}
	}()
}

// добавляет токен и app-name к запросу
func prepareOutgoingContext(ctx context.Context, token string, appName string) context.Context {
	md := metadata.New(map[string]string{
		"Authorization": "Bearer " + token,
		"x-app-name":    appName,
	})
	return metadata.NewOutgoingContext(ctx, md)
}

// извлекает сообщение об ошибке из API
func extractRequestError(md ...*metadata.MD) error {
	for _, m := range md {
		if errMessages, ok := (*m)["message"]; ok {
			return xerrors.Errorf("request error: %s", errMessages[0])
		}
	}
	return nil
}

// извлекает trackingId из ответа API
func extractTrackingId(md ...*metadata.MD) string {
	for _, m := range md {
		if trackingIds, ok := (*m)["x-tracking-id"]; ok {
			return trackingIds[0]
		}
	}
	return ""
}
