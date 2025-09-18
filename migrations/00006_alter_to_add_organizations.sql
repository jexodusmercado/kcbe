-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN organization_id UUID NOT NULL REFERENCES organizations(id) ;

ALTER TABLE categories ADD COLUMN organization_id UUID NOT NULL REFERENCES organizations(id);

ALTER TABLE items ADD COLUMN organization_id UUID NOT NULL REFERENCES organizations(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN organization_id;

ALTER TABLE categories DROP COLUMN organization_id;

ALTER TABLE items DROP COLUMN organization_id;
-- +goose StatementEnd
