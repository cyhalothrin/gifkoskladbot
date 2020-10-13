package publish

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Arman92/go-tdlib"

	"github.com/cyhalothrin/gifkoskladbot/bot"
	"github.com/cyhalothrin/gifkoskladbot/config"
	"github.com/cyhalothrin/gifkoskladbot/favchannel/tdlibclient"
	fileStorage "github.com/cyhalothrin/gifkoskladbot/storage"
)

// GifTagsPublisher publish tagged gifs from fav channel and adds Tags to storage
type GifTagsPublisher struct {
	client publisherClient
	conf   config.Config
}

// NewGifTagsPublisher creates GifTagsPublisher
func NewGifTagsPublisher(conf config.Config, client publisherClient) (*GifTagsPublisher, error) {
	return &GifTagsPublisher{
		client: client,
		conf:   conf,
	}, nil
}

func (g *GifTagsPublisher) collect() error {
	favChatID, err := g.client.GetFavChannelID()
	if err != nil {
		return err
	}

	hIter := tdlibclient.NewHistoryIterator(g.client, favChatID)
	info := gifsInfo{
		Messages: make(map[string]animationTagInfo),
	}
	uniqueTags := make(map[string]bool)

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
			if msgAnimation.Caption.Text == "" {
				continue
			}

			fileID := msgAnimation.Animation.Animation.Remote.ID
			tags, desc := g.parseTags(msgAnimation.Caption.Text)

			if len(tags) == 0 {
				fmt.Printf("no Tags: %s\n", msgAnimation.Caption.Text)

				continue
			}

			// checking tags of same gifs
			if gifInfo, ok := info.Messages[fileID]; ok {
				tagsIsChanged := false
				for _, tag := range tags {
					isExist := false
					for _, existingTag := range gifInfo.Tags {
						if existingTag == tag {
							isExist = true
							break
						}
					}

					if !isExist {
						tagsIsChanged = true
						gifInfo.Tags = append(gifInfo.Tags, tag)
						sort.Strings(gifInfo.Tags)
					}
				}

				if tagsIsChanged {
					fmt.Printf("tags changed: %s => %s\n", info.Messages[fileID].Tags, gifInfo.Tags)
				}

				if desc != "" && desc != gifInfo.Description {
					if gifInfo.Description != "" {
						gifInfo.Description += ", " + desc
					} else {
						gifInfo.Description = desc
					}

					fmt.Printf("description changed: %s => %s\n", info.Messages[fileID].Description, gifInfo.Description)
				}

				info.Messages[fileID] = gifInfo
			} else {
				info.Messages[fileID] = animationTagInfo{
					FileID:      fileID,
					Tags:        tags,
					ID:          msg.ID,
					Description: desc,
				}
			}

			for _, tag := range tags {
				uniqueTags[tag] = true
			}
		}
	}

	for tag := range uniqueTags {
		info.Tags = append(info.Tags, tag)
	}
	sort.Strings(info.Tags)

	return g.saveInfo(info)
}

func (g *GifTagsPublisher) publishMessages(storage bot.GifkoskladMetaStorage) (err error) {
	newSentAnimations := make(map[string]*fileStorage.SentAnimation)
	info, err := g.readInfo()
	if err != nil {
		return err
	}

	defer func() {
		if len(newSentAnimations) > 0 {
			storage.AddSentAnimations(newSentAnimations)
			g.saveSentTags(storage, newSentAnimations)
		}
		if saveErr := g.saveInfo(info); saveErr != nil {
			if err == nil {
				err = saveErr
			} else {
				fmt.Println("save info:", saveErr)
			}
		}

		if r := recover(); r != nil {
			fmt.Println("panic recovered:", r) // that's enough here
		}
	}()

	msgCh := make(chan *fileStorage.SentAnimation)
	sentMsgCh := g.listenMessagesToSend(msgCh)
	sentAnimations := storage.GetSentAnimations()

	go func() {
		defer close(msgCh)

		for fileID, gifInfo := range info.Messages {
			if _, ok := sentAnimations[fileID]; ok {
				continue
			}
			if gifInfo.IsSent {
				continue
			}

			msgCh <- &fileStorage.SentAnimation{
				FileID: fileID,
				Tags:   g.addDescriptionToTags(gifInfo.Tags, gifInfo.Description),
			}
		}
	}()

	for msg := range sentMsgCh {
		newSentAnimations[msg.FileID] = msg

		gifInfo := info.Messages[msg.FileID]
		gifInfo.IsSent = true
		gifInfo.ChannelMessageID = int64(msg.MessageID)
		info.Messages[msg.FileID] = gifInfo
	}

	fmt.Println("sent animations:", len(newSentAnimations))

	return nil
}

func (g *GifTagsPublisher) listen(sentMsgCh <-chan *fileStorage.SentAnimation) {
	pendingMessagesIDs := make(map[int64]*fileStorage.SentAnimation)
	sendFailChan := g.client.AddUpdatesListener(
		tdlib.NewUpdateMessageSendFailed(nil, 0, 0, ""),
	)
	sendSucceededChan := g.client.AddUpdatesListener(
		tdlib.NewUpdateMessageSendSucceeded(nil, 0),
	)

	for {
		select {
		case upd := <-sendFailChan:

		case upd := <-sendSucceededChan:
			succeededUpd, ok := upd.(*tdlib.UpdateMessageSendSucceeded)
			if !ok {
				continue
			}

			msg := pendingMessagesIDs[succeededUpd.OldMessageID]
			if msg != nil {
				msg.MessageID = int(succeededUpd.Message.ID)
			}
		case msg := <-sentMsgCh:
			pendingMessagesIDs[int64(msg.MessageID)] = msg
		case <-time.After(30 * time.Second):
			return
		}
	}
}

