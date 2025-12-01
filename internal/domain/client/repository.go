package client

type Repository interface {
	SaveClient(client *Client) (*Client, error)
	GetClient(ID int) (*Client, error)
	GetAllClients() ([]*Client, error)
}
