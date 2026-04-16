package streampublishsessions

import (
	"errors"
	"net/http"
	"strconv"

	"demo-streaming/internal/auth"
	"demo-streaming/internal/middleware"
	streampublishsessionsservice "demo-streaming/internal/services/streampublishsessions"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	CreateService   streampublishsessionsservice.CreateService
	ListService     streampublishsessionsservice.ListService
	StartService    streampublishsessionsservice.StartService
	StopLiveService streampublishsessionsservice.StopLiveService
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type CreateRequest struct {
	Title string `json:"title"`
}

type CreateIngestResponse struct {
	Provider  string `json:"provider" example:"srs"`
	RTMPURL   string `json:"rtmp_url" example:"rtmp://localhost:1935/live"`
	StreamKey string `json:"stream_key"`
}

type CreateResponse struct {
	SessionID      int64                `json:"session_id"`
	PlaybackID     string               `json:"playback_id"`
	Status         string               `json:"status"`
	PlaybackURLCDN string               `json:"playback_url_cdn"`
	Ingest         CreateIngestResponse `json:"ingest"`
}

type StartResponse struct {
	SessionID int64  `json:"session_id"`
	Status    string `json:"status"`
	StartedAt int64  `json:"started_at"`
}

type StopResponse struct {
	SessionID int64  `json:"session_id"`
	Status    string `json:"status"`
	EndedAt   int64  `json:"ended_at"`
}

type StopAllResponse struct {
	StoppedSessionIDs []int64 `json:"stopped_session_ids"`
	Status            string  `json:"status"`
	EndedAt           int64   `json:"ended_at"`
}

type ListItemResponse struct {
	SessionID      int64                   `json:"session_id"`
	PlaybackID     string                  `json:"playback_id"`
	Title          string                  `json:"title"`
	Status         string                  `json:"status"`
	PlaybackURLCDN string                  `json:"playback_url_cdn"`
	CreatedAt      int64                   `json:"created_at"`
	StartedAt      *int64                  `json:"started_at,omitempty"`
	EndedAt        *int64                  `json:"ended_at,omitempty"`
	Renditions     []ListRenditionResponse `json:"renditions"`
}

type ListRenditionResponse struct {
	Name         string `json:"name"`
	PlaylistPath string `json:"playlist_path"`
	Status       string `json:"status"`
}

type ListResponse struct {
	Items []ListItemResponse `json:"items"`
	Page  int                `json:"page"`
	Limit int                `json:"limit"`
	Total int64              `json:"total"`
}

// Create allocates a publish session and returns playback info.
//
// @Summary Create stream publish session
// @Tags stream-publish-sessions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateRequest true "create publish session"
// @Success 201 {object} CreateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/stream-publish-sessions [post]
func (h *Handler) Create(c *gin.Context) {
	claims, ok := authClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth claims"})
		return
	}

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	out, err := h.CreateService.Execute(c.Request.Context(), streampublishsessionsservice.CreateInput{
		StreamerUserID: claims.UserID,
		Title:          req.Title,
	})
	if err != nil {
		switch {
		case errors.Is(err, streampublishsessionsservice.ErrPublishSessionLiveExists):
			c.JSON(http.StatusConflict, gin.H{"error": "a live session already exists; stop it before creating a new one"})
		case errors.Is(err, streampublishsessionsservice.ErrMediaProviderNotFound):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "media provider not configured"})
		case errors.Is(err, streampublishsessionsservice.ErrMediaProviderMisconfigured):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "media provider misconfigured"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create publish session"})
		}
		return
	}

	c.JSON(http.StatusCreated, CreateResponse{
		SessionID:      out.SessionID,
		PlaybackID:     out.PlaybackID,
		Status:         out.Status,
		PlaybackURLCDN: out.PlaybackURLCDN,
		Ingest: CreateIngestResponse{
			Provider:  out.Ingest.Provider,
			RTMPURL:   out.Ingest.RTMPURL,
			StreamKey: out.Ingest.StreamKey,
		},
	})
}

