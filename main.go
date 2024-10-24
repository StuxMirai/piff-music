package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
)

type NowPlaying struct {
	SongName         string `json:"song_name"`
	Artist           string `json:"artist"`
	CurrentTimestamp string `json:"current_timestamp"`
	EndTimestamp     string `json:"end_timestamp"`
}

var (
	currentTrack NowPlaying
	mu           sync.RWMutex
)

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Now Playing</title>
    <style>
        body, html {
            margin: 0;
            padding: 0;
            font-family: Arial, sans-serif;
            height: 100%;
            overflow: hidden;
        }
        .container {
            width: 100%;
            height: 100%;
            display: flex;
            justify-content: center;
            align-items: center;
        }
        .now-playing {
            background-color: rgba(230, 230, 250, 0.7);
            border-radius: 15px;
            padding: 20px;
            width: 80%;
            max-width: 600px;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        }
        .song-name {
            font-size: 2.5em;
            font-weight: bold;
            margin: 0;
            color: white;
            text-shadow: 
                -1px -1px 0 #000,
                1px -1px 0 #000,
                -1px 1px 0 #000,
                1px 1px 0 #000;
        }
        .artist-name {
            font-size: 1.5em;
            color: white;
            margin: 10px 0;
            text-shadow: 
                -1px -1px 0 #000,
                1px -1px 0 #000,
                -1px 1px 0 #000,
                1px 1px 0 #000;
        }
        .progress-bar {
            width: 100%;
            height: 10px;
            background-color: rgba(255, 255, 255, 0.5);
            border-radius: 5px;
            overflow: hidden;
            margin: 15px 0;
        }
        .progress {
            width: 0%;
            height: 100%;
            background-color: #4B0082;
            transition: width 0.5s ease-in-out;
        }
        .timestamp {
            font-size: 0.9em;
            color: white;
            text-shadow: 
                -1px -1px 0 #000,
                1px -1px 0 #000,
                -1px 1px 0 #000,
                1px 1px 0 #000;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="now-playing">
            <h1 class="song-name" id="songName">Waiting for track...</h1>
            <p class="artist-name" id="artistName">Unknown Artist</p>
            <div class="progress-bar">
                <div class="progress" id="progressBar"></div>
            </div>
            <p class="timestamp" id="timestamp"></p>
        </div>
    </div>

    <script>
        function updateNowPlaying() {
            fetch('/now-playing')
                .then(response => response.json())
                .then(data => {
                    if (data.song_name) {
                        document.getElementById('songName').textContent = data.song_name;
                        document.getElementById('artistName').textContent = data.artist;
                        document.getElementById('timestamp').textContent = data.current_timestamp + ' / ' + data.end_timestamp;
                        updateProgressBar(data.current_timestamp, data.end_timestamp);
                    } else {
                        document.getElementById('songName').textContent = 'Waiting for track...';
                        document.getElementById('artistName').textContent = 'Unknown Artist';
                        document.getElementById('timestamp').textContent = '';
                        document.getElementById('progressBar').style.width = '0%';
                    }
                })
                .catch(error => console.error('Error:', error));
        }

        function updateProgressBar(current, end) {
            const currentTime = timeToSeconds(current);
            const endTime = timeToSeconds(end);
            const progress = (currentTime / endTime) * 100;
            document.getElementById('progressBar').style.width = progress + '%';
        }

        function timeToSeconds(timeString) {
            const [minutes, seconds] = timeString.split(':').map(Number);
            return minutes * 60 + seconds;
        }

        updateNowPlaying();
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
	mu.RLock()
	defer mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentTrack)
}
