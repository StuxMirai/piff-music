package main

import (
	"encoding/json"
	"fmt"
	"html/template"
    "io"
	"log"
	"net/http"
    "net/url"
    "strings"
    "time"
	"sync"
)

type NowPlaying struct {
	SongName         string `json:"song_name"`
	Artist           string `json:"artist"`
	CurrentTimestamp string `json:"current_timestamp"`
	EndTimestamp     string `json:"end_timestamp"`
    AlbumArtURL      string `json:"album_art_url"`
    AlbumArtVersion  int    `json:"album_art_version,omitempty"`
    CurrentSeconds   int    `json:"current_seconds,omitempty"`
    EndSeconds       int    `json:"end_seconds,omitempty"`
}

var (
	currentTrack NowPlaying
	mu           sync.RWMutex

    currentArtURL         string
    currentArtBytes       []byte
    currentArtContentType string
    currentArtVersion     int
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
            font-family: system-ui, -apple-system, Segoe UI, Roboto, Helvetica, Arial, sans-serif;
            height: 100%;
            overflow: hidden;
            background: #0b0b0b;
        }
        .container {
            width: 100%;
            height: 100%;
            display: flex;
            justify-content: center;
            align-items: center;
        }
        .now-playing {
            position: relative;
            background-color: #000;
            border-radius: 18px;
            padding: 0;
            width: 90%;
            max-width: 800px;
            overflow: hidden;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.6);
            border: 1px solid rgba(255,255,255,0.08);
        }
        .now-playing::before {
            content: "";
            position: absolute;
            inset: 0;
            background-image: var(--album-url, none);
            background-size: cover;
            background-position: center;
            background-repeat: no-repeat;
            filter: blur(6px) brightness(0.82) saturate(1.15);
            transform: scale(1.1);
            z-index: 0;
        }
        .now-playing::after {
            content: "";
            position: absolute;
            inset: 0;
            background:
                linear-gradient(to top, rgba(0,0,0,0.4), transparent 35%),
                linear-gradient(to bottom, rgba(0,0,0,0.4), transparent 35%),
                linear-gradient(to left, rgba(0,0,0,0.4), transparent 35%),
                linear-gradient(to right, rgba(0,0,0,0.4), transparent 35%);
            z-index: 1;
            pointer-events: none;
        }
        .content {
            position: relative;
            z-index: 2;
            padding: 28px 28px 22px;
            display: flex;
            flex-direction: column;
            gap: 10px;
            background: none;
            backdrop-filter: none;
        }
        .song-name {
            font-size: 3em;
            font-weight: bold;
            margin: 0;
            color: white;
            text-shadow: 
                -1px -1px 0 #000,
                1px -1px 0 #000,
                -1px 1px 0 #000,
                1px 1px 0 #000;
            overflow: hidden;
            white-space: nowrap;
            will-change: transform;
        }
        .artist-name {
            font-size: 1.25em;
            color: white;
            margin: 10px 0;
            text-shadow: 
                -1px -1px 0 #000,
                1px -1px 0 #000,
                -1px 1px 0 #000,
                1px 1px 0 #000;
            opacity: 0.95;
            overflow: hidden;
            white-space: nowrap;
            will-change: transform;
        }
        .progress-bar {
            width: 100%;
            height: 12px;
            background-color: rgba(255, 255, 255, 0.35);
            border-radius: 5px;
            overflow: hidden;
            margin: 15px 0;
            border: 1px solid rgba(255,255,255,0.25);
            box-shadow: inset 0 2px 8px rgba(0,0,0,0.4);
        }
        .progress {
            width: 0%;
            height: 100%;
            background: linear-gradient(90deg, #9b59b6, #8e44ad, #6c5ce7, #9b59b6);
            background-size: 200% 100%;
            transition: width 0.5s ease-in-out;
            animation: progressFlow 10s linear infinite;
        }
        @keyframes progressFlow {
            0% { background-position: 0% 0; }
            100% { background-position: 200% 0; }
        }
        .timestamp {
            font-size: 0.9em;
            color: white;
            text-shadow: 
                -1px -1px 0 #000,
                1px -1px 0 #000,
                -1px 1px 0 #000,
                1px 1px 0 #000;
            opacity: 0.9;
            align-self: flex-end;
        }
        
        .marquee {
            animation: marquee 12s linear infinite;
        }
        @keyframes marquee {
            0% { transform: translateX(0); }
            100% { transform: translateX(-100%); }
        }
        .song-name, .artist-name { position: relative; z-index: 3; }

        @media (prefers-reduced-motion: reduce) {
            .marquee { animation: none; }
            .progress { animation: none; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="now-playing">
            <div class="content">
                <h1 class="song-name" id="songName">Waiting for track...</h1>
                <p class="artist-name" id="artistName">Unknown Artist</p>
                <div class="progress-bar">
                    <div class="progress" id="progressBar"></div>
                </div>
                <p class="timestamp" id="timestamp"></p>
            </div>
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
                        updateProgressBar(data.current_timestamp, data.end_timestamp, data.current_seconds, data.end_seconds);
                        updateBackground(data.album_art_url, data.album_art_version);
                        updateProgressThemeFromAlbumArt(data.album_art_version);
                        applyMarqueeIfOverflow('songName');
                        applyMarqueeIfOverflow('artistName');
                    } else {
                        document.getElementById('songName').textContent = 'Waiting for track...';
                        document.getElementById('artistName').textContent = 'Unknown Artist';
                        document.getElementById('timestamp').textContent = '';
                        document.getElementById('progressBar').style.width = '0%';
                        updateBackground(null);
                        removeMarquee('songName');
                        removeMarquee('artistName');
                    }
                })
                .catch(error => console.error('Error:', error));
        }

        function updateProgressBar(current, end, currentSeconds, endSeconds) {
            const currentTime = (Number.isFinite(currentSeconds) && currentSeconds >= 0) ? currentSeconds : timeToSeconds(current);
            const endTime = (Number.isFinite(endSeconds) && endSeconds >= 0) ? endSeconds : timeToSeconds(end);
            const progress = endTime > 0 ? (currentTime / endTime) * 100 : 0;
            document.getElementById('progressBar').style.width = Math.min(100, Math.max(0, progress)) + '%';
        }

        function timeToSeconds(timeString) {
            if (!timeString || typeof timeString !== 'string' || !timeString.includes(':')) return 0;
            const parts = timeString.split(':').map(Number);
            if (parts.length !== 2 || isNaN(parts[0]) || isNaN(parts[1])) return 0;
            const [minutes, seconds] = parts;
            return minutes * 60 + seconds;
        }

        function updateBackground(albumUrl, version) {
            const container = document.querySelector('.now-playing');
            if (albumUrl) {
                const v = (Number.isFinite(version) ? version : 0);
                const localUrl = '/album-art?v=' + v;
                container.style.setProperty('--album-url', 'url(' + "'" + localUrl + "'" + ')');
            } else {
                container.style.setProperty('--album-url', 'none');
            }
        }

        let lastPaletteVersion = -1;
        function updateProgressThemeFromAlbumArt(version) {
            if (!Number.isFinite(version) || version === lastPaletteVersion) return;
            lastPaletteVersion = version;
            const imgUrl = '/album-art?v=' + version;
            const img = new Image();
            img.crossOrigin = 'anonymous';
            img.onload = function () {
                try {
                    const canvas = document.createElement('canvas');
                    const ctx = canvas.getContext('2d');
                    const target = 64;
                    canvas.width = target;
                    canvas.height = target;
                    ctx.drawImage(img, 0, 0, target, target);
                    const { data } = ctx.getImageData(0, 0, target, target);
                    const palette = extractPalette(data);
                    applyProgressGradient(palette);
                } catch (e) {
                    // ignore
                }
            };
            img.src = imgUrl;
        }

        function applyProgressGradient(palette) {
            const el = document.getElementById('progressBar');
            if (!el) return;
            const { base, accent1, accent2 } = palette;
            el.style.background = 'linear-gradient(90deg, ' + base + ', ' + accent1 + ', ' + accent2 + ')';
        }

        function extractPalette(bytes) {
            // Build hue histogram for saturated pixels and compute base color
            const bins = new Array(12).fill(0);
            const hAcc = new Array(12).fill(0);
            const sAcc = new Array(12).fill(0);
            const lAcc = new Array(12).fill(0);
            let avgR = 0, avgG = 0, avgB = 0, count = 0;
            for (let i = 0; i < bytes.length; i += 4) {
                const r = bytes[i] / 255, g = bytes[i+1] / 255, b = bytes[i+2] / 255;
                const a = bytes[i+3] / 255;
                if (a < 0.5) continue;
                avgR += r; avgG += g; avgB += b; count++;
                const hsl = rgbToHsl(r, g, b);
                const h = hsl[0], s = hsl[1], l = hsl[2];
                if (s > 0.4 && l > 0.2 && l < 0.8) {
                    const bin = Math.floor((h * 360) / 30) % 12;
                    bins[bin]++;
                    hAcc[bin] += h; sAcc[bin] += s; lAcc[bin] += l;
                }
            }
            let baseRgb;
            if (count > 0) {
                const avgColor = [avgR / count, avgG / count, avgB / count];
                // pick dominant hue bin if available
                let maxIdx = -1, maxVal = 0;
                for (let i = 0; i < 12; i++) if (bins[i] > maxVal) { maxVal = bins[i]; maxIdx = i; }
                if (maxVal > 0) {
                    const h = (hAcc[maxIdx] / maxVal);
                    const s = Math.min(1, (sAcc[maxIdx] / maxVal) * 1.05);
                    const l = lAcc[maxIdx] / maxVal;
                    baseRgb = hslToRgb(h, s, l);
                } else {
                    baseRgb = avgColor;
                }
            } else {
                baseRgb = [0.6, 0.4, 0.8];
            }
            // Ensure minimum brightness via HSL lightness clamps
            const baseHsl = rgbToHsl(baseRgb[0], baseRgb[1], baseRgb[2]);
            const baseHslAdj = [ baseHsl[0], clamp01(baseHsl[1] * 1.05), clamp01(Math.max(0.50, baseHsl[2])) ];
            const acc1Hsl = [ rotateHue(baseHslAdj[0], 20/360), clamp01(baseHslAdj[1] * 1.05), clamp01(Math.max(0.56, baseHslAdj[2])) ];
            const acc2Hsl = [ rotateHue(baseHslAdj[0], -20/360), clamp01(baseHslAdj[1] * 0.95), clamp01(Math.max(0.48, baseHslAdj[2] * 0.95)) ];
            const baseCss = rgbTupleToCss(hslToRgb(baseHslAdj[0], baseHslAdj[1], baseHslAdj[2]));
            const acc1Css = rgbTupleToCss(hslToRgb(acc1Hsl[0], acc1Hsl[1], acc1Hsl[2]));
            const acc2Css = rgbTupleToCss(hslToRgb(acc2Hsl[0], acc2Hsl[1], acc2Hsl[2]));
            return { base: baseCss, accent1: acc1Css, accent2: acc2Css };
        }

        function rotateHue(h, delta) {
            let x = h + delta; while (x < 0) x += 1; while (x >= 1) x -= 1; return x;
        }
        function clamp01(x) { return Math.max(0, Math.min(1, x)); }
        function rgbTupleToCss(rgb) {
            const r = Math.round(rgb[0] * 255), g = Math.round(rgb[1] * 255), b = Math.round(rgb[2] * 255);
            return 'rgb(' + r + ', ' + g + ', ' + b + ')';
        }
        function rgbToHsl(r, g, b) {
            const max = Math.max(r, g, b), min = Math.min(r, g, b);
            let h, s, l = (max + min) / 2;
            if (max === min) { h = s = 0; }
            else {
                const d = max - min;
                s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
                switch (max) {
                    case r: h = (g - b) / d + (g < b ? 6 : 0); break;
                    case g: h = (b - r) / d + 2; break;
                    case b: h = (r - g) / d + 4; break;
                }
                h /= 6;
            }
            return [h, s, l];
        }
        function hslToRgb(h, s, l) {
            let r, g, b;
            if (s === 0) { r = g = b = l; }
            else {
                const q = l < 0.5 ? l * (1 + s) : l + s - l * s;
                const p = 2 * l - q;
                const hk = h;
                const t = [hk + 1/3, hk, hk - 1/3];
                const out = [0,0,0];
                for (let i = 0; i < 3; i++) {
                    let tc = t[i];
                    if (tc < 0) tc += 1; if (tc > 1) tc -= 1;
                    if (tc < 1/6) out[i] = p + (q - p) * 6 * tc;
                    else if (tc < 1/2) out[i] = q;
                    else if (tc < 2/3) out[i] = p + (q - p) * (2/3 - tc) * 6;
                    else out[i] = p;
                }
                r = out[0]; g = out[1]; b = out[2];
            }
            return [r, g, b];
        }

        function applyMarqueeIfOverflow(elementId) {
            const el = document.getElementById(elementId);
            if (!el) return;
            // Force layout before measuring
            el.offsetWidth;
            if (el.scrollWidth > el.clientWidth) {
                el.classList.add('marquee');
            } else {
                el.classList.remove('marquee');
            }
        }

        function removeMarquee(elementId) {
            const el = document.getElementById(elementId);
            if (!el) return;
            el.classList.remove('marquee');
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
    http.HandleFunc("/album-art", albumArtHandler)

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
    // CORS preflight
    if r.Method == http.MethodOptions {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        w.WriteHeader(http.StatusNoContent)
        return
    }
    if r.Method != http.MethodPost {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

	var newTrack NowPlaying
	err := json.NewDecoder(r.Body).Decode(&newTrack)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

    if newTrack.SongName != "" && newTrack.Artist != "" {
        mu.Lock()
        urlChanged := newTrack.AlbumArtURL != "" && newTrack.AlbumArtURL != currentArtURL
        currentTrack = newTrack
        mu.Unlock()
        if urlChanged {
            go fetchAndCacheAlbumArt(newTrack.AlbumArtURL)
        }
    }

    w.Header().Set("Access-Control-Allow-Origin", "*")
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
    // Include current art version so client can bust cache
    out := currentTrack
    out.AlbumArtVersion = currentArtVersion
    json.NewEncoder(w).Encode(out)
}

func albumArtHandler(w http.ResponseWriter, r *http.Request) {
    mu.RLock()
    bytes := currentArtBytes
    ctype := currentArtContentType
    mu.RUnlock()

    if len(bytes) == 0 {
        http.NotFound(w, r)
        return
    }
    if ctype == "" {
        ctype = "image/jpeg"
    }
    w.Header().Set("Content-Type", ctype)
    // Ensure the browser refetches when version changes via query param
    w.Header().Set("Cache-Control", "no-store, must-revalidate")
    w.Write(bytes)
}

func fetchAndCacheAlbumArt(src string) {
    normalized := normalizeGoogleImageSize(src)
    // Try up to 3 sizes: preferred -> 800 -> 544
    candidates := []string{normalized, replaceSize(normalized, 800), replaceSize(normalized, 544)}

    client := &http.Client{Timeout: 10 * time.Second}

    for _, u := range candidates {
        req, err := http.NewRequest("GET", u, nil)
        if err != nil {
            continue
        }
        // Spoof headers to match browser context
        req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:141.0) Gecko/20100101 Firefox/141.0")
        req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/*,*/*;q=0.8")
        req.Header.Set("Referer", "https://music.youtube.com/")

        resp, err := client.Do(req)
        if err != nil {
            continue
        }
        func() {
            defer resp.Body.Close()
            if resp.StatusCode != http.StatusOK {
                return
            }
            data, err := io.ReadAll(resp.Body)
            if err != nil || len(data) == 0 {
                return
            }
            mu.Lock()
            currentArtURL = src
            currentArtBytes = data
            currentArtContentType = resp.Header.Get("Content-Type")
            currentArtVersion++
            mu.Unlock()
        }()
        // If we were successful, break
        mu.RLock()
        ok := len(currentArtBytes) > 0
        mu.RUnlock()
        if ok {
            return
        }
    }
}

