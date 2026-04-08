package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/z3vxo/vantage/internal/server"
	"github.com/z3vxo/vantage/internal/tools"
)

func setupLogs() *os.File {
	home, _ := os.UserHomeDir()
	logDir := home + "/.recon/logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile(logDir+"/recon.log",
		os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(f,
		&slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))

	return f
}

func main() {
	// tools.RunWorkFlow("clovr.dev")
	// return
	f := setupLogs()
	defer f.Close()
	go tools.StartTeleGramBot()
	server.Run()
}
