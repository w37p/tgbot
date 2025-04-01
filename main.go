package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/telebot.v3"
)

func main() {
	// Загружаем токен из .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("Токен не найден, проверь .env файл")
	}

	// Настройка бота
	bot, err := telebot.NewBot(telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Обработчик команды /start
	bot.Handle("/start", func(c telebot.Context) error {
		return c.Send("Привет! Я ваш Telegram-бот на Go!")
	})

	// Обработчик любого текста
	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		return c.Send("Привет!")
	})

	// Запуск бота
	bot.Start()
}
