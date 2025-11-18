package project

type Service struct {
	repo Repository
}

func NewService(r Repository) *Service {
	return &Service{repo: r}
} 

func (s *Service) GetProjectByID(ID int)(*Project, error) {
	return s.repo.GetProjectByID(ID)
}

func (s *Service) GetProjects(page int)([]*Project, error) {
	return s.repo.GetProjects(page)
}