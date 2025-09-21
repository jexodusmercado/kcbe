-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS stock_levels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    location_id UUID NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    quantity_physical INT NOT NULL,
    quantity_available INT NOT NULL,
    quantity_reserved INT NOT NULL DEFAULT 0,
    reorder_level INT NOT NULL DEFAULT 0,
    max_stock_level INT NOT NULL DEFAULT 0,
    last_counted_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    version INT NOT NULL DEFAULT 1,
    UNIQUE(location_id, item_id),
    CONSTRAINT check_positive_physical CHECK (quantity_physical >= 0),
    CONSTRAINT check_positive_reserved CHECK (quantity_reserved >= 0),
    CONSTRAINT check_valid_reservation CHECK (quantity_reserved <= quantity_physical)
);

CREATE INDEX idx_stock_item_location ON stock_levels(item_id, location_id);
CREATE INDEX idx_stock_available ON stock_levels(quantity_available) WHERE quantity_available > 0;
CREATE INDEX idx_stock_reorder ON stock_levels(item_id) WHERE quantity_available <= reorder_level;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS stock_levels;
DROP INDEX IF EXISTS idx_stock_item_location;
DROP INDEX IF EXISTS idx_stock_available;
DROP INDEX IF EXISTS idx_stock_reorder;
-- +goose StatementEnd
