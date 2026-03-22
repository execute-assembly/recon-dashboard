package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/execute-assembly/recon-dashboard/internal/database"
	"github.com/go-chi/chi/v5"
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

	r.Get("/api/{domain}/hosts", Host_Handler)
	r.Get("/api/{domain}/hits", Juicy_Handler)

	r.Patch("/api/{domain}/host/{hostURL}/triage", Triage_Handler)
	r.Patch("/api/{domain}/host/{hostURL}/notes", Notes_Handler)

	//r.Post("/api/{domain}/host/{hostURL}/screenshot", ScreenShot_Handler)
	// r.Get("/api/{domain}/host/{hostURL}/screenshot/status, ScreenShotStatus_Handler)
	// r.Post( "/api/{domain}/host/{hostURL}/portscan", PortScan_Handler)

	r.Post("/api/import/{domain}", ImportHandler)
	r.Delete("/api/delete/{domain}", deleteTargetHandler)

	r.Post("/api/targets/new", NewTargetHandler)
	r.Get("/api/targets", Targets_Handler)

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

func writeJSON(w http.ResponseWriter, status int, msg any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(msg)
}

func Triage_Handler(w http.ResponseWriter, r *http.Request) {
	domain := chi.URLParam(r, "domain")
	hostURL, _ := url.QueryUnescape(chi.URLParam(r, "hostURL"))

	var data TriageData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"status": "Failed to decode json"})
		return
	}

	err := database.UpdateTriage(domain, hostURL, data.Status)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"status": "failed to insert"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "Status updated!"})
	return

}

func Notes_Handler(w http.ResponseWriter, r *http.Request) {
	domain := chi.URLParam(r, "domain")
	hostURL, _ := url.QueryUnescape(chi.URLParam(r, "hostURL"))

	var data NoteStruct
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		slog.Error("Failed To Insert Json in Notes", "hostURL", hostURL)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"status": "Failed to decode json"})
		return
	}

	err := database.WriteNote(domain, hostURL, data.Note)
	if err != nil {
		slog.Error("Failed To Insert Note", "hostURL", hostURL)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"status": "failed to insert"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "Note added!"})
}

func Host_Handler(w http.ResponseWriter, r *http.Request) {
	domain := chi.URLParam(r, "domain")
	data, err := database.ReadHosts(domain)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"status": error.Error(err)})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func Juicy_Handler(w http.ResponseWriter, r *http.Request) {
	domain := chi.URLParam(r, "domain")
	data, err := database.ReadHits(domain)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"status": error.Error(err)})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"hits": data})
}

// handles retreving the active targets from /databases/<domain>_db.sql, comes from targets.html
func Targets_Handler(w http.ResponseWriter, r *http.Request) {
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
		writeJSON(w, http.StatusInternalServerError, map[string]string{"status": error.Error(err)})
		return
	}

	defer r.Body.Close()

	err = database.CreateNewTarget(domain.Domain)

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"status": error.Error(err)})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"domain": domain.Domain})
	return

}

// Handles importing data for a target, reads json from disk and stores in DB
func ImportHandler(w http.ResponseWriter, r *http.Request) {
	domain := chi.URLParam(r, "domain")

	err := database.ImportData(domain)

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"status": error.Error(err)})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"domain": "good"})
	return
}

func deleteTargetHandler(w http.ResponseWriter, r *http.Request) {
	domain := chi.URLParam(r, "domain")

	if err := database.DeleteData(domain); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"status": "Failed Deleting Data."})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "Data Deleted Succesfully!"})

	return
}
