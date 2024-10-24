package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"
)

type NowPlaying struct {
	SongName  string `json:"song_name"`
	Artist    string `json:"artist"`
	Timestamp string `json:"timestamp"`
}

var (
	currentTrack NowPlaying
	mu           sync.RWMutex
)

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Now Playing</title>
    <style>
        body { font-family: Arial, sans-serif; background-color: #333; color: #fff; }
        #now-playing { padding: 10px; }
    </style>
</head>
<body>
    <div id="now-playing">Waiting for track information...</div>
    <script>
        function updateNowPlaying() {
            fetch('/now-playing')
                .then(response => response.text())
                .then(data => {
                    document.getElementById('now-playing').innerHTML = data;
                })
                .catch(error => console.error('Error:', error));
        }
        setInterval(updateNowPlaying, 1000);
    </script>
</body>
</html>
`

const nowPlayingTemplate = `
{{if .SongName}}
Now Playing: {{.SongName}} by {{.Artist}} ({{.Timestamp}})
{{else}}
Waiting for track information...
{{end}}
`

func main() {
	http.HandleFunc("/webhook", webhookHandler)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/now-playing", nowPlayingHandler)

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newTrack NowPlaying
	err := json.NewDecoder(r.Body).Decode(&newTrack)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	mu.Lock()
	currentTrack = newTrack
	mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("index").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, nil)
}

func nowPlayingHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(100 * time.Millisecond) // Simulate some processing time

	mu.RLock()
	defer mu.RUnlock()

	if currentTrack.SongName != "" {
		fmt.Fprintf(w, "Now Playing: %s by %s (%s)", currentTrack.SongName, currentTrack.Artist, currentTrack.Timestamp)
	} else {
		fmt.Fprint(w, "Waiting for track information...")
	}
}
