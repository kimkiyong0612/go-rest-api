package model

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"

	"github.com/guregu/sqlx"
)

type Repository interface {
	UserRepository
	Tx(context.Context, func(Repository) error) error
}

type sqlxDB interface {
	sqlx.Ext
	sqlx.ExtContext
	sqlx.Preparer
	sqlx.PreparerContext
	Get(dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	Rebind(query string) string
}

// SqlxRepository sqlxリポジトリ実装
type SqlxRepository struct {
	db   sqlxDB
	root *sqlx.DB
}

// NewSqlxRepository リポジトリ実装を初期化して生成します
func NewSqlxRepository(db *sqlx.DB) (Repository, error) {
	repo := &SqlxRepository{
		db:   db,
		root: db,
	}
	return repo, nil
}

func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	bytes, err := GenerateRandomBytes(n)
	if err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes), nil
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (repo *SqlxRepository) Tx(ctx context.Context, do func(Repository) error) error {
	tx, err := repo.root.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	child := &SqlxRepository{
		db:   tx,
		root: repo.root,
	}
	if err := do(child); err != nil {
		if innerErr := tx.Rollback(); innerErr != nil {
			return fmt.Errorf("tx: rollback error: %w (outer error: %v)", innerErr, err)
		}
		return err
	}
	return tx.Commit()
}
