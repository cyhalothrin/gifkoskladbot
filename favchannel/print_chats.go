package favchannel

import (
	"fmt"
	"log"
	"math"

	"github.com/Arman92/go-tdlib"

	"github.com/cyhalothrin/gifkoskladbot/config"
	"github.com/cyhalothrin/gifkoskladbot/favchannel/tdlibclient"
)

func PrintChatsList(conf config.TDLibClient) error {
	var offset tdlib.JSONInt64
	offset = math.MaxInt64
	client, err := tdlibclient.NewClient(conf)
	if err != nil {
		return err
	}

	var lastChatID int64
	for i := 0; i < 10; i++ {
		fmt.Printf(
			`========================
=         %d           =
========================\n`, lastChatID)
		chats, err := client.GetChats(offset, lastChatID, 100)
		if err != nil {
			return err
		}
		if len(chats.ChatIDs) == 0 {
			return nil
		}

		fmt.Printf("chats returned: %d\n", len(chats.ChatIDs))

		for _, chatID := range chats.ChatIDs {
			chat, err := client.GetChat(chatID)
			if err != nil {
				log.Printf("get chant #%d: %s\n", chatID, err)
				continue
			}

			fmt.Printf("%s #%d\n", chat.Title, chat.ID)
		}

		lastChatID = chats.ChatIDs[len(chats.ChatIDs)-1]
	}

	return nil
}
