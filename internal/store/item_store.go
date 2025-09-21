package store

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Item struct {
	ID             uuid.UUID   `json:"id"`
	SKU            *string     `json:"sku"`
	OrganizationID uuid.UUID   `json:"organization_id"`
	CategoryID     uuid.UUID   `json:"category_id"`
	Name           string      `json:"name"`
	Description    *string     `json:"description"`
	Color          *string     `json:"color"`
	Weight         *float64    `json:"weight"`
	Length         *float64    `json:"length"`
	Width          *float64    `json:"width"`
	Height         *float64    `json:"height"`
	UnitPrice      int         `json:"unit_price"`
	CostPrice      int         `json:"cost_price"`
	IsActive       bool        `json:"is_active"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	Stock          []ItemStock `json:"stock,omitempty"`
}

type ItemStock struct {
	ID                uuid.UUID `json:"id"`
	LocationID        uuid.UUID `json:"location_id"`
	ItemID            uuid.UUID `json:"item_id"`
	QuantityPhysical  int       `json:"quantity_physical"`
	QuantityAvailable int       `json:"quantity_available"`
	QuantityReserved  int       `json:"quantity_reserved"`
	ReorderLevel      int       `json:"reorder_level"`
	MaxStockLevel     int       `json:"max_stock_level"`
	LastCountedAt     time.Time `json:"last_counted_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	Version           int       `json:"version"`
}

