package publish

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"

	"github.com/cyhalothrin/gifkoskladbot/bot"
	"github.com/cyhalothrin/gifkoskladbot/config"
	fileStorage "github.com/cyhalothrin/gifkoskladbot/storage"
)

func setup() {
	// copy data file
	source, err := os.Open("testdata/messages.init.json")
	if err != nil {
		panic(err)
	}
	defer source.Close()

	destination, err := os.Create("testdata/messages.json")
	if err != nil {
		panic(err)
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		panic(err)
	}
}

func teardown() {
	err := os.Remove("testdata/messages.json")
	if err != nil {
		if os.IsNotExist(err) {
			return
		}

		fmt.Println("Failed remove data file:", err)
	}
}

func TestGifTagsPublisher_publishMessages(t *testing.T) {
	setup()
	defer teardown()

	mc := minimock.NewController(t)
	defer mc.Finish()

	channelID := int64(1013)
	conf := config.Config{
		FavChannelMigration: config.FavChannelMigration{
			GifsWithTagsListPath: "testdata/messages.json",
		},
		ChannelID: channelID,
	}

	type fields struct {
		client publisherClient
	}
	type args struct {
		storage bot.GifkoskladMetaStorage
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantErr         bool
		wantFileContent gifsInfo
	}{
		{
			"publish messages",
			fields{
				client: NewPublisherClientMock(mc).
					SendAnimationMock.
					When(
						channelID,
						"CgACAgIAAxkBAAEDBfhfXjb61m1eQc1Wmb626tmS2BgTNwAClwAD5Im4SfoGWydN2QgMGAQ",
						"#непонятно сложно!",
					).
					Then(101, nil).
					SendAnimationMock.
					When(
						channelID,
						"CgACAgIAAx0ETm6cZwACA9BfY6IgM6ZaGFh89Erp6-G6547K6wAC-gMAAgeIOUuCh-0pbyC76BgE",
						"#aaaaaa #fuuuu #котики",
					).
					Then(102, nil).
					GetPinnedMessageIDMock.Expect(channelID).Return(100500, nil).
					EditMessageCaptionMock.
					Expect(channelID, 100500, "#aaaaaa\n#fuuuu\n#tag1\n#tag2\n#котики\n#непонятно").
					Return(nil),
			},
			args{
				storage: NewGifkoskladMetaStorageMock(t).
					GetSentAnimationsMock.
					Return(map[string]*fileStorage.SentAnimation{
						"CgACAgIAAx0ER7jZmwACB29fY6IkFe1uAcGWG1slegjj9SLIZwAC0AADQREZCsJQJe2uMLiHGAQ": nil,
					}).
					AddSentAnimationsMock.
					Expect(map[string]*fileStorage.SentAnimation{
						"CgACAgIAAxkBAAEDBfhfXjb61m1eQc1Wmb626tmS2BgTNwAClwAD5Im4SfoGWydN2QgMGAQ": {
							MessageID: 101,
							FileID:    "CgACAgIAAxkBAAEDBfhfXjb61m1eQc1Wmb626tmS2BgTNwAClwAD5Im4SfoGWydN2QgMGAQ",
							Tags:      []string{"#непонятно", "сложно!"},
						},
						"CgACAgIAAx0ETm6cZwACA9BfY6IgM6ZaGFh89Erp6-G6547K6wAC-gMAAgeIOUuCh-0pbyC76BgE": {
							MessageID: 102,
							FileID:    "CgACAgIAAx0ETm6cZwACA9BfY6IgM6ZaGFh89Erp6-G6547K6wAC-gMAAgeIOUuCh-0pbyC76BgE",
							Tags:      []string{"#aaaaaa", "#fuuuu", "#котики"},
						},
					}).
					Return().
					GetTagsMock.
					Return([]string{"#aaaaaa", "#tag1", "#tag2"}).
					SetTagsMock.
					Expect([]string{"#aaaaaa", "#fuuuu", "#tag1", "#tag2", "#котики", "#непонятно"}).
					Return(),
			},
			false,
			gifsInfo{
				Messages: map[string]animationTagInfo{
					"CgACAgIAAx0ER7jZmwACB29fY6IkFe1uAcGWG1slegjj9SLIZwAC0AADQREZCsJQJe2uMLiHGAQ": {
						FileID: "CgACAgIAAx0ER7jZmwACB29fY6IkFe1uAcGWG1slegjj9SLIZwAC0AADQREZCsJQJe2uMLiHGAQ",
						Tags:   []string{"#lol"},
						ID:     207842443264,
					},
					"CgACAgIAAxkBAAEDBfhfXjb61m1eQc1Wmb626tmS2BgTNwAClwAD5Im4SfoGWydN2QgMGAQ": {
						FileID:           "CgACAgIAAxkBAAEDBfhfXjb61m1eQc1Wmb626tmS2BgTNwAClwAD5Im4SfoGWydN2QgMGAQ",
						Tags:             []string{"#непонятно"},
						Description:      "сложно!",
						ID:               207760654336,
						IsSent:           true,
						ChannelMessageID: 101,
					},
					"CgACAgIAAx0ETm6cZwACA9BfY6IgM6ZaGFh89Erp6-G6547K6wAC-gMAAgeIOUuCh-0pbyC76BgE": {
						FileID:           "CgACAgIAAx0ETm6cZwACA9BfY6IgM6ZaGFh89Erp6-G6547K6wAC-gMAAgeIOUuCh-0pbyC76BgE",
						Tags:             []string{"#aaaaaa", "#fuuuu", "#котики"},
						ID:               207861317632,
						IsSent:           true,
						ChannelMessageID: 102,
					},
					"CgACAgIAAxkBAAEDBU9fXjb76uuhZkONrEXHA3BVxb66xwAC6AIAAg0IUEuo9FFl_K-mRxgE": {
						FileID: "CgACAgIAAxkBAAEDBU9fXjb76uuhZkONrEXHA3BVxb66xwAC6AIAAg0IUEuo9FFl_K-mRxgE",
						Tags:   []string{"#овечка"},
						ID:     207583444992,
						IsSent: true,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GifTagsPublisher{
				client: tt.fields.client,
				conf:   conf,
			}
			if err := g.publishMessages(tt.args.storage); (err != nil) != tt.wantErr {
				t.Errorf("publishMessages() error = %v, wantErr %v", err, tt.wantErr)
			}

			content, err := ioutil.ReadFile(conf.FavChannelMigration.GifsWithTagsListPath)
			if err != nil {
				t.Errorf("read file: %s", err)
			}

			fileGifInfo := &gifsInfo{}
			if err := json.Unmarshal(content, fileGifInfo); err != nil {
				t.Errorf("parse json: %s", err)
			}

			assert.Equal(t, tt.wantFileContent.Messages, fileGifInfo.Messages)
		})
	}
}

func TestGifTagsPublisher_addDescriptionToTags(t *testing.T) {
	for i := 0; i < 10; i++ {
		length := rand.Intn(9) + 1
		tags := make([]string, 0, length)
		for i := 1; i <= length; i++ {
			tags = append(tags, fmt.Sprintf("#tag%d", i))
		}

		desc := ""
		if rand.Intn(2) > 0 {
			desc = "gif description"
		}

		g := &GifTagsPublisher{}
		newTags := g.addDescriptionToTags(tags, desc)
		if desc == "" {
			assert.ElementsMatch(t, newTags, tags)
		} else {
			assert.Equal(t, len(newTags), length+1)
			assert.Equal(t, len(tags), length, "original slice should not changed")
			assert.Equal(t, newTags[len(newTags)-1], desc)
		}
	}
}

func TestGifTagsPublisher_listenMessagesToSend(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

	var errCount int64

	g := &GifTagsPublisher{
		client: NewPublisherClientMock(mc).SendAnimationMock.Set(
			func(chatID int64, fileID string, caption string) (i1 int64, err error) {
				if rand.Float64() < 0.3 {
					atomic.AddInt64(&errCount, 1)
					return 0, errors.New("sendAnimationErr")
				}

				id, err := strconv.ParseInt(fileID, 10, 64)
				return id, nil
			},
		),
	}
	msgInChan := make(chan *fileStorage.SentAnimation, 10)
	msgOutChan := g.listenMessagesToSend(msgInChan)

	msgCount := int64(10)
	for i := int64(1); i <= msgCount; i++ {
		msg := &fileStorage.SentAnimation{
			FileID: strconv.FormatInt(i, 10),
			Tags:   []string{fmt.Sprintf("#tag%d", i)},
		}

		msgInChan <- msg
	}
	close(msgInChan)

	var successCount int64
	for msg := range msgOutChan {
		assert.True(t, msg.MessageID > 0, "should be more than 0, got=%d", msg.MessageID)
		successCount++
	}

	assert.Equal(t, msgCount-errCount, successCount)
}

func TestGifTagsPublisher_postToChannel(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

	conf := config.Config{
		ChannelID: 1010,
	}

	type fields struct {
		client publisherClient
	}
	type args struct {
		msg *fileStorage.SentAnimation
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantErr       bool
		wantMessageID int
	}{
		{
			"should set message id",
			fields{
				client: NewPublisherClientMock(mc).
					SendAnimationMock.
					Expect(conf.ChannelID, "file_1", "#tag1 #tag2 description").
					Return(1001, nil),
			},
			args{
				msg: &fileStorage.SentAnimation{
					FileID: "file_1",
					Tags:   []string{"#tag1", "#tag2", "description"},
				},
			},
			false,
			1001,
		},
		{
			"should return err",
			fields{
				client: NewPublisherClientMock(mc).
					SendAnimationMock.
					Expect(conf.ChannelID, "file_1", "#tag1 #tag2 description").
					Return(0, errors.New("SendAnimationErr")),
			},
			args{
				msg: &fileStorage.SentAnimation{
					FileID: "file_1",
					Tags:   []string{"#tag1", "#tag2", "description"},
				},
			},
			true,
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GifTagsPublisher{
				client: tt.fields.client,
				conf:   conf,
			}
			if err := g.postToChannel(tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("postToChannel() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.Equal(t, tt.wantMessageID, tt.args.msg.MessageID)
		})
	}
}

func TestGifTagsPublisher_updateTagsList(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

	conf := config.Config{
		ChannelID: 1010,
	}

	tags := []string{"#tag1", "#tag2", "#tag3", "#tag4"}
	caption := strings.Join(tags, "\n")
	pinnedMessageID := int64(20304050)

	type fields struct {
		client publisherClient
	}
	type args struct {
		tags []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"should edit existing pinned message",
			fields{
				client: NewPublisherClientMock(mc).GetPinnedMessageIDMock.
					Expect(conf.ChannelID).Return(pinnedMessageID, nil).
					EditMessageCaptionMock.
					Expect(conf.ChannelID, pinnedMessageID, caption).Return(nil),
			},
			args{
				tags: tags,
			},
			false,
		},
		{
			"should creat new message and pin it",
			fields{
				client: NewPublisherClientMock(mc).
					GetPinnedMessageIDMock.
					Expect(conf.ChannelID).Return(0, nil).
					SendTextMessageMock.
					Expect(conf.ChannelID, caption).Return(1020, nil).
					PinMessageMock.Expect(conf.ChannelID, 1020).Return(nil),
			},
			args{
				tags: tags,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GifTagsPublisher{
				client: tt.fields.client,
				conf:   conf,
			}
			if err := g.updateTagsList(tt.args.tags); (err != nil) != tt.wantErr {
				t.Errorf("updateTagsList() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
