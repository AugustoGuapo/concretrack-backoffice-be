package storage

import (
	"github.com/AugustoGuapo/concretrack-backoffice-be/internal/domain/client"
	"github.com/jmoiron/sqlx"
)

type clientRepository struct {
	db *sqlx.DB
}

func NewClientRepository(db *sqlx.DB) *clientRepository {
	return &clientRepository{db: db}
}

func (r *clientRepository) SaveClient(c *client.Client) (*client.Client, error) {
	result, err := r.db.Exec("INSERT INTO clients (name) VALUES (?)", c.Name)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &client.Client{
		ID:   int(id),
		Name: c.Name,
	}, nil
}

func (r *clientRepository) GetClient(ID int) (*client.Client, error) {
	clientRow := r.db.QueryRowx("SELECT id, name FROM clients WHERE id = ?", ID)
	client := &client.Client{}
	if err := clientRow.StructScan(client); err != nil {
		return nil, err
	}
	return client, nil
}

func (r *clientRepository) GetAllClients() ([]*client.Client, error) {
	rows, err := r.db.Queryx("SELECT id, name FROM clients")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []*client.Client

	for rows.Next() {
		c := &client.Client{}
		if err := rows.StructScan(c); err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}

	// Siempre chequear errores del iterador
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return clients, nil
}
