package middlewares

import (
	"context"
	"errors"
	"net/http"

	"github.com/ptflp/go-light/types"

	light "github.com/ptflp/go-light"

	"github.com/ptflp/go-light/session"

	"github.com/ptflp/go-light/respond"
)

type Token struct {
	respond.Responder
	jwt *session.JWTKeys
}

func NewCheckToken(responder respond.Responder, jwt *session.JWTKeys) *Token {
	return &Token{
		Responder: responder,
		jwt:       jwt,
	}
}

func (t *Token) CheckStrict(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, err := t.jwt.ExtractAccessToken(r)
		if err != nil && (err.Error() == "token expired" || err.Error() == "Token is expired") {
			t.ErrorUnauthorized(w, errors.New("token expired"))
			return
		}
		if err != nil {
			t.ErrorForbidden(w, err)
			return
		}
		ctx := context.WithValue(r.Context(), types.User{}, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (t *Token) Check(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, err := t.jwt.ExtractAccessToken(r)
		if err != nil {
			u = &light.User{}
		}
		ctx := context.WithValue(r.Context(), types.User{}, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
