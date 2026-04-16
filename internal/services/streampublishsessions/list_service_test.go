package streampublishsessions

import (
	"context"
	"testing"
	"time"

	"demo-streaming/internal/database"
)

func TestListService_IncludesTranscodeRenditions(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	now := time.Now().UTC()
	if err := db.Exec(`INSERT INTO users(id, email, display_name, password_hash) VALUES (1, 'streamer@example.com', 'streamer', 'hash')`).Error; err != nil {
		t.Fatalf("seed user failed: %v", err)
	}
	if err := db.Exec(`INSERT INTO media_providers(id, code, display_name, config) VALUES (1, 'srs', 'SRS', '{}')`).Error; err != nil {
		t.Fatalf("seed media provider failed: %v", err)
	}
	if err := db.Exec(`INSERT INTO stream_keys(id, owner_user_id, stream_key_secret, media_provider_id, label) VALUES (1, 1, 'playback-101', 1, 'default')`).Error; err != nil {
		t.Fatalf("seed stream key failed: %v", err)
	}

	session := database.StreamPublishSession{
		ID:              101,
		StreamerUserID:  1,
		MediaProviderID: 1,
		StreamKeyID:     1,
		PlaybackID:      "65",
		Title:           "demo",
		Status:          "created",
		PlaybackURLCDN:  "http://localhost/live/65/master.m3u8",
		CreatedAt:       now,
	}
	if err := db.Create(&session).Error; err != nil {
		t.Fatalf("create session failed: %v", err)
	}
	rendition := database.TranscodeRendition{
		PublishSessionID: session.ID,
		PlaybackID:       session.PlaybackID,
		RenditionName:    "720p",
		PlaylistPath:     "65/720p/index.m3u8",
		Status:           "ready",
	}
	if err := db.Create(&rendition).Error; err != nil {
		t.Fatalf("create rendition failed: %v", err)
	}

	svc := NewGormListService(db)
	out, err := svc.Execute(context.Background(), ListInput{
		StreamerUserID: 1,
		Page:           1,
		Limit:          20,
	})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(out.Items) == 0 {
		t.Fatalf("expected items")
	}
	if len(out.Items[0].Renditions) != 1 {
		t.Fatalf("expected 1 rendition, got %d", len(out.Items[0].Renditions))
	}
	if out.Items[0].Renditions[0].Name != "720p" {
		t.Fatalf("unexpected rendition name: %s", out.Items[0].Renditions[0].Name)
	}
}
