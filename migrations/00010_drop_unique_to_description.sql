-- +goose Up
-- +goose StatementBegin
ALTER TABLE items DROP CONSTRAINT items_description_key;
ALTER TABLE categories DROP CONSTRAINT categories_description_key;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE items ALTER COLUMN description SET UNIQUE;
ALTER TABLE categories ALTER COLUMN description SET UNIQUE;
-- +goose StatementEnd