package repo

import (
	"Makhkets/internal/configs"
	"Makhkets/internal/database/postgres"
	"Makhkets/internal/database/redis"
	"Makhkets/pkg/logging"
	"Makhkets/pkg/utils"
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v4"
	"strconv"
	"time"
)

type Repository interface {
	CreateUser(ctx context.Context, user *UserDTO) (*UserDTO, error)
	FindOne(ctx context.Context, id string) (*User, error)
	Delete(ctx context.Context, id string) error
	FindLoginUser(ctx context.Context, username, password string) (*User, error)

	UpdateUsername(ctx context.Context, id, username string) error
	UpdatePassword(ctx context.Context, id, oldPassword, newPassword string) error

	GetUser(ctx context.Context, id string) (*User, error)

	ChangeStatus(ctx context.Context, id, status string) error
	ChangePermission(ctx context.Context, id string, permission bool) error

	//ChangeRefreshInCache(ctx context.Context, fingerprint, newRefreshToken string) error
	GetRefreshSession(ctx context.Context, fingerprint string) (*RefreshSession, error)
	SaveRefreshSession(ctx context.Context, rs *RefreshSession) error
	DeleteRefreshSession(ctx context.Context, key, id string) error
	GetRefreshSessionsByUserId(ctx context.Context, userId string) ([]*RefreshSession, error)
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
	// First, check if an array exists under the key "id"
	key := rs.UserId
	exists, err := r.rdb.Exists(ctx, key).Result()
	if err != nil {
		return err
	}

	if exists == 0 {
		// If the array doesn't exist, create a new one with "fingerprint" as the first element
		response := r.rdb.SAdd(ctx, utils.GetIdField(key), rs.Fingerprint)
		if response.Err() != nil {
			return response.Err()
		}
	} else {
		// If the array exists, add "fingerprint" to it
		response := r.rdb.SAdd(ctx, utils.GetIdField(key), rs.Fingerprint)
		if response.Err() != nil {
			return response.Err()
		}
	}

	// Save the refresh session data
	response := r.rdb.HMSet(ctx, rs.Fingerprint, map[string]interface{}{
		"refreshToken": rs.RefreshToken,
		"userId":       rs.UserId,
		"ua":           rs.Ua,
		"ip":           rs.Ip,
		"fingerprint":  rs.Fingerprint,
		"expiresIn":    rs.ExpiresIn,
		"createdAt":    rs.CreatedAt,
	}).Err()

	r.rdb.Expire(ctx, rs.Fingerprint, time.Duration(r.cfg.Jwt.Refresh)*time.Minute)
	r.rdb.Expire(ctx, utils.GetIdField(key), time.Duration(r.cfg.Jwt.Refresh)*time.Minute)

	return response
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

func (r *repository) GetRefreshSessionsByUserId(ctx context.Context, id string) ([]*RefreshSession, error) {
	// Делаем запрос, на получение всех ключей, на сессии
	var refreshSessions []*RefreshSession = []*RefreshSession{}
	result := r.rdb.SMembers(ctx, utils.GetIdField(id))
	if err := result.Err(); err != nil {
		return nil, err
	}

	for _, key := range result.Val() {
		refreshSession, err := r.GetRefreshSession(ctx, key)
		if err != nil {
			return nil, err
		}
		refreshSessions = append(refreshSessions, refreshSession)
	}

	return refreshSessions, nil
}

func (r *repository) DeleteRefreshSession(ctx context.Context, key, id string) error {
	if err := r.rdb.Del(ctx, key).Err(); err != nil {
		return err
	}

	if err := r.rdb.LRem(ctx, id, 0, key).Err(); err != nil {
		return err
	}

	return nil
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
	var q = `SELECT id, username, password, created_at, updated_at, status, is_banned FROM users WHERE id=$1`
	r.logger.Debug(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))
	err := r.client.QueryRow(ctx, q, id).
		Scan(&dto.Id, &dto.Username, &dto.PasswordHash, &dto.CreatedAt, &dto.UpdatedAt, &dto.Status, &dto.IsBanned)
	return dto, err
}

func (r *repository) FindLoginUser(ctx context.Context, username, password string) (*User, error) {
	var dto = &User{}
	var q = `SELECT id, username, password, created_at, updated_at, status, is_banned FROM users WHERE username=$1 AND password=$2`
	r.logger.Debug(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	if err := r.client.QueryRow(ctx, q, username, password).
		Scan(&dto.Id, &dto.Username, &dto.PasswordHash, &dto.CreatedAt, &dto.UpdatedAt, &dto.Status, &dto.IsBanned); err != nil {
		return nil, err
	}

	return dto, nil
}

func (r *repository) FindAll() {
	panic("implement me")
}

func (r *repository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM users WHERE id = $1 RETURNING id"
	r.logger.Debug(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(query)))
	var deletedID int
	err := r.client.QueryRow(ctx, query, id).Scan(&deletedID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("user with ID %s doesn't exist", id)
		}
		return err
	}
	return nil
}

func (r *repository) UpdateUsername(ctx context.Context, id, username string) error {
	// Update the username for the specified user, checking for uniqueness in a single query
	q := `
		UPDATE users
		SET username = $1
		WHERE id = $2 AND NOT EXISTS (
			SELECT 1 FROM users WHERE username = $1 AND id != $2
		)
	`
	r.logger.Debug(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))
	result, err := r.client.Exec(ctx, q, username, id)
	if err != nil {
		return err
	}

	// Check how many rows were affected by the update
	rowsAffected := result.RowsAffected()

	// If no rows were affected by the update, it means either the user doesn't exist or the new username is already taken
	if rowsAffected == 0 {
		return fmt.Errorf("user not found or username already taken")
	}

	return nil
}

func (r *repository) UpdatePassword(ctx context.Context, id, oldPassword, newPassword string) error {
	var dbPassword string
	q := "UPDATE users SET password = $1 WHERE id = $2 AND password = $3 RETURNING password"
	r.logger.Debug(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))
	err := r.client.QueryRow(ctx, q, newPassword, id, oldPassword).Scan(&dbPassword)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("user with id %s not found or wrong passsword", id)
		}
		return fmt.Errorf("error updating password in database: %w", err)
	}
	if dbPassword != newPassword {
		return fmt.Errorf("old password does not match current password")
	}
	return nil
}

func (r *repository) ChangeStatus(ctx context.Context, id, status string) error {
	// Выполнение SQL запроса на изменение роли пользователя
	q := "UPDATE users SET status = $1 WHERE id = $2"
	r.logger.Debug(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))
	_, err := r.client.Exec(ctx, q, status, id)
	return err
}

func (r *repository) ChangePermission(ctx context.Context, id string, permission bool) error {
	q := "UPDATE users SET is_banned=$1 WHERE id=$2"
	r.logger.Debug(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))
	_, err := r.client.Exec(ctx, q, permission, id)
	return err
}

func (r *repository) GetUser(ctx context.Context, id string) (*User, error) {
	q := "SELECT id, username, is_banned, status FROM users WHERE id = $1"
	r.logger.Debug(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))

	row := r.client.QueryRow(ctx, q, id)

	u := &User{}
	err := row.Scan(&u.Id, &u.Username, &u.IsBanned, &u.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return u, nil
}
