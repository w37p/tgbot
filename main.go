package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Загружаем переменные окружения из .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Создаем бота с токеном из .env
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Настройка обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	// Обработка входящих сообщений
	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Получаем информацию о пользователе
		user := update.Message.From
		greeting := generateGreeting(user)

		// Создаем ответное сообщение
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, greeting)
		msg.ReplyToMessageID = update.Message.MessageID

		// Отправляем сообщение
		if _, err := bot.Send(msg); err != nil {
			log.Println("Error sending message:", err)
		}
	}
}

// generateGreeting создает персонализированное приветствие
func generateGreeting(user *tgbotapi.User) string {
	var name string

	// Проверяем наличие username
	if user.UserName != "" {
		name = "@" + user.UserName
	} else {
		name = user.FirstName
		if user.LastName != "" {
			name += " " + user.LastName
		}
	}

	// Если вообще нет имени
	if name == "" {
		name = "друг"
	}

	return "Привет, " + name + "! 😊\nРад тебя видеть!"
}
