package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/telebot.v3"
)

func main() {
	// Загружаем .env файл
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	// Чтение токена из переменной окружения
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("Токен не найден, проверь .env файл")
	}

	// Настройка бота
	bot, err := telebot.NewBot(telebot.Settings{
		Token: token,
	})
	if err != nil {
		log.Fatal("Ошибка создания бота: ", err)
	}

	// Обработчик команды /start
	bot.Handle("/start", func(c telebot.Context) error {
		// Получаем имя пользователя
		user := c.Sender()
		message := "Привет, " + user.FirstName + " " + user.LastName + "!"
		return c.Send(message)
	})

	// Запуск бота
	log.Println("Бот запущен...")
	bot.Start()
}
