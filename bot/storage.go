package bot

import "github.com/cyhalothrin/gifkoskladbot/storage"

type GifkoskladMetaStorage interface {
	GetTags() []string
	SetTags([]string)
	GetTagsAliases() map[string]string
	SetTagsAliases(map[string]string)
	GetTagsListMessageID() int
	SetTagsListMessageID(int)
	GetSentAnimations() map[string]*storage.SentAnimation
	AddSentAnimations(map[string]*storage.SentAnimation)
}
