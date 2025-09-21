package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type StockLevel struct {
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

type PostgresStockLevelStore struct {
	db *sql.DB
}

func NewPostgresStockLevelStore(db *sql.DB) *PostgresStockLevelStore {
	return &PostgresStockLevelStore{db: db}
}

type StockLevelStore interface {
	CreateStockLevel(stockLevel *StockLevel) (*StockLevel, error)
	GetStockLevelByID(id uuid.UUID) (*StockLevel, error)
}
