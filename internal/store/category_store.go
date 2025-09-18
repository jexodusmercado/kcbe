package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Description    *string   `json:"description"`
	OrganizationID uuid.UUID `json:"organization_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type PostgresCategoryStore struct {
	db *sql.DB
}

func NewPostgresCategoryStore(db *sql.DB) *PostgresCategoryStore {
	return &PostgresCategoryStore{db: db}
}

type CategoryStore interface {
	CreateCategory(category *Category) (*Category, error)
	GetCategoryByID(id uuid.UUID) (*Category, error)
	UpdateCategory(category *Category) (*Category, error)
	DeleteCategory(id uuid.UUID) error
	GetCategoryByOrganization(page, pageSize int, organizationID uuid.UUID) ([]*Category, error)
	CountCategoriesByOrganization(organizationID uuid.UUID) (int, error)
}

func (s *PostgresCategoryStore) CreateCategory(category *Category) (*Category, error) {
	query := `
		INSERT INTO categories (name, description, organization_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := s.db.QueryRow(
		query,
		category.Name,
		category.Description,
		category.OrganizationID,
	).Scan(
		&category.ID,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return category, nil
}

func (s *PostgresCategoryStore) GetCategoryByID(id uuid.UUID) (*Category, error) {
	query := `
		SELECT id, name, description, organization_id, created_at, updated_at
		FROM categories
		WHERE id = $1
	`

	category := &Category{}
	err := s.db.QueryRow(query, id).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.OrganizationID,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return category, nil
}

func (s *PostgresCategoryStore) UpdateCategory(category *Category) (*Category, error) {
	query := `
		UPDATE categories
		SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING updated_at
	`

	err := s.db.QueryRow(
		query,
		category.Name,
		category.Description,
		category.ID,
	).Scan(&category.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return category, nil
}

func (s *PostgresCategoryStore) DeleteCategory(id uuid.UUID) error {
	query := `
		DELETE FROM categories
		WHERE id = $1
	`

	_, err := s.db.Exec(query, id)
	return err
}

func (s *PostgresCategoryStore) GetCategoryByOrganization(page, pageSize int, organizationID uuid.UUID) ([]*Category, error) {
	query := `
		SELECT id, name, description, organization_id, created_at, updated_at
		FROM categories
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	offset := page * pageSize

	rows, err := s.db.Query(query, organizationID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*Category
	for rows.Next() {
		category := &Category{}
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Description,
			&category.OrganizationID,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *PostgresCategoryStore) CountCategoriesByOrganization(organizationID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM categories
		WHERE organization_id = $1
	`

	var count int
	err := s.db.QueryRow(query, organizationID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
