package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"demo-streaming/internal/config"
)

func main() {
	config.LoadDotEnv()
	systemCfg := config.LoadSystemConfig()

	var (
		action  = flag.String("action", "", "migration action: up|down|force|version|create")
		steps   = flag.Int("steps", 1, "steps used by down action")
		version = flag.Int("version", -1, "version used by force action")
		name    = flag.String("name", "", "name used by create action")
	)
	flag.Parse()

	switch *action {
	case "create":
		if err := createMigrationFiles(systemCfg.MigrationsDir, *name); err != nil {
			log.Fatalf("create migration files failed: %v", err)
		}
		log.Printf("created migration files for %q", *name)
		return
	case "up", "down", "force", "version":
		if err := runDatabaseMigration(systemCfg, *action, *steps, *version); err != nil {
			log.Fatalf("migration action %q failed: %v", *action, err)
		}
		return
	default:
		log.Fatalf("invalid action: %q", *action)
	}
}

func runDatabaseMigration(systemCfg config.SystemConfig, action string, steps int, forceVersion int) error {
	sourceURL := fmt.Sprintf("file://%s", systemCfg.MigrationsDir)
	m, err := migrate.New(sourceURL, config.DatabaseURL(systemCfg))
	if err != nil {
		return err
	}
	defer func() {
		_, _ = m.Close()
	}()

	switch action {
	case "up":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
		log.Println("migration up completed")
	case "down":
		if err := m.Steps(-steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
		log.Printf("migration down completed (%d step)", steps)
	case "force":
		if forceVersion < 0 {
			return errors.New("force action requires -version")
		}
		if err := m.Force(forceVersion); err != nil {
			return err
		}
		log.Printf("migration force completed (version=%d)", forceVersion)
	case "version":
		v, dirty, err := m.Version()
		if errors.Is(err, migrate.ErrNilVersion) {
			log.Println("current migration version: nil (no migration applied yet)")
			return nil
		}
		if err != nil {
			return err
		}
		log.Printf("current migration version: %d (dirty=%v)", v, dirty)
	default:
		return fmt.Errorf("unsupported action: %s", action)
	}

	return nil
}
