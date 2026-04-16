package transcode

import (
	"testing"

	"demo-streaming/internal/config"
)

func TestTranscodeInputURL(t *testing.T) {
	job := PublishJob{PlaybackID: "a", StreamKeySecret: "deadbeefcafe"}

	t.Run("rtmp when base set", func(t *testing.T) {
		u, wait, err := transcodeInputURL(config.AppConfig{
			TranscodeInputRTMPBaseURL: "rtmp://srs:1935/live",
			TranscodeInputBaseURL:     "http://srs:8080/live",
		}, job)
		if err != nil {
			t.Fatal(err)
		}
		if wait {
			t.Fatalf("expected no HLS wait for RTMP input")
		}
		if u != "rtmp://srs:1935/live/deadbeefcafe" {
			t.Fatalf("got %q", u)
		}
	})

	t.Run("hls when rtmp empty", func(t *testing.T) {
		u, wait, err := transcodeInputURL(config.AppConfig{
			TranscodeInputRTMPBaseURL: "",
			TranscodeInputBaseURL:     "http://srs:8080/live",
		}, job)
		if err != nil {
			t.Fatal(err)
		}
		if !wait {
			t.Fatalf("expected HLS wait")
		}
		if u != "http://srs:8080/live/deadbeefcafe.m3u8" {
			t.Fatalf("got %q", u)
		}
	})

	t.Run("hls error when bases empty", func(t *testing.T) {
		_, _, err := transcodeInputURL(config.AppConfig{
			TranscodeInputRTMPBaseURL: "",
			TranscodeInputBaseURL:     "",
		}, job)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
