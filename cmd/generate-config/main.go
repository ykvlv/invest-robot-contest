package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strings"

	"github.com/fatih/color"

	"tinkoff-invest-bot/internal/config"
	"tinkoff-invest-bot/internal/rule-strategy"
	"tinkoff-invest-bot/investapi"
	"tinkoff-invest-bot/pkg/sdk"
	"tinkoff-invest-bot/pkg/utils"
)

var (
	scanner = bufio.NewScanner(os.Stdin)
	bold    = color.New(color.Bold).SprintfFunc()
)

const (
	configsPath     = "./configs/generated/"
	robotConfigPath = "./configs/robot.yaml"
	defaultQuantity = 1
)

func main() {
	fmt.Println(color.GreenString("🤖 Генератор конфига для торгового робота запущен!"))
	fmt.Println("Робот создан для торговли", color.MagentaString("базовыми акциями 📈"), "в Тинькофф Инвестиции")
	fmt.Println("Сгенерировать", color.MagentaString("конфигурации"), "вы можете с помощью инструкций ниже 💫")

	// Инициализация SDK
	robotConfig := config.LoadRobotConfig(robotConfigPath)
	if robotConfig.TinkoffAccessToken == "" {
		log.Fatalf("Токен доступа (TINKOFF_ACCESS_TOKEN) не был найден в .env")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s, err := sdk.New(robotConfig.TinkoffApiEndpoint, robotConfig.TinkoffAccessToken, robotConfig.AppName, ctx)
	if err != nil {
		log.Fatalf("Не удается инициализировать SDK: %v", err)
	}

	// Формирование информации об аккаунтах
	// TODO может это привести к общему виду? (isSandbox)
	isSandbox := utils.RequestBool("⏳ Сконфигурировать робота для работы в Sandbox?", scanner)
	var accounts []*investapi.Account
	if isSandbox {
		accounts, _, err = s.GetSandboxAccounts()
	} else {
		accounts, _, err = s.GetAccounts()
	}
	if err != nil {
		log.Fatalf("Не удается получить информацию об аккаунтах: %v", err)
	}
	invalidAccounts := 0
	var validAccounts []*investapi.Account
	var accountsInfo []string
	for _, account := range accounts {
		// Фильтрация аккаунтов на валидные и нет
		if account.GetType() == investapi.AccountType_ACCOUNT_TYPE_UNSPECIFIED ||
			account.GetStatus() != investapi.AccountStatus_ACCOUNT_STATUS_OPEN ||
			account.GetAccessLevel() != investapi.AccessLevel_ACCOUNT_ACCESS_LEVEL_FULL_ACCESS {
			invalidAccounts++
			continue
		}
		// Получение краткой информации об аккаунте
		var accountInfo string
		switch account.GetType() {
		// TODO можно ли торговать на инвест копилке? (бред)
		case investapi.AccountType_ACCOUNT_TYPE_INVEST_BOX:
			accountInfo += "🐷 "
		case investapi.AccountType_ACCOUNT_TYPE_TINKOFF_IIS:
			accountInfo += "🏦 "
		case investapi.AccountType_ACCOUNT_TYPE_TINKOFF:
			accountInfo += "💰 "
		}
		if account.GetName() != "" {
			accountInfo += account.GetName()
		} else {
			accountInfo += account.GetId()
		}
		var portfolio *investapi.PortfolioResponse
		// TODO может это привести к общему виду? (isSandbox)
		if isSandbox {
			portfolio, _, err = s.GetSandboxPortfolio(account.GetId())
		} else {
			portfolio, _, err = s.GetPortfolio(account.GetId())
		}
		if err != nil {
			log.Fatalf("Не удается получить портфолио аккаунта %s: %v", account.GetId(), err)
		}
		accountInfo += " " + portfolioReport(portfolio)
		accountsInfo = append(accountsInfo, accountInfo)
		validAccounts = append(validAccounts, account)
	}

	// Выбор аккаунта для торговли
	if invalidAccounts >= len(accounts) {
		log.Fatalln("По данному токену не найдено аккаунтов с доступом к торговле")
	} else if invalidAccounts > 0 {
		fmt.Printf(color.YellowString("Найдено аккаунтов без доступа к торговле")+": %d\n", invalidAccounts)
	}
	n := utils.RequestChoice("👤 Выберите аккаунт для торговли", accountsInfo, scanner)
	account := validAccounts[n]

	// Конфигурация стратегии
	var ruleStrategyNames []string
	for name := range rule_strategy.List {
		ruleStrategyNames = append(ruleStrategyNames, name)
	}
	n = utils.RequestChoice("🕹 Выберите стратегию из предложенных", ruleStrategyNames, scanner)
	ruleStrategyName := ruleStrategyNames[n]
	n = utils.RequestChoice("🕯 Выберите свечной интервал", sdk.Intervals, scanner)
	interval := sdk.Intervals[n]

	// Задание дополнительных параметров для стратегии
	requiredParameters := rule_strategy.RequiredParameters[ruleStrategyName]
	other := make(map[string]int, len(requiredParameters))
	for _, requiredParameter := range requiredParameters {
		requestInt := utils.RequestInt(fmt.Sprintf("📏 Введите параметр \"%s\" для %s", requiredParameter, ruleStrategyName), scanner)
		other[requiredParameter] = requestInt
	}

	strategyConfig := config.StrategyConfig{
		Name:     ruleStrategyName,
		Interval: interval,
		Quantity: defaultQuantity,
		Other:    other,
	}

	// Выбор акций для торговли
	responseShares, _, err := s.GetShares()
	if err != nil {
		log.Fatalf("Не удается получить информацию об акциях: %v", err)
	}

	// Создание конфигурации для каждого тикера
	isTryAgain := false
	for {
		var input string
		if isTryAgain {
			isTryAgain = false
			input = utils.RequestString("🏷 Уточните тикеры акций введенные неверно (через пробел)", scanner)
		} else {
			input = utils.RequestString("🛍 Введите тикеры акций для торговли (через пробел)", scanner)
		}
		inputTickers := strings.Split(input, " ")
	tickerLoop:
		for _, inputTicker := range inputTickers {
			for _, share := range responseShares {
				if share.GetTicker() == strings.ToUpper(inputTicker) {
					tradingConfig := config.TradingConfig{
						AccountId:      account.GetId(),
						IsSandbox:      isSandbox,
						Ticker:         share.GetTicker(),
						Figi:           share.GetFigi(),
						Exchange:       share.GetExchange(),
						Currency:       share.GetCurrency(),
						StrategyConfig: strategyConfig,
					}
					filename := share.GetTicker() + "_" + account.GetId() + ".yaml"
					if err = config.WriteTradingConfig(configsPath, filename, &tradingConfig); err != nil {
						color.Yellow("Торговая конфигурация %s не была записана %v", filename, err)
						isTryAgain = true
					}
					color.Green("Торговая конфигурация %s успешно записана", filename)
					continue tickerLoop
				}
			}
			color.Yellow("Инструмент с тикером \"%s\" не найден!", inputTicker)
			isTryAgain = true
		}
		if !isTryAgain {
			break
		}
	}
	fmt.Println("Вы можете изменять конфигурации вручную, если понимаете что делаете")
	color.Green("👍 Удачной торговли!")
}

func portfolioReport(portfolio *investapi.PortfolioResponse) string {
	totalAmount := sdk.MoneyValueToFloat(portfolio.GetTotalAmountCurrencies()) +
		sdk.MoneyValueToFloat(portfolio.GetTotalAmountBonds()) +
		sdk.MoneyValueToFloat(portfolio.GetTotalAmountShares()) +
		sdk.MoneyValueToFloat(portfolio.GetTotalAmountEtf()) +
		sdk.MoneyValueToFloat(portfolio.GetTotalAmountFutures())

	report := bold("%.2f₽ ", totalAmount)
	if portfolio.ExpectedYield != nil {
		expectedYield := sdk.QuotationToFloat(portfolio.ExpectedYield)

		income := fmt.Sprintf("%.2f₽ (%.2f%%)", totalAmount*expectedYield/100, math.Abs(expectedYield))
		switch {
		case expectedYield < 0:
			report += color.RedString(income)
		case expectedYield > 0:
			report += color.GreenString(income)
		default:
			report += color.WhiteString(income)
		}
		return report
	}
	return report
}
