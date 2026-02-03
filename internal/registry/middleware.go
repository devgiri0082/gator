package registry

import (
	"context"
	"fmt"
	"log"

	"github.com/devgiri0082/gator/internal/database"
)

type userCtxKeyType struct{}

var userCtxKey = userCtxKeyType{}

func (r *Registry) middlewareLoggedIn(handler CommandFunc) CommandFunc {
	return func(ctx context.Context, args []string) error {
		username := r.config.Getuser()
		user, err := r.db.GetUser(ctx, username)
		if err != nil {
			return fmt.Errorf("user not logged in: %w", err)
		}
		ctx = context.WithValue(ctx, userCtxKey, &user)
		return handler(ctx, args)
	}
}

func UserFromContext(ctx context.Context) *database.User {
	user, ok := ctx.Value(userCtxKey).(*database.User)
	fmt.Print()
	if !ok {
		log.Fatal("user should exist")
	}
	return user
}
