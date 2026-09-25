package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"project.com/momentum-service/internal/model"
)

type ActivityRepository struct {
	DB *sql.DB
}

type ActivityFilter struct {
	Page        int
	Limit       int
	Category    string
	Priority    string
	IsCompleted *bool
	Search      string
	StartDate   *time.Time
	EndDate     *time.Time
	SortBy      string
	SortOrder   string
}

func NewActivityRepository(db *sql.DB) *ActivityRepository {
	return &ActivityRepository{DB: db}
}

var ErrNotFound = errors.New("Aktivitas tidak ditemukan")

func (r *ActivityRepository) GetAllFiltered(filter ActivityFilter) ([]model.Activity, int, error) {
	var conditions []string
	var args []interface{}
	argID := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if filter.Category != "" {
		conditions = append(conditions, fmt.Sprintf("category ILIKE $%d", argID))
		args = append(args, filter.Category)
		argID++
	}

	if filter.Priority != "" {
		conditions = append(conditions, fmt.Sprintf("priority ILIKE $%d", argID))
		args = append(args, filter.Priority)
		argID++
	}

	if filter.IsCompleted != nil {
		conditions = append(conditions, fmt.Sprintf("is_completed = $%d", argID))
		args = append(args, *filter.IsCompleted)
		argID++
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR COALESCE(notes, '') ILIKE $%d)", argID, argID))
		args = append(args, "%"+filter.Search+"%")
		argID++
	}

	if filter.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argID))
		args = append(args, *filter.StartDate)
		argID++
	}

	if filter.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argID))
		args = append(args, *filter.EndDate)
		argID++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	sortBy := "created_at"
	switch strings.ToLower(filter.SortBy) {
	case "title", "category", "priority", "estimated_minutes", "is_completed":
		sortBy = strings.ToLower(filter.SortBy)
	}

	sortOrder := "DESC"
	if strings.ToLower(filter.SortOrder) == "asc" {
		sortOrder = "ASC"
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM activities %s", whereClause)
	var totalItems int
	err := r.DB.QueryRow(countQuery, args...).Scan(&totalItems)
	if err != nil {
		return nil, 0, fmt.Errorf("count query failed: %w", err)
	}

	offset := (filter.Page - 1) * filter.Limit
	dataQuery := fmt.Sprintf(`
		SELECT 
			id, 
			title, 
			COALESCE(notes, '') AS notes, 
			is_completed, 
			COALESCE(category, 'General') AS category, 
			COALESCE(priority, 'medium') AS priority, 
			COALESCE(estimated_minutes, 0) AS estimated_minutes, 
			created_at, 
			completed_at
		FROM activities
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortBy, sortOrder, argID, argID+1)

	dataArgs := append(args, filter.Limit, offset)

	rows, err := r.DB.Query(dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	activities := make([]model.Activity, 0)
	for rows.Next() {
		var act model.Activity
		err := rows.Scan(
			&act.ID,
			&act.Title,
			&act.Notes,
			&act.IsCompleted,
			&act.Category,
			&act.Priority,
			&act.EstimatedMinutes,
			&act.CreatedAt,
			&act.CompletedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		activities = append(activities, act)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return activities, totalItems, nil
}

func (r *ActivityRepository) GetByID(id string) (*model.Activity, error) {
	query := `
		SELECT 
			id, 
			title, 
			COALESCE(notes, '') AS notes, 
			is_completed, 
			COALESCE(category, 'General') AS category, 
			COALESCE(priority, 'medium') AS priority, 
			COALESCE(estimated_minutes, 0) AS estimated_minutes, 
			created_at, 
			completed_at
		FROM activities
		WHERE id = $1 AND deleted_at IS NULL
	`
	var act model.Activity
	err := r.DB.QueryRow(query, id).Scan(
		&act.ID,
		&act.Title,
		&act.Notes,
		&act.IsCompleted,
		&act.Category,
		&act.Priority,
		&act.EstimatedMinutes,
		&act.CreatedAt,
		&act.CompletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err

	}

	return &act, nil
}

func (r *ActivityRepository) Create(act *model.Activity) error {
	query := `
		INSERT INTO activities (title, notes, category, priority, estimated_minutes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, is_completed, created_at
	`
	err := r.DB.QueryRow(
		query,
		act.Title,
		act.Notes,
		act.Category,
		act.Priority,
		act.EstimatedMinutes,
	).Scan(&act.ID, &act.IsCompleted, &act.CreatedAt)

	return err
}

func (r *ActivityRepository) Update(id string, act *model.Activity) error {
	var completedAt *time.Time
	if act.IsCompleted {
		now := time.Now()
		completedAt = &now
	}

	query := `
		UPDATE activities 
		SET title = $1, 
			notes = $2, 
			is_completed = $3, 
			category = $4, 
			priority = $5, 
			estimated_minutes = $6, 
			completed_at = $7	
		WHERE id = $8 AND deleted_at IS NULL
		RETURNING created_at
	`

	err := r.DB.QueryRow(
		query,
		act.Title,
		act.Notes,
		act.IsCompleted,
		act.Category,
		act.Priority,
		act.EstimatedMinutes,
		completedAt,
		id,
	).Scan(&act.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	act.CompletedAt = completedAt
	return nil
}

func (r *ActivityRepository) Delete(id string) error {
	query := `
		UPDATE activities 
		SET deleted_at = $1 
		WHERE id = $2 
		AND deleted_at IS NULL
	`

	result, err := r.DB.Exec(query, time.Now(), id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return ErrNotFound
	}

	return nil
}
