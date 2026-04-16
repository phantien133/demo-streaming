package streampublishsessions

import (
	"errors"
	"testing"
	"time"

	"demo-streaming/internal/database"
)

func TestStopLiveService_StopByID(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	now := time.Now().UTC()

	u := database.User{Email: "a@example.com", DisplayName: "A", CreatedAt: now}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("seed user failed: %v", err)
	}
	p := database.MediaProvider{Code: "srs", DisplayName: "SRS", Config: []byte(`{"rtmp_base_url":"rtmp://x/live","playback_base_url":"http://x/live"}`), CreatedAt: now}
	if err := db.Create(&p).Error; err != nil {
		t.Fatalf("seed provider failed: %v", err)
	}
	k := database.StreamKey{OwnerUserID: u.ID, StreamKeySecret: "pb1", MediaProviderID: p.ID, Label: "default", CreatedAt: now}
	if err := db.Create(&k).Error; err != nil {
		t.Fatalf("seed key failed: %v", err)
	}

	startedAt := now.Add(-time.Minute)
	s := database.StreamPublishSession{
		StreamerUserID:  u.ID,
		MediaProviderID: p.ID,
		StreamKeyID:     k.ID,
		PlaybackID:      "pb1",
		Status:          "live",
		PlaybackURLCDN:  "http://x/live/pb1/master.m3u8",
		CreatedAt:       now,
		StartedAt:       &startedAt,
	}
	if err := db.Create(&s).Error; err != nil {
		t.Fatalf("seed session failed: %v", err)
	}
	vs := database.ViewSession{
		PublishSessionID: s.ID,
		ViewerRef:        "anon",
		ClientType:       "web",
		JoinedAt:         now,
		LastSeenAt:       now,
	}
	if err := db.Create(&vs).Error; err != nil {
		t.Fatalf("seed view session failed: %v", err)
	}

	svc := NewGormStopLiveService(db)
	out, err := svc.Execute(t.Context(), StopLiveInput{StreamerUserID: u.ID, SessionID: &s.ID})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if len(out.StoppedSessionIDs) != 1 || out.StoppedSessionIDs[0] != s.ID {
		t.Fatalf("unexpected stopped ids: %+v", out.StoppedSessionIDs)
	}

	var got database.StreamPublishSession
	if err := db.First(&got, "id = ?", s.ID).Error; err != nil {
		t.Fatalf("load session failed: %v", err)
	}
	if got.Status != "ended" {
		t.Fatalf("expected ended, got %s", got.Status)
	}
	if got.EndedAt == nil {
		t.Fatalf("expected ended_at to be set")
	}

	var gotVS database.ViewSession
	if err := db.First(&gotVS, "id = ?", vs.ID).Error; err != nil {
		t.Fatalf("load view session failed: %v", err)
	}
	if gotVS.LeftAt == nil {
		t.Fatalf("expected view_session.left_at to be set")
	}
}

func TestStopLiveService_StopAllLiveForStreamer(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	now := time.Now().UTC()

	u := database.User{Email: "a2@example.com", DisplayName: "A2", CreatedAt: now}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("seed user failed: %v", err)
	}
	p := database.MediaProvider{Code: "srs", DisplayName: "SRS", Config: []byte(`{"rtmp_base_url":"rtmp://x/live","playback_base_url":"http://x/live"}`), CreatedAt: now}
	if err := db.Create(&p).Error; err != nil {
		t.Fatalf("seed provider failed: %v", err)
	}
	k1 := database.StreamKey{OwnerUserID: u.ID, StreamKeySecret: "pb1", MediaProviderID: p.ID, Label: "default", CreatedAt: now}
	k2 := database.StreamKey{OwnerUserID: u.ID, StreamKeySecret: "pb2", MediaProviderID: p.ID, Label: "default", CreatedAt: now}
	if err := db.Create(&k1).Error; err != nil {
		t.Fatalf("seed key1 failed: %v", err)
	}
	if err := db.Create(&k2).Error; err != nil {
		t.Fatalf("seed key2 failed: %v", err)
	}

	s1 := database.StreamPublishSession{StreamerUserID: u.ID, MediaProviderID: p.ID, StreamKeyID: k1.ID, PlaybackID: "pb1", Status: "live", PlaybackURLCDN: "http://x/live/pb1/master.m3u8", CreatedAt: now}
	s2 := database.StreamPublishSession{StreamerUserID: u.ID, MediaProviderID: p.ID, StreamKeyID: k2.ID, PlaybackID: "pb2", Status: "live", PlaybackURLCDN: "http://x/live/pb2/master.m3u8", CreatedAt: now}
	if err := db.Create(&s1).Error; err != nil {
		t.Fatalf("seed s1 failed: %v", err)
	}
	if err := db.Create(&s2).Error; err != nil {
		t.Fatalf("seed s2 failed: %v", err)
	}
	vs := database.ViewSession{PublishSessionID: s1.ID, ViewerRef: "anon", ClientType: "web", JoinedAt: now, LastSeenAt: now}
	if err := db.Create(&vs).Error; err != nil {
		t.Fatalf("seed view session failed: %v", err)
	}

	svc := NewGormStopLiveService(db)
	out, err := svc.Execute(t.Context(), StopLiveInput{StreamerUserID: u.ID, SessionID: nil})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if len(out.StoppedSessionIDs) != 2 {
		t.Fatalf("expected 2 stopped sessions, got %d", len(out.StoppedSessionIDs))
	}

	var count int64
	if err := db.Model(&database.StreamPublishSession{}).
		Where("streamer_user_id = ? AND status = ?", u.ID, "live").
		Count(&count).Error; err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 live sessions, got %d", count)
	}
}

func TestStopLiveService_Errors(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	now := time.Now().UTC()
	u := database.User{Email: "a3@example.com", DisplayName: "A3", CreatedAt: now}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("seed user failed: %v", err)
	}
	p := database.MediaProvider{Code: "srs", DisplayName: "SRS", Config: []byte(`{"rtmp_base_url":"rtmp://x/live","playback_base_url":"http://x/live"}`), CreatedAt: now}
	if err := db.Create(&p).Error; err != nil {
		t.Fatalf("seed provider failed: %v", err)
	}
	k := database.StreamKey{OwnerUserID: u.ID, StreamKeySecret: "pb3", MediaProviderID: p.ID, Label: "default", CreatedAt: now}
	if err := db.Create(&k).Error; err != nil {
		t.Fatalf("seed key failed: %v", err)
	}
	s := database.StreamPublishSession{StreamerUserID: u.ID, MediaProviderID: p.ID, StreamKeyID: k.ID, PlaybackID: "pb3", Status: "created", PlaybackURLCDN: "http://x/live/pb3/master.m3u8", CreatedAt: now}
	if err := db.Create(&s).Error; err != nil {
		t.Fatalf("seed session failed: %v", err)
	}

	svc := NewGormStopLiveService(db)

	t.Run("bad state when not live", func(t *testing.T) {
		t.Parallel()
		_, err := svc.Execute(t.Context(), StopLiveInput{StreamerUserID: u.ID, SessionID: &s.ID})
		if !errors.Is(err, ErrPublishSessionBadState) {
			t.Fatalf("expected ErrPublishSessionBadState, got %v", err)
		}
	})

	t.Run("forbidden when different owner", func(t *testing.T) {
		t.Parallel()
		_, err := svc.Execute(t.Context(), StopLiveInput{StreamerUserID: 999, SessionID: &s.ID})
		if !errors.Is(err, ErrPublishSessionForbidden) {
			t.Fatalf("expected ErrPublishSessionForbidden, got %v", err)
		}
	})
}

