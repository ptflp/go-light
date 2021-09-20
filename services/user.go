package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"math/rand"
	"net/url"
	"time"

	"github.com/ptflp/go-light/email"
	"github.com/ptflp/go-light/types"

	"github.com/ptflp/go-light/utils"

	"github.com/ptflp/go-light/components"
	"go.uber.org/zap"

	"github.com/ptflp/go-light/validators"

	"github.com/ptflp/go-light/decoder"

	"github.com/ptflp/go-light/hasher"

	light "github.com/ptflp/go-light"
	"github.com/ptflp/go-light/request"
)

const (
	PhoneRecoverKey      = "phone:recover:%s"
	PasswordRecoveryUUID = "R"
	RecoveryIDKey        = "recover:id:%s"
)

type User struct {
	*decoder.Decoder
	userRepository light.UserRepository
	components.Componenter
}

func NewUserService(rs light.Repositories, cmps components.Componenter) *User {
	return &User{userRepository: rs.Users, Decoder: decoder.NewDecoder(), Componenter: cmps}
}

func (u *User) CheckEmailPass(ctx context.Context, user light.User) bool {
	uDB, err := u.userRepository.FindByEmail(ctx, user)
	if err != nil {
		return false
	}

	return hasher.CheckPasswordHash(user.Password.String, uDB.Password.String)
}

func (u *User) CreateByEmailPassword(ctx context.Context, user light.User) error {
	passHash, err := hasher.HashPassword(user.Password.String)
	if err != nil {
		return err
	}

	user.Password = types.NewNullString(passHash)
	return u.userRepository.CreateUser(ctx, user)
}

func (u *User) GetProfile(ctx context.Context) (request.UserData, error) {
	user, err := extractUser(ctx)
	if err != nil {
		return request.UserData{}, err
	}

	user, err = u.userRepository.Find(ctx, user)
	if err != nil {
		return request.UserData{}, err
	}

	userData := request.UserData{}
	err = u.MapStructs(&userData, &user)
	if err != nil {
		return request.UserData{}, err
	}

	userData.Counts = &request.UserDataCounts{
		Friends: 377,
	}

	userData.PasswordSet = &user.Password.Valid

	return userData, nil
}

func (u *User) UpdateProfile(ctx context.Context, profileUpdateReq request.ProfileUpdateReq, user light.User) (request.UserData, error) {
	user, err := u.userRepository.Find(ctx, user)
	if err != nil {
		return request.UserData{}, err
	}

	err = u.MapStructs(&user, &profileUpdateReq)
	if err != nil {
		return request.UserData{}, err
	}

	user.Active = types.NewNullBool(true)
	err = u.userRepository.Update(ctx, user)
	if err != nil {
		return request.UserData{}, err
	}

	userData := request.UserData{}
	err = u.MapStructs(&userData, &user)
	if err != nil {
		return request.UserData{}, err
	}

	return userData, nil
}

func (u *User) SetPassword(ctx context.Context, setPasswordReq request.SetPasswordReq) error {
	user, err := extractUser(ctx)
	if err != nil {
		return err
	}
	user, err = u.userRepository.Find(ctx, user)
	if err != nil {
		return err
	}
	if user.Password.Valid {
		if setPasswordReq.OldPassword == nil {
			return fmt.Errorf("old password is required")
		}
		if !hasher.CheckPasswordHash(*setPasswordReq.OldPassword, user.Password.String) {
			return fmt.Errorf("wrong old password")
		}
	}

	passHash, err := hasher.HashPassword(setPasswordReq.Password)
	if err != nil {
		return err
	}
	user.Password = types.NewNullString(passHash)

	return u.userRepository.SetPassword(ctx, user)
}

func (u *User) prepareRecoveryTemplate(recoverUrl string) (bytes.Buffer, error) {
	tmpl, err := template.ParseFiles("./templates/password_recovery.html")
	if err != nil {
		u.Logger().Error("recover password template parse", zap.Error(err))
		return bytes.Buffer{}, err
	}
	type PasswordRecover struct {
		PasswordRecover string
	}

	b := bytes.Buffer{}

	err = tmpl.Execute(&b, PasswordRecover{PasswordRecover: recoverUrl})

	return b, err
}

