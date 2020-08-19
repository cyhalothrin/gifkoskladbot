package bot

import (
	"reflect"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/gojuno/minimock/v3"

	"github.com/cyhalothrin/gifkoskladbot/config"
	"github.com/cyhalothrin/gifkoskladbot/storage"
)

func TestBot_parseTags(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

	type fields struct {
		store GifkoskladMetaStorage
	}
	type args struct {
		text string
	}
	type want struct {
		tags               []string
		uniqueTags         map[string]bool
		hasTagsListChanges bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			"should parse tags",
			fields{
				store: NewGifkoskladMetaStorageMock(mc).
					GetTagsAliasesMock.Return(nil).
					GetSentAnimationsMock.Return(nil).
					GetTagsMock.Return(nil),
			},
			args{
				text: "tag1 f  1tag 11tag with  space11 00just description i 00",
			},
			want{
				[]string{"#tag1", "#f", "#1tag", "#tag_with_space", "just description i"},
				map[string]bool{"#tag1": true, "#f": true, "#1tag": true, "#tag_with_space": true},
				true,
			},
		},
		{
			"should parse tags with spaces",
			fields{
				store: NewGifkoskladMetaStorageMock(mc).
					GetTagsAliasesMock.Return(nil).
					GetSentAnimationsMock.Return(nil).
					GetTagsMock.Return(nil),
			},
			args{
				text: "11like a boss11 00description00",
			},
			want{
				[]string{"#like_a_boss", "description"},
				map[string]bool{"#like_a_boss": true},
				true,
			},
		},
		{
			"should replace tag with alias and unique tags not changed",
			fields{
				store: NewGifkoskladMetaStorageMock(mc).
					GetTagsAliasesMock.Return(map[string]string{"#lab": "#like_a_boss"}).
					GetSentAnimationsMock.Return(nil).
					GetTagsMock.Return([]string{"#like_a_boss", "#existing_tag"}),
			},
			args{
				text: "lab 11existing tag11 00description00",
			},
			want{
				[]string{"#like_a_boss", "#existing_tag", "description"},
				map[string]bool{"#like_a_boss": true, "#existing_tag": true},
				false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			b := NewUpdatesHandler(config.Config{}, tt.fields.store, NewAlerterMock(mc), NewTelegramBotAPIMock(mc))
			if got := b.parseTags(tt.args.text); !reflect.DeepEqual(got, tt.want.tags) {
				t.Errorf("parseTags() = %v, want %v", got, tt.want.tags)
			}

			if !reflect.DeepEqual(b.uniqueTags, tt.want.uniqueTags) {
				t.Errorf("uniqueTags = %v, want %v", b.uniqueTags, tt.want.uniqueTags)
			}

			if !reflect.DeepEqual(b.hasTagsListChanges, tt.want.hasTagsListChanges) {
				t.Errorf("hasTagsListChanges = %v, want %v", b.hasTagsListChanges, tt.want.hasTagsListChanges)
			}
		})
	}
}

//func TestUpdatesHandler_createTagsList(t *testing.T) {
//	mc := minimock.NewController(t)
//	defer mc.Finish()
//
//	type fields struct {
//		storage GifkoskladMetaStorage
//	}
//	type args struct {
//		tags map[string]bool
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//		want   string
//	}{
//		{
//			"",
//			fields{
//				storage: NewGifkoskladMetaStorageMock(mc).
//					GetTagsMock.
//					Return([]string{"#tagA", "#tagB"}).
//					GetTagsAliasesMock.
//					Return(nil),
//			},
//			args{
//				tags: map[string]bool{"#tagA": true, "#tagD": true, "#tagC": true},
//			},
//			"#tagA\n#tagB\n#tagC\n#tagD",
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			b := NewUpdatesHandler(config.Config{}, tt.fields.storage, nil, nil)
//			if got := b.createTagsList(tt.args.tags); got != tt.want {
//				t.Errorf("createTagsList() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

