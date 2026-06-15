package addon

const migrationTemplate = `-- Golider 数据库迁移占位：创建 messages 表
-- 使用方法：psql $DATABASE_URL -f migrations/001_create_messages.sql

CREATE TABLE IF NOT EXISTS messages (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    content     TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'active',
    version     INTEGER NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ,
    archived_at TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at DESC);
`
