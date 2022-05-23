package config

import (
	"io/fs"
	"log"

	"github.com/joho/godotenv"
)

const (
	writeMode = fs.FileMode(0755)
)

// Загружает параметры из .env файла
func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("Ошибка загрузки .env файла")
	}
}
