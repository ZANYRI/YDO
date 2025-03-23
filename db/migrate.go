package db

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func buildPostgresURL(config *Config) string {
    return fmt.Sprintf(
        "postgresql://%s:%s@%s:%d/%s?sslmode=disable",
        config.Database.User,
        config.Database.Password,
        config.Database.Host,
        config.Database.Port,
        config.Database.DBName,
    )
}

func formatMigrationSourcePath(migrationsPath string) (string, error) {

    absPath, err := filepath.Abs(migrationsPath)
    if err != nil {
        return "", fmt.Errorf("failed to get absolute path: %w", err)
    }

    stat, err := os.Stat(absPath)
    if err != nil {
        return "", fmt.Errorf("migrations path does not exist: %w", err)
    }
    if !stat.IsDir() {
        return "", fmt.Errorf("migrations path is not a directory: %s", absPath)
    }

    cleanPath := filepath.ToSlash(absPath)

    return fmt.Sprintf("file://%s", cleanPath), nil
}

func RunMigrations(config *Config, migrationsPath string) error {

    sourceURL, err := formatMigrationSourcePath(migrationsPath)
    if err != nil {
        return fmt.Errorf("failed to format migrations path: %w", err)
    }

    m, err := migrate.New(
        sourceURL,
        buildPostgresURL(config),
    )
    if err != nil {
        return fmt.Errorf("failed to create migrate instance: %w", err)
    }
    defer func() {
        sourceErr, dbErr := m.Close()
        if sourceErr != nil {
            err = fmt.Errorf("failed to close migrate source: %w", sourceErr)
        }
        if dbErr != nil {
            err = fmt.Errorf("failed to close migrate database: %w", dbErr)
        }
    }()

    if err := m.Up(); err != nil {
        return fmt.Errorf("failed to run migrations: %w", err)
    }

    return nil
}

func RollbackMigrations(config *Config, migrationsPath string) error {

    sourceURL, err := formatMigrationSourcePath(migrationsPath)
    if err != nil {
        return fmt.Errorf("failed to format migrations path: %w", err)
    }

    m, err := migrate.New(
        sourceURL,
        buildPostgresURL(config),
    )
    if err != nil {
        return fmt.Errorf("failed to create migrate instance: %w", err)
    }
    defer func() {
        sourceErr, dbErr := m.Close()
        if sourceErr != nil {
            err = fmt.Errorf("failed to close migrate source: %w", sourceErr)
        }
        if dbErr != nil {
            err = fmt.Errorf("failed to close migrate database: %w", dbErr)
        }
    }()

    if err := m.Down(); err != nil {
        return fmt.Errorf("failed to rollback migrations: %w", err)
    }

    return nil
}