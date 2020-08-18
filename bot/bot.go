package bot

import (
	"github.com/cyhalothrin/gifkoskladbot/api"
	"github.com/cyhalothrin/gifkoskladbot/config"
	"github.com/cyhalothrin/gifkoskladbot/storage"
)

func HandleNewMessages() error {
	gbot, err := newGifkoSkladBot()
	if err != nil {
		return err
	}

	return gbot.handleNewMessages()
}

type gsBot struct {
	conf  config.Config
	store *storage.FileMetaStorage
	tgAPI *api.TelegramBotAPI
}

func newGifkoSkladBot() (*gsBot, error) {
	conf, err := config.ReadConfig()
	if err != nil {
		return nil, err
	}

	store, err := storage.NewFileMetaStorage(conf.StoragePath)
	if err != nil {
		return nil, err
	}

	tgAPI, err := api.NewTelegramBotAPI(conf)
	if err != nil {
		return nil, err
	}

	return &gsBot{
		conf:  conf,
		store: store,
		tgAPI: tgAPI,
	}, nil
}

func (g *gsBot) handleNewMessages() error {
	defer g.close()

	return nil
	handler := NewUpdatesHandler(g.conf, g.store, NewTgAlert(g.conf, g.tgAPI), g.tgAPI)

	updates, err := g.tgAPI.GetUpdates()
	if err != nil {
		return err
	}

	return handler.HandleUpdates(updates)
}

func (g *gsBot) close() {
	g.store.Close()
}
