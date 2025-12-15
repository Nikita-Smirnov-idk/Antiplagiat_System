package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Nikita-Smirnov-idk/antiplagiat-service/internal/domain"
	"github.com/Nikita-Smirnov-idk/antiplagiat-service/internal/infrastructure/repositories"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FileRepo struct {
	pool *pgxpool.Pool
}

func NewFileRepository(pool *pgxpool.Pool) *FileRepo {
	return &FileRepo{pool: pool}
}
