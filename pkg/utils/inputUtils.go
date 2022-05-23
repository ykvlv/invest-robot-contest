package utils

import (
	"bufio"
	"fmt"
	"strconv"
	"time"

	"github.com/fatih/color"
)

// RequestString Запросить у пользователя параметр в виде строки
func RequestString(msg string, scanner *bufio.Scanner) string {
	for {
		fmt.Printf(color.BlueString(msg) + ": ")
		if !scanner.Scan() {
			if scanner.Err() == nil {
				panic("Ввод из консоли принудительно завершен")
			} else {
				color.Yellow("Не удалось прочитать из консоли: %v", scanner.Err())
				continue
			}
		}
		input := scanner.Text()
		if input != "" {
			return input
		}
	}
}

// RequestChoice Запросить у пользователя выбор строки из предложенных строк
func RequestChoice(msg string, a []string, scanner *bufio.Scanner) int {
	if len(a) <= 0 {
		panic("Ошибка, в RequestChoice передано 0 возможных значений")
	}
	for i, aa := range a {
		fmt.Printf("%d. %s\n", i, aa)
	}
	for {
		if n := RequestInt(msg, scanner); n >= 0 && n < len(a) {
			return n
		} else {
			color.Yellow("Введите число в промежутке [%d, %d]", 0, len(a)-1)
		}
	}
}

const yes, no = "y", "n"

// RequestBool Запросить у пользователя yes or no
func RequestBool(msg string, scanner *bufio.Scanner) bool {
	for {
		input := RequestString(fmt.Sprintf("%s (%s/%s)", msg, yes, no), scanner)
		if input == yes {
			return true
		} else if input == no {
			return false
		}
		color.Yellow("Есть только %s и %s", yes, no)
	}
}

// RequestInt Запросить у пользователя параметр в виде целого числа
func RequestInt(msg string, scanner *bufio.Scanner) int {
	for {
		input := RequestString(msg, scanner)
		if n, err := strconv.Atoi(input); err != nil {
			color.Yellow("Ошибка конвертации в целое число: %v", err)
		} else {
			return n
		}
	}
}

const layout = "02-01-06"

// RequestDate Запросить у пользователя дату
func RequestDate(msg string, scanner *bufio.Scanner) time.Time {
	for {
		input := RequestString(msg, scanner)
		if t, err := time.Parse(layout, input); err != nil {
			color.Yellow("Ошибка конвертации во время: %v", err)
		} else {
			return t
		}
	}
}
