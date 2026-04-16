package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func createMigrationFiles(dir, name string) error {
	if name == "" {
		return errors.New("create action requires -name")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	ts := time.Now().UTC().Format("20060102150405")
	upPath := filepath.Join(dir, fmt.Sprintf("%s_%s.up.sql", ts, name))
	downPath := filepath.Join(dir, fmt.Sprintf("%s_%s.down.sql", ts, name))

	upContent := []byte("-- Write UP migration here.\n")
	downContent := []byte("-- Write DOWN migration here.\n")

	if err := os.WriteFile(upPath, upContent, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(downPath, downContent, 0o644); err != nil {
		return err
	}
	return nil
}
