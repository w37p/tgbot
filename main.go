package main

import (
	"fmt"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	// Получаем токен из системных переменных окружения
	token := os.Getenv("TELEGRAM_BOT_API_KEY")
	if token == "" {
		log.Fatal("Токен бота не найден в переменных окружения")
	}

	// Создаем объект бота
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	fmt.Println("Бот запущен:", bot.Self.UserName)

	// Укажите chatID, куда бот должен отправлять сообщения
	chatID := int64(123456789) // Замените на реальный ID

	// Запускаем бесконечный цикл отправки сообщений
	for {
		msg := tgbotapi.NewMessage(chatID, "Привет! Я активен и отправляю сообщение каждые 10 секунд.")
		bot.Send(msg)

		time.Sleep(10 * time.Second)
	}
}
