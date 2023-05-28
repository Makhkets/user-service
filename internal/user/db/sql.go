package user

import (
	"Makhkets/database/postgres"
	"Makhkets/pkg/logging"
	"Makhkets/pkg/utils"
	"context"
	"fmt"
	"strconv"
)

type Repository interface {
	Create(ctx context.Context, user *UserDTO) (*UserDTO, error)
}

type repository struct {
	logger *logging.Logger
	client postgres.Client
}

func NewStorage(logger *logging.Logger, client postgres.Client) Repository {
	return &repository{
		logger: logger,
		client: client,
	}
}

func (r *repository) FindOne() {
	panic("implement me")
}

func (r *repository) FindAll() {
	panic("implement me")
}

func (r *repository) Create(ctx context.Context, user *UserDTO) (*UserDTO, error) {
	var id int
	q := `INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id`
	r.logger.Debug(fmt.Sprintf("SQL Query: %s", utils.FormatQuery(q)))
	err := r.client.QueryRow(ctx, q, user.Username, user.Password).Scan(&id)

	if err != nil {
		return nil, err
	}

	return &UserDTO{Id: strconv.Itoa(id), Username: user.Username}, nil
}

func (r *repository) Update() {
	panic("implement me")
}

func (r *repository) Delete() {
	panic("implement me")
}