func TestUpdatesHandler_captionsIsEqual(t *testing.T) {
	type args struct {
		tagsA []string
		tagsB []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"equal tags list",
			args{
				tagsA: []string{"#tag1", "#tag2"},
				tagsB: []string{"#tag1", "#tag2"},
			},
			true,
		},
		{
			"not equal tags list",
			args{
				tagsA: []string{"#tag1", "#tag2", "#tag3"},
				tagsB: []string{"#tag1", "#tag2"},
			},
			false,
		},
		{
			"not equal tags list",
			args{
				tagsA: []string{"#tag1", "#tag2", "#tag3"},
				tagsB: []string{"#tag1", "#tag2", "#tag4"},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UpdatesHandler{}
			if got := u.captionsIsEqual(tt.args.tagsA, tt.args.tagsB); got != tt.want {
				t.Errorf("captionsIsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdatesHandler_handleAnimationCaption(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

	emptyStorage := NewGifkoskladMetaStorageMock(mc).
		GetTagsAliasesMock.Return(nil).
		GetSentAnimationsMock.Return(nil).
		GetTagsMock.Return(nil)
	conf := config.Config{
		AllowedUsers: []string{"cyhalothrin"},
	}

	type fields struct {
		storage GifkoskladMetaStorage
		config  config.Config
	}
	type args struct {
		update tgbotapi.Update
	}
	tests := []struct {
		name                      string
		fields                    fields
		args                      args
		want                      bool
		wantErr                   bool
		wantAnimationsNewCaptions map[string]*storage.SentAnimation
	}{
		{
			"handler should return false, coz update not has message",
			fields{storage: emptyStorage},
			args{
				update: tgbotapi.Update{},
			},
			false,
			false,
			make(map[string]*storage.SentAnimation),
		},
		{
			"should be rejected due not allowed user",
			fields{
				storage: emptyStorage,
				config:  conf,
			},
			args{
				update: tgbotapi.Update{
					Message: &tgbotapi.Message{
						From: &tgbotapi.User{
							UserName: "not_cyhalothrin",
						},
					},
				},
			},
			false,
			false,
			make(map[string]*storage.SentAnimation),
		},
		{
			"not reply message should be skipped",
			fields{
				storage: emptyStorage,
				config:  conf,
			},
			args{
				update: tgbotapi.Update{
					Message: &tgbotapi.Message{
						From: &tgbotapi.User{
							UserName: "cyhalothrin",
						},
					},
				},
			},
			false,
			false,
			make(map[string]*storage.SentAnimation),
		},
		{
			"without animation message should be skipped",
			fields{
				storage: emptyStorage,
				config:  conf,
			},
			args{
				update: tgbotapi.Update{
					Message: &tgbotapi.Message{
						From: &tgbotapi.User{
							UserName: "cyhalothrin",
						},
						ReplyToMessage: &tgbotapi.Message{},
					},
				},
			},
			false,
			false,
			make(map[string]*storage.SentAnimation),
		},
		{
			"without text message should be skipped",
			fields{
				storage: emptyStorage,
				config:  conf,
			},
			args{
				update: tgbotapi.Update{
					Message: &tgbotapi.Message{
						From: &tgbotapi.User{
							UserName: "cyhalothrin",
						},
						ReplyToMessage: &tgbotapi.Message{
							Animation: &tgbotapi.ChatAnimation{},
						},
					},
				},
			},
			false,
			false,
			make(map[string]*storage.SentAnimation),
		},
		{
			"should add new caption",
			fields{
				storage: emptyStorage,
				config:  conf,
			},
			args{
				update: tgbotapi.Update{
					Message: &tgbotapi.Message{
						From: &tgbotapi.User{
							UserName: "cyhalothrin",
						},
						ReplyToMessage: &tgbotapi.Message{
							Animation: &tgbotapi.ChatAnimation{
								FileID: "animation_file_id_1",
							},
						},
						Text: "tag1 tag2 11tag tree11 00not a tag00",
					},
				},
			},
			true,
			false,
			map[string]*storage.SentAnimation{
				"animation_file_id_1": {
					FileID: "animation_file_id_1",
					Tags:   []string{"#tag1", "#tag2", "#tag_tree", "not a tag"},
				},
			},
		},
		{
			"should add caption with old message id",
			fields{
				storage: NewGifkoskladMetaStorageMock(mc).
					GetTagsAliasesMock.Return(nil).
					GetSentAnimationsMock.Return(map[string]*storage.SentAnimation{
					"animation_file_id_1": {
						MessageID: 101,
						FileID:    "animation_file_id_1",
						Tags:      []string{"#tag1", "#tag2", "not a tag"},
					},
				}).
					GetTagsMock.Return(nil),
				config: conf,
			},
			args{
				update: tgbotapi.Update{
					Message: &tgbotapi.Message{
						From: &tgbotapi.User{
							UserName: "cyhalothrin",
						},
						ReplyToMessage: &tgbotapi.Message{
							Animation: &tgbotapi.ChatAnimation{
								FileID: "animation_file_id_1",
							},
						},
						Text: "tag1 tag2 11tag tree11 00not a tag00",
					},
				},
			},
			true,
			false,
			map[string]*storage.SentAnimation{
				"animation_file_id_1": {
					MessageID: 101,
					FileID:    "animation_file_id_1",
					Tags:      []string{"#tag1", "#tag2", "#tag_tree", "not a tag"},
				},
			},
		},
		{
			"should not add caption with same tags",
			fields{
				storage: NewGifkoskladMetaStorageMock(mc).
					GetTagsAliasesMock.Return(nil).
					GetSentAnimationsMock.Return(map[string]*storage.SentAnimation{
					"animation_file_id_1": {
						MessageID: 101,
						FileID:    "animation_file_id_1",
						Tags:      []string{"#tag1", "#tag2", "not a tag"},
					},
				}).
					GetTagsMock.Return(nil),
				config: conf,
			},
			args{
				update: tgbotapi.Update{
					Message: &tgbotapi.Message{
						From: &tgbotapi.User{
							UserName: "cyhalothrin",
						},
						ReplyToMessage: &tgbotapi.Message{
							Animation: &tgbotapi.ChatAnimation{
								FileID: "animation_file_id_1",
							},
						},
						Text: "tag1 tag2 00not a tag00",
					},
				},
			},
			true,
			false,
			map[string]*storage.SentAnimation{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUpdatesHandler(
				tt.fields.config,
				tt.fields.storage,
				NewAlerterMock(mc),
				NewTelegramBotAPIMock(mc),
			)

			got, err := u.handleAnimationCaption(tt.args.update)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleAnimationCaption() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("handleAnimationCaption() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(u.animationsNewCaptions, tt.wantAnimationsNewCaptions) {
				t.Errorf("animationsNewCaptions = %v, want %v", u.animationsNewCaptions, tt.wantAnimationsNewCaptions)
			}
		})
	}
}

func TestUpdatesHandler_publishAnimations(t *testing.T) {
	t.Parallel()

	mc := minimock.NewController(t)
	defer mc.Finish()

	conf := config.Config{
		ChannelID: 1000,
	}

	type fields struct {
		api     telegramBotAPI
		storage GifkoskladMetaStorage
	}
	type args struct {
		animationsNewCaptions map[string]*storage.SentAnimation
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			"should send animations",
			fields{
				api: NewTelegramBotAPIMock(mc).
					EditMessageMock.
					Expect(conf.ChannelID, 10, "#tag3 #tag4 description").
					Return(nil).
					SendAnimationMock.
					Expect(conf.ChannelID, "new_file_id", "#tag1 #tag2 description").
					Return(20, nil),
				storage: NewGifkoskladMetaStorageMock(mc).
					GetTagsAliasesMock.Return(nil).
					GetSentAnimationsMock.Return(nil).
					GetTagsMock.Return(nil).
					AddSentAnimationsMock.
					Expect(map[string]*storage.SentAnimation{
						"new_file_id": {
							MessageID: 20,
							FileID:    "new_file_id",
							Tags:      []string{"#tag1", "#tag2", "description"},
						},
						"old_file_id": {
							MessageID: 10,
							FileID:    "old_file_id",
							Tags:      []string{"#tag3", "#tag4", "description"},
						},
					}).
					Return(),
			},
			args{
				animationsNewCaptions: map[string]*storage.SentAnimation{
					"new_file_id": {
						MessageID: 0,
						FileID:    "new_file_id",
						Tags:      []string{"#tag1", "#tag2", "description"},
					},
					"old_file_id": {
						MessageID: 10,
						FileID:    "old_file_id",
						Tags:      []string{"#tag3", "#tag4", "description"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUpdatesHandler(conf, tt.fields.storage, NewAlerterMock(mc), tt.fields.api)
			u.animationsNewCaptions = tt.args.animationsNewCaptions

			u.publishAnimations()
		})
	}
}

func TestUpdatesHandler_updateTagsList(t *testing.T) {
	t.Parallel()

	mc := minimock.NewController(t)
	defer mc.Finish()

	conf := config.Config{
		ChannelID: 10001,
	}
	type fields struct {
		api     telegramBotAPI
		storage GifkoskladMetaStorage
	}
	type args struct {
		uniqueTags         map[string]bool
		hasTagsListChanges bool
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		wantErr bool
	}{
		{
			"should create tags list",
			args{
				uniqueTags:         map[string]bool{"#tag3": true, "#tag1": true, "#tag2": true},
				hasTagsListChanges: true,
			},
			fields{
				storage: NewGifkoskladMetaStorageMock(mc).
					GetTagsAliasesMock.Return(nil).
					GetSentAnimationsMock.Return(nil).
					GetTagsMock.Return(nil).
					SetTagsMock.Expect([]string{"#tag1", "#tag2", "#tag3"}).Return().
					GetTagsListMessageIDMock.Return(10).
					SetTagsListMessageIDMock.Expect(100).Return(),
				api: NewTelegramBotAPIMock(mc).
					SendMessageMock.Expect(conf.ChannelID, "#tag1\n#tag2\n#tag3").Return(100, nil).
					DeleteMessageMock.Expect(conf.ChannelID, 10).Return(nil).
					PinMessageMock.Expect(conf.ChannelID, 100).Return(nil),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUpdatesHandler(conf, tt.fields.storage, NewAlerterMock(mc), tt.fields.api)
			u.hasTagsListChanges = tt.args.hasTagsListChanges
			u.uniqueTags = tt.args.uniqueTags

			if err := u.updateTagsList(); (err != nil) != tt.wantErr {
				t.Errorf("updateTagsList() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && u.hasTagsListChanges {
				t.Error("hasTagsListChanges should be reset")
			}
		})
	}
}
