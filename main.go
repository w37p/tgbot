package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var cart = make(map[int64]string) // userID -> товар

func main() {
	botToken := os.Getenv("BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	go startBot(bot)

	r := gin.Default()
	r.Static("/static", "./static")
	r.POST("/api/cart", func(c *gin.Context) {
		var req struct {
			UserID int64  `json:"user_id"`
			Item   string `json:"item"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid"})
			return
		}
		cart[req.UserID] = req.Item
		c.JSON(http.StatusOK, gin.H{"status": "added"})
	})
	r.Run(":8080")
}

func startBot(bot *tgbotapi.BotAPI) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			url := "https://8124458d.pyansee.pages.dev"
			webApp := tgbotapi.WebAppInfo{URL: url}
			btn := tgbotapi.NewKeyboardButtonWebApp("🛒 Открыть корзину", &webApp)
			keyboard := tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{btn})

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Открой мини-приложение для добавления товара:")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
		}
	}
}

