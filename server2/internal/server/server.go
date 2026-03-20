package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"github.com/go-chi/chi/v5"
	"github.com/execute-assembly/recon-dashboard/internal/database"
)

type NewTargetJson struct {
	Domain string `json:"domain"`
}

type NoteStruct struct {
	Domain string `json:"domain"`
	Note   string `json:"notes"`
}


//  {domain} -> target level e.g domain.com
//  {hostURL} -> host level, e.g https://domain.com:443 -> routes need url decoding!

func Run() {
	r := chi.NewRouter()

	r.Get("/api/{domain}/hosts", HostHandler)
	r.Get("/api/{domain}/hits", JuicyHandler)

	r.Patch("/api/{domain}/host/{hostURL}/triage", TriageHandler)
	r.Patch("/api/{domain}/host/{hostURL}/notes", NotesHandler)
	// r.Post( "/api/{domain}/host/{hostURL}/screenshot", ScreenShotHandler)
	// r.Post( "/api/{domain}/host/{hostURL}/portscan", PortScanHandler)

	r.Post("/api/import/{domain}", ImportHandler)
	//r.Post("/api/delete/{domain}", deleteTargetHandler)

	


	r.Post("/api/targets/new", NewTargetHandler)
	r.Get("/api/targets", TargetHandler)


	r.Get("/index.html", serveHTML("static/index.html"))
	r.Get("/*", serveHTML("static/target.html"))

	// Middleware wrapper serves /css/* and /js/* before Chi route matching.
	fs := http.FileServer(http.Dir("static"))
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		p := req.URL.Path
		if strings.HasPrefix(p, "/css/") || strings.HasPrefix(p, "/js/") || strings.HasPrefix(p, "/images/") || strings.HasSuffix(p, ".png") || strings.HasSuffix(p, ".jpg") || strings.HasSuffix(p, ".jpeg") || strings.HasSuffix(p, ".gif") || strings.HasSuffix(p, ".webp") {
			fs.ServeHTTP(w, req)
			return
		}
		r.ServeHTTP(w, req)
	})

	fmt.Println("[+] Server running on http://127.0.0.1:8080")
	http.ListenAndServe(":8080", handler)
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


type TriageData struct {
	Domain string `json:"domain"`
	Status string `json:"status"`
}

func TriageHandler(w http.ResponseWriter, r *http.Request) {
	domain := chi.URLParam(r, "domain")
	hostURL, _ := url.QueryUnescape(chi.URLParam(r, "hostURL"))

// 	Triage sends:
  // - PATCH /api/domains/{hostURL}/triage
  // - Body: { domain: "<target>", status: "<none|to-test|dead-end|tested>" }
  // - Expects: any 2xx, doesn't read the body at all


	var data TriageData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"status": "Failed to decode json"})
		return
	}

	err := database.UpdateTriage(domain, hostURL, data.Status)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"status": "failed to insert"})
		return
	}


	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "Status updated!"})
	return 



}


func NotesHandler(w http.ResponseWriter, r *http.Request) {
	domain := chi.URLParam(r, "domain")
	var data NoteStruct
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"status": "Failed to decode json"})
		return
	}

	hostURL, _ := url.QueryUnescape(chi.URLParam(r, "hostURL"))

	fmt.Printf("[+] data.Domain: %s\n[+] hostURL: %s\n[+] data.Note: %s\n", domain, hostURL, data.Note)
	err := database.WriteNote(data.Domain, hostURL, data.Note)
	if err != nil {
		fmt.Println(err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"status": "failed to insert"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "Note added!"})
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