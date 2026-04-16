package streampublishsessions

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"demo-streaming/internal/config"
	"demo-streaming/internal/database"
	streamkeysservice "demo-streaming/internal/services/streamkeys"

	"gorm.io/gorm"
)

type GormCreateService struct {
	db        *gorm.DB
	appConfig config.AppConfig
}

func NewGormCreateService(db *gorm.DB, appConfig config.AppConfig) *GormCreateService {
	return &GormCreateService{db: db, appConfig: appConfig}
}

func (s *GormCreateService) Execute(ctx context.Context, input CreateInput) (CreateOutput, error) {
	var out CreateOutput
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var provider database.MediaProvider
		if err := tx.Where("code = ?", "srs").First(&provider).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrMediaProviderNotFound
			}
			return err
		}

		cfg, cfgErr := parseProviderConfig(provider.Config)
		if cfgErr != nil {
			return cfgErr
		}

		rtmpBaseURL := strings.TrimSpace(cfg.RTMPBaseURL)
		playbackBaseURL := strings.TrimSpace(cfg.PlaybackBaseURL)
		if rtmpBaseURL == "" || playbackBaseURL == "" {
			return ErrMediaProviderMisconfigured
		}

		// Fail fast if streamer already has a LIVE session that hasn't been closed.
		var liveCount int64
		if err := tx.Model(&database.StreamPublishSession{}).
			Where("streamer_user_id = ? AND status = ?", input.StreamerUserID, "live").
			Count(&liveCount).Error; err != nil {
			return err
		}
		if liveCount > 0 {
			return ErrPublishSessionLiveExists
		}

		// RTMP stream name (secret) is independent from playback_id so viewers cannot infer the ingest key.
		// playback_id is hex(session id) and is updated after insert.
		streamKeySecret, keyErr := streamkeysservice.GenerateStreamKey()
		if keyErr != nil {
			return fmt.Errorf("failed to generate stream key: %w", keyErr)
		}
		placeholderPlaybackID, phErr := generatePlaybackID()
		if phErr != nil {
			return phErr
		}
		streamKey := database.StreamKey{
			OwnerUserID:     input.StreamerUserID,
			StreamKeySecret: streamKeySecret,
			MediaProviderID: provider.ID,
			Label:           "publish-session",
		}
		if createErr := tx.Create(&streamKey).Error; createErr != nil {
			return createErr
		}

		session := database.StreamPublishSession{
			StreamerUserID:  input.StreamerUserID,
			MediaProviderID: provider.ID,
			StreamKeyID:     streamKey.ID,
			PlaybackID:      placeholderPlaybackID,
			Title:           strings.TrimSpace(input.Title),
			Status:          "created",
			PlaybackURLCDN:  buildPlaybackURL(playbackBaseURL, placeholderPlaybackID),
		}
		if err := tx.Create(&session).Error; err != nil {
			return err
		}

		playbackID := fmt.Sprintf("%x", session.ID)
		playbackURLCDN := buildPlaybackURL(playbackBaseURL, playbackID)
		if err := tx.Model(&session).Updates(map[string]interface{}{
			"playback_id":      playbackID,
			"playback_url_cdn": playbackURLCDN,
		}).Error; err != nil {
			return err
		}

		out = CreateOutput{
			SessionID:      session.ID,
			PlaybackID:     playbackID,
			Status:         session.Status,
			PlaybackURLCDN: playbackURLCDN,
			Ingest: CreateOutputIngest{
				Provider:  provider.Code,
				RTMPURL:   rtmpBaseURL,
				StreamKey: streamKey.StreamKeySecret,
			},
		}
		return nil
	})
	if err != nil {
		return CreateOutput{}, err
	}
	return out, nil
}

type providerConfig struct {
	RTMPBaseURL     string `json:"rtmp_base_url"`
	PlaybackBaseURL string `json:"playback_base_url"`
}

func parseProviderConfig(raw []byte) (providerConfig, error) {
	var cfg providerConfig
	if len(raw) == 0 {
		return providerConfig{}, ErrMediaProviderMisconfigured
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return providerConfig{}, ErrMediaProviderMisconfigured
	}
	return cfg, nil
}

func generatePlaybackID() (string, error) {
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate playback id: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// buildPlaybackURL returns the CDN master playlist for multi-bitrate output:
// {base}/live/<playback_id>/master.m3u8 — variant relative paths resolve under .../live/<playback_id>/.
func buildPlaybackURL(baseURL, playbackID string) string {
	if strings.TrimSpace(baseURL) == "" {
		return ""
	}
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	return trimmed + "/" + playbackID + "/master.m3u8"
}
