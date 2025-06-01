-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS question_bank(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  topic TEXT,
  question_count INT NOT NULL,
  question_data JSONB NOT NULL,
  is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
  created_on TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_on TIMESTAMP NOT NULL DEFAULT NOW(),
  created_by TEXT NOT NULL,
  updated_by TEXT NOT NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists question_bank;
-- +goose StatementEnd
