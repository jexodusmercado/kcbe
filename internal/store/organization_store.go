package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PostgresOrganizationStore struct {
	db *sql.DB
}

func NewPostgresOrganizationStore(db *sql.DB) *PostgresOrganizationStore {
	return &PostgresOrganizationStore{db: db}
}

type OrganizationStore interface {
	CreateOrganization(org *Organization) (*Organization, error)
	GetOrganizationByID(id uuid.UUID) (*Organization, error)
	UpdateOrganization(org *Organization) (*Organization, error)
	DeleteOrganization(id uuid.UUID) error
}

func (pg *PostgresOrganizationStore) CreateOrganization(org *Organization) (*Organization, error) {
	tx, err := pg.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO organizations (name)
		VALUES ($1)
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRow(
		query,
		org.Name,
	).Scan(
		&org.ID,
		&org.CreatedAt,
		&org.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return org, nil
}

func (pg *PostgresOrganizationStore) GetOrganizationByID(id uuid.UUID) (*Organization, error) {
	org := &Organization{}
	query := `
		SELECT id, name, created_at, updated_at
		FROM organizations
		WHERE id = $1
	`
	err := pg.db.QueryRow(query, id).Scan(
		&org.ID,
		&org.Name,
		&org.CreatedAt,
		&org.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // Not found
	}

	if err != nil {
		return nil, err
	}

	return org, nil
}

func (pg *PostgresOrganizationStore) UpdateOrganization(org *Organization) (*Organization, error) {
	tx, err := pg.db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	query := `
		UPDATE organizations
		SET name = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING updated_at
	`
	results, err := tx.Exec(
		query,
		org.Name,
		org.ID,
	)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := results.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, sql.ErrNoRows
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return org, nil
}

func (pg *PostgresOrganizationStore) DeleteOrganization(id uuid.UUID) error {
	tx, err := pg.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		DELETE FROM organizations
		WHERE id = $1
	`
	result, err := tx.Exec(query, id)
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

	return tx.Commit()
}
