package tracker

import (
	"errors"
	"fmt"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// initialize and validate bot
func InitBotInstance() (*tgbotapi.BotAPI, error) {
	BotToken := os.Getenv("TOKEN")
	if len(BotToken) == 0 {
		return nil, errors.New("getTBot: could not find bot token")
	}
	bot, err := tgbotapi.NewBotAPI(BotToken)
	//bot.Debug = true
	if err != nil {
		return nil, fmt.Errorf("getTBot: error initializing bot: %v", err)
	}
	return bot, err
}

// sends message to registered id
func SendMessage(Info string, userid int64) error {
	link := "https://selfregistration.cowin.gov.in"
	msg := tgbotapi.NewMessage(userid, Info)
	btn := tgbotapi.NewInlineKeyboardButtonURL("Book on cowin.gov", link)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{btn})
	//msg1 := tgbotapi.NewMessage(GROUPID, Info)
	msg.ParseMode = "markdown"
	_, err := Bot.Send(msg)
	//_, err = bot.Send(msg1)
	if err != nil {
		return fmt.Errorf("sendmessage: message sending failed: %v", err)
	}
	time.Sleep(30 * time.Second)
	return nil
}
