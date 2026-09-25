package database

import (
	"database/sql"
	"log"
)

func RunAutoMigration(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS activities (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),		
		title VARCHAR(255) NOT NULL,
		notes TEXT,
		is_completed BOOLEAN NOT NULL DEFAULT FALSE,
		category VARCHAR(50) NOT NULL DEFAULT 'General',
		priority VARCHAR(20) NOT NULL DEFAULT 'medium',
		estimated_minutes INT DEFAULT 0,
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		completed_at TIMESTAMPTZ NULL,
		deleted_at TIMESTAMPTZ NULL
	);

	CREATE INDEX IF NOT EXISTS idx_activities_status_created 
	ON activities (is_completed, created_at DESC) 
	WHERE deleted_at IS NULL;

	CREATE INDEX IF NOT EXISTS idx_activities_category 
	ON activities (category) 
	WHERE deleted_at IS NULL;
	`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	log.Println("Database migration completed successfully.")
	return nil
}
