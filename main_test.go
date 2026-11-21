package main

import (
	"bytes"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/skip2/go-qrcode"
)

func TestYouTubeVideoID(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://www.youtube.com/watch?v=iGZQaJZ14X0&t=10", "iGZQaJZ14X0"},
		{"https://youtu.be/iGZQaJZ14X0?si=abc", "iGZQaJZ14X0"},
		{"https://youtube.com/live/iGZQaJZ14X0", "iGZQaJZ14X0"},
		{"https://m.youtube.com/shorts/iGZQaJZ14X0", "iGZQaJZ14X0"},
		{"https://www.youtube.com/embed/iGZQaJZ14X0", "iGZQaJZ14X0"},
		{"", ""},
		{"youtube.com/watch?v=iGZQaJZ14X0", ""},
		{"https://youtube.com.evil.example/watch?v=iGZQaJZ14X0", ""},
		{"https://youtube.com@evil.example/watch?v=iGZQaJZ14X0", ""},
		{"https://youtube.com:8080/watch?v=iGZQaJZ14X0", ""},
		{"https://youtu.be/iGZQaJZ14X0/extra", ""},
		{"https://youtube.com/watch?v=invalid", ""},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got, err := youtubeVideoID(tt.url)
			if got != tt.want || (err != nil) != (tt.want == "") {
				t.Fatalf("youtubeVideoID(%q) = %q, %v; want %q", tt.url, got, err, tt.want)
			}
		})
	}
}

func TestHandler(t *testing.T) {
	streamURL := "https://youtu.be/iGZQaJZ14X0?si=abc"
	handler, err := newHandler(streamURL)
	if err != nil {
		t.Fatal(err)
	}

	page := httptest.NewRecorder()
	handler.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/", nil))
	if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), "https://www.youtube.com/embed/iGZQaJZ14X0?autoplay=1&amp;mute=1") || !strings.Contains(page.Body.String(), `src="/qr.png"`) {
		t.Fatalf("unexpected page: status %d, body %s", page.Code, page.Body.String())
	}
	qr := httptest.NewRecorder()
	handler.ServeHTTP(qr, httptest.NewRequest(http.MethodGet, "/qr.png", nil))
	expectedQR, err := qrcode.Encode(streamURL, qrcode.Medium, 512)
	if err != nil {
		t.Fatal(err)
	}
	if qr.Code != http.StatusOK || qr.Header().Get("Content-Type") != "image/png" || !bytes.Equal(qr.Body.Bytes(), expectedQR) {
		t.Fatal("QR response does not encode the supplied URL")
	}
	if _, err := png.Decode(qr.Body); err != nil {
		t.Fatalf("invalid QR PNG: %v", err)
	}
	missing := httptest.NewRecorder()
	handler.ServeHTTP(missing, httptest.NewRequest(http.MethodGet, "/missing", nil))
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing path returned %d", missing.Code)
	}
	if _, err := newHandler("not a URL"); err == nil {
		t.Fatal("invalid URL was accepted")
	}
}
