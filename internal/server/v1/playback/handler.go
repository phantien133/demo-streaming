package playback

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"demo-streaming/internal/config"
	"demo-streaming/internal/database"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	DB        *gorm.DB
	AppConfig config.AppConfig
}

func (h *Handler) Playlist(c *gin.Context) {
	playbackID := strings.TrimSpace(c.Param("playback_id"))
	if playbackID == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	session, streamKeySecret, err := h.lookup(c, playbackID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "playback not found"})
		return
	}
	_ = session

	originBase := strings.TrimRight(strings.TrimSpace(h.AppConfig.SRSInternalPlaybackOriginBaseURL), "/")
	if originBase == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "origin not configured"})
		return
	}

	playlistURL := fmt.Sprintf("%s/%s.m3u8", originBase, streamKeySecret)
	if q := strings.TrimSpace(c.Request.URL.RawQuery); q != "" {
		playlistURL += "?" + q
	}
	resp, err := http.Get(playlistURL)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "origin fetch failed"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "origin read failed"})
		return
	}
	if resp.StatusCode != http.StatusOK {
		c.Data(resp.StatusCode, "text/plain; charset=utf-8", body)
		return
	}

	rewritten := rewritePlaylist(string(body), playbackID, streamKeySecret)
	c.Data(http.StatusOK, "application/vnd.apple.mpegurl", []byte(rewritten))
}

func (h *Handler) Segment(c *gin.Context) {
	playbackID := strings.TrimSpace(c.Param("playback_id"))
	name := strings.TrimSpace(c.Param("name"))
	if playbackID == "" || name == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	_, streamKeySecret, err := h.lookup(c, playbackID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "playback not found"})
		return
	}

	originBase := strings.TrimRight(strings.TrimSpace(h.AppConfig.SRSInternalPlaybackOriginBaseURL), "/")
	if originBase == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "origin not configured"})
		return
	}

	// Map playback-id-prefixed filenames back to stream-key-prefixed filenames.
	originName := strings.Replace(name, playbackID, streamKeySecret, 1)
	originURL := fmt.Sprintf("%s/%s", originBase, originName)
	if q := strings.TrimSpace(c.Request.URL.RawQuery); q != "" {
		originURL += "?" + q
	}

	resp, err := http.Get(originURL)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "origin fetch failed"})
		return
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "origin read failed"})
		return
	}
	c.Data(resp.StatusCode, contentType, data)
}

func (h *Handler) lookup(c *gin.Context, playbackID string) (database.StreamPublishSession, string, error) {
	var session database.StreamPublishSession
	if err := h.DB.WithContext(c.Request.Context()).
		Preload("StreamKey").
		Where("playback_id = ?", playbackID).
		Order("id DESC").
		First(&session).Error; err != nil {
		return database.StreamPublishSession{}, "", err
	}
	secret := strings.TrimSpace(session.StreamKey.StreamKeySecret)
	if secret == "" {
		return database.StreamPublishSession{}, "", gorm.ErrRecordNotFound
	}
	return session, secret, nil
}

func rewritePlaylist(raw, playbackID, streamKeySecret string) string {
	// Replace stream-key-based filenames with playback-id-based filenames.
	// Example:
	// - streamKeySecret-123.ts -> playbackID-123.ts
	// - /live/streamKeySecret.m3u8 (rare inside playlist) -> /live/playbackID.m3u8
	out := raw
	out = strings.ReplaceAll(out, streamKeySecret+".m3u8", playbackID+".m3u8")
	out = strings.ReplaceAll(out, streamKeySecret+"-", playbackID+"-")

	// Rewrite segment lines to hit our proxy path.
	lines := strings.Split(out, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		// Only rewrite known media extensions; leave absolute URLs untouched.
		if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
			continue
		}
		if strings.Contains(trimmed, ".ts") || strings.Contains(trimmed, ".m4s") || strings.Contains(trimmed, ".mp4") || strings.Contains(trimmed, ".aac") {
			lines[i] = fmt.Sprintf("/api/v1/playback/%s/segments/%s", playbackID, trimmed)
		}
	}
	return strings.Join(lines, "\n")
}

