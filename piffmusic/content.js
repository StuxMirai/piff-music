function getAlbumArtUrl() {
    try {
        const artwork = navigator.mediaSession && navigator.mediaSession.metadata && navigator.mediaSession.metadata.artwork;
        if (artwork && artwork.length) {
            const sorted = [...artwork].sort((a, b) => (b.sizes?.split('x')[0] || 0) - (a.sizes?.split('x')[0] || 0));
            if (sorted[0] && sorted[0].src) return sorted[0].src;
        }
    } catch (e) { }

    const selectors = [
        'ytmusic-player-page ytmusic-player #song-image yt-img-shadow#thumbnail img#img',
        'ytmusic-player #song-image img#img',
        'ytmusic-player img#img.style-scope.yt-img-shadow',
        'ytmusic-player img#img',
        'ytmusic-player-page img#img',
        'ytmusic-player-bar #song-image img#img',
        'ytmusic-player-bar img#img',
        'img#img.style-scope.yt-img-shadow',
        'img#img'
    ];

    for (const sel of selectors) {
        const el = document.querySelector(sel);
        if (el) {
            const srcset = el.getAttribute('srcset');
            const src = el.getAttribute('src');
            if (srcset) {
                const best = srcset.split(',').map(s => s.trim().split(' ')[0]).filter(Boolean).pop();
                if (best) return best;
            }
            if (src) return src;
        }
    }

    const shadowSelectors = [
        'ytmusic-player-page ytmusic-player #song-image yt-img-shadow#thumbnail',
        'ytmusic-player #song-image yt-img-shadow#thumbnail',
        'ytmusic-player yt-img-shadow#thumbnail',
        'ytmusic-player-page yt-img-shadow#thumbnail',
        'ytmusic-player-bar yt-img-shadow#thumbnail'
    ];
    for (const sel of shadowSelectors) {
        const el = document.querySelector(sel);
        if (!el) continue;
        const styleBg = el.style && el.style.backgroundImage;
        const compBg = !styleBg ? (window.getComputedStyle ? getComputedStyle(el).backgroundImage : '') : styleBg;
        const url = extractUrlFromCss(compBg);
        if (url) return url;
    }
    return null;
}

