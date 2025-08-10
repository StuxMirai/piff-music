package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type NowPlaying struct {
	SongName         string `json:"song_name"`
	Artist           string `json:"artist"`
	CurrentTimestamp string `json:"current_timestamp"`
	EndTimestamp     string `json:"end_timestamp"`
    AlbumArtURL      string `json:"album_art_url"`
}

var songs = []string{"Bohemian Rhapsody", "Stairway to Heaven", "Imagine", "Smells Like Teen Spirit", "Billie Jean"}
var artists = []string{"Queen", "Led Zeppelin", "John Lennon", "Nirvana", "Michael Jackson"}
var art = []string{
    "https://i.imgur.com/SGP2XjL.jpeg",
}

func main() {
	rand.Seed(time.Now().UnixNano())

	for {
		track := generateRandomTrack()
		sendTrackData(track)
		time.Sleep(1 * time.Second)
	}
}

func generateRandomTrack() NowPlaying {
	songIndex := rand.Intn(len(songs))
	artistIndex := rand.Intn(len(artists))
    artIndex := rand.Intn(len(art))

	totalSeconds := rand.Intn(300) + 60
	currentSeconds := rand.Intn(totalSeconds)

	endTimestamp := fmt.Sprintf("%02d:%02d", totalSeconds/60, totalSeconds%60)
	currentTimestamp := fmt.Sprintf("%02d:%02d", currentSeconds/60, currentSeconds%60)

	return NowPlaying{
		SongName:         songs[songIndex],
		Artist:           artists[artistIndex],
		CurrentTimestamp: currentTimestamp,
        EndTimestamp:     endTimestamp,
        AlbumArtURL:      art[artIndex],
	}
}

func sendTrackData(track NowPlaying) {
	jsonData, err := json.Marshal(track)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	resp, err := http.Post("http://localhost:8080/webhook", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error sending data:", err)
		return
	}
	defer resp.Body.Close()

    if resp.StatusCode == http.StatusOK {
        fmt.Printf("Sent: %s by %s (%s) [art=%s]\n", track.SongName, track.Artist, track.CurrentTimestamp, track.AlbumArtURL)
	} else {
		fmt.Printf("Failed to send data. Status code: %d\n", resp.StatusCode)
	}
}
