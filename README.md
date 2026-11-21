# share-stream

Share a YouTube video or livestream on a full-screen web page with a QR code
that opens the same URL on another device. The application is a single Go
binary; it generates the QR code locally and embeds the YouTube player.

## Download

Download `share-stream-linux-amd64.tar.gz` from the
[releases page](https://github.com/engineerjoe440/share-stream/releases).
The **stable** release tracks `main`, and the **latest** prerelease tracks
`develop`. These are rolling releases: their archives and tags are updated
when the corresponding branch is built. The published binary targets Linux
on amd64 (x86-64). Rolling releases require release immutability to be disabled
for this repository so the workflow can replace their tags and assets.

```sh
tar -xzf share-stream-linux-amd64.tar.gz
./share-stream -url 'https://www.youtube.com/watch?v=iGZQaJZ14X0' -port 8080
```

## Build from source

Install Go 1.22 or newer, then run:

```sh
git clone https://github.com/engineerjoe440/share-stream.git
cd share-stream
go build -o share-stream .
./share-stream -url 'https://www.youtube.com/watch?v=iGZQaJZ14X0' -port 8080
```

Run the tests with `go test ./...`.

## Usage

```text
share-stream -url URL [-port PORT] [-open]
```

| Option | Description |
| --- | --- |
| `-url` | Required fully qualified YouTube video URL. The QR code contains this exact URL. |
| `-port` | HTTP port from 1 to 65535; defaults to `80`. |
| `-open` | Open the page in the operating system's default browser; defaults to false. |

For example:

```sh
./share-stream -url 'https://youtu.be/iGZQaJZ14X0' -port 8080 -open
```

Open `http://localhost:8080/` in a browser, or use the host machine's IP
address and port from another device on the same network. The server listens
on all interfaces, so network and firewall settings determine who can access
it. Port 80 may require elevated privileges; use a higher port if binding
fails. Press Ctrl+C to stop the server. On Linux, `-open` requires a desktop
environment and `xdg-open`; browser launch failures are logged without
stopping the server.

Supported URLs include YouTube `watch?v=`, `live/`, `embed/`, `shorts/`,
and `v/` links, as well as `youtu.be` links. A specific 11-character video ID
is required; channel-only or playlist-only URLs are not supported. The player
starts muted to allow browser autoplay. YouTube must allow embedding for the
selected video, and the viewing browser needs internet access to YouTube.
