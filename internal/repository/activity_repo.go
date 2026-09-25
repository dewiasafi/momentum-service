package repository

import (
	"database/sql"
	"errors"

	"project.com/momentum-service/internal/model"
)

type ActivityRepository struct {
	DB *sql.DB
}

func NewActivityRepository(db *sql.DB) *ActivityRepository {
	return &ActivityRepository{DB: db}
}

func (r *ActivityRepository) GetAll() ([]model.Activity, error) {
	rows, err := r.DB.Query("SELECT id, title, is_completed, COALESCE(notes, '') FROM activities ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []model.Activity{}
	for rows.Next() {
		var a model.Activity
		if err := rows.Scan(&a.ID, &a.Title, &a.IsCompleted, &a.Notes); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (r *ActivityRepository) Create(act *model.Activity) error {
	query := `INSERT INTO activities (title, notes) VALUES ($1, $2) RETURNING id, is_completed`
	return r.DB.QueryRow(query, act.Title, act.Notes).Scan(&act.ID, &act.IsCompleted)
}

func (r *ActivityRepository) Update(id string, act *model.Activity) error {
	query := `UPDATE activities SET title = $1, is_completed = $2, notes = $3 WHERE id = $4`
	result, err := r.DB.Exec(query, act.Title, act.IsCompleted, act.Notes, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("activity not found")
	}

	return nil
}

func (r *ActivityRepository) Delete(id string) error {
	query := `DELETE FROM activities WHERE id = $1`
	result, err := r.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("activity not found")
	}

	return nil
}
