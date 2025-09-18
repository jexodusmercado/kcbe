package store

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type password struct {
	plainText *string
	hash      []byte
}

func (p *password) Set(plainText string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainText), 12)
	if err != nil {
		return err
	}
	p.plainText = &plainText
	p.hash = hash
	return nil
}

func (p *password) Matches(plainText string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plainText))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

type User struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	PasswordHash   password  `json:"-"`
	Bio            string    `json:"bio,omitempty"`
	Role           string    `json:"role,omitempty"` // e.g., "admin", "user"
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

var AnonymousUser = &User{}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

type PostgresUserStore struct {
	db *sql.DB
}

func NewPostgresUserStore(db *sql.DB) *PostgresUserStore {
	return &PostgresUserStore{db: db}
}

type UserStore interface {
	CreateUser(user *User) (*User, error)
	GetUserByUsername(username string) (*User, error)
	UpdateUser(user *User) (*User, error)
	GetUserToken(scope, tokenPlaintext string) (*User, error)
}

func (pg *PostgresUserStore) CreateUser(user *User) (*User, error) {
	query := `
		INSERT INTO users (organization_id, username, email, password_hash, bio, role)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	err := pg.db.QueryRow(
		query,
		user.OrganizationID,
		user.Username,
		user.Email,
		user.PasswordHash.hash,
		user.Bio,
		user.Role,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (pg *PostgresUserStore) GetUserByUsername(username string) (*User, error) {
	user := &User{
		PasswordHash: password{},
	}

	query := `
		SELECT id, organization_id, username, email, password_hash, bio, role, created_at, updated_at
		FROM users
		WHERE username = $1
	`
	err := pg.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.OrganizationID,
		&user.Username,
		&user.Email,
		&user.PasswordHash.hash,
		&user.Bio,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // User not found
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (pg *PostgresUserStore) UpdateUser(user *User) (*User, error) {
	query := `
		UPDATE users
		SET email = $1, bio = $2, role = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
		RETURNING updated_at
	`
	results, err := pg.db.Exec(
		query,
		user.Email,
		user.Bio,
		user.Role,
		user.ID,
	)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := results.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, sql.ErrNoRows // User not found
	}

	err = pg.db.QueryRow(`SELECT updated_at FROM users WHERE id = $1`, user.ID).Scan(&user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (pg *PostgresUserStore) GetUserToken(scope, tokenPlaintext string) (*User, error) {

	query := `
		SELECT u.id, u.organization_id, u.username, u.email, u.password_hash, u.bio, u.role, u.created_at, u.updated_at
		FROM users u
		INNER JOIN tokens t ON t.user_id = u.id
		WHERE t.hash = $1 AND t.scope = $2 AND t.expiry > $3
	`
	user := &User{
		PasswordHash: password{},
	}

	err := pg.db.QueryRow(query, []byte(tokenPlaintext), scope, time.Now()).Scan(
		&user.ID,
		&user.OrganizationID,
		&user.Username,
		&user.Email,
		&user.PasswordHash.hash,
		&user.Bio,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Token not found or expired
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}
