package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"go.uber.org/zap"

	"tinkoff-invest-bot/internal/config"
	"tinkoff-invest-bot/internal/engine"
	"tinkoff-invest-bot/pkg/graphics"
	"tinkoff-invest-bot/pkg/sdk"
)

func main() {
	logConf := zap.NewProductionConfig()
	err := config.CreateDirIfNotExist("./logs")
	if err != nil {
		log.Fatalf("Cant create dir: %v", err)
	}
	logConf.OutputPaths = []string{
		"stdout", "./logs/run-robot-stdout.log",
	}
	logConf.ErrorOutputPaths = []string{
		"stderr", "./logs/run-robot-stderr.log",
	}
	logger, err := logConf.Build()

	if err != nil {
		log.Fatalf("Cant create production logger: %v", err)
	}

	robotConfig := config.LoadRobotConfig("./configs/robot.yaml")
	tradingConfigs := config.LoadTradingConfigsFromDir("./configs/generated/")

	fmt.Println("Parsed trading configs:")
	for _, conf := range tradingConfigs {
		fmt.Printf("%v\n", conf)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s, err := sdk.New(robotConfig.TinkoffApiEndpoint, robotConfig.TinkoffAccessToken, robotConfig.AppName, ctx)
	if err != nil {
		logger.Fatal("Can't init SDK", zap.Error(err))
	}
	s.Run()

	go serveGraphics(8080, logger)

	var wg sync.WaitGroup
	for _, conf := range tradingConfigs {
		wg.Add(1)
		robotInstance, err := engine.New(robotConfig, conf, s, logger)
		if err != nil {
			logger.Fatal("Cant create robot instance", zap.Error(err))
		}
		go func() {
			robotInstance.Run()
			wg.Done()
		}()
	}
	wg.Wait()
}

func serveGraphics(port int, logger *zap.Logger) {
	for {
		http.Handle("/", graphics.NewGraphHandler(logger))
		fmt.Printf("http server listen http://localhost:%d/\n", port)
		err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
		if err != nil {
			logger.Error("Error in http server", zap.Error(err))
		}
	}
}
