package repository

import (
	"context"
	"database/sql"
	"errors"

	"project.com/momentum-service/internal/model"
)

type AuthRepository interface {
	CreateUserWithVerificationToken(ctx context.Context, user *model.User, token *model.VerificationToken) error
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error

	CreateVerificationToken(ctx context.Context, token *model.VerificationToken) error
	GetVerificationToken(ctx context.Context, tokenHash string) (*model.VerificationToken, error)
	MarkEmailVerified(ctx context.Context, userID string, tokenID string) error

	CreateSession(ctx context.Context, session *model.Session) error
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (*model.Session, error)
	DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error

	CreatePasswordResetToken(ctx context.Context, token *model.PasswordResetToken) error
	GetPasswordResetToken(ctx context.Context, tokenHash string) (*model.PasswordResetToken, error)
	ResetPassword(ctx context.Context, userID string, tokenID string, hashedPassword string) error
}

type authRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, email, password_hash, display_name, email_verified_at, created_at, updated_at FROM users WHERE email = $1`
	var user model.User
	err := r.db.QueryRowContext(ctx, query, email).
		Scan(&user.ID, &user.Email, &user.Password, &user.Name, &user.EmailVerifiedAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	query := `SELECT id, email, password_hash, display_name, email_verified_at, created_at, updated_at FROM users WHERE id = $1`
	var user model.User
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&user.ID, &user.Email, &user.Password, &user.Name, &user.EmailVerifiedAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) UpdateUser(ctx context.Context, user *model.User) error {
	query := `UPDATE users SET display_name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, user.Name, user.ID)
	return err
}

func (r *authRepository) CreateVerificationToken(ctx context.Context, token *model.VerificationToken) error {
	query := `
		INSERT INTO verification_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(ctx, query, token.UserID, token.TokenHash, token.ExpiresAt).
		Scan(&token.ID, &token.CreatedAt)
}

func (r *authRepository) GetVerificationToken(ctx context.Context, tokenHash string) (*model.VerificationToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM verification_tokens
		WHERE token_hash = $1
	`
	var token model.VerificationToken
	err := r.db.QueryRowContext(ctx, query, tokenHash).
		Scan(&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.UsedAt, &token.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &token, nil
}

func (r *authRepository) MarkEmailVerified(ctx context.Context, userID string, tokenID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updateUserQuery := `UPDATE users SET email_verified_at = CURRENT_TIMESTAMP WHERE id = $1`
	if _, err := tx.ExecContext(ctx, updateUserQuery, userID); err != nil {
		return err
	}

	updateTokenQuery := `UPDATE verification_tokens SET used_at = CURRENT_TIMESTAMP WHERE id = $1`
	if _, err := tx.ExecContext(ctx, updateTokenQuery, tokenID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *authRepository) CreateSession(ctx context.Context, session *model.Session) error {
	query := `
		INSERT INTO sessions (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := r.db.QueryRowContext(ctx, query, session.UserID, session.TokenHash, session.ExpiresAt).
		Scan(&session.ID, &session.CreatedAt)
	return err
}

func (r *authRepository) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*model.Session, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at 
		FROM sessions 
		WHERE token_hash = $1 AND expires_at > NOW()
	`
	var session model.Session
	err := r.db.QueryRowContext(ctx, query, tokenHash).
		Scan(&session.ID, &session.UserID, &session.TokenHash, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &session, nil
}

func (r *authRepository) DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM sessions WHERE token_hash = $1`
	_, err := r.db.ExecContext(ctx, query, tokenHash)
	return err
}

func (r *authRepository) CreateUserWithVerificationToken(
	ctx context.Context,
	user *model.User,
	token *model.VerificationToken,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queryUser := `
        INSERT INTO users (email, password_hash, display_name)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at
    `

	err = tx.QueryRowContext(
		ctx,
		queryUser,
		user.Email,
		user.Password,
		user.Name,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return err
	}

	token.UserID = user.ID

	queryToken := `
        INSERT INTO verification_tokens (user_id, token_hash, expires_at)
        VALUES ($1, $2, $3)
        RETURNING id, created_at
    `

	err = tx.QueryRowContext(
		ctx,
		queryToken,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
	).Scan(&token.ID, &token.CreatedAt)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *authRepository) CreatePasswordResetToken(
	ctx context.Context,
	token *model.PasswordResetToken,
) error {
	query := `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	return r.db.QueryRowContext(
		ctx,
		query,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
	).Scan(&token.ID, &token.CreatedAt)
}

func (r *authRepository) GetPasswordResetToken(
	ctx context.Context,
	tokenHash string,
) (*model.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1
	`

	var token model.PasswordResetToken

	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.UsedAt,
		&token.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &token, nil
}

func (r *authRepository) ResetPassword(
	ctx context.Context,
	userID string,
	tokenID string,
	hashedPassword string,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var lockedTokenID string

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT id
		FROM password_reset_tokens
		WHERE id = $1
		  AND user_id = $2
		  AND used_at IS NULL
		  AND expires_at > NOW()
		FOR UPDATE
		`,
		tokenID,
		userID,
	).Scan(&lockedTokenID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sql.ErrNoRows
		}
		return err
	}

	result, err := tx.ExecContext(
		ctx,
		`
		UPDATE users
		SET password_hash = $1,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		`,
		hashedPassword,
		userID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	result, err = tx.ExecContext(
		ctx,
		`
		UPDATE password_reset_tokens
		SET used_at = CURRENT_TIMESTAMP
		WHERE id = $1
		  AND used_at IS NULL
		`,
		tokenID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	_, err = tx.ExecContext(
		ctx,
		`DELETE FROM sessions WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
