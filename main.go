package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	movieList     []string         // Хранилище фильмов
	waitingForAdd map[int64]bool = make(map[int64]bool) // Флаг ожидания ввода фильма
)

func main() {
	// Загружаем .env файл
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	// Получаем токен из .env
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		log.Fatal("Токен не найден в .env файле")
	}

	// Создаем бота
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	fmt.Println("Бот запущен:", bot.Self.UserName)

	// Получаем обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // Обрабатываем текстовые сообщения
			handleMessage(bot, update.Message)
		} else if update.CallbackQuery != nil { // Обрабатываем кнопки
			handleCallback(bot, update.CallbackQuery)
		}
	}
}

// Обрабатываем текстовые команды
func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// Если ожидаем ввода фильма, добавляем его в список
	if waitingForAdd[chatID] {
		movie := strings.TrimSpace(message.Text)
		if movie == "" {
			msg := tgbotapi.NewMessage(chatID, "Название фильма не может быть пустым. Попробуйте еще раз.")
			bot.Send(msg)
			return
		}

		movieList = append(movieList, movie)
		waitingForAdd[chatID] = false // Сбрасываем режим ожидания

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Фильм '%s' добавлен!", movie))
		bot.Send(msg)
		return
	}

	switch message.Text {
	case "/start":
		sendMainMenu(bot, chatID)
	default:
		msg := tgbotapi.NewMessage(chatID, "Неизвестная команда. Используйте меню ниже.")
		bot.Send(msg)
	}
}

// Отправляем главное меню с кнопками
func sendMainMenu(bot *tgbotapi.BotAPI, chatID int64) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Добавить фильм", "add"),
			tgbotapi.NewInlineKeyboardButtonData("❌ Удалить фильм", "remove"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Показать список", "list"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "Выберите действие:")
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

// Обрабатываем кнопки
func handleCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID

	switch callback.Data {
	case "add":
		waitingForAdd[chatID] = true // Включаем режим ожидания ввода
		msg := tgbotapi.NewMessage(chatID, "Введите название фильма для добавления:")
		bot.Send(msg)

	case "remove":
		if len(movieList) == 0 {
			msg := tgbotapi.NewMessage(chatID, "Список фильмов пуст.")
			bot.Send(msg)
			return
		}

		// Создаем кнопки для удаления фильмов
		var rows [][]tgbotapi.InlineKeyboardButton
		for _, movie := range movieList {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(movie, "del_"+movie),
			))
		}
		keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

		msg := tgbotapi.NewMessage(chatID, "Выберите фильм для удаления:")
		msg.ReplyMarkup = keyboard
		bot.Send(msg)

	case "list":
		if len(movieList) == 0 {
			msg := tgbotapi.NewMessage(chatID, "Список фильмов пуст.")
			bot.Send(msg)
			return
		}

		msg := tgbotapi.NewMessage(chatID, "Ваши фильмы:\n" + strings.Join(movieList, "\n"))
		bot.Send(msg)

	default:
		if strings.HasPrefix(callback.Data, "del_") {
			movie := strings.TrimPrefix(callback.Data, "del_")
			removeMovie(movie)
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Фильм '%s' удален!", movie))
			bot.Send(msg)
		}
	}
}

// Функция удаления фильма
func removeMovie(movie string) {
	for i, m := range movieList {
		if m == movie {
			movieList = append(movieList[:i], movieList[i+1:]...)
			break
		}
	}
}
