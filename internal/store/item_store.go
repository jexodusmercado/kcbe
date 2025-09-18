package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Item struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	CategoryID     uuid.UUID `json:"category_id"`
	Name           string    `json:"name"`
	Description    *string   `json:"description"`
	SKU            string    `json:"sku"`
	UnitPrice      int64     `json:"unit_price"`
	ReorderLevel   int64     `json:"reorder_level"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type PostgresItemStore struct {
	db *sql.DB
}

func NewPostgresItemStore(db *sql.DB) *PostgresItemStore {
	return &PostgresItemStore{db: db}
}

type ItemStore interface {
	CreateItem(item *Item) (*Item, error)
	GetItemByID(id uuid.UUID) (*Item, error)
	UpdateItem(item *Item) (*Item, error)
	DeleteItem(id uuid.UUID) error
	GetItemsByOrganization(page, pageSize int, organizationID uuid.UUID) ([]*Item, error)
	CountItemsByOrganization(organizationID uuid.UUID) (int, error)
	GetItemOrgID(id uuid.UUID) (uuid.UUID, error)
}

func (s *PostgresItemStore) CreateItem(item *Item) (*Item, error) {
	query := `
		INSERT INTO items (organization_id, category_id, name, description, sku, unit_price, reorder_level)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	err := s.db.QueryRow(
		query,
		item.OrganizationID,
		item.CategoryID,
		item.Name,
		item.Description,
		item.SKU,
		item.UnitPrice,
		item.ReorderLevel,
	).Scan(
		&item.ID,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (s *PostgresItemStore) GetItemByID(id uuid.UUID) (*Item, error) {
	item := &Item{}
	query := `
		SELECT id, organization_id, category_id, name, description, sku, unit_price, reorder_level, created_at, updated_at
		FROM items
		WHERE id = $1
	`
	err := s.db.QueryRow(query, id).Scan(
		&item.ID,
		&item.OrganizationID,
		&item.CategoryID,
		&item.Name,
		&item.Description,
		&item.SKU,
		&item.UnitPrice,
		&item.ReorderLevel,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func (s *PostgresItemStore) UpdateItem(item *Item) (*Item, error) {
	query := `
		UPDATE items
		SET category_id = $1, name = $2, description = $3, sku = $4, unit_price = $5, reorder_level = $6, updated_at = $7
		WHERE id = $8
		RETURNING created_at, updated_at
	`

	err := s.db.QueryRow(
		query,
		item.CategoryID,
		item.Name,
		item.Description,
		item.SKU,
		item.UnitPrice,
		item.ReorderLevel,
		time.Now(),
		item.ID,
	).Scan(
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return item, nil
}

func (s *PostgresItemStore) DeleteItem(id uuid.UUID) error {
	query := `
		DELETE FROM items
		WHERE id = $1
	`
	result, err := s.db.Exec(query, id)
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

	return err
}

func (s *PostgresItemStore) GetItemsByOrganization(page, pageSize int, organizationID uuid.UUID) ([]*Item, error) {
	query := `
		SELECT id, organization_id, category_id, name, description, sku, unit_price, reorder_level, created_at, updated_at
		FROM items
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

	var items []*Item
	for rows.Next() {
		item := &Item{}
		err := rows.Scan(
			&item.ID,
			&item.OrganizationID,
			&item.CategoryID,
			&item.Name,
			&item.Description,
			&item.SKU,
			&item.UnitPrice,
			&item.ReorderLevel,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (s *PostgresItemStore) CountItemsByOrganization(organizationID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM items
		WHERE organization_id = $1
	`
	var count int
	err := s.db.QueryRow(query, organizationID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *PostgresItemStore) GetItemOrgID(id uuid.UUID) (uuid.UUID, error) {
	var orgID uuid.UUID
	query := `
		SELECT organization_id
		FROM items
		WHERE id = $1
	`
	err := s.db.QueryRow(query, id).Scan(&orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, nil
		}
		return uuid.Nil, err
	}
	return orgID, nil
}
