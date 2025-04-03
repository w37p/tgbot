// Обрабатываем текстовые команды
func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID

	if waitingForAdd[chatID] {
		movie := strings.TrimSpace(message.Text)
		if movie == "" {
			msg := tgbotapi.NewMessage(chatID, "Название фильма не может быть пустым. Попробуйте еще раз.")
			bot.Send(msg)
			return
		}

		movieList = append(movieList, movie)
		waitingForAdd[chatID] = false 

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Фильм '%s' добавлен!", movie))
		bot.Send(msg)

		// Показываем меню снова
		sendMainMenu(bot, chatID)
		return
	}

	switch message.Text {
	case "/start":
		sendMainMenu(bot, chatID)
	default:
		msg := tgbotapi.NewMessage(chatID, "Неизвестная команда. Используйте меню ниже.")
		bot.Send(msg)

		// Показываем меню снова
		sendMainMenu(bot, chatID)
	}
}

// Обрабатываем нажатие кнопок
func handleCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID

	switch callback.Data {
	case "add":
		waitingForAdd[chatID] = true
		msg := tgbotapi.NewMessage(chatID, "Введите название фильма для добавления:")
		bot.Send(msg)

	case "remove":
		if len(movieList) == 0 {
			msg := tgbotapi.NewMessage(chatID, "Список фильмов пуст.")
			bot.Send(msg)
		} else {
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
		}

	case "watched":
		if len(movieList) == 0 {
			msg := tgbotapi.NewMessage(chatID, "Список фильмов пуст.")
			bot.Send(msg)
		} else {
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
		}

	case "list":
		if len(movieList) == 0 {
			msg := tgbotapi.NewMessage(chatID, "Список фильмов пуст.")
			bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(chatID, "Ваши фильмы:\n" + strings.Join(movieList, "\n"))
			bot.Send(msg)
		}

	case "refresh":
		sendMainMenu(bot, chatID) // Перерисовываем меню

	default:
		if strings.HasPrefix(callback.Data, "del_") {
			movie := strings.TrimPrefix(callback.Data, "del_")
			removeMovie(movie)
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Фильм '%s' удален!", movie))
			bot.Send(msg)
		} else if strings.HasPrefix(callback.Data, "watch_") {
			movie := strings.TrimPrefix(callback.Data, "watch_")
			markMovieWatched(movie)
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Фильм '%s' отмечен как просмотренный ✅", movie))
			bot.Send(msg)
		}
	}

	// Показываем меню снова
	sendMainMenu(bot, chatID)
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
