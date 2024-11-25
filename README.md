# Youtube Music Now Playing Display

This tool allows you to display your currently playing music from Youtube Music in your OBS stream. It consists of a browser extension and a local server that work together to show real-time "Now Playing" information.

## Setup Instructions

### Step 1: Install the Browser Extension

1. Visit the [PiffMusic Firefox Add-on](https://addons.mozilla.org/en-US/firefox/addon/piffmusic/) page
2. Click "Add to Firefox" to install the extension
3. After installation, you should see the Piff Music icon in your browser toolbar

### Step 2: Set Up the Local Server

1. Download and install [Go](https://golang.org/dl/) if you haven't already
2. Click the green "Code" button above and select "Download ZIP"
3. Extract the downloaded ZIP file to a location of your choice
4. Open a terminal/command prompt
5. Navigate to the extracted folder using the `cd` command, for example:
   ```bash
   cd C:\Users\YourName\Downloads\PiffMusic-main
   ```
6. Run the following command to compile the server:
   ```bash
   go build -o PiffMusic.exe main.go
   ```
7. Copy or move PiffMusic.exe to your Desktop
8. Double-click PiffMusic.exe on your Desktop to start the server
9. You should see the message "Server is running on http://localhost:8080"

### Step 3: Add to OBS

1. Open OBS Studio
2. In your desired scene, add a new "Browser" source
3. Configure the Browser source with these settings:
   - URL: `http://localhost:8080`
   - Width: 1920 (recommended, adjust as needed)
   - Height: 1080 (recommended, adjust as needed)
4. Click "OK" to add the source

## Usage

1. Make sure the local server is running (`PiffMusic.exe`)
2. Play music on Youtube Music in Firefox
3. The "Now Playing" information will automatically update in your OBS scene

## Troubleshooting

- If no track information appears, ensure:
  - The Firefox extension is installed and active
  - The local server is running
  - You're playing music on Youtube Music
  - The browser source in OBS is properly configured
  - Only use one tab of Youtube Music at a time