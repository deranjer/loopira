-- +goose Up
ALTER TABLE issues DROP CONSTRAINT issues_status_check;
ALTER TABLE issues ADD CONSTRAINT issues_status_check
    CHECK (status IN ('backlog', 'todo', 'in_progress', 'blocked', 'done', 'canceled'));

-- +goose Down
UPDATE issues SET status = 'todo' WHERE status = 'blocked';
ALTER TABLE issues DROP CONSTRAINT issues_status_check;
ALTER TABLE issues ADD CONSTRAINT issues_status_check
    CHECK (status IN ('backlog', 'todo', 'in_progress', 'done', 'canceled'));
