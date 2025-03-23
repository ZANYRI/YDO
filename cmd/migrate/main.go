package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"ydo/db"
)

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found in any parent directory")
		}
		dir = parent
	}
}

type migrateConfig struct {
	configPath     string
	migrationsPath string
	command        string
}

func parseFlags(args []string) (migrateConfig, error) {
	fs := flag.NewFlagSet("migrate", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to config.yaml")
	migrationsPath := fs.String("migrations", "", "path to migrations directory")

	err := fs.Parse(args)
	if err != nil {
		return migrateConfig{}, fmt.Errorf("failed to parse flags: %w", err)
	}

	if fs.NArg() < 1 {
		return migrateConfig{}, fmt.Errorf("usage: migrate [--config path] [--migrations path] <up|down>")
	}

	command := fs.Arg(0)
	if command != "up" && command != "down" {
		return migrateConfig{}, fmt.Errorf("invalid command: must be 'up' or 'down'")
	}

	return migrateConfig{
		configPath:     *configPath,
		migrationsPath: *migrationsPath,
		command:        command,
	}, nil
}

func run(args []string) error {
	cfg, err := parseFlags(args)
	if err != nil {
		return err
	}

	root, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	defaultConfigPath := filepath.Join(root, "config.yaml")
	defaultMigrationsPath := filepath.Join(root, "db", "migrations")

	configPath := defaultConfigPath
	if cfg.configPath != "" {
		configPath = cfg.configPath
	}

	migrationsPath := defaultMigrationsPath
	if cfg.migrationsPath != "" {
		migrationsPath = cfg.migrationsPath
	}

	config, err := db.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.command == "up" {
		err = db.RunMigrations(config, migrationsPath)
		if err != nil {
			return fmt.Errorf("failed to apply migrations: %w", err)
		}
		fmt.Println("Migrations applied successfully")
	} else {
		err = db.RollbackMigrations(config, migrationsPath)
		if err != nil {
			return fmt.Errorf("failed to rollback migrations: %w", err)
		}
		fmt.Println("Migrations rolled back successfully")
	}

	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatalf("Error: %v", err)
	}
}