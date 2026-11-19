//go:build !solution

package ledger

import (
	"context"
	"database/sql"
	"errors"
	"slices"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type MyLedger struct {
	db *sql.DB
}

func New(ctx context.Context, dsn string) (Ledger, error) {
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
	CREATE TABLE IF NOT EXISTS accounts (
		id TEXT PRIMARY KEY,
		balance BIGINT NOT NULL DEFAULT 0
		CONSTRAINT accounts_balance_nonnegative CHECK (balance >= 0)
	)`)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &MyLedger{db: db}, nil
}

func (l *MyLedger) CreateAccount(ctx context.Context, id ID) error {
	_, err := l.db.ExecContext(ctx, "INSERT INTO accounts (id) VALUES ($1)", id)
	return err
}

func (l *MyLedger) GetBalance(ctx context.Context, id ID) (Money, error) {
	var m Money

	err := l.db.QueryRowContext(ctx, "SELECT balance FROM accounts WHERE id = $1", id).
		Scan(&m)

	return m, err
}

func (l *MyLedger) Deposit(ctx context.Context, id ID, amount Money) error {
	if amount < 0 {
		return ErrNegativeAmount
	}

	result, err := l.db.ExecContext(ctx,
		"UPDATE accounts SET balance = balance + $1 WHERE id = $2",
		amount,
		id,
	)
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

func (l *MyLedger) Withdraw(ctx context.Context, id ID, amount Money) error {
	if amount < 0 {
		return ErrNegativeAmount
	}

	result, err := l.db.ExecContext(ctx,
		"UPDATE accounts SET balance = balance - $1 WHERE id = $2",
		amount,
		id,
	)
	if err != nil {
		var pqErr *pgconn.PgError

		if errors.As(err, &pqErr) && pqErr.Code == "23514" &&
			pqErr.ConstraintName == "accounts_balance_nonnegative" {
			return ErrNoMoney
		}

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

func (l *MyLedger) Transfer(ctx context.Context, from, to ID, amount Money) error {
	if amount < 0 {
		return ErrNegativeAmount
	}

	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ids := []ID{from, to}
	slices.Sort(ids)

	balances := make(map[ID]int64, 2)

	for _, id := range ids {
		var balance int64

		err = tx.QueryRowContext(ctx, "SELECT balance FROM accounts WHERE id = $1 FOR UPDATE", id).
			Scan(&balance)
		if err != nil {
			return err
		}

		balances[id] = balance
	}

	if balances[from] < int64(amount) {
		return ErrNoMoney
	}

	_, err = tx.ExecContext(
		ctx,
		"UPDATE accounts SET balance = balance - $1 WHERE id = $2",
		amount,
		from,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		"UPDATE accounts SET balance = balance + $1 WHERE id = $2",
		amount,
		to,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (l *MyLedger) Close() error {
	return l.db.Close()
}
