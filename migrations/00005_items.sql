-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS items (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	SKU VARCHAR(100) NULL,
	organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
	category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
	name VARCHAR(50) UNIQUE NOT NULL,
	description VARCHAR(255) NULL,
	color VARCHAR(30) NULL,
	weight DECIMAL(10,2) NULL,
	length DECIMAL(10,2) NULL,
	width DECIMAL(10,2) NULL,
	height DECIMAL(10,2) NULL,
	unit_price BIGINT NOT NULL,
	cost_price BIGINT NOT NULL,
	is_active BOOLEAN DEFAULT TRUE,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_items_name ON items(name);
CREATE INDEX idx_items_organization_id ON items(organization_id);
CREATE INDEX idx_items_category_id ON items(category_id);
CREATE INDEX idx_items_is_active ON items(is_active);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS items;

DROP INDEX IF EXISTS idx_items_name;
DROP INDEX IF EXISTS idx_items_organization_id;
DROP INDEX IF EXISTS idx_items_category_id;
DROP INDEX IF EXISTS idx_items_is_active;
-- +goose StatementEnd
