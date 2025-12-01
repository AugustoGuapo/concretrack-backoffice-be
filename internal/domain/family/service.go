package family

type Service struct {
	repo Repository
}

func NewFamilyService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SaveFamily(family Family) (*Family, error) {
	return s.repo.SaveFamily(&family)
}
