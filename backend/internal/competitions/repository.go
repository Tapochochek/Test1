package competitions

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, upcomingOnly bool) ([]Competition, error) {
	query := `SELECT id,title,discipline,competition_type,age_group,starts_at,location,status,participants FROM competitions ORDER BY (starts_at < NOW()) ASC, starts_at ASC`
	if upcomingOnly {
		query = `SELECT id,title,discipline,competition_type,age_group,starts_at,location,status,participants FROM competitions WHERE starts_at >= NOW() - INTERVAL '1 day' ORDER BY starts_at ASC LIMIT 4`
	}
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Competition{}
	for rows.Next() {
		c, err := scanCompetition(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func (r *Repository) Create(ctx context.Context, input CreateInput, starts time.Time) (Competition, error) {
	var c Competition
	var stored time.Time
	err := r.db.QueryRowContext(ctx, `INSERT INTO competitions(title,discipline,competition_type,age_group,starts_at,location,status,participants)
		VALUES($1,$2,$3,$4,$5,$6,'Черновик',0) RETURNING id,title,discipline,competition_type,age_group,starts_at,location,status,participants`,
		input.Title, input.Discipline, input.CompetitionType, input.AgeGroup, starts, input.Location).
		Scan(&c.ID, &c.Title, &c.Discipline, &c.CompetitionType, &c.AgeGroup, &stored, &c.Location, &c.Status, &c.Participants)
	if err != nil {
		return Competition{}, err
	}
	c.StartsAt = stored.Format(time.RFC3339)
	return c, nil
}

func IsDuplicateTitle(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "duplicate key")
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanCompetition(row rowScanner) (Competition, error) {
	var c Competition
	var starts time.Time
	if err := row.Scan(&c.ID, &c.Title, &c.Discipline, &c.CompetitionType, &c.AgeGroup, &starts, &c.Location, &c.Status, &c.Participants); err != nil {
		return Competition{}, err
	}
	c.StartsAt = starts.Format(time.RFC3339)
	return c, nil
}
