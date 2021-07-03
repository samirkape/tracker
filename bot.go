package tracker

import (
	"errors"
	"fmt"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// initialize and validate bot
func getTBot() (*tgbotapi.BotAPI, error) {
	BotToken := os.Getenv("TOKEN")
	if len(BotToken) == 0 {
		return nil, errors.New("getTBot: could not find bot token")
	}
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		return nil, fmt.Errorf("getTBot: error initializing bot: %v", err)
	}
	return bot, err
}

// sends message to registered id
func SendMessage(Info string, userid int64) error {
	link := "https://selfregistration.cowin.gov.in"
	msg := tgbotapi.NewMessage(userid, Info)
	btn := tgbotapi.NewInlineKeyboardButtonURL("Book Vaccine", link)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{btn})
	msg.ParseMode = "markdown"
	_, err := Bot.Send(msg)
	if err != nil {
		return fmt.Errorf("sendmessage: message sending failed: %v", err)
	}
	return nil
}
