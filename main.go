package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	movieList     []string
	userChats     map[int64]bool = make(map[int64]bool)
	waitingForAdd map[int64]bool = make(map[int64]bool)
	mu            sync.Mutex
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

// Обрабатываем команды
func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// Запоминаем пользователя
	userChats[chatID] = true

	if waitingForAdd[chatID] {
		movie := strings.TrimSpace(message.Text)
		if movie == "" {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Название фильма не может быть пустым. Попробуйте снова."))
			return
		}

		mu.Lock()
		movieList = append(movieList, movie)
		mu.Unlock()

		waitingForAdd[chatID] = false

		// Отправляем сообщение только добавившему пользователю
		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ Фильм '%s' добавлен!", movie)))

		// Оповещаем всех остальных пользователей
		notifyAllUsers(bot, fmt.Sprintf("🎬 *Добавлен новый фильм:* _%s_", movie), chatID)
		return
	}

	if message.Text == "/start" {
		sendMainMenu(bot, chatID)
	}
}

// Отправляем главное меню (только при /start)
func sendMainMenu(bot *tgbotapi.BotAPI, chatID int64) {
	// Добавляем постоянную кнопку /start внизу экрана
	replyKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/start"),
		),
	)

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
	)

	msg := tgbotapi.NewMessage(chatID, "🎬 *Ваш список фильмов:*")
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = replyKeyboard // Добавляем кнопку /start
	bot.Send(msg)

	menuMsg := tgbotapi.NewMessage(chatID, "Выберите действие:")
	menuMsg.ReplyMarkup = keyboard
	bot.Send(menuMsg)
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
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Список фильмов пуст."))
			return
		}

		var rows [][]tgbotapi.InlineKeyboardButton
		mu.Lock()
		for _, movie := range movieList {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(movie, "del_"+movie),
			))
		}
		mu.Unlock()

		keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

		msg := tgbotapi.NewMessage(chatID, "🗑 Выберите фильм для удаления:")
		msg.ReplyMarkup = keyboard
		bot.Send(msg)

	case "watched":
		if len(movieList) == 0 {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Список фильмов пуст."))
			return
		}

		var rows [][]tgbotapi.InlineKeyboardButton
		mu.Lock()
		for _, movie := range movieList {
			if !strings.Contains(movie, "✅") {
				rows = append(rows, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(movie, "watch_"+movie),
				))
			}
		}
		mu.Unlock()

		keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

		msg := tgbotapi.NewMessage(chatID, "🎬 Выберите фильм, который вы посмотрели:")
		msg.ReplyMarkup = keyboard
		bot.Send(msg)

	case "list":
		if len(movieList) == 0 {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Список фильмов пуст."))
		} else {
			mu.Lock()
			movieText := "📋 *Список фильмов:*\n" + strings.Join(movieList, "\n")
			mu.Unlock()
			msg := tgbotapi.NewMessage(chatID, movieText)
			msg.ParseMode = "Markdown"
			bot.Send(msg)
		}

	default:
		if strings.HasPrefix(callback.Data, "del_") {
			movie := strings.TrimPrefix(callback.Data, "del_")
			removeMovie(movie)
			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("🗑 Фильм '%s' удален!", movie)))
		} else if strings.HasPrefix(callback.Data, "watch_") {
			movie := strings.TrimPrefix(callback.Data, "watch_")
			markMovieWatched(movie)
			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ Фильм '%s' отмечен как просмотренный!", movie)))
		}
	}
}

// 🔹 Рассылка уведомлений (кроме автора)
func notifyAllUsers(bot *tgbotapi.BotAPI, message string, excludeChatID int64) {
	mu.Lock()
	defer mu.Unlock()
	for chatID := range userChats {
		if chatID != excludeChatID { // Исключаем отправителя
			msg := tgbotapi.NewMessage(chatID, message)
			msg.ParseMode = "Markdown"
			bot.Send(msg)
		}
	}
}

// Удаляем фильм
func removeMovie(movie string) {
	mu.Lock()
	defer mu.Unlock()
	for i, m := range movieList {
		if m == movie {
			movieList = append(movieList[:i], movieList[i+1:]...)
			break
		}
	}
}

// Отмечаем фильм просмотренным
func markMovieWatched(movie string) {
	mu.Lock()
	defer mu.Unlock()
	for i, m := range movieList {
		if m == movie {
			movieList[i] = movie + " ✅"
			break
		}
	}
}
