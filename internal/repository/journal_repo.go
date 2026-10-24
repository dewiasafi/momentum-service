package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"project.com/momentum-service/internal/model"
)

type JournalRepository struct {
	DB *sql.DB
}

func NewJournalRepository(db *sql.DB) *JournalRepository {
	return &JournalRepository{DB: db}
}

var ErrJournalNotFound = errors.New("Jurnal tidak ditemukan")

func (r *JournalRepository) Create(
	ctx context.Context,
	userID string,
	journal *model.Journal,
) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, `
		INSERT INTO journals (user_id, title, content)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`,
		userID,
		strings.TrimSpace(journal.Title),
		journal.Content,
	).Scan(
		&journal.ID,
		&journal.CreatedAt,
		&journal.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("membuat jurnal: %w", err)
	}

	journal.UserID = userID

	_, err = tx.ExecContext(ctx, `
		INSERT INTO journal_versions
			(journal_id, version_number, title, content)
		VALUES ($1, 1, $2, $3)
	`,
		journal.ID,
		journal.Title,
		journal.Content,
	)
	if err != nil {
		return fmt.Errorf("menyimpan versi awal: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit jurnal: %w", err)
	}

	return nil
}

func (r *JournalRepository) GetByID(
	ctx context.Context,
	userID string,
	journalID string,
) (*model.Journal, error) {
	journal := &model.Journal{}

	err := r.DB.QueryRowContext(ctx, `
		SELECT id, user_id, title, content, created_at, updated_at
		FROM journals
		WHERE id = $1
		  AND user_id = $2
		  AND deleted_at IS NULL
	`, journalID, userID).Scan(
		&journal.ID,
		&journal.UserID,
		&journal.Title,
		&journal.Content,
		&journal.CreatedAt,
		&journal.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrJournalNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("mengambil jurnal: %w", err)
	}

	categories, err := r.getJournalCategories(ctx, r.DB, journal.ID)
	if err != nil {
		return nil, err
	}

	journal.Categories = categories
	return journal, nil
}

func (r *JournalRepository) List(
	ctx context.Context,
	userID string,
	limit int,
	offset int,
) ([]model.Journal, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, user_id, title, content, created_at, updated_at
		FROM journals
		WHERE user_id = $1
		  AND deleted_at IS NULL
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("daftar jurnal: %w", err)
	}
	defer rows.Close()

	journals := make([]model.Journal, 0)

	for rows.Next() {
		var j model.Journal

		if err := rows.Scan(
			&j.ID,
			&j.UserID,
			&j.Title,
			&j.Content,
			&j.CreatedAt,
			&j.UpdatedAt,
		); err != nil {
			return nil, err
		}

		journals = append(journals, j)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range journals {
		categories, err := r.getJournalCategories(
			ctx, r.DB, journals[i].ID,
		)
		if err != nil {
			return nil, err
		}
		journals[i].Categories = categories
	}

	return journals, nil
}