function extractUrlFromCss(bg) {
    if (!bg || typeof bg !== 'string') return null;
    const match = bg.match(/url\(("|'|)(.*?)\1\)/i);
    return match && match[2] ? match[2] : null;
}

function getNowPlaying() {
    const title = document.querySelector('yt-formatted-string.title.style-scope.ytmusic-player-bar')?.textContent?.trim() || '';
    const artist = document.querySelector('yt-formatted-string.byline.style-scope.ytmusic-player-bar.complex-string')?.textContent?.trim() || '';

    const { startTime, endTime, currentSeconds, endSeconds } = getTimes();

    const albumArtUrl = capAlbumArtResolution(upgradeAlbumArtResolution(getAlbumArtUrl()));

    return {
        song_name: title,
        artist: artist,
        current_timestamp: startTime,
        end_timestamp: endTime,
        album_art_url: albumArtUrl,
        current_seconds: currentSeconds,
        end_seconds: endSeconds,
        progress_pct: computeProgressPct(currentSeconds, endSeconds)
    };
}

function upgradeAlbumArtResolution(url) {
    if (!url || typeof url !== 'string') return url;
    try {
        const u = new URL(url, location.href);
        if (u.hostname.includes('googleusercontent.com')) {
            return url.replace(/w\d+-h\d+/g, 'w1600-h1600');
        }
    } catch (e) {
    }
    return url;
}

function capAlbumArtResolution(url) {
    if (!url || typeof url !== 'string') return url;
    try {
        const u = new URL(url, location.href);
        if (u.hostname.includes('googleusercontent.com')) {
            return url.replace(/w\d+-h\d+/g, 'w800-h800');
        }
    } catch (e) { }
    return url;
}

function postNowPlaying() {
    try {
        const nowPlaying = getNowPlaying();
        const payload = {
            song_name: nowPlaying.song_name,
            artist: nowPlaying.artist,
            current_timestamp: nowPlaying.current_timestamp,
            end_timestamp: nowPlaying.end_timestamp,
            album_art_url: nowPlaying.album_art_url,
            current_seconds: nowPlaying.current_seconds,
            end_seconds: nowPlaying.end_seconds
        };
        fetch("http://localhost:8080/webhook", {
            method: "POST",
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        }).catch(() => {
            try {
                (typeof browser !== 'undefined' ? browser : chrome).runtime.sendMessage({ type: 'postNowPlaying', payload });
            } catch (_) {}
        });
    } catch (e) {
    }
}

setTimeout(() => {
    postNowPlaying();
    setInterval(postNowPlaying, 1000);
}, 1500);

function getTimes() {
    try {
        const pos = navigator.mediaSession && navigator.mediaSession.positionState;
        if (pos && Number.isFinite(pos.position) && Number.isFinite(pos.duration)) {
            const curN = Math.max(0, Math.floor(pos.position));
            const durN = Math.max(0, Math.floor(pos.duration));
            return { startTime: secondsToMMSS(curN), endTime: secondsToMMSS(durN), currentSeconds: curN, endSeconds: durN };
        }
    } catch (_) { }

    const infoLeftControls = document.querySelector('ytmusic-player-bar div#left-controls.left-controls span.time-info.style-scope.ytmusic-player-bar, ytmusic-player-bar #left-controls span.time-info');
    const timesFromInfoLeftControls = readTimesFromInfo(infoLeftControls);
    if (timesFromInfoLeftControls) return timesFromInfoLeftControls;

    const slider = document.querySelector('tp-yt-paper-slider#progress-bar');
    const timesFromSlider = readTimesFromSlider(slider);
    if (timesFromSlider) return timesFromSlider;

    const infoLight = document.querySelector('ytmusic-player-bar span.time-info, ytmusic-player-page span.time-info, span.time-info');
    const timesFromInfoLight = readTimesFromInfo(infoLight);
    if (timesFromInfoLight) return timesFromInfoLight;

    const bar = document.querySelector('ytmusic-player-bar');
    if (bar && bar.shadowRoot) {
        const sliderShadow = bar.shadowRoot.querySelector('tp-yt-paper-slider#progress-bar, tp-yt-paper-slider');
        const timesFromSliderShadow = readTimesFromSlider(sliderShadow);
        if (timesFromSliderShadow) return timesFromSliderShadow;

        const infoShadow = bar.shadowRoot.querySelector('span.time-info');
        const timesFromInfoShadow = readTimesFromInfo(infoShadow);
        if (timesFromInfoShadow) return timesFromInfoShadow;
    }

    const page = document.querySelector('ytmusic-player-page');
    if (page && page.shadowRoot) {
        const infoShadow2 = page.shadowRoot.querySelector('span.time-info');
        const timesFromInfoShadow2 = readTimesFromInfo(infoShadow2);
        if (timesFromInfoShadow2) return timesFromInfoShadow2;
    }
    const media = document.querySelector('video, audio');
    if (media && Number.isFinite(media.duration) && media.duration > 0) {
        const curN = Math.max(0, Math.floor(media.currentTime || 0));
        const durN = Math.max(0, Math.floor(media.duration || 0));
        return { startTime: secondsToMMSS(curN), endTime: secondsToMMSS(durN), currentSeconds: curN, endSeconds: durN };
    }
    return { startTime: '0:00', endTime: '0:00', currentSeconds: 0, endSeconds: 0 };
}

function readTimesFromSlider(slider) {
    if (!slider) return null;
    const text = slider.getAttribute && slider.getAttribute('aria-valuetext');
    if (text && text.includes(' of ')) {
        const [cur, dur] = text.split(' of ').map(s => s.trim());
        const curN = timeToSeconds(cur);
        const durN = timeToSeconds(dur);
        return { startTime: normalizeTime(cur), endTime: normalizeTime(dur), currentSeconds: curN, endSeconds: durN };
    }
    const now = slider.getAttribute && slider.getAttribute('aria-valuenow');
    const max = slider.getAttribute && slider.getAttribute('aria-valuemax');
    if (now && max) {
        const curN = Number(now);
        const durN = Number(max);
        return { startTime: secondsToMMSS(curN), endTime: secondsToMMSS(durN), currentSeconds: curN, endSeconds: durN };
    }
    return null;
}

function readTimesFromInfo(infoEl) {
    if (!infoEl || !infoEl.textContent) return null;
    const text = infoEl.textContent;
    if (text.includes('/')) {
        const [cur, dur] = text.split('/').map(s => s.trim());
        const curN = timeToSeconds(cur);
        const durN = timeToSeconds(dur);
        return { startTime: normalizeTime(cur), endTime: normalizeTime(dur), currentSeconds: curN, endSeconds: durN };
    }
    return null;
}

function secondsToMMSS(totalSeconds) {
    if (!Number.isFinite(totalSeconds) || totalSeconds < 0) return '0:00';
    const minutes = Math.floor(totalSeconds / 60);
    const seconds = Math.floor(totalSeconds % 60);
    return `${String(minutes)}:${String(seconds).padStart(2, '0')}`;
}

function normalizeTime(str) {
    if (!str) return '0:00';
    const parts = str.split(':').map(s => s.trim());
    if (parts.length === 2) return `${Number(parts[0])}:${String(Number(parts[1])||0).padStart(2,'0')}`;
    if (parts.length === 3) {
        const hours = Number(parts[0]) || 0;
        const minutes = Number(parts[1]) || 0;
        const seconds = Number(parts[2]) || 0;
        const total = hours * 3600 + minutes * 60 + seconds;
        return secondsToMMSS(total);
    }
    return '0:00';
}

function timeToSeconds(s) {
    if (!s || typeof s !== 'string' || !s.includes(':')) return 0;
    const parts = s.split(':').map(x => Number(x.trim()));
    if (parts.some(isNaN)) return 0;
    if (parts.length === 2) return parts[0] * 60 + parts[1];
    if (parts.length === 3) return parts[0] * 3600 + parts[1] * 60 + parts[2];
    return 0;
}

function computeProgressPct(cur, dur) {
    if (!Number.isFinite(cur) || !Number.isFinite(dur) || dur <= 0) return 0;
    return Math.max(0, Math.min(100, (cur / dur) * 100));
}
