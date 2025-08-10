# Youtube Music Now Playing Display

This tool displays your currently playing music from YouTube Music in OBS. It consists of:
- A Firefox add-on that scrapes the now-playing info from `music.youtube.com`
- A local Windows executable that hosts a widget at `http://localhost:8080` for OBS

## Install the Firefox Add-on

- Install from AMO: [PiffMusic (Firefox Add-on)](https://addons.mozilla.org/en-US/firefox/addon/piffmusic/). Once installed, it will automatically run on `music.youtube.com`.

## Download and Run the Windows EXE

- Download the latest `piff-music-windows-*.exe` from the [GitHub Releases](https://github.com/StuxMirai/piff-music/releases).
- Double-click the EXE to run it.
- You should see a console window with: `Server is running on http://localhost:8080`.

Notes:
- Keep this EXE running while streaming. Close it to stop the widget.
- The EXE also proxies and caches album art to avoid rate limits.

## Add the Widget to OBS

1. Open OBS Studio
2. Add a new Source → Browser
3. Set URL to `http://localhost:8080`
4. Recommended size: Width 1280–1920, Height 720–1080 (adjust to your scene)
5. Press OK

The widget will show song title, artist, a progress bar, and blurred album art with a subtle edge fade.

## How It Works

- The add-on posts now-playing data to `http://localhost:8080/webhook` once per second (title, artist, time, album art URL)
- The EXE stores the latest payload and serves a live-updating widget at `/`
- Album art is fetched once by the EXE and served locally at `/album-art` for stability

## Development

- Run the server locally (Go):
  ```bash
  go run .
  ```
- Mock sender:
  ```bash
  go run mock/mock.go
  ```
- Load the add-on temporarily for development:
  - Firefox → about:debugging → This Firefox → Load Temporary Add-on → select `piffmusic/manifest.json`

## Releases

- GitHub Actions builds Windows x64/ARM64 EXEs and signs the Firefox add-on.
- Releases are named from `piffmusic/manifest.json` version.