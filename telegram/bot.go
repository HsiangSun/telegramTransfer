package telegram

import (
	tb "gopkg.in/telebot.v3"
	"log"
	"telgramTransfer/utils/config"
	"time"
)

var Bot *tb.Bot

func BootRun() {
	setting := tb.Settings{
		//Token:   "5815075296:AAEA_XGTHpp57Tefb0kQhvoIKCtCht68g-Q",
		Token:   config.BootC.Token,
		Updates: 100,
		Poller:  &tb.LongPoller{Timeout: 2 * time.Second},
		//Poller:  &tb.Webhook{Timeout: 10 * time.Second},
		OnError: func(err error, context tb.Context) {
			log.Printf("Error:%+v \n", err)
		},
	}

	b, err := tb.NewBot(setting)
	if err != nil {
		log.Fatal(err)
	}
	Bot = b
	RegisterHandle()
	b.Start()
}

func RegisterHandle() {
	Bot.Handle(tb.OnText, OnTextMessage)
}