func (u *User) List(ctx context.Context) ([]request.UserData, error) {
	users, err := u.userRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	usersData := []request.UserData{}
	for _, user := range users {
		userData := request.UserData{}
		err = u.MapStructs(&userData, &user)
		if err != nil {
			return nil, err
		}
		usersData = append(usersData, userData)
	}

	return usersData, nil
}

func (u *User) Listx(ctx context.Context, condition light.Condition) ([]light.User, error) {
	return u.userRepository.Listx(ctx, condition)
}

func (u *User) TempList(ctx context.Context, req request.LimitOffsetReq) ([]request.UserData, error) {
	users, err := u.userRepository.FindLimitOffset(ctx, uint64(req.Limit), uint64(req.Offset))
	if err != nil {
		return nil, err
	}

	usersData := []request.UserData{}

	for _, user := range users {
		userData := request.UserData{}
		err = u.MapStructs(&userData, &user)
		if err != nil {
			return nil, err
		}
		usersData = append(usersData, userData)
	}

	return usersData, nil
}

func (u *User) Recommends(ctx context.Context, req request.LimitOffsetReq) ([]request.UserData, error) {
	user, err := extractUser(ctx)
	if err != nil {
		return nil, err
	}
	condition := light.Condition{
		NotIn: &light.In{
			Field: "uuid",
			Args:  []interface{}{user.UUID},
		},
		Order: &light.Order{
			Field: "likes",
		},
		Other: &light.Other{
			Condition: "nickname IS NOT null",
			Args:      nil,
		},
	}
	users, err := u.userRepository.Listx(ctx, condition)
	if err != nil {
		return nil, err
	}

	var usersData []request.UserData

	err = u.MapStructs(&usersData, &users)
	if err != nil {
		return nil, err
	}

	return usersData, nil
}

func (u *User) GetUsersByCondition(ctx context.Context, condition light.Condition) ([]request.UserData, error) {
	users, err := u.userRepository.Listx(ctx, condition)
	if err != nil {
		return nil, err
	}

	var usersData []request.UserData

	err = u.MapStructs(&usersData, &users)
	if err != nil {
		return nil, err
	}

	return usersData, nil
}

func (u *User) Get(ctx context.Context, req request.UserIDNickRequest) (request.UserData, error) {
	user := light.User{}
	var err error
	if req.UUID != nil {
		user.UUID = types.NewNullUUID(*req.UUID)
		user, err = u.userRepository.Find(ctx, user)
		if err != nil {
			return request.UserData{}, err
		}
	}
	if req.NickName != nil {
		user.NickName = types.NewNullString(*req.NickName)
		user, err = u.userRepository.FindByNickname(ctx, user)
		if err != nil {
			return request.UserData{}, err
		}
	}

	userData := request.UserData{}
	err = u.MapStructs(&userData, &user)
	if err != nil {
		return request.UserData{}, err
	}

	return userData, nil
}

func (u *User) Autocomplete(ctx context.Context, req request.UserNicknameRequest) ([]request.UserData, error) {
	users, err := u.userRepository.FindLikeNickname(ctx, req.Nickname)
	if err != nil {
		return nil, err
	}
	usersData := make([]request.UserData, 0, len(users))
	for i := range users {
		var userData request.UserData
		err = u.MapStructs(&userData, users[i])
		if err != nil {
			return nil, err
		}
		usersData = append(usersData, userData)
	}

	return usersData, nil
}

func (u *User) PasswordRecover(ctx context.Context, req request.PasswordRecoverRequest) error {
	user := light.User{}

	err := u.MapStructs(&user, &req)
	if err != nil {
		return err
	}

	if user.Email.Valid {
		err = validators.CheckEmailFormat(user.Email.String)
		if err != nil {
			return err
		}
		user, err = u.userRepository.FindByEmail(ctx, user)
		if err != nil {
			return err
		}
		// send email
		var recoverUrl string
		recoverUrl, _, err = u.generateRecoverUrl(user)
		if err != nil {
			return err
		}

		var body bytes.Buffer
		body, err = u.prepareRecoveryTemplate(recoverUrl)
		if err != nil {
			return err
		}

		msg := email.NewMessage()
		msg.SetSubject("Восстановление пароля")
		msg.SetType(email.TypeHtml)
		msg.SetReceiver(user.Email.String)
		msg.SetBody(body)

		err = u.Email().Send(msg)
		if err != nil {
			return err
		}
	}

	if user.Phone.Valid {
		user.Phone.String, err = validators.CheckPhoneFormat(user.Phone.String)
		if err != nil {
			return err
		}
		user, err = u.userRepository.FindByPhone(ctx, user)
		if err != nil {
			return err
		}

		code := genCode()
		if u.Config().SMSC.Dev {
			code = 3455
		}
		u.Cache().Set(fmt.Sprintf(PhoneRecoverKey, user.Phone.String), &code, 15*time.Minute)
		if u.Config().SMSC.Dev {
			return nil
		}

		err = u.Componenter.SMS().Send(ctx, user.Phone.String, fmt.Sprintf("Ваш код: %d", code))
		if err != nil {
			u.Logger().Error("send sms err", zap.String("user.Phone.String", user.Phone.String), zap.Int("code", code))
		}

		return err
	}

	return errors.New("bad request params")
}

