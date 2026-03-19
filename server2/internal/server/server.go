package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"github.com/go-chi/chi/v5"
	"github.com/execute-assembly/recon-dashboard/internal/database"
)

type NewTargetJson struct {
	Domain string `json:"domain"`
}

func Run() {
	r := chi.NewRouter()

	r.Get("/api/{domain}/hosts", HostHandler)
	r.Get("/api/{domain}/hits", JuicyHandler)
	r.Post("/api/import/{domain}", ImportHandler)


	r.Post("/api/targets/new", NewTargetHandler)
	r.Get("/api/targets", TargetHandler)


	r.Get("/index.html", serveHTML("static/index.html"))
	r.Get("/*", serveHTML("static/target.html"))

	fmt.Println("[+] Server running on http://127.0.0.1:8080")
	http.ListenAndServe(":8080", r)
}


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
    domain := chi.URLParam(r, "domain")
    data, err := database.ReadHosts(domain)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(data)
}

func JuicyHandler(w http.ResponseWriter, r *http.Request) {
	domain := chi.URLParam(r, "domain")
	data, err := database.ReadHits(domain)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"hits": data})
}


// handles retreving the active targets from /databases/<domain>_db.sql, comes from targets.html
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



// handles the creation of a new target from /target.html
func NewTargetHandler(w http.ResponseWriter, r *http.Request) {

	var domain NewTargetJson
	err := json.NewDecoder(r.Body).Decode(&domain)
	if err != nil {
		http.Error(w, "domain cannot be empty", http.StatusBadRequest) 
		return
	}

	defer r.Body.Close()

	fmt.Println(domain.Domain)

	err = database.CreateNewTarget(domain.Domain)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"domain": domain.Domain})
	return 


}

// Handles importing data for a target, reads json from disk and stores in DB
func ImportHandler(w http.ResponseWriter, r *http.Request) {
	domain := chi.URLParam(r, "domain")

	fmt.Printf("[+] Importing data for %s\n", domain)

	err := database.ImportData(domain)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"domain": "good"})
	return
}