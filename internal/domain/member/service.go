package member

type Service struct {
	repo Repository
}

func NewMemberService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SaveMembers(members []*Member) ([]*Member, error) {
	return s.repo.SaveMembers(members)
}