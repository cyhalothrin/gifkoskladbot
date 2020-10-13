package extractor

import (
	"log"

	"github.com/cyhalothrin/gifkoskladbot/config"
	"github.com/cyhalothrin/gifkoskladbot/favchannel/tdlibclient"
	fileStorage "github.com/cyhalothrin/gifkoskladbot/storage"
)

func ForwardWithEmptyCaption() error {
	conf, err := config.ReadConfig()
	if err != nil {
		return err
	}

	store, err := fileStorage.NewFileMetaStorage(conf.StoragePath)
	if err != nil {
		return err
	}
	defer store.Close()

	client, err := tdlibclient.NewClient(conf.TDLib)
	if err != nil {
		return err
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Println("destroy telegram client:", err)
		}
	}()

	gifExt, err := NewGifExtractor(conf, store, client)
	if err != nil {
		return err
	}

	return gifExt.moveMessagesWithoutCaptionToBotChannel()
}
