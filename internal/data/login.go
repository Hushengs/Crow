package data

import (
	"context"
	"database/sql"

	"crow/internal/biz"
)

type loginRepo struct {
	data *Data
}

type loginUser struct {
	ID           int64
	Account      string
	PasswordHash string
}

// NewLoginRepo creates a new LoginRepo instance.
func NewLoginRepo(data *Data) biz.LoginRepo {
	return &loginRepo{data: data}
}

func (r *loginRepo) FindByAccount(ctx context.Context, account string) (*biz.LoginUser, error) {
	row := &loginUser{}
	err := r.data.db.QueryRowContext(
		ctx,
		"SELECT id, account, password_hash FROM users WHERE account = ? LIMIT 1",
		account,
	).Scan(&row.ID, &row.Account, &row.PasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return toLoginBiz(row), nil
}

func toLoginBiz(in *loginUser) *biz.LoginUser {
	if in == nil {
		return nil
	}
	return &biz.LoginUser{
		ID:           in.ID,
		Account:      in.Account,
		PasswordHash: in.PasswordHash,
	}
}
