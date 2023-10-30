package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	Togo "github.com/pya-h/togo4bot/Togo"
)

const (
	MaximumInlineButtonTextLength = 24
	MaximumNumberOfRowItems       = 3
	NumberOfSeparatorSpaces       = 2
)

// ---------------------- Telegram Response Struct & Interfaces --------------------------------
type TelegramAPI interface {
	CallAPI(res *http.ResponseWriter)
}

type TelegramResponse struct {
	TextMsg            string      `json:"text,omitempty"`
	TargetChatID       int64       `json:"chat_id"`
	Method             string      `json:"method"`
	ReplyMarkup        ReplyMarkup `json:"reply_markup,omitempty"`
	MessageRepliedTo   int         `json:"reply_to_message_id,omitempty"`
	MessageBeingEdited int         `json:"message_id,omitempty"` // for edit message & etc
	// file/photo?
}

type ReplyMarkup struct {
	ResizeKeyboard bool                       `json:"resize_keyboard,omitempty"`
	OneTime        bool                       `json:"one_time_keyboard,omitempty"`
	Keyboard       [][]string                 `json:"keyboard,omitempty"`
	InlineKeyboard [][]InlineKeyboardMenuItem `json:"inline_keyboard,omitempty"`
}

type InlineKeyboardMenuItem struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
	URL          string `json:"url,omitempty"`
}

func (response *TelegramResponse) CallAPI(res *http.ResponseWriter) {
	msg, _ := json.Marshal(response)
	log.Printf("Response %s", string(msg))
	//	fmt.Fprintf(*res,string(msg))
	(*res).Write(msg)
}

// ---------------------- Callback Structs & Functions --------------------------------
type UserAction uint8

const (
	None UserAction = iota
	TickTogo
	UpdateTogo
	RemoveTogo
	// ...
)

type CallbackData struct {
	Action  UserAction  `json:"A"`
	ID      int64       `json:"ID,omitempty"`
	Data    interface{} `json:"D,omitempty"`
	AllDays bool        `json:"AD,omitempty"`
}

func (callbackData CallbackData) Json() string {
	if res, err := json.Marshal(callbackData); err == nil {
		return string(res)
	} else {
		return fmt.Sprint(err)
	}
}

func LoadCallbackData(jsonString string) (data CallbackData) {
	json.Unmarshal([]byte(jsonString), &data)
	return
}

// ---------------------- Telegram Response Related Functions ------------------------------
func InlineKeyboardMenu(togos Togo.TogoList, action UserAction, allDays bool) (menu ReplyMarkup) {
	var (
		count     = len(togos)
		col       = 0
		row       = 0
		rowsCount = int(count / MaximumNumberOfRowItems)
	) // calculate the number of rows needed
	if count%MaximumNumberOfRowItems != 0 {
		rowsCount++
	}

	menu.InlineKeyboard = make([][]InlineKeyboardMenuItem, rowsCount)

	for i := range togos {
		if col == 0 {
			// calculting the number of column needed in each row
			if row < rowsCount-1 {
				menu.InlineKeyboard[row] = make([]InlineKeyboardMenuItem, MaximumNumberOfRowItems)
			} else {
				menu.InlineKeyboard[row] = make([]InlineKeyboardMenuItem, count-row*MaximumNumberOfRowItems)
			}
			row++
		}
		status := ""
		if togos[i].Progress >= 100 {
			status = "✅ "
		}
		var togoTitle string = fmt.Sprint(status, togos[i].Title)
		if len(togoTitle) >= MaximumInlineButtonTextLength {
			togoTitle = fmt.Sprint(togoTitle[:MaximumInlineButtonTextLength], "...")
		}
		menu.InlineKeyboard[row-1][col] = InlineKeyboardMenuItem{Text: togoTitle,
			CallbackData: (CallbackData{Action: action, ID: int64(togos[i].Id), AllDays: allDays}).Json()}
		col = (col + 1) % MaximumNumberOfRowItems
	}
	return
}

func MainKeyboardMenu() ReplyMarkup {
	return ReplyMarkup{ResizeKeyboard: true,
		OneTime:  false,
		Keyboard: [][]string{{"#", "%"}, {"#  -a", "%  -a"}, {"✅"}, {"❌", "❌  -a"}}}
}

// ---------------------- tgbotapi Related Functions ------------------------------
func GetTgBotApiFunction(update *tgbotapi.Update) func(data string) string {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	return func(data string) string {
		if err == nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, data)
			// msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
			return "✅!"
		}
		return fmt.Sprintln("Fuck: ", err)
	}
}

// ---------------------- Togos Related Functions ------------------------------
func LoadForToday(chatId int64, togos *Togo.TogoList) {
	tg, err := Togo.Load(chatId, true) // load today's togos,  make(Togo.TogoList, 0)
	if err != nil {
		fmt.Println("Loading failed: ", err)
	}
	*togos = tg

	/*
		today := time.Now()
		// mainTaskScheduler.Schedule(func(ctx context.Context) { LoadForToday(togos) },
		// 	chrono.WithStartTime(today.Year(), today.Month(), today.Day()+1, 0, 0, 0))
	*/
}
func SplitArguments(statement string) []string {
	result := make([]string, 0)
	numOfSpaces := 0
	segmentStartIndex := 0

	for i := range statement {
		if statement[i] == ' ' {
			numOfSpaces++
		} else if numOfSpaces > 0 {
			if numOfSpaces == NumberOfSeparatorSpaces {
				result = append(result, statement[segmentStartIndex:i-NumberOfSeparatorSpaces])
				segmentStartIndex = i
			}
			numOfSpaces = 0
		}

	}
	result = append(result, statement[segmentStartIndex:])
	return result
}

