package addon

const postgresMessageRepositoryTemplate = `package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"{{ .ModulePath }}/internal/service"
)

// PostgresMessageRepository 是 MessageRepository 的 PostgreSQL 实现，作为数据库可切换仓储占位。
// 使用 ` + "`go get github.com/lib/pq`" + ` 安装 PostgreSQL 驱动后即可启用。
type PostgresMessageRepository struct {
	db *sql.DB
}

func NewPostgresMessageRepository(db *sql.DB) *PostgresMessageRepository {
	return &PostgresMessageRepository{db: db}
}

// NewDatabaseMessageService 创建一个连接到 PostgreSQL 的消息服务实例。
// 如果连接失败则返回错误；连接成功后数据库就绪检查也会通过。
func NewDatabaseMessageService(databaseURL string) (*service.MessageService, *sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("打开数据库连接失败：%w", err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("数据库探活失败：%w", err)
	}

	repo := NewPostgresMessageRepository(db)
	return service.NewMessageService(repo), db, nil
}

func (r *PostgresMessageRepository) List(ctx context.Context) ([]service.Message, error) {
	const query = "SELECT id, title, content, status, version, created_at, updated_at, archived_at, deleted_at FROM messages WHERE deleted_at IS NULL ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查询消息列表失败：%w", err)
	}
	defer rows.Close()

	var messages []service.Message
	for rows.Next() {
		var m service.Message
		if err := scanMessage(rows, &m); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历消息列表失败：%w", err)
	}

	return messages, nil
}

func (r *PostgresMessageRepository) FindByID(ctx context.Context, id string) (service.Message, bool, error) {
	const query = "SELECT id, title, content, status, version, created_at, updated_at, archived_at, deleted_at FROM messages WHERE id = $1 AND deleted_at IS NULL"

	row := r.db.QueryRowContext(ctx, query, id)

	var m service.Message
	if err := scanMessage(row, &m); err != nil {
		if err == sql.ErrNoRows {
			return service.Message{}, false, nil
		}
		return service.Message{}, false, fmt.Errorf("查询消息失败：%w", err)
	}
	return m, true, nil
}

func (r *PostgresMessageRepository) Save(ctx context.Context, message service.Message) error {
	const query = "INSERT INTO messages (id, title, content, status, version, created_at, updated_at, archived_at, deleted_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, content = EXCLUDED.content, status = EXCLUDED.status, version = EXCLUDED.version, updated_at = EXCLUDED.updated_at, archived_at = EXCLUDED.archived_at, deleted_at = EXCLUDED.deleted_at"

	_, err := r.db.ExecContext(ctx, query,
		message.ID, message.Title, message.Content, message.Status, message.Version,
		message.CreatedAt, message.UpdatedAt, message.ArchivedAt, message.DeletedAt,
	)
	if err != nil {
		return fmt.Errorf("保存消息失败：%w", err)
	}
	return nil
}

func (r *PostgresMessageRepository) SaveVersioned(ctx context.Context, message service.Message, expectedVersion int) (bool, error) {
	const query = "UPDATE messages SET title = $1, content = $2, status = $3, version = $4, updated_at = $5, archived_at = $6, deleted_at = $7 WHERE id = $8 AND version = $9 AND deleted_at IS NULL"

	result, err := r.db.ExecContext(ctx, query,
		message.Title, message.Content, message.Status, message.Version,
		message.UpdatedAt, message.ArchivedAt, message.DeletedAt,
		message.ID, expectedVersion,
	)
	if err != nil {
		return false, fmt.Errorf("版本更新失败：%w", err)
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

func (r *PostgresMessageRepository) NextID(ctx context.Context) (string, error) {
	const query = "SELECT COALESCE(MAX(CAST(SUBSTRING(id FROM 'msg_([0-9]+)$') AS INTEGER)), 0) + 1 FROM messages"

	var nextID int
	err := r.db.QueryRowContext(ctx, query).Scan(&nextID)
	if err != nil {
		return "", fmt.Errorf("生成 ID 失败：%w", err)
	}
	return fmt.Sprintf("msg_%d", nextID), nil
}

// scanMessage 提供统一的数据库行扫描逻辑。
type scanner interface {
	Scan(dest ...any) error
}

func scanMessage(row scanner, m *service.Message) error {
	return row.Scan(&m.ID, &m.Title, &m.Content, &m.Status, &m.Version,
		&m.CreatedAt, &m.UpdatedAt, &m.ArchivedAt, &m.DeletedAt)
}
`
