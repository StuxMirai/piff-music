function getNowPlaying() {
    let title = document.querySelector('yt-formatted-string.title.style-scope.ytmusic-player-bar')?.textContent?.trim();

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


setInterval(() => {
    const nowPlaying = getNowPlaying();
    // console.log("Now Playing:", nowPlaying);

    fetch("http://localhost:8080/webhook", {
        method: "POST",
        body: JSON.stringify({
            song_name: nowPlaying.song_name,
            artist: nowPlaying.artist,
            current_timestamp: nowPlaying.current_timestamp,
            end_timestamp: nowPlaying.end_timestamp
        })
    });

}, 1000);
