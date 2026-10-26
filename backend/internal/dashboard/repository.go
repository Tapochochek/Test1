package dashboard

import (
	"context"
	"database/sql"
	"time"

	"orenkod/backend/internal/competitions"
	"orenkod/backend/internal/ratings"
)

type Stats struct {
	Athletes     int `json:"athletes"`
	Competitions int `json:"competitions"`
	Applications int `json:"applications"`
	Judges       int `json:"judges"`
}

type Snapshot struct {
	Stats             Stats                      `json:"stats"`
	CompetitionsTotal int                        `json:"competitions_total"`
	Competitions      []competitions.Competition `json:"competitions"`
	TopAthletes       []ratings.RatingEntry      `json:"top_athletes"`
	UpdatedAt         string                     `json:"updated_at"`
}

type Repository struct {
	db           *sql.DB
	competitions *competitions.Repository
	ratings      *ratings.Repository
}

func NewRepository(db *sql.DB, competitionsRepo *competitions.Repository, ratingsRepo *ratings.Repository) *Repository {
	return &Repository{db: db, competitions: competitionsRepo, ratings: ratingsRepo}
}

func (r *Repository) Snapshot(ctx context.Context) (Snapshot, error) {
	var athletes, totalCompetitions, activeCompetitions, participantSum int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM athletes`).Scan(&athletes); err != nil {
		return Snapshot{}, err
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM competitions`).Scan(&totalCompetitions); err != nil {
		return Snapshot{}, err
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM competitions WHERE status IN ('Регистрация','Планируется')`).Scan(&activeCompetitions); err != nil {
		return Snapshot{}, err
	}
	// participants is a stored count on competitions, not a separate applications entity.
	if err := r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(participants),0) FROM competitions WHERE status IN ('Регистрация','Планируется')`).Scan(&participantSum); err != nil {
		return Snapshot{}, err
	}
	items, err := r.competitions.List(ctx, true)
	if err != nil {
		return Snapshot{}, err
	}
	top, err := r.ratings.List(ctx, "", 5)
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{
		Stats: Stats{
			Athletes:     athletes,
			Competitions: activeCompetitions,
			Applications: participantSum,
			Judges:       24,
		},
		CompetitionsTotal: totalCompetitions,
		Competitions:      items,
		TopAthletes:       top,
		UpdatedAt:         time.Now().Format(time.RFC3339),
	}, nil
}
