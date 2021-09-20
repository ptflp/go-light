package db

import (
	"context"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	light "github.com/ptflp/go-light"
)

const (
	setPassword               = "UPDATE users SET password = ? WHERE uuid = ?"
	createUserByEmailPassword = "INSERT INTO users (uuid, email, password, active, email_verified) VALUES (?, ?, ?, 1, 1)"
)

type userRepository struct {
	db *sqlx.DB
	crud
}

func NewUserRepository(db *sqlx.DB) light.UserRepository {
	return &userRepository{db: db, crud: crud{db: db}}
}

func (u *userRepository) FindByEmail(ctx context.Context, user light.User) (light.User, error) {

	fields, err := light.GetFields(&light.User{})
	if err != nil {
		return light.User{}, err
	}

	query, args, err := sq.Select(fields...).From("users").Where(sq.Eq{"email": user.Email}).ToSql()
	if err != nil {
		return light.User{}, err
	}

	if err := u.db.QueryRowxContext(ctx, query, args...).StructScan(&user); err != nil {
		return light.User{}, err
	}

	return user, nil
}

func (u *userRepository) FindByPhone(ctx context.Context, user light.User) (light.User, error) {

	fields, err := light.GetFields(&light.User{})
	if err != nil {
		return light.User{}, err
	}

	query, args, err := sq.Select(fields...).From("users").Where(sq.Eq{"phone": user.Phone}).ToSql()
	if err != nil {
		return light.User{}, err
	}

	if err = u.db.QueryRowxContext(ctx, query, args...).StructScan(&user); err != nil {
		return light.User{}, err
	}

	return user, nil
}

func (u *userRepository) CreateUser(ctx context.Context, user light.User) error {
	createFields, err := light.GetFields(&light.User{}, "create")
	if err != nil {
		return err
	}
	createFieldsPointers := light.GetFieldsPointers(&user, "create")

	queryRaw := sq.Insert("users").Columns(createFields...).Values(createFieldsPointers...)
	query, args, err := queryRaw.ToSql()
	if err != nil {
		return err
	}

	_, err = u.db.QueryxContext(ctx, query, args...)

	return err
}

func (u *userRepository) CreateUserByEmailPassword(ctx context.Context, user light.User) error {
	err := u.db.QueryRowContext(ctx, createUserByEmailPassword, user.UUID, user.Email, user.Password).Err()

	return err
}

func (u *userRepository) Update(ctx context.Context, user light.User) error {
	updateFields, err := light.GetUpdateFields("users")
	if err != nil {
		return err
	}
	updateFieldsPointers := light.GetFieldsPointers(&user, "update")

	queryRaw := sq.Update("users").Where(sq.Eq{"uuid": user.UUID})
	for i := range updateFields {
		queryRaw = queryRaw.Set(updateFields[i], updateFieldsPointers[i])
	}

	query, args, err := queryRaw.ToSql()
	if err != nil {
		return err
	}
	res, err := u.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()

	return err
}

func (u *userRepository) SetPassword(ctx context.Context, user light.User) error {
	_, err := u.db.MustExecContext(ctx, setPassword, user.Password, user.UUID).RowsAffected()

	return err
}

func (u *userRepository) Find(ctx context.Context, user light.User) (light.User, error) {
	fields, err := light.GetFields(&light.User{})
	if err != nil {
		return light.User{}, err
	}

	query, args, err := sq.Select(fields...).From("users").Where(sq.Eq{"uuid": user.UUID}).ToSql()
	if err != nil {
		return light.User{}, err
	}

	if err := u.db.QueryRowxContext(ctx, query, args...).StructScan(&user); err != nil {
		return light.User{}, err
	}

	return user, nil
}

func (u *userRepository) FindByNickname(ctx context.Context, user light.User) (light.User, error) {
	fields, err := light.GetFields(&light.User{})
	if err != nil {
		return light.User{}, err
	}

	query, args, err := sq.Select(fields...).From("users").Where(sq.Eq{"nickname": user.NickName}).ToSql()
	if err != nil {
		return light.User{}, err
	}

	if err := u.db.QueryRowxContext(ctx, query, args...).StructScan(&user); err != nil {
		return light.User{}, err
	}

	return user, nil
}

func (u *userRepository) FindLikeNickname(ctx context.Context, nickname string) ([]light.User, error) {
	fields, err := light.GetFields(&light.User{})
	if err != nil {
		return nil, err
	}

	query, args, err := sq.Select(fields...).From("users").Where(sq.Like{"nickname": strings.Join([]string{"%", nickname, "%"}, "")}).ToSql()
	if err != nil {
		return nil, err
	}

	var users []light.User

	if err := u.db.SelectContext(ctx, &users, query, args...); err != nil {
		return nil, err
	}

	return users, nil
}

func (u *userRepository) FindByFacebook(ctx context.Context, user light.User) (light.User, error) {
	if !user.FacebookID.Valid {
		return light.User{}, fmt.Errorf("invalid facebook_id %d", user.FacebookID.Int64)
	}
	fields, err := light.GetFields(&light.User{})
	if err != nil {
		return light.User{}, err
	}

	query, args, err := sq.Select(fields...).From("users").Where(sq.Eq{"facebook_id": user.FacebookID}).ToSql()
	if err != nil {
		return light.User{}, err
	}

	if err := u.db.QueryRowxContext(ctx, query, args...).StructScan(&user); err != nil {
		return light.User{}, err
	}

	return user, nil
}

func (u *userRepository) FindByGoogle(ctx context.Context, user light.User) (light.User, error) {
	if !user.GoogleID.Valid {
		return light.User{}, fmt.Errorf("invalid facebook_id %s", user.GoogleID.String)
	}
	fields, err := light.GetFields(&light.User{})
	if err != nil {
		return light.User{}, err
	}

	query, args, err := sq.Select(fields...).From("users").Where(sq.Eq{"google_id": user.GoogleID}).ToSql()
	if err != nil {
		return light.User{}, err
	}

	if err := u.db.QueryRowxContext(ctx, query, args...).StructScan(&user); err != nil {
		return light.User{}, err
	}

	return user, nil
}

func (u *userRepository) FindAll(ctx context.Context) ([]light.User, error) {
	fields, err := light.GetFields(&light.User{})
	if err != nil {
		return nil, err
	}

	query, args, err := sq.Select(fields...).From("users").Where(sq.NotEq{"nickname": "null"}).ToSql()
	if err != nil {
		return nil, err
	}

	var users []light.User
	if err = u.db.Select(&users, query, args...); err != nil {
		return nil, err
	}

	return users, nil
}

func (u *userRepository) FindLimitOffset(ctx context.Context, limit, offset uint64) ([]light.User, error) {
	fields, err := light.GetFields(&light.User{})
	if err != nil {
		return nil, err
	}

	query, args, err := sq.Select(fields...).From("users").Limit(limit).Offset(offset).ToSql()
	if err != nil {
		return nil, err
	}

	var users []light.User
	if err = u.db.Select(&users, query, args...); err != nil {
		return nil, err
	}

	return users, nil
}

func (u *userRepository) Count(ctx context.Context, user light.User, field, ops string) (light.User, error) {
	err := u.count(ctx, &user, field, ops)
	if err != nil {
		return light.User{}, err
	}

	return user, nil
}

func (u *userRepository) Listx(ctx context.Context, condition light.Condition) ([]light.User, error) {
	var users []light.User
	err := u.crud.listx(ctx, &users, light.User{}, condition)
	if err != nil {
		return nil, err
	}

	return users, nil
}
