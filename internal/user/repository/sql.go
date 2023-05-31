package user

import (
	"Makhkets/database/postgres"
	rdb "Makhkets/database/redis"
	"Makhkets/internal/configs"
	"Makhkets/pkg/logging"
	"Makhkets/pkg/utils"
	"context"
	"fmt"
	"strconv"
	"time"
)

type Repository interface {
	CreateUser(ctx context.Context, user *UserDTO) (*UserDTO, error)
	FindOne(ctx context.Context, id string) (*User, error)
	FindLoginUser(ctx context.Context, username, password string) (*User, error)

	//ChangeRefreshInCache(ctx context.Context, fingerprint, newRefreshToken string) error
	GetRefreshSession(ctx context.Context, fingerprint string) (*RefreshSession, error)
	SaveRefreshSession(ctx context.Context, rs *RefreshSession) error
	DeleteRefreshSession(ctx context.Context, key string) error
}

type repository struct {
	logger *logging.Logger
	cfg    *configs.Config
	client postgres.Client
	rdb    rdb.Client
}

func NewStorage(logger *logging.Logger, client postgres.Client, rdb rdb.Client) Repository {
	return &repository{
		logger: logger,
		client: client,
		rdb:    rdb,
		cfg:    configs.GetConfig(),
	}
}

func (r *repository) SaveRefreshSession(ctx context.Context, rs *RefreshSession) error {
	return r.rdb.HMSet(ctx, rs.Fingerprint, map[string]interface{}{
		"refreshToken": rs.RefreshToken,
		"userId":       rs.UserId,
		"ua":           rs.Ua,
		"ip":           rs.Ip,
		"fingerprint":  rs.Fingerprint,
		"expiresIn":    rs.ExpiresIn,
		"createdAt":    rs.CreatedAt,
	}).Err()
}

func (r *repository) GetRefreshSession(ctx context.Context, fingerprint string) (*RefreshSession, error) {
	result, err := r.rdb.HMGet(
		ctx,
		fingerprint,
		"expiresIn", "ua", "fingerprint", "refreshToken", "createdAt", "userId", "ip",
	).Result()
	if err != nil {
		return nil, err
	}

	if utils.HasNil(result) {
		return nil, fmt.Errorf("refresh session not found")
	}

	d, _ := time.ParseDuration(result[0].(string))
	t, _ := time.Parse(time.RFC3339Nano, result[4].(string))
	return &RefreshSession{
		ExpiresIn:    d,
		Ua:           result[1].(string),
		Fingerprint:  result[2].(string),
		RefreshToken: result[3].(string),
		CreatedAt:    t,
		UserId:       result[5].(string),
		Ip:           result[6].(string),
	}, nil
}

//func (r *repository) ChangeRefreshInCache(ctx context.Context, fingerprint, newRefreshToken string) error {
//	_, err := r.rdb.HMSet(ctx, fingerprint, map[string]interface{}{
//		"refreshToken": newRefreshToken,
//		"expiresIn":    time.Now().Add(time.Minute * time.Duration(r.cfg.Jwt.Refresh)).Unix(),
//	}).Result()
//	if err != nil {
//		panic(err)
//	}
//
//	return nil
//}

func (r *repository) DeleteRefreshSession(ctx context.Context, key string) error {
	return r.rdb.Del(ctx, key).Err()
}

func (r *repository) CreateUser(ctx context.Context, user *UserDTO) (*UserDTO, error) {
	var id int
	q := `INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id`
	r.logger.Debug(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))
	if err := r.client.QueryRow(ctx, q, user.Username, user.Password).Scan(&id); err != nil {
		return nil, err
	}

	return &UserDTO{Id: strconv.Itoa(id), Username: user.Username}, nil
}

func (r *repository) FindOne(ctx context.Context, id string) (*User, error) {
	var dto = &User{}
	var q = `SELECT id, username, password, created_at, updated_at, is_admin, is_banned FROM users WHERE id=$1`
	r.logger.Debug(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))
	err := r.client.QueryRow(ctx, q, id).
		Scan(&dto.Id, &dto.Username, &dto.PasswordHash, &dto.CreatedAt, &dto.UpdatedAt, &dto.IsAdmin, &dto.IsBanned)
	return dto, err
}

func (r *repository) FindLoginUser(ctx context.Context, username, password string) (*User, error) {
	var dto = &User{}
	var q = `SELECT id, username, password, created_at, updated_at, is_admin, is_banned FROM users WHERE username=$1 AND password=$2`
	r.logger.Debug(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	if err := r.client.QueryRow(ctx, q, username, password).
		Scan(&dto.Id, &dto.Username, &dto.PasswordHash, &dto.CreatedAt, &dto.UpdatedAt, &dto.IsAdmin, &dto.IsBanned); err != nil {
		return nil, err
	}

	return dto, nil
}

func (r *repository) FindAll() {
	panic("implement me")
}

func (r *repository) Update() {
	panic("implement me")
}

func (r *repository) Delete() {
	panic("implement me")
}
