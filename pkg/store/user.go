package store

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// UserRepo contain user's interact method
type UserRepo interface {
	GetByFBID(ctx context.Context, field, id string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	List(ctx context.Context) ([]model.User, error)
}

type userDB struct {
	dbconn *sql.DB
}

// NewUserStore return user interfaces
func newUserRepo(dbconn *sql.DB) *userDB {
	return &userDB{dbconn: dbconn}
}

// field is either psid or appid
func (u *userDB) GetByFBID(ctx context.Context, field, id string) (*model.User, error) {
	query := "WHERE " + field + "=$1"
	users, err := u.getBySQL(ctx, query, id)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get user with %s: %s", field, id)
	}

	if len(users) == 0 {
		return &model.User{}, util.ErrNotFound
	}

	return &users[0], nil
}

func (u *userDB) Create(ctx context.Context, user *model.User) error {

	// query := `INSERT INTO users (name, psid, appid, profile_pic) VALUES ($1, $2, $3, $4)
	// 			ON CONFLICT (psid) DO NOTHING
	// 			RETURNING psid`

	// err := db.WithTransaction(ctx, u.dbconn, func(tx db.Transaction) error {
	// 	return tx.QueryRowContext(ctx, query, user.Name, user.PSID, user.AppID, user.ProfilePic).Scan(&user.PSID)
	// })

	// if err == sql.ErrNoRows {
	// 	return nil
	// }

	return nil

}

func (u *userDB) List(ctx context.Context) ([]model.User, error) {

	return u.getBySQL(ctx, "")
}

func (u *userDB) getBySQL(ctx context.Context, query string, args ...interface{}) ([]model.User, error) {
	rows, err := u.dbconn.QueryContext(ctx, "SELECT * FROM users "+query, args...)
	if err != nil {
		return nil, err
	}

	users := []model.User{}
	defer rows.Close()
	for rows.Next() {
		user := model.User{}
		err := rows.Scan(&user.Name, &user.PSID, &user.AppID, &user.ProfilePic)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
