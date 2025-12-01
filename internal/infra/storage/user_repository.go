package storage

import (
	"log"

	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/user"
	"github.com/jmoiron/sqlx"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) user.Repository {
	return &userRepository{db}
}

func (r *userRepository) GetByUsername(username string) (*user.User, error) {
	row := r.db.QueryRowx("SELECT id, username, first_name, last_name, role, password, is_active FROM users WHERE username = ? LIMIT 1", username)
	u := &user.User{}

	if err := row.StructScan(u); err != nil {
		log.Printf("error: %+v\n", err)
		return nil, err
	}

	return u, nil
}
