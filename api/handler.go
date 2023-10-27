package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	Togo "github.com/pya-h/togo4bot/Togo"
)

type KeyboardMenu interface {
	HttpSendMessage(res *http.ResponseWriter, chatID int64, text string, messageID int)
}

const (
	MaximumInlineButtonTextLength uint8 = 16
	MaximumNumberOfRowItems             = 3
)

type Response struct {
	Msg              string      `json:"text"`
	ChatID           int64       `json:"chat_id"`
	Method           string      `json:"method"`
	ReplyMarkup      ReplyMarkup `json:"reply_markup"`
	ReplyToMessageID int         `json:"reply_to_message_id"`
}

type ReplyMarkup struct {
	ResizeKeyboard bool       `json:"resize_keyboard"`
	OneTime        bool       `json:"one_time_keyboard"`
	Keyboard       [][]string `json:"keyboard"`
	InlineKeyboard [][]InlineKeyboardMenuItem `json:"inline_keyboard"`
}

type InlineKeyboardMenuItem struct {
	Text string `json:"text"`
	CallbackData CallbackData `json:"callback_data"`
	URL string `json:"url"`
}

type UserAction uint8

const (
	TickTogo UserAction = iota
	UpdateTogo
	DeleteTogo
	// ...
)

type CallbackData struct {
	Action UserAction
	Id int64
	Data interface{}
}

func InlineKeyboardMenu(togos Togo.TogoList) (menu ReplyMarkup, action UserAction) {
	col := 0
	row := 0
	menu.InlineKeyboard = make([][]InlineKeyboardMenuItem, int(len(togos) / 3) + 1)

	for _, togo := range togos {
		if col == 0 {
			menu.InlineKeyboard[row] = make([]InlineKeyboardMenuItem, MaximumNumberOfRowItems)
			row++
		}
		var togoTitle string = togo.Title
		if len(togoTitle) >= int(MaximumInlineButtonTextLength) {
			togoTitle = fmt.Sprint(togoTitle[:MaximumInlineButtonTextLength], "...")
		}
		menu.InlineKeyboard[row - 1][col] = InlineKeyboardMenuItem{Text: togoTitle,
			CallbackData: CallbackData{Action: action, Id: togo.Id}}
		col = (col + 1) % MaximumNumberOfRowItems
	}
	return
}

func MainKeyboardMenu() ReplyMarkup {
	return ReplyMarkup{ResizeKeyboard: true,
		OneTime:  false,
		Keyboard: [][]string{{"#", "%"}, {"#   -a", "%   -a"}, {"✅"}}}
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

func (replyMarkup ReplyMarkup) HttpSendMessage(res *http.ResponseWriter, chatID int64, text string, messageID int) {
	data := Response{Msg: text,
		Method:           "sendMessage",
		ReplyMarkup:      replyMarkup,
		ChatID:           chatID,
		ReplyToMessageID: messageID}

	msg, _ := json.Marshal(data)
	log.Printf("Response %s", string(msg))
	//	fmt.Fprintf(*res,string(msg))
	(*res).Write(msg)
}

func GetBotFunction(update *tgbotapi.Update) func(data string) string {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	return func(data string) string {
		if err == nil {
			msg := tgbotapi.NewMessage((*update).Message.Chat.ID, data)
			// msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
			return "✅!"
		}
		return fmt.Sprintln("Fuck: ", err)
	}
}

func Handler(res http.ResponseWriter, r *http.Request) {
	var togos Togo.TogoList

	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	var update tgbotapi.Update
	if err := json.Unmarshal(body, &update); err != nil {
		log.Fatal("Error en el update →", err)
	}

	res.Header().Add("Content-Type", "application/json")

	//if update.Message.IsCommand() {
	if update.Message != nil { // If we got a message
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		defer func() {
			err := recover()
			if err != nil {
				HttpSendMessage(&res, update.Message.Chat.ID, fmt.Sprint(err), update.Message.MessageID)
			}
		}()

		autoLoad(update.Message.Chat.ID, &togos)

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		input := update.Message.Text[:len(update.Message.Text)]
		terms := strings.Split(input, "   ")
		numOfTerms := len(terms)
		var response string = "What?"
		var now Togo.Date = Togo.Today()
		var menu = MainKeyboardMenu()
		for i := 0; i < numOfTerms; i++ {
			switch terms[i] {
			case "+":
				if numOfTerms > 1 {

					togo := Togo.Extract(update.Message.Chat.ID, terms[i+1:])
					togo.Id = togo.Save()
					if togo.Date.Short() == now.Short() {
						togos = togos.Add(&togo)
						if togo.Date.After(now.Time) {
							togo.Schedule()
						}
					}

					response = fmt.Sprint(now.Get(), ": DONE!")
				} else {
					response = "You must provide some values!!"
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
				sendMessage := GetBotFunction(&update)
				if len(result) > 0 {
					for _, tg := range result {
						sendMessage(tg)
					}
					response = "✅!"
				} else {
					response = "Nothing!"
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
				if i+1 < numOfTerms {
					if terms[i+1] == "-a" {
						if i+2 < numOfTerms {
							allTogos, err := Togo.Load(update.Message.Chat.ID, false)
							if err != nil {
								panic(err)
							}
							response = allTogos.Update(update.Message.Chat.ID, terms[i+2:])
						} else {
							response = "Insufficient number of parameters!"

						}
					} else {
						response = togos.Update(update.Message.Chat.ID, terms[i+1:])

					}
				} else {
					response = "Insufficient number of parameters!"
				}
			case "✅":
				menu = InlineKeyboardMenu(togos, TickTogo)
			case "/now":
				response = now.Get()

			}

		}

		menu.HttpSendMessage(&res, update.Message.Chat.ID, response, update.Message.MessageID)
	}

}
