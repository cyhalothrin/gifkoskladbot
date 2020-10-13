package tdlibclient

import (
	"errors"
	"fmt"

	"github.com/Arman92/go-tdlib"

	"github.com/cyhalothrin/gifkoskladbot/config"
)

//type tgClient interface {
//	TgMessageRemover
//	ChatHistorier
//	//forwardMessagesSilently(messageIDs []int64, fromChatID, toChatID int64) error
//}

type TgMessageRemover interface {
	RemoveMessages(chatID int64, messageIDs []int64) error
}

type FavChannelFinder interface {
	GetFavChannelID() (int64, error)
}

type ClientCloser interface {
	Close() error
}

type TdLibClient struct {
	*tdlib.Client
}

// NewClient create new instance of TdLibClient
func NewClient(conf config.TDLibClient) (*TdLibClient, error) {
	tdlib.SetLogVerbosityLevel(conf.TDLogVerbosity)
	if conf.TDLogsFile != "" {
		tdlib.SetFilePath(conf.TDLogsFile)
	}

	tdClient := tdlib.NewClient(tdlib.Config{
		APIID:               conf.APIID,
		APIHash:             conf.APIHash,
		SystemLanguageCode:  "en",
		DeviceModel:         "GifExtractor",
		SystemVersion:       "1.0.0",
		ApplicationVersion:  "1.0.0",
		UseMessageDatabase:  true,
		UseFileDatabase:     true,
		UseChatInfoDatabase: true,
		UseTestDataCenter:   false,
		DatabaseDirectory:   conf.DatabaseDirectory,
		FileDirectory:       conf.FileDirectory,
		IgnoreFileNames:     false,
	})

	client := &TdLibClient{
		Client: tdClient,
	}

	if err := authorize(client, conf); err != nil {
		return nil, fmt.Errorf("auhtorization failed: %w", err)
	}

	return client, nil
}

func (t *TdLibClient) getCurrentUserID() (int32, error) {
	user, err := t.GetMe()
	if err != nil {
		return 0, fmt.Errorf("getting current user: %w", err)
	}

	return user.ID, nil
}

func (t *TdLibClient) GetFavChannelID() (int64, error) {
	id, err := t.getCurrentUserID()
	if err != nil {
		return 0, fmt.Errorf("getting fav channel id: %w", err)
	}

	return int64(id), nil
}

func (t *TdLibClient) Close() error {
	_, err := t.Destroy()

	return err
}

func (t *TdLibClient) GetChatHistoryRemote(chatID int64, fromMessageID int64, offset int32, limit int32) (*tdlib.Messages, error) {
	return t.Client.GetChatHistory(chatID, fromMessageID, offset, limit, false)
}

func (t *TdLibClient) forwardMessagesSilently(messageIDs []int64, fromChatID, toChatID int64) error {
	messages, err := t.Client.ForwardMessages(toChatID, fromChatID, messageIDs, true, true, false)
	if err != nil {
		return err
	}

	if messages == nil {
		return errors.New("can't forward messages")
	}

	return nil
}

func (t *TdLibClient) RemoveMessages(chatID int64, messageIDs []int64) error {
	_, err := t.Client.DeleteMessages(chatID, messageIDs, true)
	if err != nil {
		return errors.New("can't delete messages")
	}

	return nil
}

func (t *TdLibClient) SendAnimation(chatID int64, fileID string, caption string) (int64, error) {
	msg, err := t.Client.SendMessage(
		chatID,
		0,
		true,
		true,
		nil,
		&tdlib.InputMessageAnimation{
			Animation: tdlib.NewInputFileRemote(fileID),
			Caption:   tdlib.NewFormattedText(caption, nil),
		},
	)

	if err != nil {
		return 0, fmt.Errorf("message not sent: %w", err)
	}

	return msg.ID, nil
}

func (t *TdLibClient) EditMessageCaption(chatID int64, messageID int64, caption string) error {
	_, err := t.Client.EditMessageCaption(chatID, messageID, nil, tdlib.NewFormattedText(caption, nil))
	if err != nil {
		return fmt.Errorf("editing message caption: %w", err)
	}

	return nil
}

func (t *TdLibClient) SendTextMessage(chatID int64, text string) (int64, error) {
	msg, err := t.Client.SendMessage(
		chatID,
		0,
		true,
		true,
		nil,
		&tdlib.InputMessageText{
			Text: tdlib.NewFormattedText(text, nil),
		},
	)
	if err != nil {
		return 0, fmt.Errorf("sending text message: %w", err)
	}

	return msg.ID, nil
}

func (t *TdLibClient) GetPinnedMessageID(chatID int64) (int64, error) {
	msg, err := t.Client.GetChatPinnedMessage(chatID)
	if err != nil {
		return 0, fmt.Errorf("getting pinned message: %w", err)
	}

	return msg.ID, nil
}

func (t *TdLibClient) PinMessage(chatID int64, messageID int64) error {
	_, err := t.Client.PinSupergroupMessage(int32(chatID), messageID, true)

	if err != nil {
		return fmt.Errorf("message pin: %w", err)
	}

	return nil
}
