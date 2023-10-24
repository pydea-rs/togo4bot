package main

import (
	"log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"Togo"
	"bufio"
	chrono "github.com/gochrono/chrono"
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

var mainTaskScheduler chrono.TaskScheduler = chrono.NewDefaultTaskScheduler()

func autoLoad(togos *Togo.TogoList) {
	tg, err := Togo.Load(true) // load today's togos,  make(Togo.TogoList, 0)
	if err != nil {
		fmt.Println("Loading failed: ", err)
	}
	*togos = tg
	today := time.Now()
	mainTaskScheduler.Schedule(func(ctx context.Context) { autoLoad(togos) },
		chrono.WithStartTime(today.Year(), today.Month(), today.Day()+1, 0, 0, 0))

}

func main() {
	defer func() {
		err := recover()
		if err != nil {
			log.Fatal("Something fucked up: ", err)
		}
	}()

	var togos Togo.TogoList
	autoLoad(&togos)
	
	bot, err := tgbotapi.NewBotAPI("6599650500:AAEVpmrVg3BAIy4x8i1W63eRCuyVckyBEug")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	if updates, err := bot.GetUpdatesChan(u); err == nil {

		for update := range updates {
			if update.Message != nil { // If we got a message
				log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
				input = update.Message.Text[:len(input)-1] // remove '\n' char from the end of string
				terms := strings.Split(input, "    ")
				num_of_terms := len(terms)
				var response string
				var now Togo.Date = Togo.Today()
				for i := 0; i < num_of_terms; i++ {
					switch terms[i] {
					case "+":
						if num_of_terms > 1 {
	
							togo := Togo.Extract(terms[i+1:], togos.NextID())
							if togo.Date.Short() == now.Short() {
								togos = togos.Add(&togo)
								if togo.Date.After(now.Time) {
									togo.Schedule()
								}
							}
	
							togo.Save()
							response = "Done!"
						} else {
							response = "You must provide some values!"
						}
					case "#":
						if i+1 < num_of_terms && terms[i+1] == "-a" {
							all_togos, err := Togo.Load(false)
							if err != nil {
								panic(err)
							}
							response = all_togos.ToString()
						} else {
							response = togos.ToString()
						}
					case "%":
						var target *Togo.TogoList = &togos
						scope := "Today's"
						if i+1 < num_of_terms && terms[i+1] == "-a" {
							all_togos, err := Togo.Load(false)
							if err != nil {
								panic(err)
							}
							target = &all_togos
							scope = "Total"
						}
						progress, completedInPercent, completed, extra, total := (*target).ProgressMade()
						response = fmt.Sprintf("%s Progress: %3.2f%% (%3.2f%% Completed),\nStatistics: %d / %d",
							scope, progress, completedInPercent, completed, total)
						if extra > 0 {
							response = fmt.Sprintf("%s[+%d]\n", response, extra)
						}
					case "$":
						// set or update a togo
						// only for today togos
						togos.Update(terms[i+1:])
						response = "Done!"

					}
				}

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
				msg.ReplyToMessageID = update.Message.MessageID

				bot.Send(msg)
	
			}
		}
	} else {
		log.Fatalln(err)
	}
}