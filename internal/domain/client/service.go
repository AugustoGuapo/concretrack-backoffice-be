package client

type Service struct {
	repo Repository
}

func NewClientService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SaveClient(client *Client) (*Client, error) {
	return s.repo.SaveClient(client)
}

func (s *Service) GetClient(ID int) (*Client, error) {
	return s.repo.GetClient(ID)
}

func (s *Service) GetAllClients() ([]*Client, error) {
	return s.repo.GetAllClients()
}