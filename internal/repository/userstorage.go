package repository

import "context"

type UserStorage interface {
	SaveUser(ctx context.Context) (int, error)
}
