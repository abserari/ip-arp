package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/abserari/ip-arp/fing"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type online struct {
	name    string
	online  int
	updated bool
}

var onlineMap = make(map[string]*online)
var mutex sync.Mutex

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

	go notify(bot, false)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "detect":
				{
					notify(bot, true)
					continue
				}
			}
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
	}

}

func notify(bot *tgbotapi.BotAPI, instance bool) {
	for {
		raw, err := LoadFile("./list.json")
		if err != nil {
			log.Println(err)
			return
		}
		peoples, err := UnmarshalListProjects(raw)
		if err != nil {
			log.Println(err)
			return
		}
		mutex.Lock()
		for _, p := range peoples {
			p.Mac = strings.ToLower(p.Mac)
			p.Mac = strings.ReplaceAll(p.Mac, "-", ":")
			// log.Println(p.Mac)
			if onlineMap[p.Mac] != nil {
				continue
			}
			var people online
			people.name = p.Login
			people.online = 0
			onlineMap[p.Mac] = &people
		}
		mutex.Unlock()

		var msgs []string
		var str string
		msg := tgbotapi.NewMessage(891630877, "")
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
				msgs = append(msgs, fmt.Sprintf("%s ❤  ---------online", v.name))
			}
			v.online = 1
			v.updated = true
		}

		for _, v := range onlineMap {
			if !v.updated {
				// offline
				if v.online == 1 {
					msgs = append(msgs, fmt.Sprintf("%s 😴 xxxxxxxxxxxoffline", v.name))
				}
				v.online = 0
			}
		}
		log.Println("updated")
		if len(msgs) == 0 {
			goto Sleep
		}
		for _, s := range msgs {
			str += s
			str += "\n"
		}
		msg.Text = str
		_, _ = bot.Send(msg)

	Sleep:
		if instance {
			return
		}
		time.Sleep(time.Minute)
	}
}

func LoadFile(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type People struct {
	Login string `json:"login"`
	Mac   string `json:"mac"`
}

// UnmarshalListProjects -
func UnmarshalListProjects(data []byte) ([]People, error) {
	var Peoples []People

	err := json.Unmarshal(data, &Peoples)
	if err != nil {
		return nil, err
	}

	return Peoples, nil
}
