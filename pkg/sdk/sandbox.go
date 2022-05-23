package sdk

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	api "tinkoff-invest-bot/investapi"
)

// GetSandboxAccounts Получает все Sandbox аккаунты
func (s *SDK) GetSandboxAccounts() ([]*api.Account, string, error) {
	var header, trailer metadata.MD
	resp, err := s.sandbox.GetSandboxAccounts(
		s.ctx,
		&api.GetAccountsRequest{},
		grpc.Header(&header),
		grpc.Trailer(&trailer),
	)

	trackingId := extractTrackingId(&header, &trailer)

	if err != nil {
		if extractedError := extractRequestError(&trailer); extractedError != nil {
			return nil, trackingId, extractedError
		}
		return nil, trackingId, err
	}
	return resp.Accounts, trackingId, nil
}

// SandboxMarketSell выставляет ордер на покупку акции в Sandbox
func (s *SDK) SandboxMarketBuy(figi string, quantity int64, accountId string, orderId string) (*api.PostOrderResponse, string, error) {
	return s.postSandboxMarketOrder(figi, quantity, api.OrderDirection_ORDER_DIRECTION_BUY, accountId, orderId)
}

// SandboxMarketSell выставляет ордер на продажу акции в Sandbox
func (s *SDK) SandboxMarketSell(figi string, quantity int64, accountId string, orderId string) (*api.PostOrderResponse, string, error) {
	return s.postSandboxMarketOrder(figi, quantity, api.OrderDirection_ORDER_DIRECTION_SELL, accountId, orderId)
}

func (s *SDK) postSandboxMarketOrder(figi string, quantity int64, direction api.OrderDirection, accountId string, orderId string) (*api.PostOrderResponse, string, error) {
	var header, trailer metadata.MD

	resp, err := s.sandbox.PostSandboxOrder(
		s.ctx,
		&api.PostOrderRequest{
			Figi:      figi,
			Quantity:  quantity,
			Price:     nil,
			Direction: direction,
			AccountId: accountId,
			OrderType: api.OrderType_ORDER_TYPE_MARKET,
			OrderId:   orderId,
		},
		grpc.Header(&header),
		grpc.Trailer(&trailer),
	)

	trackingId := extractTrackingId(&header, &trailer)

	if err != nil {
		if extractedError := extractRequestError(&trailer); extractedError != nil {
			return nil, trackingId, extractedError
		}
		return nil, trackingId, err
	}
	return resp, trackingId, nil
}

// GetSandboxPositions Получает все активные позиции Sandbox аккаунта
func (s *SDK) GetSandboxPositions(accountId string) (*api.PositionsResponse, string, error) {
	var header, trailer metadata.MD
	resp, err := s.sandbox.GetSandboxPositions(
		s.ctx,
		&api.PositionsRequest{AccountId: accountId},
		grpc.Header(&header),
		grpc.Trailer(&trailer),
	)

	trackingId := extractTrackingId(&header, &trailer)

	if err != nil {
		if extractedError := extractRequestError(&trailer); extractedError != nil {
			return nil, trackingId, extractedError
		}
		return nil, trackingId, err
	}
	return resp, trackingId, nil
}

// GetSandboxPortfolio Получает портфолио Sandbox аккаунта
func (s *SDK) GetSandboxPortfolio(accountId string) (*api.PortfolioResponse, string, error) {
	var header, trailer metadata.MD

	resp, err := s.sandbox.GetSandboxPortfolio(
		s.ctx,
		&api.PortfolioRequest{
			AccountId: accountId,
		},
		grpc.Header(&header),
		grpc.Trailer(&trailer),
	)

	trackingId := extractTrackingId(&header, &trailer)

	if err != nil {
		if extractedError := extractRequestError(&trailer); extractedError != nil {
			return nil, trackingId, extractedError
		}
		return nil, trackingId, err
	}
	return resp, trackingId, nil
}
