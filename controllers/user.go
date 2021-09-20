package controllers

import (
	"context"
	"errors"
	"fmt"
	"github.com/ptflp/go-light"
	"github.com/ptflp/go-light/types"
	"net/http"

	"github.com/ptflp/go-light/decoder"

	"github.com/ptflp/go-light/request"

	"github.com/ptflp/go-light/services"

	"github.com/ptflp/go-light/respond"
	"go.uber.org/zap"
)

type usersController struct {
	*decoder.Decoder
	respond.Responder
	user   *services.User
	logger *zap.Logger
}

func NewUsersController(responder respond.Responder, user *services.User, logger *zap.Logger) *usersController {
	return &usersController{
		Decoder:   decoder.NewDecoder(),
		Responder: responder,
		user:      user,
		logger:    logger,
	}
}

func (u *usersController) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		usersData, err := u.user.List(r.Context())

		if err != nil {
			u.ErrorBadRequest(w, err)
			return
		}

		u.SendJSON(w, request.Response{
			Success: true,
			Data: struct {
				Users []request.UserData `json:"users"`
			}{
				Users: usersData,
			},
		})
	}
}

func (u *usersController) TempList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var limitOffsetReq request.LimitOffsetReq

		err := u.Decode(r.Body, &limitOffsetReq)
		if err != nil {
			u.ErrorBadRequest(w, err)
			return
		}

		usersData, err := u.user.TempList(r.Context(), limitOffsetReq)

		if err != nil {
			u.ErrorBadRequest(w, err)
			return
		}

		u.SendJSON(w, request.Response{
			Success: true,
			Data: struct {
				Users []request.UserData `json:"users"`
			}{
				Users: usersData,
			},
		})
	}
}

func (u *usersController) Recommends() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var limitOffsetReq request.LimitOffsetReq

		err := u.Decode(r.Body, &limitOffsetReq)
		if err != nil {
			u.ErrorBadRequest(w, err)
			return
		}

		usersData, err := u.user.Recommends(r.Context(), limitOffsetReq)

		if err != nil {
			u.ErrorBadRequest(w, err)
			return
		}

		u.SendJSON(w, request.Response{
			Success: true,
			Data: struct {
				Users []request.UserData `json:"users"`
			}{
				Users: usersData,
			},
		})
	}
}

func (u *usersController) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var userIDNickReq request.UserIDNickRequest

		err := u.Decode(r.Body, &userIDNickReq)
		if err != nil {
			u.ErrorBadRequest(w, err)
			return
		}

		userData, err := u.user.Get(r.Context(), userIDNickReq)

		if err != nil {
			u.ErrorBadRequest(w, err)
			return
		}

		u.SendJSON(w, request.Response{
			Success: true,
			Data: struct {
				Users request.UserData `json:"user"`
			}{
				Users: userData,
			},
		})
	}
}

func (u *usersController) Autocomplete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var nickNameReq request.UserNicknameRequest

		err := u.Decode(r.Body, &nickNameReq)
		if err != nil {
			u.ErrorBadRequest(w, err)
			return
		}

		if len(nickNameReq.Nickname) < 3 {
			u.SendJSON(w, request.Response{
				Success: true,
				Data: struct {
					Users []request.UserData `json:"user"`
				}{
					Users: []request.UserData{},
				},
			})

			return
		}

		userData, err := u.user.Autocomplete(r.Context(), nickNameReq)

		if err != nil {
			u.ErrorBadRequest(w, err)
			return
		}

		u.SendJSON(w, request.Response{
			Success: true,
			Data: struct {
				Users []request.UserData `json:"user"`
			}{
				Users: userData,
			},
		})
	}
}

func (u *usersController) RecoverPassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var passwordRecReq request.PasswordRecoverRequest

		err := u.Decode(r.Body, &passwordRecReq)

		if err != nil {
			u.ErrorBadRequest(w, err)
			return
		}

		err = u.user.PasswordRecover(r.Context(), passwordRecReq)

		if err != nil {
			u.ErrorBadRequest(w, err)
			return
		}

		u.SendJSON(w, request.Response{
			Success: true,
		})
	}
}

func (u *usersController) CheckPhoneCode() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var checkPhoneCodeReq request.CheckPhoneCodeRequest

		err := u.Decode(r.Body, &checkPhoneCodeReq)

		if err != nil {
			u.ErrorBadRequest(w, err)
			return
		}

		res, err := u.user.CheckPhoneCode(r.Context(), checkPhoneCodeReq)

		if err != nil {
			u.ErrorBadRequest(w, err)
			return
		}

		u.SendJSON(w, res)
	}
}

func (u *usersController) PasswordReset() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var passwordResetReq request.PasswordResetRequest

		err := u.Decode(r.Body, &passwordResetReq)

		if err != nil {
			u.ErrorBadRequest(w, err)
			return
		}

		err = u.user.PasswordReset(r.Context(), passwordResetReq)

		if err != nil {
			u.ErrorBadRequest(w, err)
			return
		}

		u.SendJSON(w, request.Response{
			Success: true,
		})
	}
}

func (u *usersController) EmailExist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var emailRequest request.EmailRequest

		err := u.Decode(r.Body, &emailRequest)

		if err != nil {
			u.ErrorBadRequest(w, err)
			return
		}

		err = u.user.EmailExist(r.Context(), emailRequest)

		if err != nil {
			u.SendJSON(w, request.Response{
				Success: false,
				Msg:     fmt.Sprintf("%s не зарегистрирована", emailRequest.Email),
			})
			return
		}

		u.SendJSON(w, request.Response{
			Success: true,
			Msg:     fmt.Sprintf("%s уже существует", emailRequest.Email),
		})
	}
}

func (u *usersController) NicknameExist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var nickNameReq request.NicknameRequest

		err := u.Decode(r.Body, &nickNameReq)

		if err != nil {
			u.ErrorBadRequest(w, err)
			return
		}

		err = u.user.NicknameExist(r.Context(), nickNameReq)

		if err != nil {
			u.SendJSON(w, request.Response{
				Success: false,
				Msg:     fmt.Sprintf("%s не зарегистрирован", nickNameReq.Nickname),
			})
			return
		}

		u.SendJSON(w, request.Response{
			Success: true,
			Msg:     fmt.Sprintf("%s уже существует", nickNameReq.Nickname),
		})
	}
}

func extractUser(ctx context.Context) (light.User, error) {
	u, ok := ctx.Value(types.User{}).(*light.User)
	if !ok {
		return light.User{}, errors.New("type assertion to user err")
	}

	return *u, nil
}
