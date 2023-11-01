package notification
 
import (
	"fmt"
	"net/http"
	"log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ---------------------- tgbotapi Related Functions ------------------------------
func GetTgBotApiFunction() func(chatID int64, data string) string {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	return func(data string) string {
		if err == nil {
			msg := tgbotapi.NewMessage(chatID, data)
			// msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
			return "✅!"
		}
		return fmt.Sprintln("Fuck: ", err)
	}
}

func Handler(res http.ResponseWriter, req *http.Request) {
	log.Println("Cron running")
	sendMessage := GetTgBotApiFunction()
	sendMessage(1137617789, "Cron Test")
	log.Println("Cron done")

	res.Write("Successfull!")
}