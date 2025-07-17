-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE templates ADD COLUMN lang VARCHAR(2) DEFAULT '';

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE templates DROP COLUMN lang;