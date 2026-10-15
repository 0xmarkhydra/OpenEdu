package auth

import (
	"context"
	"database/sql"
	"errors"
)

// ChangePassword verifies the current password and increments the account
// auth_version in the same transaction, revoking every previously issued JWT.
func (i *Identity) ChangePassword(ctx context.Context, subject, current, newPassword string) error {
	if newPassword == current {
		return ErrPassword
	}
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	tx, err := i.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var oldHash string
	err = tx.QueryRowContext(ctx, `SELECT password_hash FROM user_credentials WHERE user_id=$1::uuid FOR UPDATE`, subject).Scan(&oldHash)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrCredentials
	}
	if err != nil {
		return err
	}
	if !VerifyPassword(oldHash, current) {
		return ErrCredentials
	}
	if _, err = tx.ExecContext(ctx, `UPDATE user_credentials SET password_hash=$2,changed_at=now() WHERE user_id=$1::uuid`, subject, hash); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE users SET auth_version=auth_version+1 WHERE id=$1::uuid`, subject); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO security_events(actor_id,event_type) VALUES($1::uuid,'password_changed')`, subject); err != nil {
		return err
	}
	return tx.Commit()
}
