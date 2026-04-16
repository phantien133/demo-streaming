package mediawebhooks

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"demo-streaming/internal/config"
	mediawebhooksservice "demo-streaming/internal/services/mediawebhooks"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	AppConfig        config.AppConfig
	SRSOnPublishService mediawebhooksservice.SRSOnPublishService
}

type SRSOnPublishRequest struct {
	App    string `json:"app"`
	Stream string `json:"stream"`
	Param  string `json:"param"`
}

func (h *Handler) OnPublish(c *gin.Context) {
	var req SRSOnPublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		payload := gin.H{"code": 1001, "error": "invalid payload"}
		raw, _ := json.Marshal(payload)
		log.Printf("[srs_on_publish] status=%d app=%q stream=%q response=%s", http.StatusOK, "", "", string(raw))
		// SRS expects a JSON object with `code`, even on deny.
		c.JSON(http.StatusOK, payload)
		return
	}

	respond := func(payload gin.H) {
		if _, ok := payload["code"]; !ok {
			// Safety: SRS requires `code` field.
			payload["code"] = 1500
		}
		// Best-effort JSON for readable logs; never block response on logging.
		raw, _ := json.Marshal(payload)
		log.Printf("[srs_on_publish] status=%d app=%q stream=%q response=%s", http.StatusOK, req.App, req.Stream, string(raw))
		// SRS `http_hooks` expects HTTP 200 with JSON {code:0} to allow.
		// Any non-0 code denies publish.
		c.JSON(http.StatusOK, payload)
	}

	out, err := h.SRSOnPublishService.Execute(c.Request.Context(), mediawebhooksservice.SRSOnPublishInput{
		WebhookSecretHeader: c.GetHeader("X-SRS-Secret"),
		ExpectedSecret:      h.AppConfig.SRSWebhookSecret,
		Stream:              req.Stream,
		Param:               req.Param,
	})
	if err != nil {
		switch {
		case errors.Is(err, mediawebhooksservice.ErrUnauthorized):
			log.Printf("[srs_on_publish] deny=unauthorized app=%q stream=%q err=%v", req.App, req.Stream, err)
			respond(gin.H{"code": 1002, "error": "unauthorized"})
		case errors.Is(err, mediawebhooksservice.ErrBadRequest):
			log.Printf("[srs_on_publish] deny=bad_request app=%q stream=%q err=%v", req.App, req.Stream, err)
			respond(gin.H{"code": 1003, "error": "bad request"})
		case errors.Is(err, mediawebhooksservice.ErrForbidden):
			log.Printf("[srs_on_publish] deny=forbidden app=%q stream=%q err=%v", req.App, req.Stream, err)
			respond(gin.H{"code": 1004, "error": "forbidden"})
		default:
			// Internal error; log full error for debugging.
			log.Printf("[srs_on_publish] deny=internal_error app=%q stream=%q err=%v", req.App, req.Stream, err)
			respond(gin.H{"code": 1500, "error": "internal error"})
		}
		return
	}

	respond(gin.H{
		"code":       0,
		"status":     out.Status,
		"session_id": out.SessionID,
	})
}