// List returns paginated publish sessions for current streamer.
//
// @Summary List stream publish sessions
// @Tags stream-publish-sessions
// @Security BearerAuth
// @Produce json
// @Param page query int false "page number (default 1)"
// @Param limit query int false "page size (default 20, max 100)"
// @Success 200 {object} ListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/stream-publish-sessions [get]
func (h *Handler) List(c *gin.Context) {
	claims, ok := authClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth claims"})
		return
	}

	page := 1
	if raw := c.Query("page"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page"})
			return
		}
		page = parsed
	}

	limit := 20
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
			return
		}
		limit = parsed
	}

	out, err := h.ListService.Execute(c.Request.Context(), streampublishsessionsservice.ListInput{
		StreamerUserID: claims.UserID,
		Page:           page,
		Limit:          limit,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list publish sessions"})
		return
	}

	items := make([]ListItemResponse, 0, len(out.Items))
	for _, item := range out.Items {
		row := ListItemResponse{
			SessionID:      item.SessionID,
			PlaybackID:     item.PlaybackID,
			Title:          item.Title,
			Status:         item.Status,
			PlaybackURLCDN: item.PlaybackURLCDN,
			CreatedAt:      item.CreatedAt.Unix(),
			Renditions:     make([]ListRenditionResponse, 0, len(item.Renditions)),
		}
		if item.StartedAt != nil {
			v := item.StartedAt.Unix()
			row.StartedAt = &v
		}
		if item.EndedAt != nil {
			v := item.EndedAt.Unix()
			row.EndedAt = &v
		}
		for _, rendition := range item.Renditions {
			row.Renditions = append(row.Renditions, ListRenditionResponse{
				Name:         rendition.Name,
				PlaylistPath: rendition.PlaylistPath,
				Status:       rendition.Status,
			})
		}
		items = append(items, row)
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"page":  out.Page,
		"limit": out.Limit,
		"total": out.Total,
	})
}

// Start marks a created publish session as started (manual step).
//
// @Summary Start stream publish session
// @Tags stream-publish-sessions
// @Security BearerAuth
// @Produce json
// @Success 200 {object} StartResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/stream-publish-sessions/{id}/start [post]
func (h *Handler) Start(c *gin.Context) {
	claims, ok := authClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth claims"})
		return
	}

	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || sessionID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid publish session id"})
		return
	}

	out, err := h.StartService.Execute(c.Request.Context(), streampublishsessionsservice.StartInput{
		SessionID:      sessionID,
		StreamerUserID: claims.UserID,
	})
	if err != nil {
		switch {
		case errors.Is(err, streampublishsessionsservice.ErrPublishSessionNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "publish session not found"})
		case errors.Is(err, streampublishsessionsservice.ErrPublishSessionForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "publish session does not belong to user"})
		case errors.Is(err, streampublishsessionsservice.ErrPublishSessionBadState):
			c.JSON(http.StatusConflict, gin.H{"error": "publish session is not in created state"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start publish session"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": out.SessionID,
		"status":     out.Status,
		"started_at": out.StartedAt,
	})
}

// Stop ends a live publish session (manual step).
//
// @Summary Stop stream publish session
// @Tags stream-publish-sessions
// @Security BearerAuth
// @Produce json
// @Success 200 {object} StopResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/stream-publish-sessions/{id}/stop [post]
func (h *Handler) Stop(c *gin.Context) {
	claims, ok := authClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth claims"})
		return
	}

	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || sessionID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid publish session id"})
		return
	}

	out, err := h.StopLiveService.Execute(c.Request.Context(), streampublishsessionsservice.StopLiveInput{
		StreamerUserID: claims.UserID,
		SessionID:      &sessionID,
	})
	if err != nil {
		switch {
		case errors.Is(err, streampublishsessionsservice.ErrPublishSessionNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "publish session not found"})
		case errors.Is(err, streampublishsessionsservice.ErrPublishSessionForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "publish session does not belong to user"})
		case errors.Is(err, streampublishsessionsservice.ErrPublishSessionBadState):
			c.JSON(http.StatusConflict, gin.H{"error": "publish session is not in live state"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to stop publish session"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"status":     "ended",
		"ended_at":   out.EndedAt.Unix(),
	})
}

// StopAll ends ALL live publish sessions for the current streamer.
//
// @Summary Stop all live stream publish sessions
// @Tags stream-publish-sessions
// @Security BearerAuth
// @Produce json
// @Success 200 {object} StopAllResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/stream-publish-sessions/stop-all [post]
func (h *Handler) StopAll(c *gin.Context) {
	claims, ok := authClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth claims"})
		return
	}

	out, err := h.StopLiveService.Execute(c.Request.Context(), streampublishsessionsservice.StopLiveInput{
		StreamerUserID: claims.UserID,
		SessionID:      nil,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to stop live sessions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stopped_session_ids": out.StoppedSessionIDs,
		"status":              "ended",
		"ended_at":            out.EndedAt.Unix(),
	})
}

func authClaimsFromContext(c *gin.Context) (*auth.Claims, bool) {
	rawClaims, ok := c.Get(middleware.AuthClaimsContextKey)
	if !ok {
		return nil, false
	}
	claims, ok := rawClaims.(*auth.Claims)
	return claims, ok
}
