package tools

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/z3vxo/recon-dashboard/internal/database"
)

func SendTelegram(msg string) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage",
		token)
	http.PostForm(url, map[string][]string{
		"chat_id": {chatID},
		"text":    {msg},
	})
}

func RunWorkFlow(baseDomain string) {
	SendTelegram(fmt.Sprintf("[*] Starting recon — %s", baseDomain))

	cmd := exec.Command("./recon.sh", baseDomain)
	cmd.Dir = ".."
	out, err := cmd.CombinedOutput()
	if err != nil {
		SendTelegram(fmt.Sprintf("[!] Recon failed — %s\n%s", baseDomain, string(out)))
		return
	}

	// create DB, ignore error if it already exists
	if err = database.CreateNewTarget(baseDomain); err != nil && err != database.ErrDomainExists {
		SendTelegram(fmt.Sprintf("[!] Failed creating database — %s", baseDomain))
		return
	}

	if err = database.ImportData(baseDomain); err != nil {
		SendTelegram(fmt.Sprintf("[!] Failed ingesting data — %s", baseDomain))
		return
	}

	stats, err := database.GetStats(baseDomain)
	if err != nil {
		SendTelegram(fmt.Sprintf("[*] Recon done — %s (stats unavailable)", baseDomain))
		return
	}

	msg := fmt.Sprintf(
		"[*] Recon Done — %s\n\n[+] Hosts: %d\n[+] 2xx: %d | 4xx: %d | 5xx: %d\n[+] Endpoint hits: %d",
		baseDomain, stats.Total, stats.S2xx, stats.S4xx, stats.S5xx, stats.Hits,
	)
	SendTelegram(msg)
}
