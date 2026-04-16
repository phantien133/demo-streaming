package streampublishsessions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authpkg "demo-streaming/internal/auth"
	streampublishsessionsservice "demo-streaming/internal/services/streampublishsessions"
	"github.com/gin-gonic/gin"
)

func TestStartRoute_Unauthorized(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	router := gin.New()
	v1 := router.Group("/api/v1")
	jwtManager := mustJWTManager(t)
	RegisterRoutes(v1, Deps{
		JWTManager: jwtManager,
		ListService: listServiceFunc(func(context.Context, streampublishsessionsservice.ListInput) (streampublishsessionsservice.ListOutput, error) {
			return streampublishsessionsservice.ListOutput{}, nil
		}),
		CreateService: createServiceFunc(func(context.Context, streampublishsessionsservice.CreateInput) (streampublishsessionsservice.CreateOutput, error) {
			return streampublishsessionsservice.CreateOutput{}, nil
		}),
		StartService: startServiceFunc(func(context.Context, streampublishsessionsservice.StartInput) (streampublishsessionsservice.StartOutput, error) {
			return streampublishsessionsservice.StartOutput{}, nil
		}),
		StopLiveService: stopLiveServiceFunc(func(context.Context, streampublishsessionsservice.StopLiveInput) (streampublishsessionsservice.StopLiveOutput, error) {
			return streampublishsessionsservice.StopLiveOutput{}, nil
		}),
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/stream-publish-sessions/1/start", strings.NewReader(""))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestStopRoute_Unauthorized(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	router := gin.New()
	v1 := router.Group("/api/v1")
	jwtManager := mustJWTManager(t)
	RegisterRoutes(v1, Deps{
		JWTManager: jwtManager,
		ListService: listServiceFunc(func(context.Context, streampublishsessionsservice.ListInput) (streampublishsessionsservice.ListOutput, error) {
			return streampublishsessionsservice.ListOutput{}, nil
		}),
		CreateService: createServiceFunc(func(context.Context, streampublishsessionsservice.CreateInput) (streampublishsessionsservice.CreateOutput, error) {
			return streampublishsessionsservice.CreateOutput{}, nil
		}),
		StartService: startServiceFunc(func(context.Context, streampublishsessionsservice.StartInput) (streampublishsessionsservice.StartOutput, error) {
			return streampublishsessionsservice.StartOutput{}, nil
		}),
		StopLiveService: stopLiveServiceFunc(func(context.Context, streampublishsessionsservice.StopLiveInput) (streampublishsessionsservice.StopLiveOutput, error) {
			return streampublishsessionsservice.StopLiveOutput{}, nil
		}),
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/stream-publish-sessions/1/stop", strings.NewReader(""))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestStopAllRoute_Unauthorized(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	router := gin.New()
	v1 := router.Group("/api/v1")
	jwtManager := mustJWTManager(t)
	RegisterRoutes(v1, Deps{
		JWTManager: jwtManager,
		ListService: listServiceFunc(func(context.Context, streampublishsessionsservice.ListInput) (streampublishsessionsservice.ListOutput, error) {
			return streampublishsessionsservice.ListOutput{}, nil
		}),
		CreateService: createServiceFunc(func(context.Context, streampublishsessionsservice.CreateInput) (streampublishsessionsservice.CreateOutput, error) {
			return streampublishsessionsservice.CreateOutput{}, nil
		}),
		StartService: startServiceFunc(func(context.Context, streampublishsessionsservice.StartInput) (streampublishsessionsservice.StartOutput, error) {
			return streampublishsessionsservice.StartOutput{}, nil
		}),
		StopLiveService: stopLiveServiceFunc(func(context.Context, streampublishsessionsservice.StopLiveInput) (streampublishsessionsservice.StopLiveOutput, error) {
			return streampublishsessionsservice.StopLiveOutput{}, nil
		}),
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/stream-publish-sessions/stop-all", strings.NewReader(""))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestCreateRoute_Mappings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		serviceErr error
		wantStatus int
	}{
		{
			name:       "invalid payload returns bad request",
			body:       `{"title":1}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "media provider missing maps to 503",
			body:       `{"title":"my live"}`,
			serviceErr: streampublishsessionsservice.ErrMediaProviderNotFound,
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "live session exists maps to 409",
			body:       `{"title":"my live"}`,
			serviceErr: streampublishsessionsservice.ErrPublishSessionLiveExists,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "success returns 201",
			body:       `{"title":"my live"}`,
			wantStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gin.SetMode(gin.TestMode)

			router := gin.New()
			v1 := router.Group("/api/v1")
			jwtManager := mustJWTManager(t)
			RegisterRoutes(v1, Deps{
				JWTManager: jwtManager,
				ListService: listServiceFunc(func(context.Context, streampublishsessionsservice.ListInput) (streampublishsessionsservice.ListOutput, error) {
					return streampublishsessionsservice.ListOutput{}, nil
				}),
				CreateService: createServiceFunc(func(context.Context, streampublishsessionsservice.CreateInput) (streampublishsessionsservice.CreateOutput, error) {
					if tt.serviceErr != nil {
						return streampublishsessionsservice.CreateOutput{}, tt.serviceErr
					}
					return streampublishsessionsservice.CreateOutput{
						SessionID:      10,
						Status:         "created",
						PlaybackURLCDN: "http://localhost:8080/live/a/master.m3u8",
						Ingest: streampublishsessionsservice.CreateOutputIngest{
							Provider:  "srs",
							RTMPURL:   "rtmp://localhost:1935/live",
							StreamKey: "a",
						},
					}, nil
				}),
				StartService: startServiceFunc(func(context.Context, streampublishsessionsservice.StartInput) (streampublishsessionsservice.StartOutput, error) {
					return streampublishsessionsservice.StartOutput{}, nil
				}),
				StopLiveService: stopLiveServiceFunc(func(context.Context, streampublishsessionsservice.StopLiveInput) (streampublishsessionsservice.StopLiveOutput, error) {
					return streampublishsessionsservice.StopLiveOutput{}, nil
				}),
			})

			req := httptest.NewRequest(http.MethodPost, "/api/v1/stream-publish-sessions", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, jwtManager))
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("unexpected status: got %d want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusCreated {
				var body map[string]any
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if body["status"] != "created" {
					t.Fatalf("unexpected create status payload: %v", body["status"])
				}
				ingest, ok := body["ingest"].(map[string]any)
				if !ok {
					t.Fatalf("expected ingest object, got %T", body["ingest"])
				}
				if ingest["provider"] != "srs" || ingest["rtmp_url"] != "rtmp://localhost:1935/live" || ingest["stream_key"] != "a" {
					t.Fatalf("unexpected ingest payload: %+v", ingest)
				}
			}
		})
	}
}

func TestStartRoute_Mappings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		target      string
		serviceErr  error
		wantStatus  int
	}{
		{
			name:       "invalid id returns bad request",
			target:     "/api/v1/stream-publish-sessions/abc/start",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not found maps to 404",
			target:     "/api/v1/stream-publish-sessions/1/start",
			serviceErr: streampublishsessionsservice.ErrPublishSessionNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "forbidden maps to 403",
			target:     "/api/v1/stream-publish-sessions/1/start",
			serviceErr: streampublishsessionsservice.ErrPublishSessionForbidden,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "bad state maps to 409",
			target:     "/api/v1/stream-publish-sessions/1/start",
			serviceErr: streampublishsessionsservice.ErrPublishSessionBadState,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "success returns 200",
			target:     "/api/v1/stream-publish-sessions/1/start",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gin.SetMode(gin.TestMode)

			router := gin.New()
			v1 := router.Group("/api/v1")
			jwtManager := mustJWTManager(t)
			RegisterRoutes(v1, Deps{
				JWTManager: jwtManager,
				ListService: listServiceFunc(func(context.Context, streampublishsessionsservice.ListInput) (streampublishsessionsservice.ListOutput, error) {
					return streampublishsessionsservice.ListOutput{}, nil
				}),
				CreateService: createServiceFunc(func(context.Context, streampublishsessionsservice.CreateInput) (streampublishsessionsservice.CreateOutput, error) {
					return streampublishsessionsservice.CreateOutput{}, nil
				}),
				StartService: startServiceFunc(func(context.Context, streampublishsessionsservice.StartInput) (streampublishsessionsservice.StartOutput, error) {
					if tt.serviceErr != nil {
						return streampublishsessionsservice.StartOutput{}, tt.serviceErr
					}
					return streampublishsessionsservice.StartOutput{
						SessionID: 1,
						Status:    "live",
						StartedAt: time.Now().UTC(),
					}, nil
				}),
				StopLiveService: stopLiveServiceFunc(func(context.Context, streampublishsessionsservice.StopLiveInput) (streampublishsessionsservice.StopLiveOutput, error) {
					return streampublishsessionsservice.StopLiveOutput{}, nil
				}),
			})

			req := httptest.NewRequest(http.MethodPost, tt.target, strings.NewReader(""))
			req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, jwtManager))
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("unexpected status: got %d want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var body map[string]any
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if body["status"] != "live" {
					t.Fatalf("unexpected status payload: %v", body["status"])
				}
			}
		})
	}
}

func TestStopRoute_Mappings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		target     string
		serviceErr error
		wantStatus int
	}{
		{
			name:       "invalid id returns bad request",
			target:     "/api/v1/stream-publish-sessions/abc/stop",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not found maps to 404",
			target:     "/api/v1/stream-publish-sessions/1/stop",
			serviceErr: streampublishsessionsservice.ErrPublishSessionNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "forbidden maps to 403",
			target:     "/api/v1/stream-publish-sessions/1/stop",
			serviceErr: streampublishsessionsservice.ErrPublishSessionForbidden,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "bad state maps to 409",
			target:     "/api/v1/stream-publish-sessions/1/stop",
			serviceErr: streampublishsessionsservice.ErrPublishSessionBadState,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "success returns 200",
			target:     "/api/v1/stream-publish-sessions/1/stop",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gin.SetMode(gin.TestMode)

			router := gin.New()
			v1 := router.Group("/api/v1")
			jwtManager := mustJWTManager(t)
			RegisterRoutes(v1, Deps{
				JWTManager: jwtManager,
				ListService: listServiceFunc(func(context.Context, streampublishsessionsservice.ListInput) (streampublishsessionsservice.ListOutput, error) {
					return streampublishsessionsservice.ListOutput{}, nil
				}),
				CreateService: createServiceFunc(func(context.Context, streampublishsessionsservice.CreateInput) (streampublishsessionsservice.CreateOutput, error) {
					return streampublishsessionsservice.CreateOutput{}, nil
				}),
				StartService: startServiceFunc(func(context.Context, streampublishsessionsservice.StartInput) (streampublishsessionsservice.StartOutput, error) {
					return streampublishsessionsservice.StartOutput{}, nil
				}),
				StopLiveService: stopLiveServiceFunc(func(context.Context, streampublishsessionsservice.StopLiveInput) (streampublishsessionsservice.StopLiveOutput, error) {
					if tt.serviceErr != nil {
						return streampublishsessionsservice.StopLiveOutput{}, tt.serviceErr
					}
					return streampublishsessionsservice.StopLiveOutput{
						StoppedSessionIDs: []int64{1},
						EndedAt:           time.Now().UTC(),
					}, nil
				}),
			})

			req := httptest.NewRequest(http.MethodPost, tt.target, strings.NewReader(""))
			req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, jwtManager))
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("unexpected status: got %d want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var body map[string]any
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if body["status"] != "ended" {
					t.Fatalf("unexpected status payload: %v", body["status"])
				}
				if _, ok := body["ended_at"].(float64); !ok {
					t.Fatalf("expected ended_at to be a number, got %T", body["ended_at"])
				}
			}
		})
	}
}

func TestStopAllRoute_Returns200(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	router := gin.New()
	v1 := router.Group("/api/v1")
	jwtManager := mustJWTManager(t)
	RegisterRoutes(v1, Deps{
		JWTManager: jwtManager,
		ListService: listServiceFunc(func(context.Context, streampublishsessionsservice.ListInput) (streampublishsessionsservice.ListOutput, error) {
			return streampublishsessionsservice.ListOutput{}, nil
		}),
		CreateService: createServiceFunc(func(context.Context, streampublishsessionsservice.CreateInput) (streampublishsessionsservice.CreateOutput, error) {
			return streampublishsessionsservice.CreateOutput{}, nil
		}),
		StartService: startServiceFunc(func(context.Context, streampublishsessionsservice.StartInput) (streampublishsessionsservice.StartOutput, error) {
			return streampublishsessionsservice.StartOutput{}, nil
		}),
		StopLiveService: stopLiveServiceFunc(func(context.Context, streampublishsessionsservice.StopLiveInput) (streampublishsessionsservice.StopLiveOutput, error) {
			return streampublishsessionsservice.StopLiveOutput{
				StoppedSessionIDs: []int64{1, 2},
				EndedAt:           time.Now().UTC(),
			}, nil
		}),
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/stream-publish-sessions/stop-all", strings.NewReader(""))
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, jwtManager))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusOK)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["status"] != "ended" {
		t.Fatalf("unexpected status payload: %v", body["status"])
	}
	if _, ok := body["ended_at"].(float64); !ok {
		t.Fatalf("expected ended_at to be a number, got %T", body["ended_at"])
	}
	if _, ok := body["stopped_session_ids"].([]any); !ok {
		t.Fatalf("expected stopped_session_ids to be array, got %T", body["stopped_session_ids"])
	}
}

func TestListRoute_Unauthorized(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	router := gin.New()
	v1 := router.Group("/api/v1")
	jwtManager := mustJWTManager(t)
	RegisterRoutes(v1, Deps{
		JWTManager: jwtManager,
		ListService: listServiceFunc(func(context.Context, streampublishsessionsservice.ListInput) (streampublishsessionsservice.ListOutput, error) {
			return streampublishsessionsservice.ListOutput{}, nil
		}),
		CreateService: createServiceFunc(func(context.Context, streampublishsessionsservice.CreateInput) (streampublishsessionsservice.CreateOutput, error) {
			return streampublishsessionsservice.CreateOutput{}, nil
		}),
		StartService: startServiceFunc(func(context.Context, streampublishsessionsservice.StartInput) (streampublishsessionsservice.StartOutput, error) {
			return streampublishsessionsservice.StartOutput{}, nil
		}),
		StopLiveService: stopLiveServiceFunc(func(context.Context, streampublishsessionsservice.StopLiveInput) (streampublishsessionsservice.StopLiveOutput, error) {
			return streampublishsessionsservice.StopLiveOutput{}, nil
		}),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/stream-publish-sessions?page=1&limit=10", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestListRoute_ReturnsPaginatedItems(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	router := gin.New()
	v1 := router.Group("/api/v1")
	jwtManager := mustJWTManager(t)
	RegisterRoutes(v1, Deps{
		JWTManager: jwtManager,
		ListService: listServiceFunc(func(context.Context, streampublishsessionsservice.ListInput) (streampublishsessionsservice.ListOutput, error) {
			return streampublishsessionsservice.ListOutput{
				Items: []streampublishsessionsservice.ListItem{
					{
						SessionID:      100,
						PlaybackID:     "playback-100",
						Title:          "demo",
						Status:         "created",
						PlaybackURLCDN: "http://localhost:8088/live/playback-100/master.m3u8",
						CreatedAt:      time.Unix(1700000000, 0).UTC(),
						Renditions: []streampublishsessionsservice.ListRenditionItem{
							{
								Name:         "720p",
								PlaylistPath: "playback-100/720p/index.m3u8",
								Status:       "ready",
							},
						},
					},
				},
				Page:  2,
				Limit: 5,
				Total: 11,
			}, nil
		}),
		CreateService: createServiceFunc(func(context.Context, streampublishsessionsservice.CreateInput) (streampublishsessionsservice.CreateOutput, error) {
			return streampublishsessionsservice.CreateOutput{}, nil
		}),
		StartService: startServiceFunc(func(context.Context, streampublishsessionsservice.StartInput) (streampublishsessionsservice.StartOutput, error) {
			return streampublishsessionsservice.StartOutput{}, nil
		}),
		StopLiveService: stopLiveServiceFunc(func(context.Context, streampublishsessionsservice.StopLiveInput) (streampublishsessionsservice.StopLiveOutput, error) {
			return streampublishsessionsservice.StopLiveOutput{}, nil
		}),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/stream-publish-sessions?page=2&limit=5", nil)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, jwtManager))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusOK)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["page"] != float64(2) {
		t.Fatalf("unexpected page: %v", body["page"])
	}
	if body["limit"] != float64(5) {
		t.Fatalf("unexpected limit: %v", body["limit"])
	}
	if body["total"] != float64(11) {
		t.Fatalf("unexpected total: %v", body["total"])
	}
	items, ok := body["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("unexpected items payload: %T %+v", body["items"], body["items"])
	}
	first, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("unexpected item type: %T", items[0])
	}
	renditions, ok := first["renditions"].([]any)
	if !ok || len(renditions) != 1 {
		t.Fatalf("unexpected renditions payload: %T %+v", first["renditions"], first["renditions"])
	}
}

type startServiceFunc func(context.Context, streampublishsessionsservice.StartInput) (streampublishsessionsservice.StartOutput, error)

type createServiceFunc func(context.Context, streampublishsessionsservice.CreateInput) (streampublishsessionsservice.CreateOutput, error)

type stopLiveServiceFunc func(context.Context, streampublishsessionsservice.StopLiveInput) (streampublishsessionsservice.StopLiveOutput, error)
type listServiceFunc func(context.Context, streampublishsessionsservice.ListInput) (streampublishsessionsservice.ListOutput, error)

func (f createServiceFunc) Execute(ctx context.Context, input streampublishsessionsservice.CreateInput) (streampublishsessionsservice.CreateOutput, error) {
	return f(ctx, input)
}

func (f startServiceFunc) Execute(ctx context.Context, input streampublishsessionsservice.StartInput) (streampublishsessionsservice.StartOutput, error) {
	return f(ctx, input)
}

func (f stopLiveServiceFunc) Execute(ctx context.Context, input streampublishsessionsservice.StopLiveInput) (streampublishsessionsservice.StopLiveOutput, error) {
	return f(ctx, input)
}

func (f listServiceFunc) Execute(ctx context.Context, input streampublishsessionsservice.ListInput) (streampublishsessionsservice.ListOutput, error) {
	return f(ctx, input)
}

func mustJWTManager(t *testing.T) *authpkg.JWTManager {
	t.Helper()
	jwtManager, err := authpkg.NewJWTManager("test-secret", "test-issuer")
	if err != nil {
		t.Fatalf("failed to create jwt manager: %v", err)
	}
	return jwtManager
}

func mustAccessToken(t *testing.T, jwtManager *authpkg.JWTManager) string {
	t.Helper()
	token, err := jwtManager.GenerateToken(123, "streamer@example.com", "streamer", time.Hour)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}
	return token
}
