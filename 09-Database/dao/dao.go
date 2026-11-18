//go:build !solution

package dao

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type MyDao struct {
	db *sql.DB
}

func CreateDao(ctx context.Context, dsn string) (Dao, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	_, err = db.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS users (
		id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		name TEXT NOT NULL
	)`)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &MyDao{db: db}, nil
}

func (d *MyDao) Create(ctx context.Context, u *User) (UserID, error) {
	var id UserID

	err := d.db.QueryRowContext(ctx, "INSERT INTO users (name) VALUES ($1) RETURNING id", u.Name).
		Scan(&id)

	return id, err
}

func (d *MyDao) Update(ctx context.Context, u *User) error {
	result, err := d.db.ExecContext(ctx, "UPDATE users SET name = $1 WHERE id = $2", u.Name, u.ID)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (d *MyDao) Delete(ctx context.Context, id UserID) error {
	_, err := d.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	return err
}

func (d *MyDao) Lookup(ctx context.Context, id UserID) (User, error) {
	var u User

	err := d.db.QueryRowContext(ctx, "SELECT id, name FROM users WHERE id = $1", id).
		Scan(&u.ID, &u.Name)

	return u, err
}

func (d *MyDao) List(ctx context.Context) ([]User, error) {
	rows, err := d.db.QueryContext(ctx, "SELECT id, name FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]User, 0)

	for rows.Next() {
		var u User
		if err = rows.Scan(&u.ID, &u.Name); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (d *MyDao) Close() error {
	return d.db.Close()
}
