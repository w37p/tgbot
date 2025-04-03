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
	movieList     []string
	waitingForAdd map[int64]bool = make(map[int64]bool)
	userChats     map[int64]bool = make(map[int64]bool) // Храним список пользователей
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		log.Fatal("Токен не найден в .env файле")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	fmt.Println("Бот запущен:", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			handleMessage(bot, update.Message)
		} else if update.CallbackQuery != nil {
			handleCallback(bot, update.CallbackQuery)
		}
	}
}

// Обрабатываем текстовые команды
func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// Добавляем пользователя в список (чтобы он получал уведомления)
	userChats[chatID] = true

	if waitingForAdd[chatID] {
		movie := strings.TrimSpace(message.Text)
		if movie == "" {
			bot.Send(tgbotapi.NewMessage(chatID, "Название фильма не может быть пустым. Попробуйте еще раз."))
			return
		}

		movieList = append(movieList, movie)
		waitingForAdd[chatID] = false 

		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Фильм '%s' добавлен!", movie)))

		// Рассылаем уведомление всем пользователям
		notifyAllUsers(bot, fmt.Sprintf("🎬 Новый фильм добавлен: *%s*", movie))

		sendMainMenu(bot, chatID)
		return
	}

	switch message.Text {
	case "/start":
		sendMainMenu(bot, chatID)
	default:
		bot.Send(tgbotapi.NewMessage(chatID, "Неизвестная команда. Используйте меню ниже."))
		sendMainMenu(bot, chatID)
	}
}

// Отправляем главное меню
func sendMainMenu(bot *tgbotapi.BotAPI, chatID int64) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Добавить фильм", "add"),
			tgbotapi.NewInlineKeyboardButtonData("❌ Удалить фильм", "remove"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Отметить просмотренные", "watched"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Показать список", "list"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Обновить меню", "refresh"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "🎬 *Ваш список фильмов*\nВыберите действие:")
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

// Обрабатываем кнопки
func handleCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID

	switch callback.Data {
	case "add":
		waitingForAdd[chatID] = true
		bot.Send(tgbotapi.NewMessage(chatID, "Введите название фильма для добавления:"))

	case "remove":
		if len(movieList) == 0 {
			bot.Send(tgbotapi.NewMessage(chatID, "Список фильмов пуст."))
			return
		}

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

	case "watched":
		if len(movieList) == 0 {
			bot.Send(tgbotapi.NewMessage(chatID, "Список фильмов пуст."))
			return
		}

		var rows [][]tgbotapi.InlineKeyboardButton
		for _, movie := range movieList {
			if !strings.Contains(movie, "✅") {
				rows = append(rows, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(movie, "watch_"+movie),
				))
			}
		}
		keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

		msg := tgbotapi.NewMessage(chatID, "Выберите фильм, который вы посмотрели:")
		msg.ReplyMarkup = keyboard
		bot.Send(msg)

	case "list":
		if len(movieList) == 0 {
			bot.Send(tgbotapi.NewMessage(chatID, "Список фильмов пуст."))
		} else {
			bot.Send(tgbotapi.NewMessage(chatID, "Ваши фильмы:\n"+strings.Join(movieList, "\n")))
		}

	case "refresh":
		sendMainMenu(bot, chatID)

	default:
		if strings.HasPrefix(callback.Data, "del_") {
			movie := strings.TrimPrefix(callback.Data, "del_")
			removeMovie(movie)
			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Фильм '%s' удален!", movie)))
		} else if strings.HasPrefix(callback.Data, "watch_") {
			movie := strings.TrimPrefix(callback.Data, "watch_")
			markMovieWatched(movie)
			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Фильм '%s' отмечен как просмотренный ✅", movie)))
		}
	}

	sendMainMenu(bot, chatID)
}

// Функция рассылки уведомлений всем пользователям
func notifyAllUsers(bot *tgbotapi.BotAPI, message string) {
	for chatID := range userChats {
		msg := tgbotapi.NewMessage(chatID, message)
		msg.ParseMode = "Markdown"
		bot.Send(msg)
	}
}

// Удаляем фильм
func removeMovie(movie string) {
	for i, m := range movieList {
		if m == movie {
			movieList = append(movieList[:i], movieList[i+1:]...)
			break
		}
	}
}

// Отмечаем фильм просмотренным
func markMovieWatched(movie string) {
	for i, m := range movieList {
		if m == movie {
			movieList[i] = movie + " ✅"
			break
		}
	}
}
