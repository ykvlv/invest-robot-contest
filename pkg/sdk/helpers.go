package sdk

import (
	"math/rand"
	"time"
)

// GenerateOrderId Генерирует уникальный id ордера на покупку или продажу
func GenerateOrderId() string {
	const length = 36
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

	var random = rand.New(rand.NewSource(time.Now().UnixNano()))

	orderId := make([]byte, length)
	for i := range orderId {
		orderId[i] = charset[random.Intn(len(charset))]
	}
	return string(orderId)
}
