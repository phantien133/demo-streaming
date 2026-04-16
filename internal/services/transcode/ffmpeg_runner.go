package transcode

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"demo-streaming/internal/config"
)

type Runner struct {
	cfg config.AppConfig
}

func NewRunner(cfg config.AppConfig) *Runner {
	return &Runner{cfg: cfg}
}

func transcodeInputURL(cfg config.AppConfig, job PublishJob) (inputURL string, waitHLS bool, err error) {
	rtmpBase := strings.TrimSpace(cfg.TranscodeInputRTMPBaseURL)
	if rtmpBase != "" {
		return strings.TrimRight(rtmpBase, "/") + "/" + job.StreamKeySecret, false, nil
	}
	inputBase := strings.TrimRight(strings.TrimSpace(cfg.TranscodeInputBaseURL), "/")
	if inputBase == "" {
		return "", false, fmt.Errorf("transcode input base url is empty")
	}
	return fmt.Sprintf("%s/%s.m3u8", inputBase, job.StreamKeySecret), true, nil
}

func waitForHLSPlaylist(ctx context.Context, playlistURL string, maxWait time.Duration) error {
	if maxWait <= 0 {
		maxWait = 90 * time.Second
	}
	deadline := time.Now().Add(maxWait)
	client := &http.Client{Timeout: 8 * time.Second}
	tick := 2 * time.Second

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, playlistURL, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			return nil
		}
		if resp != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("hls playlist not ready after %s: %s", maxWait, playlistURL)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(tick):
		}
	}
}

func (r *Runner) Execute(ctx context.Context, job PublishJob) (TranscodeResult, error) {
	if !r.cfg.TranscodeEnabled {
		return TranscodeResult{}, nil
	}
	inputURL, waitHLS, err := transcodeInputURL(r.cfg, job)
	if err != nil {
		return TranscodeResult{}, err
	}
	if waitHLS {
		maxWait := time.Duration(r.cfg.TranscodeInputWaitMaxSeconds) * time.Second
		if err := waitForHLSPlaylist(ctx, inputURL, maxWait); err != nil {
			return TranscodeResult{}, err
		}
	}
	outputRoot := strings.TrimSpace(r.cfg.TranscodeOutputDir)
	if outputRoot == "" {
		return TranscodeResult{}, fmt.Errorf("transcode output directory is empty")
	}
	outputDir := filepath.Join(outputRoot, job.PlaybackID)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return TranscodeResult{}, err
	}

	// Minimal ABR output (720p + 480p) for local demo verification.
	outputPattern := filepath.Join(outputDir, "%v", "index.m3u8")
	if err := os.MkdirAll(filepath.Join(outputDir, "720p"), 0o755); err != nil {
		return TranscodeResult{}, err
	}
	if err := os.MkdirAll(filepath.Join(outputDir, "480p"), 0o755); err != nil {
		return TranscodeResult{}, err
	}

	// Two video renditions + duplicate audio via asplit. Reusing the same output audio index in
	// -var_stream_map for both HLS variants makes the muxer error:
	// "Same elementary stream found more than once in two different variant definitions".
	filterComplex := strings.Join([]string{
		"[0:v]split=2[v720in][v480in]",
		"[v720in]scale=w=1280:h=720:force_original_aspect_ratio=decrease:force_divisible_by=2[v720]",
		"[v480in]scale=w=854:h=480:force_original_aspect_ratio=decrease:force_divisible_by=2[v480]",
		"[0:a]asplit=2[a720][a480]",
	}, ";")

	cmd := exec.CommandContext(
		ctx,
		r.cfg.TranscodeFFmpegBin,
		"-hide_banner",
		"-y",
		"-i", inputURL,
		"-filter_complex", filterComplex,
		"-map", "[v720]", "-map", "[a720]",
		"-map", "[v480]", "-map", "[a480]",
		"-c:v", "libx264",
		"-preset", r.cfg.TranscodePreset,
		"-c:a", "aac",
		"-ar", "48000",
		"-b:a", "128k",
		"-b:v:0", "2800k",
		"-maxrate:v:0", "2996k",
		"-bufsize:v:0", "4200k",
		"-b:v:1", "1400k",
		"-maxrate:v:1", "1498k",
		"-bufsize:v:1", "2100k",
		"-f", "hls",
		"-hls_time", "4",
		"-hls_list_size", "0",
		"-hls_flags", "independent_segments",
		"-master_pl_name", "master.m3u8",
		// v:N / a:N are per-type indices (1st video, 1st audio, …), not global output stream indices.
		"-var_stream_map", "v:0,a:0,name:720p v:1,a:1,name:480p",
		outputPattern,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return TranscodeResult{}, err
	}

	return TranscodeResult{
		Renditions: []RenditionOutput{
			{Name: "720p", PlaylistPath: filepath.ToSlash(filepath.Join(job.PlaybackID, "720p", "index.m3u8")), Status: "ready"},
			{Name: "480p", PlaylistPath: filepath.ToSlash(filepath.Join(job.PlaybackID, "480p", "index.m3u8")), Status: "ready"},
		},
	}, nil
}
