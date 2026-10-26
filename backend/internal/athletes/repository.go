package athletes

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context) ([]Athlete, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,full_name,organization,discipline,age_group,points,coach_name,status FROM athletes ORDER BY full_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Athlete{}
	for rows.Next() {
		var a Athlete
		if err := rows.Scan(&a.ID, &a.FullName, &a.Organization, &a.Discipline, &a.AgeGroup, &a.Points, &a.CoachName, &a.Status); err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	return items, rows.Err()
}
