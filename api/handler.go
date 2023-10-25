package handler

import (
    "fmt"
    "net/http"
    "encoding/json"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
    "io/ioutil"
	Togo "github.com/pya-h/togo4bot/Togo"
	"log"
	"strings"
    "os"
)

type Response struct {
    Msg string `json:"text"`
    ChatID int64 `json:"chat_id"`
    Method string `json:"method"`
}

func autoLoad(chatId int64, togos *Togo.TogoList) {
	tg, err := Togo.Load(chatId, true) // load today's togos,  make(Togo.TogoList, 0)
	if err != nil {
		fmt.Println("Loading failed: ", err)
	}
	*togos = tg
	/*
	today := time.Now()
	// mainTaskScheduler.Schedule(func(ctx context.Context) { autoLoad(togos) },
	// 	chrono.WithStartTime(today.Year(), today.Month(), today.Day()+1, 0, 0, 0))
	*/
}

func SendMessage(res *http.ResponseWriter, chatID int64, text string) {
	data := Response{ Msg: text,
		Method: "sendMessage",
		ChatID: chatID}
	msg, _ := json.Marshal( data )
	log.Printf("Response %s", string(msg))
//	fmt.Fprintf(*res,string(msg))
	(*res).Write(msg)
}

func Handler(res http.ResponseWriter, r *http.Request) {
	var togos Togo.TogoList

    defer r.Body.Close()
    body, _ := ioutil.ReadAll(r.Body)
    var update tgbotapi.Update
    if err := json.Unmarshal(body,&update); err != nil {
        log.Fatal("Error en el update →", err)
    }

	res.Header().Add("Content-Type", "application/json")
	
    //if update.Message.IsCommand() {
	if update.Message != nil { // If we got a message
		autoLoad(update.Message.Chat.ID, &togos)
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		defer func() {
			err := recover()
			if err != nil {
				SendMessage(&res, update.Message.Chat.ID, fmt.Sprint(err))
			}
		}()

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		input := update.Message.Text[:len(update.Message.Text)]
		terms := strings.Split(input, "   ")
		numOfTerms := len(terms)
		var response string = "What?"
		var now Togo.Date = Togo.Today()
		for i := 0; i < numOfTerms; i++ {
			switch terms[i] {
			case "+":
				if numOfTerms > 1 {

					togo := Togo.Extract(update.Message.Chat.ID, terms[i+1:], togos.NextID())
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
				var result []string
				if i+1 < numOfTerms && terms[i+1] == "-a" {
					allTogos, err := Togo.Load(update.Message.Chat.ID, false)
					if err != nil {
						panic(err)
					}
					result = allTogos.ToString()
				} else {
					result = togos.ToString()
				}
				bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
				if err == nil {
					if len(result) > 0 {

						for _, tg := range result {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, tg)
							// msg.ReplyToMessageID = update.Message.MessageID
							bot.Send(msg)
						}
					} else {
						response = "Nothing!"
					}
					return
				} else {
					response = fmt.Sprint(err)// "I can't send you the details!"
				}
			case "%":
				var target *Togo.TogoList = &togos
				scope := "Today's"
				if i+1 < numOfTerms && terms[i+1] == "-a" {
					allTogos, err := Togo.Load(update.Message.Chat.ID, false)
					if err != nil {
						panic(err)
					}
					target = &allTogos
					scope = "Total"
				}
				progress, completedInPercent, completed, extra, total := (*target).ProgressMade()
				response = fmt.Sprintf("%s Progress: %3.2f%% \n%3.2f%% Completed\nStatistics: %d / %d",
					scope, progress, completedInPercent, completed, total)
				if extra > 0 {
					response = fmt.Sprintf("%s[+%d]\n", response, extra)
				}
			case "$":
				// set or update a togo
				// only for today togos
				response = togos.Update(terms[i+1:])

			}

		}

		SendMessage(&res, update.Message.Chat.ID, response)
	}

}