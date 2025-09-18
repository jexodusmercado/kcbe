-- +goose Up
-- +goose StatementBegin
ALTER TABLE categories ALTER COLUMN description DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE categories ALTER COLUMN description SET NOT NULL;
-- +goose StatementEnd