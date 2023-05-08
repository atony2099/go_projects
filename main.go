package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/atony2099/time_manager/config"
	"github.com/atony2099/time_manager/db"
	"github.com/atony2099/time_manager/router"
)

func main() {

	// test
	config.LoadEnv()
	cfg := config.GetConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db.Open(cfg.DatabaseDSN)
	// go bot.RunBot(cfg.TelegramToken, cfg.ChatID, ctx)
	// go scheduler.StartScheduler(ctx)

	server := startSever(cfg.Port)

	waitForSignal()
	cancel() //stop bot,scheduler
	db.Close()
	closeServer(server, ctx)

	log.Print("Server exiting gracefully")

}

func waitForSignal() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-stop
}

func closeServer(srv *http.Server, ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 3)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown Failded:", err)
	}
}
func startSever(port string) *http.Server {
	srv := &http.Server{Addr: fmt.Sprintf(":%s", port), Handler: router.SetupRouter()}
	go func() {
		log.Printf("Server started on port %s", port)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()
	return srv
}
