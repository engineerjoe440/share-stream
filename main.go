package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/skip2/go-qrcode"
)

var videoIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

func youtubeVideoID(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil || u == nil || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || u.Host == "" || u.Port() != "" {
		return "", errors.New("url must be a fully qualified YouTube video URL")
	}

	var id string
	switch strings.ToLower(u.Hostname()) {
	case "youtube.com", "www.youtube.com", "m.youtube.com", "music.youtube.com":
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) == 1 && parts[0] == "watch" {
			id = u.Query().Get("v")
		} else if len(parts) == 2 && (parts[0] == "live" || parts[0] == "embed" || parts[0] == "shorts" || parts[0] == "v") {
			id = parts[1]
		}
	case "youtu.be", "www.youtu.be":
		id = strings.Trim(u.Path, "/")
	}
	if !videoIDPattern.MatchString(id) {
		return "", errors.New("url must identify a YouTube video")
	}
	return id, nil
}

var pageTemplate = template.Must(template.New("page").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Livestream + QR Code</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
html, body { width: 100%; height: 100%; overflow: hidden; background: #000; font-family: Arial, sans-serif; }
.video-wrapper { position: fixed; inset: 0; }
.video-wrapper iframe { width: 100%; height: 100%; border: 0; }
.qr-container { position: fixed; right: 20px; bottom: 20px; z-index: 1000; width: 240px; background: rgba(255,255,255,.95); padding: 12px; border-radius: 12px; box-shadow: 0 6px 20px rgba(0,0,0,.45); }
.qr-container img { display: block; width: 100%; height: auto; }
@media (max-width: 900px) { .qr-container { width: 180px; } }
@media (max-width: 600px) { .qr-container { width: 140px; right: 10px; bottom: 10px; padding: 8px; } }
</style>
</head>
<body>
<div class="video-wrapper"><iframe src="{{.EmbedURL}}" title="Livestream player" allow="autoplay; encrypted-media; picture-in-picture" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe></div>
<div class="qr-container"><img src="/qr.png" alt="QR code for the YouTube stream"></div>
</body>
</html>`))

func newHandler(rawURL string) (http.Handler, error) {
	id, err := youtubeVideoID(rawURL)
	if err != nil {
		return nil, err
	}
	qr, err := qrcode.Encode(rawURL, qrcode.Medium, 512)
	if err != nil {
		return nil, fmt.Errorf("generate QR code: %w", err)
	}
	var page bytes.Buffer
	if err := pageTemplate.Execute(&page, struct{ EmbedURL string }{
		EmbedURL: "https://www.youtube.com/embed/" + id + "?autoplay=1&mute=1",
	}); err != nil {
		return nil, fmt.Errorf("render page: %w", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(page.Bytes())
	})
	mux.HandleFunc("/qr.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(qr)
	})
	return mux, nil
}

func openBrowser(address string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", address)
	case "darwin":
		cmd = exec.Command("open", address)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", address)
	default:
		return fmt.Errorf("opening a browser is not supported on %s", runtime.GOOS)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Printf("Could not open browser: %v", err)
		}
	}()
	return nil
}

func main() {
	streamURL := flag.String("url", "", "fully qualified YouTube video URL (required)")
	port := flag.Int("port", 80, "HTTP port (1-65535)")
	open := flag.Bool("open", false, "open the page in the default browser")
	flag.Parse()
	if flag.NArg() != 0 {
		log.Fatal("unexpected positional arguments")
	}
	if *port < 1 || *port > 65535 {
		log.Fatal("port must be between 1 and 65535")
	}
	handler, err := newHandler(*streamURL)
	if err != nil {
		log.Fatal(err)
	}
	listener, err := net.Listen("tcp", net.JoinHostPort("", strconv.Itoa(*port)))
	if err != nil {
		log.Fatal(err)
	}
	address := "http://localhost:" + strconv.Itoa(*port) + "/"
	log.Printf("Serving %s", address)
	if *open {
		if err := openBrowser(address); err != nil {
			log.Printf("Could not open browser: %v", err)
		}
	}
	server := &http.Server{Handler: handler, ReadHeaderTimeout: 10 * time.Second}
	log.Fatal(server.Serve(listener))
}
