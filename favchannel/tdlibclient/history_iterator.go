package tdlibclient

import (
	"fmt"

	"github.com/Arman92/go-tdlib"
)

// HistoryIterator iterates over chat history
type HistoryIterator struct {
	chatID    int64
	lastMsgID int64
	client    ChatHistorier
}

// NewHistoryIterator creates HistoryIterator
func NewHistoryIterator(client ChatHistorier, chatID int64, options ...HistoryIteratorOption) *HistoryIterator {
	iter := &HistoryIterator{
		chatID: chatID,
		client: client,
	}

	for _, option := range options {
		option(iter)
	}

	return iter
}

func (h *HistoryIterator) Next() (*tdlib.Messages, error) {
	msgs, err := h.client.GetChatHistoryRemote(h.chatID, h.lastMsgID, 0, 100)
	if err != nil {
		return nil, fmt.Errorf("getting chat history: %w", err)
	}

	if len(msgs.Messages) > 0 {
		h.lastMsgID = msgs.Messages[len(msgs.Messages)-1].ID
	}

	return msgs, nil
}

type ChatHistorier interface {
	GetChatHistoryRemote(chatID int64, fromMessageID int64, offset int32, limit int32) (*tdlib.Messages, error)
}

type HistoryIteratorOption func(iter *HistoryIterator)

func HistoryIteratorWithLastMessageID(id int64) HistoryIteratorOption {
	return func(iter *HistoryIterator) {
		iter.lastMsgID = id
	}
}
