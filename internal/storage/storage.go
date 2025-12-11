package storage

import (
	"context"
	"time"
)

// ContentType - тип контенту (фільм, серіал, аніме)
type ContentType string

const (
	ContentTypeMovie             ContentType = "movie"
	ContentTypeSeries            ContentType = "series"
	ContentTypeAnime             ContentType = "anime"
	ContentTypeAnimeSeries       ContentType = "anime_series"
	ContentTypeCartoon           ContentType = "cartoon"
	ContentTypeCartoonSeries     ContentType = "cartoon_series"
	ContentTypeDocumentary       ContentType = "documentary"
	ContentTypeDocumentarySeries ContentType = "documentary_series"
)

type Movie struct {
	ID          uint        `gorm:"primaryKey"`
	Title       string      `gorm:"size:255;not null"`
	ChatID      int         `gorm:"not null;index"`
	TMDBID      int         `gorm:"index"`
	Overview    string      `gorm:"type:text"`
	ReleaseDate string      `gorm:"size:10"`
	Rating      float64     `gorm:"type:decimal(3,1)"`
	PosterURL   string      `gorm:"size:500"`
	Genres      []string    `gorm:"type:jsonb;serializer:json"`
	ContentType ContentType `gorm:"type:content_type;default:movie"`
	CreatedAt   time.Time   `gorm:"autoCreateTime"`
	UpdatedAt   time.Time   `gorm:"autoUpdateTime"`
}

func (Movie) TableName() string {
	return "movies"
}

type Storage interface {
	Init(ctx context.Context) error
	Save(ctx context.Context, m *Movie) error
	PickRandom(ctx context.Context, chatID int) (*Movie, error)
	Remove(ctx context.Context, m *Movie) error
	List(ctx context.Context, chatID int) ([]*Movie, error)
	IsExists(ctx context.Context, m *Movie) (bool, error)
	Close() error
}
