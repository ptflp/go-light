package providers

import (
	"net/http"

	light "github.com/ptflp/go-light"
)

type Socials interface {
	Callback(r *http.Request) (light.User, error)
	RedirectUrl() string
}
