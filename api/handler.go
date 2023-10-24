package handler

import (
    "fmt"
    "net/http"
    "encoding/json"
    "github.com/go-telegram-bot-api/telegram-bot-api"
    "io/ioutil"
	Togo "github.com/pya-h/Togo"
	"log"
	"strings"
)

type Response struct {
    Msg string `json:"text"`
    ChatID int64 `json:"chat_id"`
    Method string `json:"method"`
}

func autoLoad(togos *Togo.TogoList) {
	tg, err := Togo.Load(true) // load today's togos,  make(Togo.TogoList, 0)
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

func Handler(w http.ResponseWriter, r *http.Request) {


	var togos Togo.TogoList
	autoLoad(&togos)

    defer r.Body.Close()
    body, _ := ioutil.ReadAll(r.Body)
    var update tgbotapi.Update
    if err := json.Unmarshal(body,&update); err != nil {
        log.Fatal("Error en el update →", err)
    }
    log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

    //if update.Message.IsCommand() {
	if update.Message != nil { // If we got a message
		defer func() {
			err := recover()
			if err != nil {
				data := Response{ Msg: fmt.Sprint(err),
					Method: "sendMessage",
					ChatID: update.Message.Chat.ID }

				msg, _ := json.Marshal( data )
				log.Printf("Response %s", string(msg))
				w.Header().Add("Content-Type", "application/json")
				fmt.Fprintf(w,string(msg))
				r.Body.Close()
			}
		}()
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		input := update.Message.Text[:len(update.Message.Text)]
		terms := strings.Split(input, "\n")
		numOfTerms := len(terms)
		var response string = "What?"
		var now Togo.Date = Togo.Today()
		for i := 0; i < numOfTerms; i++ {
			switch terms[i] {
			case "+":
				if numOfTerms > 1 {

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
				if i+1 < numOfTerms && terms[i+1] == "-a" {
					allTogos, err := Togo.Load(false)
					if err != nil {
						panic(err)
					}
					response = allTogos.ToString()
				} else {
					response = togos.ToString()
				}
			case "%":
				var target *Togo.TogoList = &togos
				scope := "Today's"
				if i+1 < numOfTerms && terms[i+1] == "-a" {
					allTogos, err := Togo.Load(false)
					if err != nil {
						panic(err)
					}
					target = &allTogos
					scope = "Total"
				}
				progress, completedInPercent, completed, extra, total := (*target).ProgressMade()
				response = fmt.Sprintf("%s Progress: %3.2f%% \n(%3.2f%% Completed),\nStatistics: %d / %d",
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
		// msg.ReplyToMessageID = update.Message.MessageID
		data := Response{ Msg: response,
			Method: "sendMessage",
			ChatID: update.Message.Chat.ID }
	
		msg, _ := json.Marshal( data )
		log.Printf("Response %s", string(msg))
		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(w,string(msg))
	}

}