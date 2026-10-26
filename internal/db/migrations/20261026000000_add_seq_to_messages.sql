-- Adds a per-insert monotonic sequence to messages so history reads can be
-- paged. created_at is second-granular, so roughly half of a busy session's
-- rows tie with their neighbour and neither LIMIT/OFFSET nor a created_at
-- cursor can page over them without dropping or repeating a row. The trigger
-- copies rowid, which SQLite assigns in insertion order, so seq orders a
-- session's rows exactly as they were written.

-- +goose Up
-- +goose StatementBegin
ALTER TABLE messages ADD COLUMN seq INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE messages SET seq = rowid WHERE seq = 0;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_messages_session_seq ON messages (session_id, seq);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER set_message_seq AFTER INSERT ON messages
BEGIN
    UPDATE messages SET seq = new.rowid WHERE id = new.id;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS set_message_seq;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_messages_session_seq;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE messages DROP COLUMN seq;
-- +goose StatementEnd
