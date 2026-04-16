package seed

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"demo-streaming/internal/config"
	"demo-streaming/internal/database"
	"demo-streaming/internal/utils/stringsutil"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SeedLocalGormService struct {
	db        *gorm.DB
	appConfig config.AppConfig
}

func NewSeedLocalGormService(db *gorm.DB, appConfig config.AppConfig) *SeedLocalGormService {
	return &SeedLocalGormService{db: db, appConfig: appConfig}
}

func (s *SeedLocalGormService) Execute(ctx context.Context, input SeedLocalInput) (SeedLocalOutput, error) {
	if input.UserCount <= 0 {
		input.UserCount = 3
	}

	if input.Reset {
		if err := s.reset(ctx); err != nil {
			return SeedLocalOutput{}, err
		}
	}

	providerID, providerEnsured, err := s.ensureSRSProvider(ctx)
	if err != nil {
		return SeedLocalOutput{}, err
	}
	_ = providerID

	var out SeedLocalOutput
	if providerEnsured {
		out.MediaProvidersEnsured = 1
	}

	users, err := s.createUsers(ctx, input.UsersFile, input.UserCount)
	if err != nil {
		return SeedLocalOutput{}, err
	}
	out.UsersCreated = len(users)

	return out, nil
}

func (s *SeedLocalGormService) reset(ctx context.Context) error {
	// Order matters because of FKs.
	stmts := []string{
		"TRUNCATE TABLE media_ingest_bindings RESTART IDENTITY CASCADE",
		"TRUNCATE TABLE view_sessions RESTART IDENTITY CASCADE",
		"TRUNCATE TABLE stream_publish_sessions RESTART IDENTITY CASCADE",
		"TRUNCATE TABLE stream_keys RESTART IDENTITY CASCADE",
		"TRUNCATE TABLE media_providers RESTART IDENTITY CASCADE",
		"TRUNCATE TABLE users RESTART IDENTITY CASCADE",
	}
	for _, stmt := range stmts {
		if err := s.db.WithContext(ctx).Exec(stmt).Error; err != nil {
			// Some tables may not exist in early stages; fail loudly so dev fixes migrations.
			return fmt.Errorf("seed reset failed (%s): %w", stmt, err)
		}
	}
	return nil
}

func (s *SeedLocalGormService) ensureSRSProvider(ctx context.Context) (providerID int64, ensured bool, err error) {
	apiBaseURL := "http://localhost:1985"

	// If provider already exists and has config, do not override it.
	var existing database.MediaProvider
	if err := s.db.WithContext(ctx).Where("code = ?", "srs").First(&existing).Error; err == nil {
		rtmpBaseURL, playbackBaseURL := parseProviderURLs(existing.Config)
		if rtmpBaseURL != "" && playbackBaseURL != "" {
			return existing.ID, false, nil
		}
	}

	cfg := map[string]any{
		"rtmp_base_url": stringsutil.FirstNonEmpty(
			s.appConfig.DevSRSRTMPPublishBaseURL,
			s.appConfig.DevSRSRTMPBaseURL,
		),
		"playback_base_url": stringsutil.FirstNonEmpty(
			s.appConfig.DevSRSPlaybackCDNBaseURL,
			s.appConfig.DevSRSPlaybackOriginBaseURL,
		),
	}
	rawCfg, _ := json.Marshal(cfg)

	provider := database.MediaProvider{
		Code:        "srs",
		DisplayName: "SRS Local",
		APIBaseURL:  &apiBaseURL,
		Config:      rawCfg,
	}

	// Upsert by code.
	if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "code"}},
		DoUpdates: clause.AssignmentColumns([]string{"display_name", "api_base_url", "config"}),
	}).Create(&provider).Error; err != nil {
		return 0, false, err
	}

	// GORM won't reliably tell whether it inserted vs updated; treat as ensured.
	return provider.ID, true, nil
}

type providerConfig struct {
	RTMPBaseURL     string `json:"rtmp_base_url"`
	PlaybackBaseURL string `json:"playback_base_url"`
}

func parseProviderURLs(raw []byte) (rtmpBaseURL, playbackBaseURL string) {
	var cfg providerConfig
	if len(raw) == 0 {
		return "", ""
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return "", ""
	}
	return strings.TrimSpace(cfg.RTMPBaseURL), strings.TrimSpace(cfg.PlaybackBaseURL)
}

type userFixture struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