func Log(update *tgbotapi.Update, values []string) {
	sendMessage := GetTgBotApiFunction(update)
	var r string
	for _, v := range values {
		r += v + "\n"
	}
	sendMessage(r)
}

// ---------------------- Serverless Function ------------------------------
func Handler(res http.ResponseWriter, r *http.Request) {
	var togos Togo.TogoList

	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)
	var update tgbotapi.Update
	if err := json.Unmarshal(body, &update); err != nil {
		log.Fatal("Error en el update →", err)
	}

	res.Header().Add("Content-Type", "application/json")

	//if update.Message.IsCommand() {
	response := TelegramResponse{TextMsg: "What?",
		Method: "sendMessage"} // default method is sendMessage

	defer func() {
		err := recover()
		if err != nil {
			response.TextMsg = fmt.Sprintln("❌: ", err)
			response.CallAPI(&res)
		}
	}()

	// ---------------------- Handling Casual Telegram text Messages ------------------------------
	if update.Message != nil { // If we got a message
		response.ReplyMarkup = MainKeyboardMenu() // default keyboard
		response.TargetChatID = update.Message.Chat.ID
		response.MessageRepliedTo = update.Message.MessageID
		// log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		LoadForToday(update.Message.Chat.ID, &togos)

		terms := SplitArguments(update.Message.Text)
		// Log(&update, terms)
		numOfTerms := len(terms)
		var now Togo.Date = Togo.Today()
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
					response.TextMsg = fmt.Sprint(now.Get(), ": DONE!")
				} else {
					response.TextMsg = "You must provide some values!!"
				}
			case "#":
				var results []string
				if i+1 < numOfTerms && terms[i+1] == "-a" {
					allTogos, err := Togo.Load(update.Message.Chat.ID, false)
					if err != nil {
						panic(err)
					}
					results = allTogos.ToString()
				} else {
					results = togos.ToString()
				}
				sendMessage := GetTgBotApiFunction(&update)
				if len(results) > 0 {
					for i := range results {
						sendMessage(results[i])
					}
					response.TextMsg = "✅!"
				} else {
					response.TextMsg = "Nothing!"
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
				response.TextMsg = fmt.Sprintf("%s Progress: %3.2f%% \n%3.2f%% Completed\nStatistics: %d / %d",
					scope, progress, completedInPercent, completed, total)
				if extra > 0 {
					response.TextMsg = fmt.Sprintf("%s[+%d]\n", response.TextMsg, extra)
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
							response.TextMsg = allTogos.Update(update.Message.Chat.ID, terms[i+2:])
						} else {
							response.TextMsg = "Insufficient number of parameters!"

						}
					} else {
						response.TextMsg = togos.Update(update.Message.Chat.ID, terms[i+1:])

					}
				} else {
					response.TextMsg = "Insufficient number of parameters!"
				}
			// TODO: write Tick command
			case "✅":
				response.TextMsg = "Here are your togos for today:"
				response.ReplyMarkup = InlineKeyboardMenu(togos, TickTogo, false)
			case "❌":
				response.TextMsg = "Here are your togos for today:"
				response.ReplyMarkup = InlineKeyboardMenu(togos, RemoveTogo, false)
			case "❌  -a":
				allTogos, err := Togo.Load(update.Message.Chat.ID, false)
				if err != nil {
					panic(err)
				}
				response.TextMsg = "Here are your ALL togos:"
				response.ReplyMarkup = InlineKeyboardMenu(allTogos, RemoveTogo, true)

			case "/now":
				response.TextMsg = now.Get()

			}

		}
		response.CallAPI(&res)

	} else if update.CallbackQuery != nil {
		response.MessageBeingEdited = update.CallbackQuery.Message.MessageID
		response.TargetChatID = update.CallbackQuery.Message.Chat.ID
		response.Method = "editMessageText"

		callbackData := LoadCallbackData(update.CallbackQuery.Data)
		if !callbackData.AllDays {
			LoadForToday(response.TargetChatID, &togos)
		} else {
			var err error
			togos, err = Togo.Load(update.Message.Chat.ID, false)
			if err != nil {
				panic(err)
			}
		}

		switch callbackData.Action {
		case TickTogo:
			togo, err := togos.Get(uint64(callbackData.ID))
			if err != nil {
				panic(err)
			} else {
				if (*togo).Progress < 100 {
					(*togo).Progress = 100
				} else {
					(*togo).Progress = 0
				}
				(*togo).Update(response.TargetChatID)
				response.ReplyMarkup = InlineKeyboardMenu(togos, TickTogo, false)
				response.TextMsg = "✅ DONE! Now select the next togo you want to tick ..."
			}
		case RemoveTogo:
			err := togos.Remove(response.TargetChatID, uint64(callbackData.ID))
			if err == nil {
				response.TextMsg = "❌ DONE! Now select the next togo you want to REMOVE ..."
				response.ReplyMarkup = InlineKeyboardMenu(togos, RemoveTogo, callbackData.AllDays)

			} else {
				panic(err)
			}
		}
		response.CallAPI(&res)
	}

}
