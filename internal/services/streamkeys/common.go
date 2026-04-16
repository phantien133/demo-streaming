package streamkeys

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

const (
	streamKeyRedisPrefix = "stream:key:"
	defaultStreamKeyTTL  = 24 * time.Hour
)

var (
	ErrStreamKeyNotFound = errors.New("stream key not found")
	ErrStreamKeyStore    = errors.New("stream key store error")
)

func redisKey(userID int64) string {
	return fmt.Sprintf("%s%d", streamKeyRedisPrefix, userID)
}

// GenerateStreamKey returns a high-entropy secret for RTMP ingest (stream name on SRS).
func GenerateStreamKey() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
