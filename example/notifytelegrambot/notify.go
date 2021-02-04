package main

import (
	"fmt"
	"log"
	"time"

	"github.com/abserari/ip-arp/fing"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	bot, err := tgbotapi.NewBotAPI("1654007818:AAEPJ2d-YZy3GshoDdD44z_dyKLX_UrPMig")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	go notify(bot)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
	}

}

type online struct {
	name    string
	online  int
	updated bool
}

var onlineMap = map[string]*online{"f8:da:0c:50:e2:25": {"杨鼎睿", 0, false}}

func notify(bot *tgbotapi.BotAPI) {
	for {
		var msgs []string
		f := new(fing.Fing)
		f.Detect()
		// f.Show()
		for _, v := range onlineMap {
			v.updated = false
		}
		for _, d := range f.Devices {
			v := onlineMap[d.Mac]
			if v == nil {
				continue
			}
			if v.online == 0 {
				msgs = append(msgs, fmt.Sprintf("%s online", v.name))
			}
			v.online = 1
			v.updated = true
		}

		for _, v := range onlineMap {
			if !v.updated {
				// offline
				v.online = 0
				msgs = append(msgs, fmt.Sprintf("%s offline", v.name))
			}
		}
		var str string
		for _, s := range msgs {
			str += s
			str += "\n"
		}
		msg := tgbotapi.NewMessage(891630877, str)
		bot.Send(msg)
		time.Sleep(time.Minute)
	}
}