func (u *User) CheckPhoneCode(ctx context.Context, req request.CheckPhoneCodeRequest) (request.RecoverChekPhoneResponse, error) {
	var code int64
	var user light.User
	err := u.Cache().Get(fmt.Sprintf(PhoneRecoverKey, req.Phone), &code)
	if err != nil {
		return request.RecoverChekPhoneResponse{}, err
	}
	if code != req.Code {
		return request.RecoverChekPhoneResponse{}, errors.New("user code error")
	}
	user.Phone = types.NewNullString(req.Phone)
	user, err = u.userRepository.FindByPhone(ctx, user)
	if err != nil {
		return request.RecoverChekPhoneResponse{}, err
	}

	recoverID, err := u.GenerateRecoverID(user)
	if err != nil {
		return request.RecoverChekPhoneResponse{}, err
	}

	return request.RecoverChekPhoneResponse{
		Success: true,
		Data: request.RecoverCheckPhoneData{
			RecoverID: recoverID,
		},
	}, nil
}

func (u *User) GenerateRecoverID(user light.User) (string, error) {
	recoverID, err := utils.ProjectUUIDGen(PasswordRecoveryUUID)
	if err != nil {
		return "", err
	}
	u.Cache().Set(fmt.Sprintf(RecoveryIDKey, recoverID), &user.UUID, 15*time.Minute)

	return recoverID, err
}

func (u *User) PasswordReset(ctx context.Context, req request.PasswordResetRequest) error {
	var user light.User
	err := u.Cache().Get(fmt.Sprintf(RecoveryIDKey, req.RecoverID), &user.UUID)
	if err != nil {
		return err
	}
	user, err = u.userRepository.Find(ctx, user)
	if err != nil {
		return err
	}
	passHash, err := hasher.HashPassword(req.Password)
	if err != nil {
		return err
	}
	user.Password = types.NewNullString(passHash)
	err = u.userRepository.SetPassword(ctx, user)
	if err != nil {
		return err
	}

	return err
}

func (u *User) EmailExist(ctx context.Context, req request.EmailRequest) error {
	var user light.User
	user.Email = types.NewNullString(req.Email)
	_, err := u.userRepository.FindByEmail(ctx, user)

	return err
}

func (u *User) NicknameExist(ctx context.Context, req request.NicknameRequest) error {
	var user light.User
	user.NickName = types.NewNullString(req.Nickname)
	_, err := u.userRepository.FindByNickname(ctx, user)

	return err
}

func (u *User) generateRecoverUrl(user light.User) (string, string, error) {

	recoverID, err := u.GenerateRecoverID(user)
	if err != nil {
		return "", "", err
	}

	uri, err := url.Parse(u.Config().App.FrontEnd)
	if err != nil {
		return "", "", err
	}
	uri.Path = fmt.Sprintf("profile/password/%s", recoverID)

	return uri.String(), recoverID, err
}

func (u *User) GetUserData(user light.User) (request.UserData, error) {

	var userData request.UserData
	err := u.MapStructs(&userData, &user)

	userData.AvatarSet = user.Avatar.Valid
	return userData, err
}

func (u *User) Count(ctx context.Context, user light.User, field, ops string) (light.User, error) {
	return u.userRepository.Count(ctx, user, field, ops)
}

func extractUser(ctx context.Context) (light.User, error) {
	u, ok := ctx.Value(types.User{}).(*light.User)
	if !ok {
		return light.User{}, errors.New("type assertion to user err")
	}

	return *u, nil
}

func genCode() int {
	rand.Seed(time.Now().UnixNano())
	code := rand.Intn(8999) + 1000

	return code
}
