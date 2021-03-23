// Code generated by sqlc. DO NOT EDIT.

package db

import (
	"context"
	"database/sql"
)

type Querier interface {
	CreateComic(ctx context.Context, arg CreateComicParams) (Comic, error)
	CreateSubscriber(ctx context.Context, arg CreateSubscriberParams) (Subscriber, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	DeleteComic(ctx context.Context, id int32) error
	DeleteSubscriber(ctx context.Context, arg DeleteSubscriberParams) error
	DeleteUser(ctx context.Context, psid sql.NullString) error
	GetComic(ctx context.Context, id int32) (Comic, error)
	GetComicByPSIDAndComicID(ctx context.Context, arg GetComicByPSIDAndComicIDParams) (Comic, error)
	GetComicByURL(ctx context.Context, url string) (Comic, error)
	GetComicForUpdate(ctx context.Context, id int32) (Comic, error)
	GetSubscriber(ctx context.Context, arg GetSubscriberParams) (Subscriber, error)
	GetUserByAppID(ctx context.Context, appid sql.NullString) (User, error)
	GetUserByPSID(ctx context.Context, psid sql.NullString) (User, error)
	ListComics(ctx context.Context) ([]Comic, error)
	ListComicsPerUserPSID(ctx context.Context, userID int32) ([]Comic, error)
	ListUsers(ctx context.Context) ([]User, error)
	ListUsersPerComic(ctx context.Context, comicID int32) ([]User, error)
	SearchComicOfUserByName(ctx context.Context, arg SearchComicOfUserByNameParams) ([]Comic, error)
	UpdateComic(ctx context.Context, arg UpdateComicParams) (Comic, error)
	UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error)
}

var _ Querier = (*Queries)(nil)
