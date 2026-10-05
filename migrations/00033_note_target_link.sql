-- +goose Up
-- 00033: TASK-011 item 5 — двусторонняя связь заметка <-> цель (задача/событие/напоминание)
-- notes.target_* позволяют создать заметку из задачи и увидеть её в списке заметок с бейджем цели;
-- обратная связь уже существует: tasks.note_id.

ALTER TABLE notes
    ADD COLUMN IF NOT EXISTS target_type VARCHAR(20),
    ADD COLUMN IF NOT EXISTS target_id   UUID;

ALTER TABLE notes DROP CONSTRAINT IF EXISTS notes_target_type_check;
ALTER TABLE notes ADD CONSTRAINT notes_target_type_check
    CHECK (target_type IS NULL OR target_type IN ('task', 'event', 'reminder'));

CREATE INDEX IF NOT EXISTS idx_notes_target ON notes (user_id, target_type, target_id);

-- +goose Down
DROP INDEX IF EXISTS idx_notes_target;
ALTER TABLE notes DROP CONSTRAINT IF EXISTS notes_target_type_check;
ALTER TABLE notes DROP COLUMN IF EXISTS target_id;
ALTER TABLE notes DROP COLUMN IF EXISTS target_type;
