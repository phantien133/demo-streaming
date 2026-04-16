package streampublishsessions

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"demo-streaming/internal/config"
	"demo-streaming/internal/database"
)

func TestCreateService_ErrLiveSessionExists(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	now := time.Now().UTC()

	u := database.User{Email: "c1@example.com", DisplayName: "C1", CreatedAt: now}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("seed user failed: %v", err)
	}
	p := database.MediaProvider{Code: "srs", DisplayName: "SRS", Config: []byte(`{"rtmp_base_url":"rtmp://x/live","playback_base_url":"http://x/live"}`), CreatedAt: now}
	if err := db.Create(&p).Error; err != nil {
		t.Fatalf("seed provider failed: %v", err)
	}
	k := database.StreamKey{OwnerUserID: u.ID, StreamKeySecret: "pb-old", MediaProviderID: p.ID, Label: "default", CreatedAt: now}
	if err := db.Create(&k).Error; err != nil {
		t.Fatalf("seed key failed: %v", err)
	}

	startedAt := now.Add(-time.Minute)
	live := database.StreamPublishSession{
		StreamerUserID:  u.ID,
		MediaProviderID: p.ID,
		StreamKeyID:     k.ID,
		PlaybackID:      "pb-old",
		Status:          "live",
		PlaybackURLCDN:  "http://x/live/pb-old/master.m3u8",
		CreatedAt:       now,
		StartedAt:       &startedAt,
	}
	if err := db.Create(&live).Error; err != nil {
		t.Fatalf("seed live session failed: %v", err)
	}
	vs := database.ViewSession{PublishSessionID: live.ID, ViewerRef: "anon", ClientType: "web", JoinedAt: now, LastSeenAt: now}
	if err := db.Create(&vs).Error; err != nil {
		t.Fatalf("seed view session failed: %v", err)
	}

	svc := NewGormCreateService(db, config.AppConfig{})
	_, err := svc.Execute(t.Context(), CreateInput{StreamerUserID: u.ID, Title: "new"})
	if !errors.Is(err, ErrPublishSessionLiveExists) {
		t.Fatalf("expected ErrPublishSessionLiveExists, got %v", err)
	}

	var reloaded database.StreamPublishSession
	if err := db.First(&reloaded, "id = ?", live.ID).Error; err != nil {
		t.Fatalf("load old session failed: %v", err)
	}
	if reloaded.Status != "live" {
		t.Fatalf("expected old session still live, got %s", reloaded.Status)
	}

	var reloadedVS database.ViewSession
	if err := db.First(&reloadedVS, "id = ?", vs.ID).Error; err != nil {
		t.Fatalf("load view session failed: %v", err)
	}
	if reloadedVS.LeftAt != nil {
		t.Fatalf("expected view session left_at to remain nil")
	}
}

func TestCreateService_PlaybackIDIsSessionIDHex_StreamKeySeparate(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	now := time.Now().UTC()

	u := database.User{Email: "c3@example.com", DisplayName: "C3", CreatedAt: now}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("seed user failed: %v", err)
	}
	p := database.MediaProvider{Code: "srs", DisplayName: "SRS", Config: []byte(`{"rtmp_base_url":"rtmp://x/live","playback_base_url":"http://x/live"}`), CreatedAt: now}
	if err := db.Create(&p).Error; err != nil {
		t.Fatalf("seed provider failed: %v", err)
	}

	svc := NewGormCreateService(db, config.AppConfig{})
	out, err := svc.Execute(t.Context(), CreateInput{StreamerUserID: u.ID, Title: "demo"})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if out.Ingest.StreamKey == "" {
		t.Fatalf("expected stream key in output")
	}
	wantPlayback := fmt.Sprintf("%x", out.SessionID)
	if out.PlaybackID != wantPlayback {
		t.Fatalf("playback_id %q want %q", out.PlaybackID, wantPlayback)
	}
	if !strings.HasSuffix(out.PlaybackURLCDN, "/"+wantPlayback+"/master.m3u8") {
		t.Fatalf("playback url %q should end with /%s/master.m3u8", out.PlaybackURLCDN, wantPlayback)
	}

	var session database.StreamPublishSession
	if err := db.Where("id = ?", out.SessionID).First(&session).Error; err != nil {
		t.Fatalf("load session: %v", err)
	}
	if session.PlaybackID != wantPlayback {
		t.Fatalf("stored playback_id %q want %q", session.PlaybackID, wantPlayback)
	}
	var sk database.StreamKey
	if err := db.Where("id = ?", session.StreamKeyID).First(&sk).Error; err != nil {
		t.Fatalf("load stream key: %v", err)
	}
	if sk.StreamKeySecret != out.Ingest.StreamKey {
		t.Fatalf("stream_key_secret should match ingest key")
	}
	if sk.StreamKeySecret == session.PlaybackID {
		t.Fatalf("stream key should not equal playback_id (collision unlikely but logic wrong)")
	}
}

func TestCreateService_ErrMisconfiguredProviderConfig(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	now := time.Now().UTC()

	u := database.User{Email: "c2@example.com", DisplayName: "C2", CreatedAt: now}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("seed user failed: %v", err)
	}
	p := database.MediaProvider{Code: "srs", DisplayName: "SRS", Config: []byte(`{}`), CreatedAt: now}
	if err := db.Create(&p).Error; err != nil {
		t.Fatalf("seed provider failed: %v", err)
	}

	svc := NewGormCreateService(db, config.AppConfig{})
	_, err := svc.Execute(t.Context(), CreateInput{StreamerUserID: u.ID, Title: "x"})
	if !errors.Is(err, ErrMediaProviderMisconfigured) {
		t.Fatalf("expected ErrMediaProviderMisconfigured, got %v", err)
	}
}

