package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cyhalothrin/gifkoskladbot/api"
	"github.com/cyhalothrin/gifkoskladbot/config"
	"github.com/cyhalothrin/gifkoskladbot/storage"
)

func HandleNewMessages() error {
	gbot, err := newGifkoSkladBot()
	if err != nil {
		return err
	}
	defer gbot.close()

	return gbot.handleNewMessages()
}

func PollUpdates() error {
	gbot, err := newGifkoSkladBot()
	if err != nil {
		return err
	}
	defer gbot.close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		defer cancel()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		fmt.Println("Понял, ща доработаю и выключусь")
		<-sigCh
	}()

	return gbot.poll(ctx)
}

type gsBot struct {
	conf    config.Config
	store   *storage.FileMetaStorage
	tgAPI   *api.TelegramBotAPI
	handler *UpdatesHandler
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
		conf:    conf,
		store:   store,
		tgAPI:   tgAPI,
		handler: NewUpdatesHandler(conf, store, NewTgAlert(conf, tgAPI), tgAPI),
	}, nil
}

func (g *gsBot) handleNewMessages() error {
	updates, err := g.tgAPI.GetUpdates()
	if err != nil {
		return err
	}

	return g.handler.HandleUpdates(updates)
}

func (g *gsBot) poll(ctx context.Context) error {
	if err := g.handleNewMessages(); err != nil {
		return err
	}

	waitDuration := 30 * time.Second

	for {
		log.Printf("Подожду %s\n", waitDuration.String())

		select {
		case <-time.After(30 * time.Second):
			if err := g.handleNewMessages(); err != nil {
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (g *gsBot) close() {
	g.store.Close()
}
