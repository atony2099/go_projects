package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/atony2099/time_manager/bot"
	"github.com/atony2099/time_manager/db"
	"github.com/atony2099/time_manager/elapse"
	"github.com/go-co-op/gocron"
)

var TodayPassRemainInterva = 120
var NoStudyrRemainInterva = 120

func StartScheduler(ctx context.Context) {
	bj, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Fatal(err)
	}
	s := gocron.NewScheduler(bj)
	doPass(s)
	doToady(s)
	s.StartAsync()
	<-ctx.Done()
	s.Stop()
}

func doPass(s *gocron.Scheduler) {
	s.Cron(fmt.Sprintf("*/%d * * * *", TodayPassRemainInterva)).Do(func() {
		beijingLocation, _ := time.LoadLocation("Asia/Shanghai")
		h := time.Now().In(beijingLocation).Hour()
		if h <= 8 {
			return
		}

		pass := elapse.Combine()
		bot.SendBotMsg(pass)
		po := db.Pomodoro()
		bot.SendBotMsg(po)

	})
}

func doToady(s *gocron.Scheduler) {
	s.Cron(fmt.Sprintf("*/%d * * * *", NoStudyrRemainInterva)).Do(func() {
		beijingLocation, _ := time.LoadLocation("Asia/Shanghai")
		h := time.Now().In(beijingLocation).Hour()
		if h <= 9 {
			return
		}
		s := db.TodayLast(20)
		bot.SendBotMsg(s)

	})
}