func (r *JournalRepository) Update(
	ctx context.Context,
	userID string,
	journalID string,
	journal *model.Journal,
	categoryIDs []string,
) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var lockedID string

	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM journals
		WHERE id = $1
		  AND user_id = $2
		  AND deleted_at IS NULL
		FOR UPDATE
	`, journalID, userID).Scan(&lockedID)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrJournalNotFound
	}
	if err != nil {
		return fmt.Errorf("mengunci jurnal: %w", err)
	}

	var version int

	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version_number), 0) + 1
		FROM journal_versions
		WHERE journal_id = $1
	`, journalID).Scan(&version)

	if err != nil {
		return fmt.Errorf("mengambil nomor versi: %w", err)
	}

	title := strings.TrimSpace(journal.Title)

	_, err = tx.ExecContext(ctx, `
		UPDATE journals
		SET title = $1,
		    content = $2,
		    updated_at = NOW()
		WHERE id = $3
		  AND user_id = $4
		  AND deleted_at IS NULL
	`, title, journal.Content, journalID, userID)

	if err != nil {
		return fmt.Errorf("memperbarui jurnal: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO journal_versions (
			journal_id,
			version_number,
			title,
			content
		)
		VALUES ($1, $2, $3, $4)
	`, journalID, version, title, journal.Content)

	if err != nil {
		return fmt.Errorf("menyimpan versi jurnal: %w", err)
	}

	if categoryIDs != nil {
		_, err = tx.ExecContext(ctx, `
			DELETE FROM journal_categories
			WHERE journal_id = $1
			  AND source = 'manual'
		`, journalID)

		if err != nil {
			return fmt.Errorf("menghapus kategori manual lama: %w", err)
		}

		for _, categoryID := range categoryIDs {
			result, err := tx.ExecContext(ctx, `
				INSERT INTO journal_categories (
					journal_id,
					category_id,
					source
				)
				SELECT $1, id, 'manual'
				FROM categories
				WHERE id = $2
				ON CONFLICT (journal_id, category_id)
				DO UPDATE SET source = 'manual'
			`, journalID, categoryID)

			if err != nil {
				return fmt.Errorf("menyimpan kategori: %w", err)
			}

			affected, err := result.RowsAffected()
			if err != nil {
				return err
			}
			if affected == 0 {
				return fmt.Errorf("kategori tidak ditemukan: %s", categoryID)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit pembaruan jurnal: %w", err)
	}

	return nil
}

func (r *JournalRepository) Delete(
	ctx context.Context,
	userID string,
	journalID string,
) error {
	result, err := r.DB.ExecContext(ctx, `
		UPDATE journals
		SET deleted_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1
		  AND user_id = $2
		  AND deleted_at IS NULL
	`, journalID, userID)
	if err != nil {
		return fmt.Errorf("menghapus jurnal: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrJournalNotFound
	}

	return nil
}

type journalQueryer interface {
	QueryContext(
		context.Context,
		string,
		...any,
	) (*sql.Rows, error)
}

func (r *JournalRepository) getJournalCategories(
	ctx context.Context,
	q journalQueryer,
	journalID string,
) ([]model.JournalCategory, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT c.id, c.name, c.slug, jc.source
		FROM journal_categories jc
		JOIN categories c ON c.id = jc.category_id
		WHERE jc.journal_id = $1
		ORDER BY c.name
	`, journalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]model.JournalCategory, 0)

	for rows.Next() {
		var c model.JournalCategory
		if err := rows.Scan(
			&c.ID, &c.Name, &c.Slug, &c.Source,
		); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *JournalRepository) ListVersions(
	ctx context.Context,
	userID string,
	journalID string,
) ([]model.JournalVersion, error) {
	var exists bool

	err := r.DB.QueryRowContext(ctx, `
        SELECT EXISTS (
            SELECT 1
            FROM journals
            WHERE id = $1
              AND user_id = $2
              AND deleted_at IS NULL
        )
    `, journalID, userID).Scan(&exists)

	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrJournalNotFound
	}

	rows, err := r.DB.QueryContext(ctx, `
        SELECT id, journal_id, version_number,
               title, content, created_at
        FROM journal_versions
        WHERE journal_id = $1
        ORDER BY version_number DESC
    `, journalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := make([]model.JournalVersion, 0)

	for rows.Next() {
		var v model.JournalVersion

		if err := rows.Scan(
			&v.ID,
			&v.JournalID,
			&v.VersionNumber,
			&v.Title,
			&v.Content,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}

		versions = append(versions, v)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return versions, nil
}

func (r *JournalRepository) ListCategories(
	ctx context.Context,
) ([]model.Category, error) {
	rows, err := r.DB.QueryContext(ctx, `
        SELECT id, name, slug, is_system
        FROM categories
        ORDER BY name
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]model.Category, 0)

	for rows.Next() {
		var c model.Category

		if err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.Slug,
			&c.IsSystem,
		); err != nil {
			return nil, err
		}

		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *JournalRepository) ApplyAICategories(
	ctx context.Context,
	userID string,
	journalID string,
	categoryIDs []string,
) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists bool

	err = tx.QueryRowContext(ctx, `
        SELECT EXISTS (
            SELECT 1
            FROM journals
            WHERE id = $1
              AND user_id = $2
              AND deleted_at IS NULL
        )
    `, journalID, userID).Scan(&exists)

	if err != nil {
		return err
	}
	if !exists {
		return ErrJournalNotFound
	}

	for _, categoryID := range categoryIDs {
		_, err := tx.ExecContext(ctx, `
            INSERT INTO journal_categories
                (journal_id, category_id, source)
            SELECT $1, id, 'ai'
            FROM categories
            WHERE id = $2
            ON CONFLICT (journal_id, category_id)
            DO UPDATE SET source = 'ai'
            WHERE journal_categories.source <> 'manual'
        `, journalID, categoryID)

		if err != nil {
			return fmt.Errorf(
				"menerapkan kategori AI: %w", err,
			)
		}
	}

	return tx.Commit()
}