type ItemStockMovement struct {
	ID            uuid.UUID `json:"id"`
	ItemID        uuid.UUID `json:"item_id"`
	LocationID    uuid.UUID `json:"location_id"`
	MovementType  string    `json:"movement_type"`
	Quantity      int       `json:"quantity"`
	ReferenceType string    `json:"reference_type"`
	ReferenceID   uuid.UUID `json:"reference_id"`
	Reason        *string   `json:"reason"`
	CreatedAt     time.Time `json:"created_at"`
	CreatedBy     uuid.UUID `json:"created_by"`
	BatchID       *string   `json:"batch_id"`
	UpdatedAt     time.Time `json:"updated_at"`
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
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO items (sku, organization_id, category_id, name, description, color, weight, length, width, height, unit_price, cost_price, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`

	err = s.db.QueryRow(
		query,
		item.SKU,
		item.OrganizationID,
		item.CategoryID,
		item.Name,
		item.Description,
		item.Color,
		item.Weight,
		item.Length,
		item.Width,
		item.Height,
		item.UnitPrice,
		item.CostPrice,
		item.IsActive,
	).Scan(
		&item.ID,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	for i := range item.Stock {

		query := `
			INSERT INTO stock_levels (location_id, item_id, quantity_physical, quantity_available, quantity_reserved, reorder_level, max_stock_level, last_counted_at, version)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id, updated_at`
		err := s.db.QueryRow(
			query,
			item.Stock[i].LocationID,
			item.ID,
			item.Stock[i].QuantityAvailable,
			item.Stock[i].QuantityAvailable,
			item.Stock[i].QuantityReserved,
			item.Stock[i].ReorderLevel,
			item.Stock[i].MaxStockLevel,
			item.Stock[i].LastCountedAt,
			1, // initial version
		).Scan(
			&item.Stock[i].ID,
			&item.Stock[i].UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (s *PostgresItemStore) GetItemByID(id uuid.UUID) (*Item, error) {
	item := &Item{}
	var stockLevelJSON []byte
	query := `
		SELECT i.id, i.sku, i.organization_id, i.category_id, i.name, i.description, i.color, i.weight, i.length, i.width, i.height, i.unit_price, i.cost_price, i.is_active, i.created_at, i.updated_at,
		COALESCE(ARRAY_AGG(
			JSON_BUILD_OBJECT(
				'id', s.id,
				'location_id', s.location_id,
				'item_id', s.item_id,
				'quantity_physical', s.quantity_physical,
				'quantity_available', s.quantity_available,
				'quantity_reserved', s.quantity_reserved,
				'reorder_level', s.reorder_level,
				'max_stock_level', s.max_stock_level,
				'last_counted_at', s.last_counted_at,
				'updated_at', s.updated_at,
				'version', s.version
			)
		) FILTER (WHERE s.id IS NOT NULL), '{}') AS stock_levels
		FROM items i
		LEFT JOIN stock_levels s ON i.id = s.item_id
		WHERE id = $1
	`
	err := s.db.QueryRow(query, id).Scan(
		&item.ID,
		&item.SKU,
		&item.OrganizationID,
		&item.CategoryID,
		&item.Name,
		&item.Description,
		&item.Color,
		&item.Weight,
		&item.Length,
		&item.Width,
		&item.Height,
		&item.UnitPrice,
		&item.CostPrice,
		&item.IsActive,
		&item.CreatedAt,
		&item.UpdatedAt,
		&stockLevelJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(stockLevelJSON, &item.Stock); err != nil {
		return nil, err
	}

	// stockQuery := `
	// 	SELECT id, location_id, item_id, quantity_physical, quantity_available, quantity_reserved, reorder_level, max_stock_level, last_counted_at, updated_at, version
	// 	FROM stock_levels
	// 	WHERE item_id = $1
	// `
	//
	// rows, err := s.db.Query(stockQuery, item.ID)
	// if err != nil {
	// 	return nil, err
	// }
	//
	// for rows.Next() {
	// 	var stock ItemStock
	// 	err := rows.Scan(
	// 		&stock.ID,
	// 		&stock.LocationID,
	// 		&stock.ItemID,
	// 		&stock.QuantityPhysical,
	// 		&stock.QuantityAvailable,
	// 		&stock.QuantityReserved,
	// 		&stock.ReorderLevel,
	// 		&stock.MaxStockLevel,
	// 		&stock.LastCountedAt,
	// 		&stock.UpdatedAt,
	// 		&stock.Version,
	// 	)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	item.Stock = append(item.Stock, stock)
	// }
	// if err = rows.Err(); err != nil {
	// 	return nil, err
	// }
	//
	return item, nil
}

func (s *PostgresItemStore) UpdateItem(item *Item) (*Item, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	query := `
		UPDATE items
		SET sku = $1, organization_id = $2, category_id = $3, name = $4, description = $5, color = $6, weight = $7, length = $8, width = $9, height = $10, unit_price = $11, cost_price = $12, is_active = $13, updated_at = $14
		WHERE id = $15
		RETURNING created_at, updated_at
	`

	err = s.db.QueryRow(
		query,
		item.SKU,
		item.OrganizationID,
		item.CategoryID,
		item.Name,
		item.Description,
		item.Color,
		item.Weight,
		item.Length,
		item.Width,
		item.Height,
		item.UnitPrice,
		item.CostPrice,
		item.IsActive,
		time.Now(),
		item.ID,
	).Scan(
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`
		DELETE FROM stock_levels
		WHERE item_id = $1
	`, item.ID)
	if err != nil {
		return nil, err
	}

	for i := range item.Stock {
		query := `
			INSERT INTO stock_levels (location_id, item_id, quantity_physical, quantity_available, quantity_reserved, reorder_level, max_stock_level, last_counted_at, version)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id, updated_at`
		err := s.db.QueryRow(
			query,
			item.Stock[i].LocationID,
			item.ID,
			item.Stock[i].QuantityPhysical,
			item.Stock[i].QuantityAvailable,
			item.Stock[i].QuantityReserved,
			item.Stock[i].ReorderLevel,
			item.Stock[i].MaxStockLevel,
			item.Stock[i].LastCountedAt,
			1, // initial version
		).Scan(
			&item.Stock[i].ID,
			&item.Stock[i].UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
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
		SELECT i.*,
			COALESCE(s.stock_levels, '{}') AS stock_levels
		FROM items i
		LEFT JOIN (
			SELECT item_id,
				JSON_AGG(
					JSON_BUILD_OBJECT(
						'id', id,
						'location_id', location_id,
						'item_id', item_id,
						'quantity_physical', quantity_physical,
						'quantity_available', quantity_available,
						'quantity_reserved', quantity_reserved,
						'reorder_level', reorder_level,
						'max_stock_level', max_stock_level,
						'last_counted_at', last_counted_at,
						'updated_at', updated_at,
						'version', version
					) ORDER BY location_id
				) AS stock_levels
			FROM stock_levels
			GROUP BY item_id
		) s ON i.id = s.item_id
		WHERE i.organization_id = $1
		ORDER BY i.created_at DESC
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
		var stockLevelJSON []byte
		err := rows.Scan(
			&item.ID,
			&item.SKU,
			&item.OrganizationID,
			&item.CategoryID,
			&item.Name,
			&item.Description,
			&item.Color,
			&item.Weight,
			&item.Length,
			&item.Width,
			&item.Height,
			&item.UnitPrice,
			&item.CostPrice,
			&item.IsActive,
			&item.CreatedAt,
			&item.UpdatedAt,
			&stockLevelJSON,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(stockLevelJSON, &item.Stock); err != nil {
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