func (s *SeedLocalGormService) createUsers(ctx context.Context, usersFile string, fallbackN int) ([]database.User, error) {
	fixtures, err := loadUsersFixture(usersFile)
	if err != nil {
		return nil, err
	}
	if len(fixtures) == 0 {
		// Backward-compatible fallback (should rarely be used).
		if fallbackN <= 0 {
			fallbackN = 3
		}
		fixtures = make([]userFixture, 0, fallbackN)
		for i := 1; i <= fallbackN; i++ {
			fixtures = append(fixtures, userFixture{
				Email:       fmt.Sprintf("streamer%02d@example.com", i),
				DisplayName: fmt.Sprintf("Streamer %02d", i),
				Password:    "password123",
			})
		}
	}

	users := make([]database.User, 0, len(fixtures))
	for _, f := range fixtures {
		hash, err := bcrypt.GenerateFromPassword([]byte(f.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		users = append(users, database.User{
			Email:        strings.ToLower(strings.TrimSpace(f.Email)),
			DisplayName:  strings.TrimSpace(f.DisplayName),
			PasswordHash: string(hash),
		})
	}

	// Upsert: keep fixtures as source of truth for local testing.
	if err := s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "email"}},
			DoUpdates: clause.AssignmentColumns([]string{"display_name", "password_hash"}),
		}).
		Create(&users).Error; err != nil {
		return nil, err
	}

	emails := make([]string, 0, len(users))
	for _, u := range users {
		emails = append(emails, u.Email)
	}
	var loaded []database.User
	if err := s.db.WithContext(ctx).Where("email IN ?", emails).Order("id asc").Find(&loaded).Error; err != nil {
		return nil, err
	}
	return loaded, nil
}

func loadUsersFixture(path string) ([]userFixture, error) {
	if strings.TrimSpace(path) == "" {
		return nil, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read users fixture failed: %w", err)
	}
	var fixtures []userFixture
	if err := json.Unmarshal(raw, &fixtures); err != nil {
		return nil, fmt.Errorf("parse users fixture failed: %w", err)
	}
	return fixtures, nil
}

func (s *SeedLocalGormService) createStreamKeys(ctx context.Context, userID, providerID int64, n int) ([]database.StreamKey, error) {
	keys := make([]database.StreamKey, 0, n)
	now := time.Now().UTC()
	for i := 1; i <= n; i++ {
		secret := fmt.Sprintf("sk_%d_%d_%d", userID, now.Unix(), i)
		keys = append(keys, database.StreamKey{
			OwnerUserID:     userID,
			StreamKeySecret: secret,
			MediaProviderID: providerID,
			Label:           fmt.Sprintf("seed-%d", i),
		})
	}

	if err := s.db.WithContext(ctx).Create(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

func (s *SeedLocalGormService) createPublishSessions(ctx context.Context, userID, providerID, streamKeyID int64, n int) (int, error) {
	sessions := make([]database.StreamPublishSession, 0, n)
	for i := 1; i <= n; i++ {
		title := strings.TrimSpace(fmt.Sprintf("Seed live %d", i))
		playbackBaseURL := stringsutil.FirstNonEmpty(s.appConfig.DevSRSPlaybackCDNBaseURL, s.appConfig.DevSRSPlaybackOriginBaseURL)
		// Placeholder until we have stream_publish_sessions.id; replaced with hex(id) after insert.
		placeholder := fmt.Sprintf("_seed_pb_%d_%d_%d", userID, time.Now().UTC().UnixNano(), i)
		playbackURL := strings.TrimRight(strings.TrimSpace(playbackBaseURL), "/") + "/" + placeholder + "/master.m3u8"
		sessions = append(sessions, database.StreamPublishSession{
			StreamerUserID:  userID,
			MediaProviderID: providerID,
			StreamKeyID:     streamKeyID,
			PlaybackID:      placeholder,
			Title:           title,
			Status:          "created",
			PlaybackURLCDN:  playbackURL,
		})
	}

	if err := s.db.WithContext(ctx).Create(&sessions).Error; err != nil {
		return 0, err
	}
	base := strings.TrimRight(strings.TrimSpace(stringsutil.FirstNonEmpty(s.appConfig.DevSRSPlaybackCDNBaseURL, s.appConfig.DevSRSPlaybackOriginBaseURL)), "/")
	for i := range sessions {
		pid := fmt.Sprintf("%x", sessions[i].ID)
		if err := s.db.WithContext(ctx).Model(&sessions[i]).Updates(map[string]interface{}{
			"playback_id":      pid,
			"playback_url_cdn": base + "/" + pid + "/master.m3u8",
		}).Error; err != nil {
			return 0, err
		}
	}
	return len(sessions), nil
}
