package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/db"
	"github.com/tinoquang/comic-notifier/pkg/model"
)

// UserInterface contain user's interact method
type UserInterface interface {
	GetByID(ctx context.Context, field, id string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	//Update
	//List
}

type userDB struct {
	dbconn *sql.DB
	cfg    *conf.Config
}

// NewUserStore return user interfaces
func NewUserStore(dbconn *sql.DB, cfg *conf.Config) UserInterface {
	return &userDB{dbconn: dbconn, cfg: cfg}
}

// field is either psid or appid
func (u *userDB) GetByID(ctx context.Context, field, id string) (*model.User, error) {
	query := "WHERE " + field + "=$1"
	users, err := u.getBySQL(ctx, query, id)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get customer with %s: %s", field, id)
	}

	if len(users) == 0 {
		return &model.User{}, errors.New(fmt.Sprintf("User not found, %s:%s", field, id))
	}

	return &users[0], nil
}

func (u *userDB) Create(ctx context.Context, user *model.User) error {

	query := "insert into users (name, psid, appid, profile_pic) values ($1, $2, $3, $4) returning id"

	err := db.WithTransaction(ctx, u.dbconn, func(tx db.Transaction) error {
		return tx.QueryRowContext(
			ctx, query, user.Name, user.PSID, user.AppID, user.ProfilePic).Scan(&user.ID)
	})
	return err

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
		err := rows.Scan(&user.ID, &user.Name, &user.PSID, &user.AppID, &user.ProfilePic)
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
