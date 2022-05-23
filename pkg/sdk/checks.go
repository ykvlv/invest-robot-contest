package sdk

import (
	"time"

	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"

	api "tinkoff-invest-bot/investapi"
)

// CanTradeNow Возможно ли торговать инструментом в данный момент времени
func (s *SDK) CanTradeNow(exchange string) (bool, string, error) {
	var header, trailer metadata.MD

	now := timestamppb.New(time.Now().UTC())
	resp, err := s.instruments.TradingSchedules(
		s.ctx, &api.TradingSchedulesRequest{
			From:     now,
			To:       now,
			Exchange: exchange,
		},
		grpc.Header(&header),
		grpc.Trailer(&trailer),
	)
	trackingId := extractTrackingId(&header, &trailer)

	if err != nil {
		if extractedError := extractRequestError(&trailer); extractedError != nil {
			return false, trackingId, extractedError
		}
		return false, trackingId, err
	}

	if len((*resp).GetExchanges()) > 0 && len((*resp).GetExchanges()[0].GetDays()) > 0 {
		schedule := (*resp).GetExchanges()[0].GetDays()[0]
		if !schedule.GetIsTradingDay() {
			return false, trackingId, nil
		}

		startTime := schedule.GetStartTime().AsTime()
		endTime := schedule.GetEndTime().AsTime()
		nowTime := time.Now().UTC()

		if startTime.Sub(nowTime) < 0 && nowTime.Sub(endTime) < 0 {
			return true, trackingId, err
		}
		return false, trackingId, err
	}
	return false, trackingId, nil
}

// IsEnoughMoneyToBuy достаточно ли денег на счёте для покупки акциий
func (s *SDK) IsEnoughMoneyToBuy(accountId string, isSandbox bool, figi string, currency string, quantity int64) (bool, string, error) {
	var positions *api.PositionsResponse
	var trackingId string
	var err error

	if isSandbox {
		positions, trackingId, err = s.GetSandboxPositions(accountId)
		if err != nil {
			return false, trackingId, xerrors.Errorf("can't receive Sandbox positions: %w", err)
		}
	} else {
		positions, trackingId, err = s.GetPositions(accountId)
		if err != nil {
			return false, trackingId, xerrors.Errorf("can't receive positions: %w", err)
		}
	}

	price, trackingId, err := s.GetLastPrice(figi)
	if err != nil {
		return false, trackingId, xerrors.Errorf("can't receive last price: %w", err)
	}

	for _, money := range positions.Money { // foreach our money
		if money.Currency == currency {
			if float64(quantity)*QuotationToFloat(price.Price) < MoneyValueToFloat(money) { // if enough to buy
				return true, trackingId, nil
			} else {
				return false, trackingId, nil // not enough money to buy
			}
		}
	}
	return false, trackingId, xerrors.Errorf("No money with currency %s", currency)
}

// IsAvailableForSale Есть ли у пользователя акции, чтобы продать их
func (s *SDK) IsAvailableForSale(accountId string, isSandbox bool, figi string, quantity int64) (bool, string, error) {
	var positions *api.PositionsResponse
	var trackingId string
	var err error

	if isSandbox {
		positions, trackingId, err = s.GetSandboxPositions(accountId)
		if err != nil {
			return false, trackingId, xerrors.Errorf("can't receive Sandbox positions: %w", err)
		}
	} else {
		positions, trackingId, err = s.GetPositions(accountId)
		if err != nil {
			return false, trackingId, xerrors.Errorf("can't receive positions: %w", err)
		}
	}

	for _, secur := range positions.GetSecurities() {
		if secur.Figi == figi {
			if secur.Balance >= quantity { // if enough to sell
				return true, trackingId, nil
			} else {
				return false, trackingId, nil
			}
		}
	}
	return false, trackingId, xerrors.Errorf("No security with figi %s", figi)
}
