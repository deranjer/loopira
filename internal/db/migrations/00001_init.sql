-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pgcrypto;
-- +goose StatementEnd

CREATE TABLE users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name          text NOT NULL,
    email         text NOT NULL UNIQUE,
    avatar_url    text,
    role          text NOT NULL DEFAULT 'member' CHECK (role IN ('admin', 'member')),
    github_id     text UNIQUE,
    password_hash text,
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE sessions (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL
);

CREATE INDEX idx_sessions_user ON sessions (user_id);

CREATE TABLE teams (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name       text NOT NULL,
    key        text NOT NULL UNIQUE, -- short prefix, e.g. "ENG"
    next_issue_number int NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE team_members (
    team_id uuid NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (team_id, user_id)
);

CREATE TABLE projects (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id     uuid NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    name        text NOT NULL,
    description text,
    status      text NOT NULL DEFAULT 'backlog' CHECK (status IN ('backlog', 'planned', 'in_progress', 'paused', 'completed', 'canceled')),
    lead_id     uuid REFERENCES users(id) ON DELETE SET NULL,
    target_date date,
    priority    smallint NOT NULL DEFAULT 0 CHECK (priority BETWEEN 0 AND 4),
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE project_members (
    project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (project_id, user_id)
);

CREATE TABLE milestones (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        text NOT NULL,
    target_date date,
    sort_order  int NOT NULL DEFAULT 0
);

CREATE TABLE cycles (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id    uuid NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    number     int NOT NULL,
    start_date date NOT NULL,
    end_date   date NOT NULL,
    UNIQUE (team_id, number)
);

CREATE TABLE labels (
    id      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id uuid REFERENCES teams(id) ON DELETE CASCADE, -- null = workspace-wide label
    name    text NOT NULL,
    color   text NOT NULL DEFAULT '#6b7280'
);

CREATE TABLE issues (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id      uuid NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    number       int NOT NULL, -- combined with team key to form e.g. ENG-123
    title        text NOT NULL,
    description  text NOT NULL DEFAULT '',
    status       text NOT NULL DEFAULT 'backlog' CHECK (status IN ('backlog', 'todo', 'in_progress', 'done', 'canceled')),
    priority     smallint NOT NULL DEFAULT 0 CHECK (priority BETWEEN 0 AND 4), -- 0=none,1=urgent,2=high,3=medium,4=low
    assignee_id  uuid REFERENCES users(id) ON DELETE SET NULL,
    parent_id    uuid REFERENCES issues(id) ON DELETE SET NULL,
    project_id   uuid REFERENCES projects(id) ON DELETE SET NULL,
    cycle_id     uuid REFERENCES cycles(id) ON DELETE SET NULL,
    due_date     date,
    created_by   uuid NOT NULL REFERENCES users(id),
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    UNIQUE (team_id, number)
);

CREATE INDEX idx_issues_team_status ON issues (team_id, status);
CREATE INDEX idx_issues_project ON issues (project_id);
CREATE INDEX idx_issues_cycle ON issues (cycle_id);
CREATE INDEX idx_issues_parent ON issues (parent_id);
CREATE INDEX idx_issues_search ON issues USING GIN (to_tsvector('english', title || ' ' || description));

CREATE TABLE issue_labels (
    issue_id uuid NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    label_id uuid NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    PRIMARY KEY (issue_id, label_id)
);

CREATE TABLE issue_relations (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    issue_id          uuid NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    related_issue_id  uuid NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    type              text NOT NULL CHECK (type IN ('blocks', 'blocked_by', 'related', 'duplicate')),
    CHECK (issue_id <> related_issue_id),
    UNIQUE (issue_id, related_issue_id, type)
);

CREATE TABLE comments (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    issue_id   uuid NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    author_id  uuid NOT NULL REFERENCES users(id),
    body       text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE comment_mentions (
    comment_id uuid NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (comment_id, user_id)
);

CREATE TABLE views (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id   uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       text NOT NULL,
    definition jsonb NOT NULL, -- filter/sort/group spec
    shared     boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE notifications (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type       text NOT NULL, -- e.g. 'mention', 'assignment', 'status_change'
    issue_id   uuid REFERENCES issues(id) ON DELETE CASCADE,
    comment_id uuid REFERENCES comments(id) ON DELETE CASCADE,
    read_at    timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_notifications_user_unread ON notifications (user_id, read_at);

CREATE TABLE api_keys (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       text NOT NULL,
    key_hash   text NOT NULL UNIQUE,
    scopes     text[] NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now(),
    last_used_at timestamptz
);

CREATE TABLE webhooks (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    target_url     text NOT NULL,
    secret         text NOT NULL,
    event_types    text[] NOT NULL,
    active         boolean NOT NULL DEFAULT true,
    created_at     timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE webhook_deliveries (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_id   uuid NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event_type   text NOT NULL,
    payload      jsonb NOT NULL,
    status       text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'delivered', 'failed')),
    attempts     int NOT NULL DEFAULT 0,
    next_attempt_at timestamptz NOT NULL DEFAULT now(),
    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE git_links (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    issue_id      uuid NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    provider      text NOT NULL DEFAULT 'github',
    pr_url        text NOT NULL,
    status        text NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'merged', 'closed')),
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE attachments (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    issue_id     uuid REFERENCES issues(id) ON DELETE CASCADE,
    comment_id   uuid REFERENCES comments(id) ON DELETE CASCADE,
    project_id   uuid REFERENCES projects(id) ON DELETE CASCADE,
    filename     text NOT NULL,
    content_type text NOT NULL,
    size_bytes   bigint NOT NULL,
    storage_key  text NOT NULL,
    uploaded_by  uuid NOT NULL REFERENCES users(id),
    created_at   timestamptz NOT NULL DEFAULT now(),
    CHECK (issue_id IS NOT NULL OR comment_id IS NOT NULL OR project_id IS NOT NULL)
);

-- +goose Down
DROP TABLE IF EXISTS attachments;
DROP TABLE IF EXISTS git_links;
DROP TABLE IF EXISTS webhook_deliveries;
DROP TABLE IF EXISTS webhooks;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS views;
DROP TABLE IF EXISTS comment_mentions;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS issue_relations;
DROP TABLE IF EXISTS issue_labels;
DROP TABLE IF EXISTS issues;
DROP TABLE IF EXISTS labels;
DROP TABLE IF EXISTS cycles;
DROP TABLE IF EXISTS milestones;
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
