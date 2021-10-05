package model

import (
	"context"
	"time"
)

type UserRepository interface {
	CreateUser(ctx context.Context, username string) (int64, error)
	GetUsers(ctx context.Context) ([]User, error)
	GetUserByID(ctx context.Context, id int64) (User, error)
	GetUserByPublicID(ctx context.Context, id string) (User, error)
	UpdateUserByID(ctx context.Context, user User) (int64, error)
	DeleteUserByID(ctx context.Context, id int64) (int64, error)
}

// User ...
type User struct {
	ID       int64  `db:"id"`
	PublicID string `db:"public_id"`
	// profile
	Username string `db:"username"`

	UpdatedAt time.Time  `db:"updated_at"`
	CreatedAt time.Time  `db:"created_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

// TODO: replace query builder(goqu)
const (
	insertUserQuery = `
		INSERT users(public_id, username)
		VALUE (?, ?);
	`

	selectUsersQuery = `
		SELECT * FROM users
		WHERE deleted_at IS NULL;
	`

	selectUserByIDQuery = `
	SELECT * FROM users
	WHERE deleted_at IS NULL AND id = ?;
	`

	selectUserByPublicIDQuery = `
	SELECT * FROM users
	WHERE deleted_at IS NULL AND public_id = ?;
	`

	updateUserByIDQuery = `
		UPDATE users
			SET username = :username
			WHERE id = :id;
	`

	deleteUserQuery = `
		UPDATE users
			SET deleted_at = NOW()
			WHERE id = ?
	`
)

func (repo *SqlxRepository) CreateUser(ctx context.Context, username string) (int64, error) {
	publicID, _ := GenerateRandomString(10)
	result, err := repo.db.ExecContext(ctx, insertUserQuery, publicID, username)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()

}
func (repo *SqlxRepository) GetUsers(ctx context.Context) ([]User, error) {
	var users []User
	err := repo.db.SelectContext(ctx, &users, selectUsersQuery)
	return users, err

}
func (repo *SqlxRepository) GetUserByID(ctx context.Context, id int64) (User, error) {
	user := User{}
	err := repo.db.GetContext(ctx, &user, selectUserByIDQuery, id)
	return user, err
}

func (repo *SqlxRepository) GetUserByPublicID(ctx context.Context, id string) (User, error) {
	user := User{}
	err := repo.db.GetContext(ctx, &user, selectUserByPublicIDQuery, id)
	return user, err
}

func (repo *SqlxRepository) UpdateUserByID(ctx context.Context, user User) (int64, error) {
	result, err := repo.db.NamedExecContext(ctx, updateUserByIDQuery, user)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()

}
func (repo *SqlxRepository) DeleteUserByID(ctx context.Context, id int64) (int64, error) {
	result, err := repo.db.ExecContext(ctx, deleteUserQuery, id)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()

}
