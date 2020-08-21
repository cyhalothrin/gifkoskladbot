package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type telegramBotAPI interface {
	SendAnimation(chatID int64, fileID string, caption string) (int, error)
	GetUpdates() ([]tgbotapi.Update, error)
	SendMessage(chatID int64, text string) (int, error)
	PinMessage(chatID int64, messageID int) error
	EditMessage(chatID int64, messageID int, text string) error
	GetChatPinnedMessageID(chatID int64) (int, error)
}
