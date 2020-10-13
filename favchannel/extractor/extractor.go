package extractor

import (
	"fmt"
	"log"

	"github.com/Arman92/go-tdlib"

	"github.com/cyhalothrin/gifkoskladbot/config"
	"github.com/cyhalothrin/gifkoskladbot/favchannel/tdlibclient"
)

type GifExtractor struct {
	client  extractorClient
	storage storage
	tags    map[string]bool
	conf    config.Config
}

// NewGifExtractor creates new instance of GifExtractor
func NewGifExtractor(conf config.Config, storage storage, client extractorClient) (*GifExtractor, error) {
	return &GifExtractor{
		client:  client,
		tags:    make(map[string]bool),
		conf:    conf,
		storage: storage,
	}, nil
}

//func (g *GifExtractor) addTags(caption string) {
//	if strings.Index(caption, "#") == -1 {
//		fmt.Println("caption without #:", caption)
//		return
//	}
//
//	for _, tag := range strings.Split(caption, " ") {
//		if strings.Index(tag, "#") == 0 && caption != "#" {
//			g.tags[tag] = true
//		}
//	}
//}

// moveMessagesWithoutCaptionToBotChannel forwards gifs without caption to bot chat
func (g *GifExtractor) moveMessagesWithoutCaptionToBotChannel() error {
	var lastSuccessfullySentMessageID int64

	defer func() {
		if lastSuccessfullySentMessageID > 0 {
			fmt.Println("Set last sent message id:", lastSuccessfullySentMessageID)
			g.storage.SetFavChannelLastForwardedMessageIDWithoutCaption(lastSuccessfullySentMessageID)
		}

		if r := recover(); r != nil {
			fmt.Println("panic recovered:", r)
		}
	}()

	favChatID, err := g.client.GetFavChannelID()
	if err != nil {
		return err
	}
	fmt.Println("Chat ID: ", favChatID)

	lastMsgID := g.storage.GetFavChannelLastForwardedMessageIDWithoutCaption()
	lastSuccessfullySentMessageID = lastMsgID
	hIter := tdlibclient.NewHistoryIterator(g.client, favChatID, tdlibclient.HistoryIteratorWithLastMessageID(lastMsgID))

	forwardedCount := 0

	messagesIDs := make([]int64, 0, 10) // send 10 messages

	for {
		msgs, err := hIter.Next()
		if err != nil {
			log.Println("iteration: ", err)

			continue
		}

		if len(msgs.Messages) == 0 {
			break
		}

		for _, msg := range msgs.Messages {
			if msg.Content.GetMessageContentEnum() != tdlib.MessageAnimationType {
				continue
			}

			msgAnimation := msg.Content.(*tdlib.MessageAnimation)
			if msgAnimation.Caption.Text != "" {
				continue
			}

			messagesIDs = append(messagesIDs, msg.ID)
			forwardedCount++

			if len(messagesIDs) < 10 {
				continue
			}

			if err := g.forwardToBotChannelAndRemove(messagesIDs, favChatID); err != nil {
				return err
			}

			lastSuccessfullySentMessageID = msg.ID
			messagesIDs = messagesIDs[:0]
		}
	}

	if len(messagesIDs) > 0 {
		if err := g.forwardToBotChannelAndRemove(messagesIDs, favChatID); err != nil {
			return err
		}

		lastSuccessfullySentMessageID = messagesIDs[len(messagesIDs)-1]
	}

	fmt.Printf("Successfully finished, %d forwarded\n", forwardedCount)

	return nil
}

// TODO: need to check that the messages were successfully sent and then delete, otherwise they will disappear
func (g *GifExtractor) forwardToBotChannelAndRemove(messagesIDs []int64, fromChatID int64) error {
	fmt.Println("going to forward:", messagesIDs)
	//if err := g.client.forwardMessagesSilently(messagesIDs, fromChatID, g.conf.BotChatID); err != nil {
	//	return err
	//}
	//
	//if err := g.client.removeMessages(fromChatID, messagesIDs); err != nil {
	//	log.Println("removing messages failed:", err)
	//} else {
	//	fmt.Println("         removed:", messagesIDs)
	//}

	return nil
}

type extractorClient interface {
	tdlibclient.ChatHistorier
	tdlibclient.FavChannelFinder
}
