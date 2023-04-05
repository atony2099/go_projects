package bot

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/atony2099/go_project/telebot/db"
	"github.com/atony2099/go_project/telebot/elapse"
	"github.com/wcharczuk/go-chart/v2"

	"github.com/atony2099/go_project/telebot/img"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var bot *tgbotapi.BotAPI
var chatID int64

func NewBot(token string, chat int64) {
	var err error
	bot, err = tgbotapi.NewBotAPI(token)
	chatID = chat
	if err != nil {
		log.Panic(err)
	}

}

func SendBotMsg(message string) {

	msg := tgbotapi.NewMessage(chatID, message)

	bot.Send(msg)

}

func HandleCommand() {

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message != nil { // If we got a message
			if command := update.Message.Command(); command != "" {
				if command == "pass" {
					fmt.Println("pass")
					msg := tgbotapi.NewMessage(chatID, elapse.Combine())
					bot.Send(msg)

				}

				if command == "d" {
					start, end, _ := HandleDetai(update.Message.Text)
					str, d, a := db.Detail(start, end)

					b := img.Image(d, a)

					msg := tgbotapi.NewMessage(chatID, str)
					photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileBytes{Name: "data", Bytes: b})
					bot.Send(msg)
					bot.Send(photo)

				}

				if command == "group" {

					list, err := db.GetTaskGroup()
					if err != nil {
						fmt.Fprintln(os.Stderr, err)
						continue
					}

					var values []chart.Value
					var str string

					for _, v := range list {
						task := fmt.Sprintf("%s (%.2f)", v.Task, v.TotalHours)
						values = append(values, chart.Value{Value: v.TotalHours, Label: task})
						s := fmt.Sprintf("task:%s	start:%s, end:%s, totalH: %.2f\n", v.Task, v.MinStartTime.Format("2006-01-02"), v.MaxEndTime.Format("2006-01-02"), v.TotalHours)
						str += s
					}

					b, err := img.Pip(values)
					if err != nil {
						fmt.Fprintln(os.Stderr, err)
						continue
					}

					msg := tgbotapi.NewMessage(chatID, str)
					bot.Send(msg)
					photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileBytes{Name: "data", Bytes: b})
					bot.Send(photo)

				}

				continue
			}

		}
	}

}

func HandleDetai(input string) (time.Time, time.Time, int) {

	strList := strings.Fields(input)

	fmt.Println(strList, "xx")

	var text string
	if len(strList) == 1 {
		text = "1"
	}

	if len(strList) >= 2 {
		text = strList[1]
	}

	fmt.Println(text, "go")

	// extract start and end times from the input
	re := regexp.MustCompile(`\d+\.\d+-\d+\.\d+`)
	match := re.FindString(text)
	fmt.Println(match, "xx")

	loc, _ := time.LoadLocation("Asia/Shanghai")
	current := time.Now().In(loc)

	var days = 1
	start := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, loc).UTC()
	end := start.Add(24 * time.Hour).Add(-1 * time.Second)

	if match != "" {
		times := regexp.MustCompile(`\d+\.\d+`).FindAllString(match, -1)
		fmt.Println(times)
		start, _ = time.Parse("1.2", times[0])
		end, _ = time.Parse("1.2", times[1])
		start = time.Date(current.Year(), start.Month(), start.Day(), 0, 0, 0, 0, loc).UTC()
		end = time.Date(current.Year(), end.Month(), end.Day(), 23, 59, 59, 0, loc).UTC()
		days = int(end.Sub(start).Hours()/24) + 1

	} else {
		re = regexp.MustCompile(`\d+`)
		match = re.FindString(text)
		if match != "" {
			n, err := fmt.Sscan(match, &days)
			fmt.Println(n, err)
			if days <= 0 {
				days = 1
			}
			start = start.Add(-time.Duration(days-1) * 24 * time.Hour)

		}

	}
	return start, end, days
}
