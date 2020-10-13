package extractor

import (
	"testing"
)

func TestGifExtractor_moveMessagesWithoutCaptionToBotChannel(t *testing.T) {
	t.Skip()

	//mc := minimock.NewController(t)
	//defer mc.Finish()
	//
	//type fields struct {
	//	client  extractorClient
	//	storage storage
	//	conf    config.TDLibClient
	//}
	//tests := []struct {
	//	name    string
	//	fields  fields
	//	wantErr bool
	//}{
	//	{
	//		"success",
	//		fields{
	//			client: NewExtractorClientMock(mc).
	//				GetFavChannelIDMock.Return(11, nil).
	//				getChatHistoryRemoteMock.
	//				When(11, 22214, 0, 100).
	//				Then(
	//					&tdlib.Messages{
	//						TotalCount: 0,
	//						Messages:   []tdlib.Message{},
	//					},
	//					nil,
	//				).
	//				getChatHistoryRemoteMock.
	//				When(11, 22211, 0, 100).
	//				Then(
	//					&tdlib.Messages{
	//						TotalCount: 1,
	//						Messages: []tdlib.Message{
	//							{
	//								ID: 22212,
	//								Content: &tdlib.MessageAnimation{
	//									Animation: nil,
	//									Caption: &tdlib.FormattedText{
	//										Text: "",
	//									},
	//								},
	//							},
	//						},
	//					},
	//					nil,
	//				).
	//				getChatHistoryRemoteMock.
	//				When(11, 22212, 0, 100).
	//				Then(
	//					&tdlib.Messages{
	//						TotalCount: 2,
	//						Messages: []tdlib.Message{
	//							{
	//								ID: 22213,
	//								Content: &tdlib.MessageAnimation{
	//									Animation: nil,
	//									Caption: &tdlib.FormattedText{
	//										Text: "",
	//									},
	//								},
	//							},
	//							{
	//								ID: 22214,
	//								Content: &tdlib.MessageAnimation{
	//									Animation: nil,
	//									Caption: &tdlib.FormattedText{
	//										Text: "#gif #has_caption",
	//									},
	//								},
	//							},
	//						},
	//					},
	//					nil,
	//				).
	//				forwardMessagesSilentlyMock.Expect([]int64{22212, 22213}, 11, 131313).Return(nil).
	//				removeMessagesMock.Expect(11, []int64{22212, 22213}).Return(nil),
	//			storage: NewStorageMock(mc).GetFavChannelLastForwardedMessageIDWithoutCaptionMock.
	//				Return(22211).
	//				SetFavChannelLastForwardedMessageIDWithoutCaptionMock.Expect(22213).Return(),
	//			conf: config.ExtractorConfig{
	//				BotChatID: 131313,
	//			},
	//		},
	//		false,
	//	},
	//}
	//for _, tt := range tests {
	//	t.Run(tt.name, func(t *testing.T) {
	//		g := &GifExtractor{
	//			client:  tt.fields.client,
	//			storage: tt.fields.storage,
	//			conf:    tt.fields.conf,
	//		}
	//		if err := g.moveMessagesWithoutCaptionToBotChannel(); (err != nil) != tt.wantErr {
	//			t.Errorf("moveMessagesWithoutCaptionToBotChannel() error = %v, wantErr %v", err, tt.wantErr)
	//		}
	//	})
	//}
}
