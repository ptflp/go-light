package providers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ptflp/go-light/types"

	"github.com/ptflp/go-light/decoder"
	"github.com/ptflp/go-light/request"

	"golang.org/x/oauth2/facebook"

	"github.com/ptflp/go-light/utils"

	"golang.org/x/oauth2"

	light "github.com/ptflp/go-light"
)

type Facebook struct {
	*decoder.Decoder
	config *oauth2.Config
}

func NewFacebookAuth(config *oauth2.Config) *Facebook {
	config.Endpoint = facebook.Endpoint
	config.Scopes = []string{"public_profile"}

	return &Facebook{config: config, Decoder: decoder.NewDecoder()}
}

func (f *Facebook) RedirectUrl() string {
	uuid, err := utils.ProjectUUIDGen("F")
	if err != nil {
		return ""
	}
	url := f.config.AuthCodeURL(uuid)

	return url
}

func (f *Facebook) Callback(r *http.Request) (light.User, error) {
	code := r.FormValue("code")

	token, err := f.config.Exchange(r.Context(), code)
	if err != nil {
		fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		return light.User{}, err
	}

	resp, err := http.Get("https://graph.facebook.com/me?access_token=" +
		url.QueryEscape(token.AccessToken))
	if err != nil {
		fmt.Printf("Get: %s\n", err)
		return light.User{}, err
	}
	defer resp.Body.Close()

	var req request.FacebookCallbackRequest
	err = f.Decode(resp.Body, &req)
	if err != nil {
		return light.User{}, err
	}
	facebookID, err := strconv.Atoi(req.FacebookID)
	if err != nil {
		return light.User{}, err
	}

	return light.User{
		FacebookID: types.NewNullInt64(int64(facebookID)),
		Name:       types.NewNullString(req.Name),
	}, nil
}
