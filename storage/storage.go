package storage

import "context"

type Movie struct {
	Title  string
	ChatID int
}

type Storage interface {
	Init(ctx context.Context) error
	Save(ctx context.Context, m *Movie) error
	PickRandom(ctx context.Context, chatID int) (*Movie, error)
	Remove(ctx context.Context, m *Movie) error
	List(ctx context.Context, chatID int) ([]*Movie, error)
	IsExists(ctx context.Context, m *Movie) (bool, error)
}