func normalizeGoogleImageSize(raw string) string {
    // If it's not a googleusercontent image, return as-is
    u, err := url.Parse(raw)
    if err != nil {
        return raw
    }
    if !strings.Contains(u.Host, "googleusercontent.com") {
        return raw
    }
    return replaceSize(raw, 800)
}

func replaceSize(raw string, size int) string {
    // Replace w###-h### with desired size
    if size <= 0 {
        size = 544
    }
    s := raw
    // common pattern
    s = sizeRegexReplace(s, size)
    return s
}

func sizeRegexReplace(s string, size int) string {
    // lightweight replace without regex package by splitting on '=' params
    // If pattern w###-h### exists after '=', try to patch it
    parts := strings.Split(s, "=")
    if len(parts) == 2 {
        suffix := parts[1]
        // find existing size token
        tokens := strings.Split(suffix, "-")
        // overwrite first two tokens if they start with w or h
        if len(tokens) >= 2 && strings.HasPrefix(tokens[0], "w") && strings.HasPrefix(tokens[1], "h") {
            tokens[0] = fmt.Sprintf("w%d", size)
            tokens[1] = fmt.Sprintf("h%d", size)
            parts[1] = strings.Join(tokens, "-")
            return strings.Join(parts, "=")
        }
    }
    return s
}
