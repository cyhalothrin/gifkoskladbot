package bot

import "github.com/cyhalothrin/gifkoskladbot/storage"

type GifkoskladMetaStorage interface {
	GetTags() []string
	SetTags([]string)
	GetTagsAliases() map[string]string
	SetTagsAliases(map[string]string)
	GetSentAnimations() map[string]*storage.SentAnimation
	// AddSentAnimations adds new sent animations to storage
	AddSentAnimations(map[string]*storage.SentAnimation)
}
