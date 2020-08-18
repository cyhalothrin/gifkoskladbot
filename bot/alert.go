package bot

import (
	"fmt"

	"github.com/cyhalothrin/gifkoskladbot/config"
)

type alerter interface {
	Send(err error) error
}

type TgAlert struct {
	api  telegramBotAPI
	conf config.Config
}

func NewTgAlert(cnf config.Config, api telegramBotAPI) *TgAlert {
	return &TgAlert{
		api:  api,
		conf: cnf,
	}
}

func (t *TgAlert) Send(err error) error {
	fmt.Println(err)

	return nil
}
