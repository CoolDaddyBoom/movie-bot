package storage

import (
	"strings"
)

func (m *Movie) GenresString() string {
	if len(m.Genres) == 0 {
		return ""
	}
	return strings.Join(m.Genres, ", ")
}

func (m *Movie) IsFromTMDB() bool {
	return m.TMDBID != 0
}