func (g *GifTagsPublisher) addDescriptionToTags(tags []string, desc string) []string {
	if desc == "" {
		return tags
	}

	newTags := make([]string, len(tags))
	copy(newTags, tags)
	newTags = append(newTags, desc)

	return newTags
}

func (g *GifTagsPublisher) listenMessagesToSend(msgCh <-chan *fileStorage.SentAnimation) <-chan *fileStorage.SentAnimation {
	sentMsgCh := make(chan *fileStorage.SentAnimation)
	var wg sync.WaitGroup

	// workers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for msg := range msgCh {
				err := g.postToChannel(msg)
				if err == nil {
					sentMsgCh <- msg

					continue
				}

				fmt.Printf("failed post gif %v: %s\n", msg.Tags, err)
			}
		}()
	}
	go func() {
		wg.Wait()
		close(sentMsgCh)
	}()

	return sentMsgCh
}

func (g *GifTagsPublisher) postToChannel(msg *fileStorage.SentAnimation) error {
	id, err := g.client.SendAnimation(g.conf.ChannelID, msg.FileID, strings.Join(msg.Tags, " "))
	if err != nil {
		return err
	}

	msg.MessageID = int(id)

	return nil
}

func (g *GifTagsPublisher) saveSentTags(
	storage bot.GifkoskladMetaStorage,
	sentAnimations map[string]*fileStorage.SentAnimation,
) {
	uniqueTags := make(map[string]bool)
	for _, tag := range storage.GetTags() {
		uniqueTags[tag] = true
	}

	for _, msg := range sentAnimations {
		for _, tag := range msg.Tags {
			if strings.Contains(tag, "#") {
				uniqueTags[tag] = true
			}
		}
	}

	tags := make([]string, 0, len(uniqueTags))
	for tag := range uniqueTags {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	storage.SetTags(tags)

	if err := g.updateTagsList(tags); err != nil {
		fmt.Println("update tags message failed:", err)
	}
}

func (g *GifTagsPublisher) updateTagsList(tags []string) error {
	if (len(tags)) == 0 {
		return nil
	}

	text := strings.Join(tags, "\n")
	msgID, err := g.client.GetPinnedMessageID(g.conf.ChannelID)
	if err != nil {
		return fmt.Errorf("getting chat pinned message id: %w", err)
	}

	if msgID != 0 {
		err := g.client.EditMessageCaption(g.conf.ChannelID, msgID, text)
		if err != nil {
			return fmt.Errorf("editing tags list message: %w", err)
		}
	} else {
		// if here no pinned message, create it
		newID, err := g.client.SendTextMessage(g.conf.ChannelID, text)
		if err != nil {
			return fmt.Errorf("sending tags list: %w", err)
		}
		if err := g.client.PinMessage(g.conf.ChannelID, newID); err != nil {
			return fmt.Errorf("pin message #%d: %w", newID, err)
		}
	}

	fmt.Printf("Tags list updated:\n%s\n", text)

	return nil
}

func (g *GifTagsPublisher) parseTags(caption string) ([]string, string) {
	chunks := strings.Split(strings.ToLower(caption), " ")
	if len(chunks) == 0 {
		return nil, ""
	}

	var tags []string
	var description string

	for _, chunk := range chunks {
		if chunk == "" || chunk == "#gif" {
			continue
		}

		if strings.HasPrefix(chunk, "#") {
			tags = append(tags, chunk)
			continue
		}

		description += " " + chunk
	}

	if len(tags) == 0 {
		return nil, ""
	}
	if description != "" {
		description = strings.TrimPrefix(description, " ")
	}

	return tags, description
}

func (g *GifTagsPublisher) saveInfo(list gifsInfo) error {
	data, err := json.Marshal(list)
	if err != nil {
		return fmt.Errorf("marshal gifs list: %w", err)
	}

	path := g.conf.FavChannelMigration.GifsWithTagsListPath
	if err := ioutil.WriteFile(path, data, 0666); err != nil {
		return fmt.Errorf("write gifs list to file '%s': %w", path, err)
	}

	fmt.Println("saved info:", path)

	return nil
}

func (g *GifTagsPublisher) readInfo() (info gifsInfo, err error) {
	content, err := ioutil.ReadFile(g.conf.FavChannelMigration.GifsWithTagsListPath)
	if err != nil {
		return info, fmt.Errorf("read file: %w", err)
	}

	if err := json.Unmarshal(content, &info); err != nil {
		return info, fmt.Errorf("parse json: %w", err)
	}

	return info, nil
}

type publisherClient interface {
	tdlibclient.ChatHistorier
	tdlibclient.FavChannelFinder
	SendAnimation(chatID int64, fileID string, caption string) (int64, error)
	EditMessageCaption(chatID int64, messageID int64, caption string) error
	SendTextMessage(chatID int64, text string) (int64, error)
	GetPinnedMessageID(chatID int64) (int64, error)
	PinMessage(chatID int64, messageID int64) error
	AddUpdatesListener(updateType tdlib.TdMessage) chan tdlib.TdMessage
}

type animationTagInfo struct {
	FileID           string
	Tags             []string
	Description      string `json:",omitempty"`
	ID               int64
	IsSent           bool  `json:",omitempty"`
	IsDeleted        bool  `json:",omitempty"`
	ChannelMessageID int64 `json:",omitempty"`
}

type gifsInfo struct {
	Messages map[string]animationTagInfo
	Tags     []string
}
