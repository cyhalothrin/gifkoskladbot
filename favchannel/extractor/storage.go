package extractor

type storage interface {
	GetTags() []string
	SetTags([]string)
	// getting and saving information on processed messages of the favorite channel
	SetFavChannelLastForwardedMessageIDWithoutCaption(int64)
	GetFavChannelLastForwardedMessageIDWithoutCaption() int64
}
