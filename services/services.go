package services

import (
	"context"

	light "github.com/ptflp/go-light"
	"github.com/ptflp/go-light/auth"
	"github.com/ptflp/go-light/components"
)

type Services struct {
	AuthService light.AuthService
	// TODO change to interface
	User *User
}

func NewServices(ctx context.Context, cmps components.Componenter, reps light.Repositories) *Services {
	var services Services
	user := NewUserService(reps, cmps)
	services.User = user

	services.AuthService = auth.NewAuthService(reps, cmps)

	return &services
}
