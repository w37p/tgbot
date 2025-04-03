package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var movieList []string

func main() {
	// Загружаем переменные из файла .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Получаем токен из переменной окружения
	token := os.Getenv("TELEGRAM_BOT_API_KEY")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_API_KEY is not set")
	}

	// Создаем объект бота
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	fmt.Println("Authorized on account", bot.Self.UserName)

	// Настроим обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)

	// Обрабатываем сообщения
	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Приветственное сообщение с кнопками
		if update.Message.Text == "/start" {
			keyboard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Обновить список"),
				),
			)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Добро пожаловать! Используйте кнопки ниже для взаимодействия с ботом.")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
			continue
		}

		// Команда для обновления списка (аналогичная /start)
		if update.Message.Text == "Обновить список" {
			showMovieList(update.Message.Chat.ID, bot)
			continue
		}

		// Команда для добавления фильма
		if strings.HasPrefix(update.Message.Text, "/add") {
			movie := strings.TrimPrefix(update.Message.Text, "/add ")
			if movie != "" {
				movieList = append(movieList, movie)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Фильм '%s' добавлен!", movie))
				bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, укажите название фильма после команды /add")
				bot.Send(msg)
			}
			continue
		}

		// Команда для удаления фильма
		if strings.HasPrefix(update.Message.Text, "/remove") {
			movie := strings.TrimPrefix(update.Message.Text, "/remove ")
			if movie != "" {
				for i, m := range movieList {
					if m == movie {
						movieList = append(movieList[:i], movieList[i+1:]...)
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Фильм '%s' удален!", movie))
						bot.Send(msg)
						break
					}
				}
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, укажите название фильма после команды /remove")
				bot.Send(msg)
			}
			continue
		}

		// Команда для вывода списка фильмов
		if update.Message.Text == "/list" {
			showMovieList(update.Message.Chat.ID, bot)
		}
	}
}

func showMovieList(chatID int64, bot *tgbotapi.BotAPI) {
	if len(movieList) == 0 {
		msg := tgbotapi.NewMessage(chatID, "Список фильмов пуст!")
		bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(chatID, "Список фильмов:\n"+strings.Join(movieList, "\n"))
		bot.Send(msg)
	}
}
