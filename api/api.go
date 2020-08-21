package api

import (
	"fmt"

	"github.com/cyhalothrin/gifkoskladbot/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type TelegramBotAPI struct {
	tg *tgbotapi.BotAPI
}

func NewTelegramBotAPI(conf config.Config) (*TelegramBotAPI, error) {
	tg, err := tgbotapi.NewBotAPI(conf.Token)
	if err != nil {
		return nil, err
	}

	return &TelegramBotAPI{
		tg: tg,
	}, nil
}

func (t *TelegramBotAPI) SendAnimation(chatID int64, fileID string, caption string) (int, error) {
	animationMsgConf := tgbotapi.AnimationConfig{
		BaseFile: tgbotapi.BaseFile{
			BaseChat: tgbotapi.BaseChat{
				ChatID: chatID,
			},
			FileID:      fileID,
			UseExisting: true,
		},
		Caption: caption,
	}

	msg, err := t.tg.Send(animationMsgConf)
	if err != nil {
		return 0, fmt.Errorf("send animation: %w", err)
	}

	return msg.MessageID, nil
}

func (t *TelegramBotAPI) EditMessage(chatID int64, messageID int, text string) error {
	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)

	if _, err := t.tg.Send(msg); err != nil {
		return fmt.Errorf("send edited message: %w", err)
	}

	return nil
}

func (t *TelegramBotAPI) GetUpdates() ([]tgbotapi.Update, error) {
	updConf := tgbotapi.NewUpdate(0)
	updConf.Timeout = 60

	return t.tg.GetUpdates(updConf)
}

func (t *TelegramBotAPI) SendMessage(chatID int64, text string) (int, error) {
	msg := tgbotapi.NewMessage(chatID, text)

	sentMsg, err := t.tg.Send(msg)
	if err != nil {
		return 0, err
	}

	return sentMsg.MessageID, nil
}

func (t *TelegramBotAPI) PinMessage(chatID int64, messageID int) error {
	_, err := t.tg.PinChatMessage(tgbotapi.PinChatMessageConfig{
		ChatID:              chatID,
		MessageID:           messageID,
		DisableNotification: true,
	})

	return err
}

func (t *TelegramBotAPI) GetChatPinnedMessageID(chatID int64) (int, error) {
	chat, err := t.tg.GetChat(tgbotapi.ChatConfig{
		ChatID: chatID,
	})
	if err != nil {
		return 0, fmt.Errorf("получение инфы о чате: %w", err)
	}

	if chat.PinnedMessage == nil {
		return 0, nil
	}

	return chat.PinnedMessage.MessageID, nil
}
