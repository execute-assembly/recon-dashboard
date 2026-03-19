package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func Run() {
	http.HandleFunc("/api/hosts", HostHandler)
	http.HandleFunc("/api/hits", JuicyHandler)

	http.HandleFunc("/api/targets/new", NewTargetHandler)
	http.HandleFunc("/api/targets", TargetHandler)

	http.HandleFunc("/index.html", serveHTML("static/index.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveHTML("static/target.html")(w, r)
	})

	fmt.Println("[+] Server running on http://127.0.0.1:8080")
	http.ListenAndServe(":8080", nil)
}

// serveHTML reads a file and writes it directly — avoids http.ServeFile's
// redirect behaviour when the request path ends in /.
func serveHTML(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		content, err := os.ReadFile(path)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(content)
	}
}

func HostHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("HOST"))
}

func JuicyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("JUICY"))
}

func TargetHandler(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir("./databases")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string][]string{"targets": []string{}})
		return
	}

	var targets []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_db.sql") {
			domain := strings.TrimSuffix(entry.Name(), "_db.sql")
			targets = append(targets, domain)
		}
	}

	if targets == nil {
		targets = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"targets": targets})
}


struct NewTargetJson {
	Domain string `json:"domain"`
}

func NewTargetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return 	http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
	}

	var domain NewTargetJson
	err := json.NewDecoder(r.Body).Decode(&domain)
	if err != nil {
		http.Error(w, "domain cannot be empty", http.StatusBadRequest) 
		return
	}

	defer r.Body.Close()

	fmt.Println(domain.Domain)

	

}