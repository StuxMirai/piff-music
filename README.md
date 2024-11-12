# Piffle's Music Overlay

A lightweight overlay system that displays YouTube Music data on your stream without requiring third-party service access. Perfect for streamers who want to show their current playing music to viewers.

## Features

- Real-time song information display
- Clean, customizable overlay for OBS
- Progress bar with current timestamp
- Lightweight Firefox extension for YouTube Music
- No third-party services required

## How It Works

1. **Browser Extension**: A Firefox extension monitors YouTube Music and captures the currently playing song information.
2. **Server**: A Go-based server receives the song data through a webhook endpoint and manages the current state.
3. **OBS Overlay**: A browser source in OBS connects to the server and displays the current song information with a sleek interface.

## Setup

### 1. Server Setup
1. Install Go if you haven't already
2. Clone this repository
3. Run the server:

```bash
go run main.go
```

The server will start on `http://localhost:8080`

### 2. Firefox Extension Setup
1. Open Firefox and go to `about:debugging`
2. Click "This Firefox" in the left sidebar
3. Click "Load Temporary Add-on"
4. Navigate to the `piffmusic` folder and select `manifest.json`

### 3. OBS Setup
1. Add a new Browser Source to your scene
2. Set the URL to `http://localhost:8080`
3. Recommended settings:
   - Width: 800
   - Height: 200
   - Custom CSS: None required (styling is handled by the overlay)

## Development

### Testing
You can use the mock client to test the overlay without YouTube Music:

```bash
cd mock
go run mock.go
```

### Project Structure
- `main.go` - Main server implementation
- `piffmusic/` - Firefox extension files
  - `manifest.json` - Extension configuration
  - `content.js` - YouTube Music scraping logic
- `mock/` - Testing utilities

## Technical Details

The project consists of three main components:

1. **Firefox Extension**: Scrapes YouTube Music data every second and sends it to the local server
2. **Go Server**: Handles webhook endpoints and serves the overlay webpage
3. **HTML/JS Overlay**: Updates in real-time with current song information

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License

## Acknowledgments

- Thanks to YouTube Music for the source data
- Icon by [attribution if needed]
```

This README references the following code blocks:

```159:166:main.go
func main() {
	http.HandleFunc("/webhook", webhookHandler)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/now-playing", nowPlayingHandler)

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```


```1:16:piffmusic/content.js
function getNowPlaying() {
    let title = document.querySelector('yt-formatted-string.title.style-scope.ytmusic-player-bar')?.textContent?.trim(); //element.class

    let artist = document.querySelector('yt-formatted-string.byline.style-scope.ytmusic-player-bar.complex-string')?.textContent?.trim();

    let progress = document.querySelector('tp-yt-paper-slider#progress-bar');
    let timeText = progress.getAttribute('aria-valuetext');
    const startTime = timeText.split(' of ')[0];
    const endTime = timeText.split(' of ')[1];
    return {
        song_name: title,
        artist: artist,
        current_timestamp: startTime,
        end_timestamp: endTime
    };
}
```


```1:28:piffmusic/manifest.json
{
  "manifest_version": 2,
  "name": "PiffMusic",
  "version": "0.1",
  "description": "Scrapes youtube music for now playing information and sends to local webhook.",
  "icons": {
    "48": "icons/stuxpup.png",
    "96": "icons/stuxpup.png"
  },
  "content_scripts": [
    {
      "matches": ["*://music.youtube.com/*"],
      "js": ["content.js"]
    }
  ],

  "permissions": [
    "activeTab",
    "<all_urls>",
    "webRequest"
  ],
  "browser_specific_settings": {   
    "gecko": {
      "id": "stux@stux.ai",
      "strict_min_version": "58.0"
    }
  }
}
```


The README provides a comprehensive overview of the project while maintaining a clean, professional format. It includes all necessary setup instructions and technical details based on the codebase provided.
```

</rewritten_file>