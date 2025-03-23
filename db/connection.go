package db

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/redis/go-redis/v9"
)

type Connections struct {
    Postgres *pgxpool.Pool
    Redis    *redis.Client
}

func NewConnections(config *Config) (*Connections, error) {
    postgresURL := fmt.Sprintf(
        "postgresql://%s:%s@%s:%d/%s?sslmode=disable",
        config.Database.User,
        config.Database.Password,
        config.Database.Host,
        config.Database.Port,
        config.Database.DBName,
    )

    postgresPool, err := pgxpool.New(context.Background(), postgresURL)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
    }

    if err := postgresPool.Ping(context.Background()); err != nil {
        postgresPool.Close()
        return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
    }

    redisClient := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port),
        Password: config.Redis.Password,
        DB:       config.Redis.DB,
    })

    if err := redisClient.Ping(context.Background()).Err(); err != nil {
        postgresPool.Close()
        return nil, fmt.Errorf("failed to connect to Redis: %w", err)
    }

    return &Connections{
        Postgres: postgresPool,
        Redis:    redisClient,
    }, nil
}

func (c *Connections) Close(ctx context.Context) error {
    var errs []error

    if c.Postgres != nil {
        c.Postgres.Close()
    }

    if c.Redis != nil {
        if err := c.Redis.Close(); err != nil {
            errs = append(errs, fmt.Errorf("failed to close Redis: %w", err))
        }
    }

    if len(errs) > 0 {
        return fmt.Errorf("errors closing connections: %v", errs)
    }
    return nil
}

func (c *Connections) GetPostgres() *pgxpool.Pool {
    return c.Postgres
}

func (c *Connections) GetRedis() *redis.Client {
    return c.Redis
}